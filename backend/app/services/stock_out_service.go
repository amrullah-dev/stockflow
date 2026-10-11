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
	stockMovementOut   = "OUT"
	stockOutReference  = "StockOut"
	stockOutPostAction = "POST_STOCK_OUT"
)

var (
	ErrStockOutAlreadyPosted      = errors.New("stock-out is already posted")
	ErrStockOutInvalidStatus      = errors.New("stock-out cannot be posted from its current status")
	ErrStockOutEmptyItems         = errors.New("stock-out must contain at least one item")
	ErrStockOutInvalidProduct     = errors.New("stock-out item product ID must be greater than zero")
	ErrStockOutInvalidQty         = errors.New("stock-out item quantity must be greater than zero")
	ErrStockOutDuplicateItem      = errors.New("stock-out contains duplicate product items")
	ErrStockOutNil                = errors.New("stock-out is nil")
	ErrStockOutInsufficientStock  = errors.New("stock-out has insufficient stock")
	ErrStockOutWarehouseNotFound  = errors.New("warehouse not found")
	ErrStockOutWarehouseInactive  = errors.New("warehouse is inactive")
	ErrStockOutProductNotFound    = errors.New("product not found")
	ErrStockOutProductInactive    = errors.New("product is inactive")
	ErrStockOutInvalidDate        = errors.New("invalid transaction date: format must be YYYY-MM-DD")
	ErrStockOutMissingWarehouseID = errors.New("warehouse ID must be greater than zero")
	ErrStockOutReferenceCollision = errors.New("unable to generate unique reference number after retries")
)

type CreateStockOutItemInput struct {
	ProductID uint64 `json:"product_id"`
	Quantity  int64  `json:"quantity"`
}

type CreateStockOutDraftInput struct {
	WarehouseID     uint64                    `json:"warehouse_id"`
	TransactionDate string                    `json:"transaction_date"`
	Destination     string                    `json:"destination"`
	Note            string                    `json:"note"`
	Items           []CreateStockOutItemInput `json:"items"`
}

type StockOutDraftItemResponse struct {
	ID         uint64 `json:"id"`
	StockOutID uint64 `json:"stock_out_id"`
	ProductID  uint64 `json:"product_id"`
	Quantity   int64  `json:"quantity"`
}

type StockOutDraftResponse struct {
	ID              uint64                      `json:"id"`
	ReferenceNumber string                      `json:"reference_number"`
	WarehouseID     uint64                      `json:"warehouse_id"`
	TransactionDate string                      `json:"transaction_date"`
	Status          string                      `json:"status"`
	Destination     string                      `json:"destination"`
	Note            string                      `json:"note"`
	CreatedBy       uint64                      `json:"created_by"`
	PostedAt        *time.Time                  `json:"posted_at"`
	Items           []StockOutDraftItemResponse `json:"items"`
}

type StockOutService struct {
	db                  contractsorm.Orm
	onHeaderCreatedHook func(tx contractsorm.Query) error
	onItemCreatedHook   func(tx contractsorm.Query) error
}

func NewStockOutService(db contractsorm.Orm) *StockOutService {
	return &StockOutService{db: db}
}

func (s *StockOutService) SetOnHeaderCreatedHook(hook func(tx contractsorm.Query) error) {
	s.onHeaderCreatedHook = hook
}

func (s *StockOutService) SetOnItemCreatedHook(hook func(tx contractsorm.Query) error) {
	s.onItemCreatedHook = hook
}

