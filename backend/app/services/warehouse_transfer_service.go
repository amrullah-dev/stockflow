package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	frameworkerrors "github.com/goravel/framework/errors"

	"goravel/app/models"
	"goravel/app/repositories"
)

const (
	warehouseTransferMovementOut = "OUT"
	warehouseTransferMovementIn  = "IN"
	warehouseTransferReference   = "WarehouseTransfer"
	warehouseTransferPostAction  = "POST_STOCK_TRANSFER"
)

var (
	ErrWarehouseTransferNil                           = errors.New("warehouse transfer is nil")
	ErrWarehouseTransferAlreadyPosted                 = errors.New("warehouse transfer is already posted")
	ErrWarehouseTransferInvalidStatus                 = errors.New("warehouse transfer cannot be posted from its current status")
	ErrWarehouseTransferEmptyItems                    = errors.New("warehouse transfer must contain at least one item")
	ErrWarehouseTransferInvalidProduct                = errors.New("warehouse transfer item product ID must be greater than zero")
	ErrWarehouseTransferInvalidQuantity               = errors.New("warehouse transfer item quantity must be greater than zero")
	ErrWarehouseTransferDuplicateItem                 = errors.New("warehouse transfer contains duplicate product items")
	ErrWarehouseTransferSameWarehouse                 = errors.New("warehouse transfer source and destination must be different")
	ErrWarehouseTransferInsufficientStock             = errors.New("warehouse transfer has insufficient stock")
	ErrWarehouseTransferSourceWarehouseNotFound       = errors.New("source warehouse not found")
	ErrWarehouseTransferDestinationWarehouseNotFound  = errors.New("destination warehouse not found")
	ErrWarehouseTransferSourceWarehouseInactive       = errors.New("source warehouse is inactive")
	ErrWarehouseTransferDestinationWarehouseInactive  = errors.New("destination warehouse is inactive")
	ErrWarehouseTransferProductNotFound               = errors.New("product not found")
	ErrWarehouseTransferProductInactive               = errors.New("product is inactive")
	ErrWarehouseTransferInvalidDate                   = errors.New("invalid transaction date: format must be YYYY-MM-DD")
	ErrWarehouseTransferMissingSourceWarehouseID      = errors.New("source warehouse ID must be greater than zero")
	ErrWarehouseTransferMissingDestinationWarehouseID = errors.New("destination warehouse ID must be greater than zero")
	ErrWarehouseTransferReferenceCollision            = errors.New("unable to generate unique reference number after retries")
)

type CreateWarehouseTransferItemInput struct {
	ProductID uint64 `json:"product_id"`
	Quantity  int64  `json:"quantity"`
}

type CreateWarehouseTransferDraftInput struct {
	FromWarehouseID uint64                             `json:"from_warehouse_id"`
	ToWarehouseID   uint64                             `json:"to_warehouse_id"`
	TransactionDate string                             `json:"transaction_date"`
	Note            string                             `json:"note"`
	Items           []CreateWarehouseTransferItemInput `json:"items"`
}

type WarehouseTransferDraftItemResponse struct {
	ID         uint64 `json:"id"`
	TransferID uint64 `json:"transfer_id"`
	ProductID  uint64 `json:"product_id"`
	Quantity   int64  `json:"quantity"`
}

type WarehouseTransferDraftResponse struct {
	ID              uint64                               `json:"id"`
	ReferenceNumber string                               `json:"reference_number"`
	FromWarehouseID uint64                               `json:"from_warehouse_id"`
	ToWarehouseID   uint64                               `json:"to_warehouse_id"`
	TransactionDate string                               `json:"transaction_date"`
	Status          string                               `json:"status"`
	Note            string                               `json:"note"`
	CreatedBy       uint64                               `json:"created_by"`
	PostedAt        *time.Time                           `json:"posted_at"`
	Items           []WarehouseTransferDraftItemResponse `json:"items"`
}

type WarehouseTransferService struct {
	db                  contractsorm.Orm
	onHeaderCreatedHook func(tx contractsorm.Query) error
	onItemCreatedHook   func(tx contractsorm.Query) error
}