func (s *StockOutService) Post(ctx context.Context, stockOutID, actorID uint64) error {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	return db.Transaction(func(tx contractsorm.Query) error {
		stockOutRepository := repositories.NewStockOutRepository(tx)
		warehouseStockRepository := repositories.NewWarehouseStockRepository(tx)
		stockMovementRepository := repositories.NewStockMovementRepository(tx)
		auditLogRepository := repositories.NewAuditLogRepository(tx)
		auditLogService := NewAuditLogService(auditLogRepository)

		stockOut, err := stockOutRepository.FindByIDForUpdate(stockOutID)
		fmt.Printf("[DEBUG SERVICE POST] Panggil Service Post untuk ID: %d, Err: %v\n", stockOutID, err)
		if err != nil {
			return err
		}

		if err := validateStockOutForPosting(stockOut); err != nil {
			return err
		}

		productIDs := make([]uint64, 0, len(stockOut.Items))
		products := make(map[uint64]struct{}, len(stockOut.Items))
		for _, item := range stockOut.Items {
			if _, exists := products[item.ProductID]; exists {
				continue
			}
			products[item.ProductID] = struct{}{}
			productIDs = append(productIDs, item.ProductID)
		}
		sort.Slice(productIDs, func(i, j int) bool {
			return productIDs[i] < productIDs[j]
		})

		lockedStocks := make(map[uint64]*models.WarehouseStock, len(productIDs))
		for _, productID := range productIDs {
			stock, err := warehouseStockRepository.FindByWarehouseAndProductForUpdate(
				stockOut.WarehouseID,
				productID,
			)
			if err != nil {
				if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
					return fmt.Errorf("%w: product ID %d", ErrStockOutInsufficientStock, productID)
				}
				return err
			}
			lockedStocks[productID] = stock
		}

		for _, item := range stockOut.Items {
			stock := lockedStocks[item.ProductID]
			if stock.Quantity < item.Quantity {
				return fmt.Errorf("%w: product ID %d", ErrStockOutInsufficientStock, item.ProductID)
			}
		}

		for _, item := range stockOut.Items {
			stock := lockedStocks[item.ProductID]

			balanceBefore := stock.Quantity

			balanceAfter := balanceBefore - item.Quantity
			if err := warehouseStockRepository.UpdateQuantity(stock, balanceAfter); err != nil {
				fmt.Printf("[DEBUG EXEC] Step Stock Update Product %d, Err: %v\n", item.ProductID, err)
				return err
			}
			fmt.Printf("[DEBUG EXEC] Step Stock Update Product %d, Err: %v\n", item.ProductID, nil)

			referenceID := uint64(stockOut.ID)
			createdBy := actorID
			movement := &models.StockMovement{
				WarehouseID:   stockOut.WarehouseID,
				ProductID:     item.ProductID,
				Type:          stockMovementOut,
				Quantity:      item.Quantity,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceAfter,
				ReferenceType: stockOutReference,
				ReferenceID:   &referenceID,
				CreatedBy:     &createdBy,
			}

			if err := stockMovementRepository.Create(movement); err != nil {
				fmt.Printf("[DEBUG EXEC] Step Movement Insert, Err: %v\n", err)
				return err
			}
			fmt.Printf("[DEBUG EXEC] Step Movement Insert, Err: %v\n", nil)
		}

		oldValues := "status=" + stockOut.Status
		newValues := "status=" + StockStatusPosted
		if err := auditLogService.Create(
			&actorID,
			stockOutPostAction,
			stockOutReference,
			uint64(stockOut.ID),
			oldValues,
			newValues,
			"",
			"",
		); err != nil {
			fmt.Printf("[DEBUG EXEC] Step Audit Log Insert, Err: %v\n", err)
			return err
		}
		fmt.Printf("[DEBUG EXEC] Step Audit Log Insert, Err: %v\n", nil)

		if err := stockOutRepository.UpdateStatus(stockOut, StockStatusPosted); err != nil {
			return err
		}

		postedAt := time.Now()
		return stockOutRepository.UpdatePostedAt(stockOut, &postedAt)
	})
}

func validateStockOutForPosting(stockOut *models.StockOut) error {
	if stockOut == nil {
		return ErrStockOutNil
	}

	if stockOut.Status == StockStatusPosted {
		return ErrStockOutAlreadyPosted
	}

	fmt.Printf("[DEBUG VALIDATE] Status: %s, Items Count: %d\n", stockOut.Status, len(stockOut.Items))
	if strings.ToUpper(stockOut.Status) != "DRAFT" {
		return fmt.Errorf("%w: %s", ErrStockOutInvalidStatus, stockOut.Status)
	}

	if len(stockOut.Items) == 0 {
		return ErrStockOutEmptyItems
	}

	products := make(map[uint64]struct{}, len(stockOut.Items))
	for _, item := range stockOut.Items {
		if item.ProductID == 0 {
			return ErrStockOutInvalidProduct
		}

		if !(item.Quantity > 0) {
			return fmt.Errorf("%w: product ID %d", ErrStockOutInvalidQty, item.ProductID)
		}

		if _, exists := products[item.ProductID]; exists {
			return fmt.Errorf("%w: product ID %d", ErrStockOutDuplicateItem, item.ProductID)
		}

		products[item.ProductID] = struct{}{}
	}

	return nil
}

// CreateDraft creates a new stock-out transaction with status DRAFT.
// It performs atomic validation and persistence without modifying warehouse inventory or checking stock levels.
func (s *StockOutService) CreateDraft(ctx context.Context, input CreateStockOutDraftInput, actorID uint64) (*StockOutDraftResponse, error) {
	if input.WarehouseID == 0 {
		return nil, ErrStockOutMissingWarehouseID
	}

	dateStr := strings.TrimSpace(input.TransactionDate)
	if dateStr == "" {
		return nil, ErrStockOutInvalidDate
	}
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, ErrStockOutInvalidDate
	}

	if len(input.Items) == 0 {
		return nil, ErrStockOutEmptyItems
	}

	seenProducts := make(map[uint64]struct{}, len(input.Items))
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return nil, ErrStockOutInvalidProduct
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: product ID %d", ErrStockOutInvalidQty, item.ProductID)
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return nil, fmt.Errorf("%w: product ID %d", ErrStockOutDuplicateItem, item.ProductID)
		}
		seenProducts[item.ProductID] = struct{}{}
	}

	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	var response *StockOutDraftResponse

	err = db.Transaction(func(tx contractsorm.Query) error {
		warehouseRepo := repositories.NewWarehouseRepository(tx)
		productRepo := repositories.NewProductRepository(tx)
		stockOutRepo := repositories.NewStockOutRepository(tx)

		warehouse, err := warehouseRepo.FindByID(input.WarehouseID)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return fmt.Errorf("%w: ID %d", ErrStockOutWarehouseNotFound, input.WarehouseID)
			}
			return err
		}
		if !warehouse.IsActive {
			return fmt.Errorf("%w: ID %d", ErrStockOutWarehouseInactive, input.WarehouseID)
		}

		for _, item := range input.Items {
			product, err := productRepo.FindByID(item.ProductID)
			if err != nil {
				if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
					return fmt.Errorf("%w: ID %d", ErrStockOutProductNotFound, item.ProductID)
				}
				return err
			}
			if !product.IsActive {
				return fmt.Errorf("%w: ID %d", ErrStockOutProductInactive, item.ProductID)
			}
		}

		var refNumber string
		var created bool
		var stockOut *models.StockOut

		for attempt := 0; attempt < 3; attempt++ {
			refNumber, err = GenerateStockOutReference()
			if err != nil {
				return err
			}

			if existing, findErr := stockOutRepo.FindByReferenceNumber(refNumber); findErr == nil && existing != nil {
				continue
			}

			stockOut = &models.StockOut{
				ReferenceNumber: refNumber,
				WarehouseID:     input.WarehouseID,
				TransactionDate: parsedDate,
				Status:          StockStatusDraft,
				Destination:     input.Destination,
				Note:            input.Note,
				CreatedBy:       actorID,
			}

			if createErr := stockOutRepo.Create(stockOut); createErr != nil {
				if isDuplicateEntryError(createErr) {
					continue
				}
				return createErr
			}

			created = true
			break
		}

		if !created {
			return ErrStockOutReferenceCollision
		}

		if s.onHeaderCreatedHook != nil {
			if err := s.onHeaderCreatedHook(tx); err != nil {
				return err
			}
		}

		itemResponses := make([]StockOutDraftItemResponse, 0, len(input.Items))
		for _, itemInput := range input.Items {
			itemModel := &models.StockOutItem{
				StockOutID: uint64(stockOut.ID),
				ProductID:  itemInput.ProductID,
				Quantity:   itemInput.Quantity,
			}

			if err := stockOutRepo.CreateItem(itemModel); err != nil {
				return err
			}

			itemResponses = append(itemResponses, StockOutDraftItemResponse{
				ID:         uint64(itemModel.ID),
				StockOutID: itemModel.StockOutID,
				ProductID:  itemModel.ProductID,
				Quantity:   itemModel.Quantity,
			})
		}

		response = &StockOutDraftResponse{
			ID:              uint64(stockOut.ID),
			ReferenceNumber: stockOut.ReferenceNumber,
			WarehouseID:     stockOut.WarehouseID,
			TransactionDate: parsedDate.Format("2006-01-02"),
			Status:          stockOut.Status,
			Destination:     stockOut.Destination,
			Note:            stockOut.Note,
			CreatedBy:       stockOut.CreatedBy,
			PostedAt:        stockOut.PostedAt,
			Items:           itemResponses,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *StockOutService) UpdateDraft(ctx context.Context, id uint64, input CreateStockOutDraftInput) (*StockOutDraftResponse, error) {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	var response *StockOutDraftResponse
	err := db.Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewStockOutRepository(tx)
		stockOut, err := repo.FindByIDForUpdate(id)
		if err != nil {
			return err
		}
		if stockOut.Status == StockStatusPosted {
			return ErrStockOutAlreadyPosted
		}
		if stockOut.Status != StockStatusDraft {
			return fmt.Errorf("%w: %s", ErrStockOutInvalidStatus, stockOut.Status)
		}
		date, err := validateStockOutDraftUpdateInput(input)
		if err != nil {
			return err
		}
		if err := validateStockOutMasters(tx, input); err != nil {
			return err
		}
		if err := repo.DeleteItemsByStockOutID(id); err != nil {
			return err
		}
		if err := repo.UpdateHeader(stockOut, repositories.UpdateStockOutHeaderData{TransactionDate: date, WarehouseID: input.WarehouseID, Destination: input.Destination, Note: input.Note}); err != nil {
			return err
		}
		items := make([]StockOutDraftItemResponse, 0, len(input.Items))
		for _, in := range input.Items {
			item := &models.StockOutItem{StockOutID: uint64(stockOut.ID), ProductID: in.ProductID, Quantity: in.Quantity}
			if err := repo.CreateItem(item); err != nil {
				return err
			}
			if s.onItemCreatedHook != nil {
				if err := s.onItemCreatedHook(tx); err != nil {
					return err
				}
			}
			items = append(items, StockOutDraftItemResponse{ID: uint64(item.ID), StockOutID: item.StockOutID, ProductID: item.ProductID, Quantity: item.Quantity})
		}
		response = &StockOutDraftResponse{ID: uint64(stockOut.ID), ReferenceNumber: stockOut.ReferenceNumber, WarehouseID: stockOut.WarehouseID, TransactionDate: date.Format("2006-01-02"), Status: stockOut.Status, Destination: stockOut.Destination, Note: stockOut.Note, CreatedBy: stockOut.CreatedBy, PostedAt: stockOut.PostedAt, Items: items}
		return nil
	})
	return response, err
}