func NewWarehouseTransferService(db contractsorm.Orm) *WarehouseTransferService {
	return &WarehouseTransferService{db: db}
}

func (s *WarehouseTransferService) SetOnHeaderCreatedHook(hook func(tx contractsorm.Query) error) {
	s.onHeaderCreatedHook = hook
}

func (s *WarehouseTransferService) SetOnItemCreatedHook(hook func(tx contractsorm.Query) error) {
	s.onItemCreatedHook = hook
}

func (s *WarehouseTransferService) Post(ctx context.Context, transferID, actorID uint64) error {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	return db.Transaction(func(tx contractsorm.Query) error {
		transferRepository := repositories.NewWarehouseTransferRepository(tx)
		stockRepository := repositories.NewWarehouseStockRepository(tx)
		movementRepository := repositories.NewStockMovementRepository(tx)
		auditLogRepository := repositories.NewAuditLogRepository(tx)
		auditLogService := NewAuditLogService(auditLogRepository)

		transfer, err := transferRepository.FindByIDForUpdate(transferID)
		fmt.Printf("[DEBUG SERVICE POST] Panggil Service Post untuk ID: %d, Err: %v\n", transferID, err)
		if err != nil {
			return err
		}

		if err := validateWarehouseTransferForPosting(transfer); err != nil {
			return err
		}

		keys := warehouseStockKeysForTransfer(transfer)
		lockedStocks := make(map[warehouseStockKey]*models.WarehouseStock, len(keys))
		for _, key := range keys {
			stock, err := stockRepository.FindByWarehouseAndProductForUpdate(
				key.warehouseID,
				key.productID,
			)
			if err != nil {
				if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
					if key.warehouseID != transfer.ToWarehouseID {
						return fmt.Errorf("%w: product ID %d", ErrWarehouseTransferInsufficientStock, key.productID)
					}
					stock = &models.WarehouseStock{WarehouseID: key.warehouseID, ProductID: key.productID, Quantity: 0}
					if err := stockRepository.Create(stock); err != nil {
						return err
					}
				} else {
					return err
				}
			}

			lockedStocks[key] = stock
		}

		for _, item := range transfer.Items {
			sourceKey := warehouseStockKey{
				warehouseID: transfer.FromWarehouseID,
				productID:   item.ProductID,
			}
			sourceStock := lockedStocks[sourceKey]
			if sourceStock.Quantity < item.Quantity {
				return fmt.Errorf("%w: product ID %d", ErrWarehouseTransferInsufficientStock, item.ProductID)
			}
		}

		referenceID := uint64(transfer.ID)
		createdBy := actorID
		for _, item := range transfer.Items {
			sourceStock := lockedStocks[warehouseStockKey{
				warehouseID: transfer.FromWarehouseID,
				productID:   item.ProductID,
			}]
			destinationStock := lockedStocks[warehouseStockKey{
				warehouseID: transfer.ToWarehouseID,
				productID:   item.ProductID,
			}]

			sourceBefore := sourceStock.Quantity
			destinationBefore := destinationStock.Quantity
			sourceAfter := sourceBefore - item.Quantity
			destinationAfter := destinationBefore + item.Quantity

			if err := stockRepository.UpdateQuantity(sourceStock, sourceAfter); err != nil {
				fmt.Printf("[DEBUG EXEC] Step Stock Update Product %d, Err: %v\n", item.ProductID, err)
				return err
			}
			fmt.Printf("[DEBUG EXEC] Step Stock Update Product %d, Err: %v\n", item.ProductID, nil)
			if err := stockRepository.UpdateQuantity(destinationStock, destinationAfter); err != nil {
				fmt.Printf("[DEBUG EXEC] Step Stock Update Product %d, Err: %v\n", item.ProductID, err)
				return err
			}
			fmt.Printf("[DEBUG EXEC] Step Stock Update Product %d, Err: %v\n", item.ProductID, nil)

			sourceMovement := &models.StockMovement{
				WarehouseID:   transfer.FromWarehouseID,
				ProductID:     item.ProductID,
				Type:          warehouseTransferMovementOut,
				Quantity:      item.Quantity,
				BalanceBefore: sourceBefore,
				BalanceAfter:  sourceAfter,
				ReferenceType: warehouseTransferReference,
				ReferenceID:   &referenceID,
				CreatedBy:     &createdBy,
			}
			if err := movementRepository.Create(sourceMovement); err != nil {
				fmt.Printf("[DEBUG EXEC] Step Movement Insert, Err: %v\n", err)
				return err
			}
			fmt.Printf("[DEBUG EXEC] Step Movement Insert, Err: %v\n", nil)

			destinationMovement := &models.StockMovement{
				WarehouseID:   transfer.ToWarehouseID,
				ProductID:     item.ProductID,
				Type:          warehouseTransferMovementIn,
				Quantity:      item.Quantity,
				BalanceBefore: destinationBefore,
				BalanceAfter:  destinationAfter,
				ReferenceType: warehouseTransferReference,
				ReferenceID:   &referenceID,
				CreatedBy:     &createdBy,
			}
			if err := movementRepository.Create(destinationMovement); err != nil {
				fmt.Printf("[DEBUG EXEC] Step Movement Insert, Err: %v\n", err)
				return err
			}
			fmt.Printf("[DEBUG EXEC] Step Movement Insert, Err: %v\n", nil)

			sourceStock.Quantity = sourceAfter
			destinationStock.Quantity = destinationAfter
		}

		oldValues := "status=" + transfer.Status
		newValues := "status=" + StockStatusPosted
		if err := auditLogService.Create(
			&actorID,
			warehouseTransferPostAction,
			warehouseTransferReference,
			uint64(transfer.ID),
			oldValues,
			newValues,
			"",
			"",
		); err != nil {
			fmt.Printf("[DEBUG EXEC] Step Audit Log Insert, Err: %v\n", err)
			return err
		}
		fmt.Printf("[DEBUG EXEC] Step Audit Log Insert, Err: %v\n", nil)

		if err := transferRepository.UpdateStatus(transfer, StockStatusPosted); err != nil {
			return err
		}

		postedAt := time.Now()
		return transferRepository.UpdatePostedAt(transfer, &postedAt)
	})
}

type warehouseStockKey struct {
	warehouseID uint64
	productID   uint64
}

func warehouseStockKeysForTransfer(transfer *models.WarehouseTransfer) []warehouseStockKey {
	keys := make(map[warehouseStockKey]struct{}, len(transfer.Items)*2)
	for _, item := range transfer.Items {
		keys[warehouseStockKey{
			warehouseID: transfer.FromWarehouseID,
			productID:   item.ProductID,
		}] = struct{}{}
		keys[warehouseStockKey{
			warehouseID: transfer.ToWarehouseID,
			productID:   item.ProductID,
		}] = struct{}{}
	}

	orderedKeys := make([]warehouseStockKey, 0, len(keys))
	for key := range keys {
		orderedKeys = append(orderedKeys, key)
	}

	sort.Slice(orderedKeys, func(i, j int) bool {
		if orderedKeys[i].warehouseID != orderedKeys[j].warehouseID {
			return orderedKeys[i].warehouseID < orderedKeys[j].warehouseID
		}

		return orderedKeys[i].productID < orderedKeys[j].productID
	})

	return orderedKeys
}

func validateWarehouseTransferForPosting(transfer *models.WarehouseTransfer) error {
	if transfer == nil {
		return ErrWarehouseTransferNil
	}

	if transfer.Status == StockStatusPosted {
		return ErrWarehouseTransferAlreadyPosted
	}

	fmt.Printf("[DEBUG VALIDATE] Status: %s, Items Count: %d\n", transfer.Status, len(transfer.Items))
	if strings.ToUpper(transfer.Status) != "DRAFT" {
		return fmt.Errorf("%w: %s", ErrWarehouseTransferInvalidStatus, transfer.Status)
	}

	if len(transfer.Items) == 0 {
		return ErrWarehouseTransferEmptyItems
	}

	if transfer.FromWarehouseID == transfer.ToWarehouseID {
		return ErrWarehouseTransferSameWarehouse
	}

	products := make(map[uint64]struct{}, len(transfer.Items))
	for _, item := range transfer.Items {
		if item.ProductID == 0 {
			return ErrWarehouseTransferInvalidProduct
		}

		if !(item.Quantity > 0) {
			return fmt.Errorf("%w: product ID %d", ErrWarehouseTransferInvalidQuantity, item.ProductID)
		}

		if _, exists := products[item.ProductID]; exists {
			return fmt.Errorf("%w: product ID %d", ErrWarehouseTransferDuplicateItem, item.ProductID)
		}

		products[item.ProductID] = struct{}{}
	}

	return nil
}