func (s *StockOutService) DeleteDraft(ctx context.Context, id uint64) error {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	return db.Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewStockOutRepository(tx)
		stockOut, err := repo.FindByIDForUpdate(id)
		if err != nil {
			return err
		}
		if stockOut.Status == StockStatusPosted {
			return ErrStockOutAlreadyPosted
		}
		if stockOut.Status != StockStatusDraft {
			return fmt.Errorf("%w: %s", ErrStockOutInvalidStatus, stockOut.Status)
		}
		if err := repo.DeleteItemsByStockOutID(id); err != nil {
			return err
		}
		return repo.DeleteByID(id)
	})
}

func validateStockOutDraftUpdateInput(input CreateStockOutDraftInput) (time.Time, error) {
	if input.WarehouseID == 0 {
		return time.Time{}, ErrStockOutMissingWarehouseID
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(input.TransactionDate))
	if err != nil {
		return time.Time{}, ErrStockOutInvalidDate
	}
	if len(input.Items) == 0 {
		return time.Time{}, ErrStockOutEmptyItems
	}
	seen := map[uint64]bool{}
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return time.Time{}, ErrStockOutInvalidProduct
		}
		if item.Quantity <= 0 {
			return time.Time{}, fmt.Errorf("%w: product ID %d", ErrStockOutInvalidQty, item.ProductID)
		}
		if seen[item.ProductID] {
			return time.Time{}, fmt.Errorf("%w: product ID %d", ErrStockOutDuplicateItem, item.ProductID)
		}
		seen[item.ProductID] = true
	}
	return date, nil
}

func validateStockOutMasters(tx contractsorm.Query, input CreateStockOutDraftInput) error {
	warehouse, err := repositories.NewWarehouseRepository(tx).FindByID(input.WarehouseID)
	if err != nil {
		if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
			return fmt.Errorf("%w: ID %d", ErrStockOutWarehouseNotFound, input.WarehouseID)
		}
		return err
	}
	if !warehouse.IsActive {
		return fmt.Errorf("%w: ID %d", ErrStockOutWarehouseInactive, input.WarehouseID)
	}
	for _, item := range input.Items {
		product, err := repositories.NewProductRepository(tx).FindByID(item.ProductID)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return fmt.Errorf("%w: ID %d", ErrStockOutProductNotFound, item.ProductID)
			}
			return err
		}
		if !product.IsActive {
			return fmt.Errorf("%w: ID %d", ErrStockOutProductInactive, item.ProductID)
		}
	}
	return nil
}