// CreateDraft creates a new warehouse transfer transaction with status DRAFT.
// It performs atomic validation and persistence without modifying warehouse inventory or checking source stock.
func (s *WarehouseTransferService) CreateDraft(ctx context.Context, input CreateWarehouseTransferDraftInput, actorID uint64) (*WarehouseTransferDraftResponse, error) {
	if input.FromWarehouseID == 0 {
		return nil, ErrWarehouseTransferMissingSourceWarehouseID
	}
	if input.ToWarehouseID == 0 {
		return nil, ErrWarehouseTransferMissingDestinationWarehouseID
	}
	if input.FromWarehouseID == input.ToWarehouseID {
		return nil, ErrWarehouseTransferSameWarehouse
	}

	dateStr := strings.TrimSpace(input.TransactionDate)
	if dateStr == "" {
		return nil, ErrWarehouseTransferInvalidDate
	}
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, ErrWarehouseTransferInvalidDate
	}

	if len(input.Items) == 0 {
		return nil, ErrWarehouseTransferEmptyItems
	}

	seenProducts := make(map[uint64]struct{}, len(input.Items))
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return nil, ErrWarehouseTransferInvalidProduct
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: product ID %d", ErrWarehouseTransferInvalidQuantity, item.ProductID)
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return nil, fmt.Errorf("%w: product ID %d", ErrWarehouseTransferDuplicateItem, item.ProductID)
		}
		seenProducts[item.ProductID] = struct{}{}
	}

	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	var response *WarehouseTransferDraftResponse

	err = db.Transaction(func(tx contractsorm.Query) error {
		warehouseRepo := repositories.NewWarehouseRepository(tx)
		productRepo := repositories.NewProductRepository(tx)
		transferRepo := repositories.NewWarehouseTransferRepository(tx)

		fromWarehouse, err := warehouseRepo.FindByID(input.FromWarehouseID)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return fmt.Errorf("%w: ID %d", ErrWarehouseTransferSourceWarehouseNotFound, input.FromWarehouseID)
			}
			return err
		}
		if !fromWarehouse.IsActive {
			return fmt.Errorf("%w: ID %d", ErrWarehouseTransferSourceWarehouseInactive, input.FromWarehouseID)
		}

		toWarehouse, err := warehouseRepo.FindByID(input.ToWarehouseID)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return fmt.Errorf("%w: ID %d", ErrWarehouseTransferDestinationWarehouseNotFound, input.ToWarehouseID)
			}
			return err
		}
		if !toWarehouse.IsActive {
			return fmt.Errorf("%w: ID %d", ErrWarehouseTransferDestinationWarehouseInactive, input.ToWarehouseID)
		}

		for _, item := range input.Items {
			product, err := productRepo.FindByID(item.ProductID)
			if err != nil {
				if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
					return fmt.Errorf("%w: ID %d", ErrWarehouseTransferProductNotFound, item.ProductID)
				}
				return err
			}
			if !product.IsActive {
				return fmt.Errorf("%w: ID %d", ErrWarehouseTransferProductInactive, item.ProductID)
			}
		}

		var refNumber string
		var created bool
		var transfer *models.WarehouseTransfer

		for attempt := 0; attempt < 3; attempt++ {
			refNumber, err = GenerateWarehouseTransferReference()
			if err != nil {
				return err
			}

			if existing, findErr := transferRepo.FindByReferenceNumber(refNumber); findErr == nil && existing != nil {
				continue
			}

			transfer = &models.WarehouseTransfer{
				ReferenceNumber: refNumber,
				FromWarehouseID: input.FromWarehouseID,
				ToWarehouseID:   input.ToWarehouseID,
				TransactionDate: parsedDate,
				Status:          StockStatusDraft,
				Note:            input.Note,
				CreatedBy:       actorID,
			}

			if createErr := transferRepo.Create(transfer); createErr != nil {
				if isDuplicateEntryError(createErr) {
					continue
				}
				return createErr
			}

			created = true
			break
		}

		if !created {
			return ErrWarehouseTransferReferenceCollision
		}

		if s.onHeaderCreatedHook != nil {
			if err := s.onHeaderCreatedHook(tx); err != nil {
				return err
			}
		}

		itemResponses := make([]WarehouseTransferDraftItemResponse, 0, len(input.Items))
		for _, itemInput := range input.Items {
			itemModel := &models.WarehouseTransferItem{
				TransferID: uint64(transfer.ID),
				ProductID:  itemInput.ProductID,
				Quantity:   itemInput.Quantity,
			}

			if err := transferRepo.CreateItem(itemModel); err != nil {
				return err
			}

			itemResponses = append(itemResponses, WarehouseTransferDraftItemResponse{
				ID:         uint64(itemModel.ID),
				TransferID: itemModel.TransferID,
				ProductID:  itemModel.ProductID,
				Quantity:   itemModel.Quantity,
			})
		}

		response = &WarehouseTransferDraftResponse{
			ID:              uint64(transfer.ID),
			ReferenceNumber: transfer.ReferenceNumber,
			FromWarehouseID: transfer.FromWarehouseID,
			ToWarehouseID:   transfer.ToWarehouseID,
			TransactionDate: parsedDate.Format("2006-01-02"),
			Status:          transfer.Status,
			Note:            transfer.Note,
			CreatedBy:       transfer.CreatedBy,
			PostedAt:        transfer.PostedAt,
			Items:           itemResponses,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *WarehouseTransferService) UpdateDraft(ctx context.Context, id uint64, input CreateWarehouseTransferDraftInput) (*WarehouseTransferDraftResponse, error) {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	var response *WarehouseTransferDraftResponse
	err := db.Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewWarehouseTransferRepository(tx)
		transfer, err := repo.FindByIDForUpdate(id)
		if err != nil {
			return err
		}
		if transfer.Status == StockStatusPosted {
			return ErrWarehouseTransferAlreadyPosted
		}
		if transfer.Status != StockStatusDraft {
			return fmt.Errorf("%w: %s", ErrWarehouseTransferInvalidStatus, transfer.Status)
		}
		date, err := validateTransferDraftUpdateInput(input)
		if err != nil {
			return err
		}
		if err := validateTransferMasters(tx, input); err != nil {
			return err
		}
		if err := repo.DeleteItemsByTransferID(id); err != nil {
			return err
		}
		if err := repo.UpdateHeader(transfer, repositories.UpdateWarehouseTransferHeaderData{TransactionDate: date, FromWarehouseID: input.FromWarehouseID, ToWarehouseID: input.ToWarehouseID, Note: input.Note}); err != nil {
			return err
		}
		items := make([]WarehouseTransferDraftItemResponse, 0, len(input.Items))
		for _, in := range input.Items {
			item := &models.WarehouseTransferItem{TransferID: uint64(transfer.ID), ProductID: in.ProductID, Quantity: in.Quantity}
			if err := repo.CreateItem(item); err != nil {
				return err
			}
			if s.onItemCreatedHook != nil {
				if err := s.onItemCreatedHook(tx); err != nil {
					return err
				}
			}
			items = append(items, WarehouseTransferDraftItemResponse{ID: uint64(item.ID), TransferID: item.TransferID, ProductID: item.ProductID, Quantity: item.Quantity})
		}
		response = &WarehouseTransferDraftResponse{ID: uint64(transfer.ID), ReferenceNumber: transfer.ReferenceNumber, FromWarehouseID: transfer.FromWarehouseID, ToWarehouseID: transfer.ToWarehouseID, TransactionDate: date.Format("2006-01-02"), Status: transfer.Status, Note: transfer.Note, CreatedBy: transfer.CreatedBy, PostedAt: transfer.PostedAt, Items: items}
		return nil
	})
	return response, err
}

func (s *WarehouseTransferService) DeleteDraft(ctx context.Context, id uint64) error {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	return db.Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewWarehouseTransferRepository(tx)
		transfer, err := repo.FindByIDForUpdate(id)
		if err != nil {
			return err
		}
		if transfer.Status == StockStatusPosted {
			return ErrWarehouseTransferAlreadyPosted
		}
		if transfer.Status != StockStatusDraft {
			return fmt.Errorf("%w: %s", ErrWarehouseTransferInvalidStatus, transfer.Status)
		}
		if err := repo.DeleteItemsByTransferID(id); err != nil {
			return err
		}
		return repo.DeleteByID(id)
	})
}

func validateTransferDraftUpdateInput(input CreateWarehouseTransferDraftInput) (time.Time, error) {
	if input.FromWarehouseID == 0 {
		return time.Time{}, ErrWarehouseTransferMissingSourceWarehouseID
	}
	if input.ToWarehouseID == 0 {
		return time.Time{}, ErrWarehouseTransferMissingDestinationWarehouseID
	}
	if input.FromWarehouseID == input.ToWarehouseID {
		return time.Time{}, ErrWarehouseTransferSameWarehouse
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(input.TransactionDate))
	if err != nil {
		return time.Time{}, ErrWarehouseTransferInvalidDate
	}
	if len(input.Items) == 0 {
		return time.Time{}, ErrWarehouseTransferEmptyItems
	}
	seen := map[uint64]bool{}
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return time.Time{}, ErrWarehouseTransferInvalidProduct
		}
		if item.Quantity <= 0 {
			return time.Time{}, fmt.Errorf("%w: product ID %d", ErrWarehouseTransferInvalidQuantity, item.ProductID)
		}
		if seen[item.ProductID] {
			return time.Time{}, fmt.Errorf("%w: product ID %d", ErrWarehouseTransferDuplicateItem, item.ProductID)
		}
		seen[item.ProductID] = true
	}
	return date, nil
}

func validateTransferMasters(tx contractsorm.Query, input CreateWarehouseTransferDraftInput) error {
	warehouses := repositories.NewWarehouseRepository(tx)
	from, err := warehouses.FindByID(input.FromWarehouseID)
	if err != nil {
		if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
			return fmt.Errorf("%w: ID %d", ErrWarehouseTransferSourceWarehouseNotFound, input.FromWarehouseID)
		}
		return err
	}
	if !from.IsActive {
		return fmt.Errorf("%w: ID %d", ErrWarehouseTransferSourceWarehouseInactive, input.FromWarehouseID)
	}
	to, err := warehouses.FindByID(input.ToWarehouseID)
	if err != nil {
		if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
			return fmt.Errorf("%w: ID %d", ErrWarehouseTransferDestinationWarehouseNotFound, input.ToWarehouseID)
		}
		return err
	}
	if !to.IsActive {
		return fmt.Errorf("%w: ID %d", ErrWarehouseTransferDestinationWarehouseInactive, input.ToWarehouseID)
	}
	products := repositories.NewProductRepository(tx)
	for _, item := range input.Items {
		product, err := products.FindByID(item.ProductID)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return fmt.Errorf("%w: ID %d", ErrWarehouseTransferProductNotFound, item.ProductID)
			}
			return err
		}
		if !product.IsActive {
			return fmt.Errorf("%w: ID %d", ErrWarehouseTransferProductInactive, item.ProductID)
		}
	}
	return nil
}
