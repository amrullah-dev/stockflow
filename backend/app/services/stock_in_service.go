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
	StockStatusDraft  = "DRAFT"
	StockStatusPosted = "POSTED"
	stockMovementIn   = "IN"
	stockInReference  = "StockIn"
	stockInPostAction = "POST_STOCK_IN"
)

var (
	ErrStockInAlreadyPosted      = errors.New("stock-in is already posted")
	ErrStockInInvalidStatus      = errors.New("stock-in cannot be posted from its current status")
	ErrStockInEmptyItems         = errors.New("stock-in must contain at least one item")
	ErrStockInInvalidProduct     = errors.New("stock-in item product ID must be greater than zero")
	ErrStockInInvalidQty         = errors.New("stock-in item quantity must be greater than zero")
	ErrStockInDuplicateItem      = errors.New("stock-in contains duplicate product items")
	ErrStockInNil                = errors.New("stock-in is nil")
	ErrStockInSupplierNotFound   = errors.New("supplier not found")
	ErrStockInSupplierInactive   = errors.New("supplier is inactive")
	ErrStockInWarehouseNotFound  = errors.New("warehouse not found")
	ErrStockInWarehouseInactive  = errors.New("warehouse is inactive")
	ErrStockInProductNotFound    = errors.New("product not found")
	ErrStockInProductInactive    = errors.New("product is inactive")
	ErrStockInInvalidDate        = errors.New("invalid transaction date: format must be YYYY-MM-DD")
	ErrStockInMissingSupplierID  = errors.New("supplier ID must be greater than zero")
	ErrStockInMissingWarehouseID = errors.New("warehouse ID must be greater than zero")
	ErrStockInInvalidUnitPrice   = errors.New("stock-in item unit price must be greater than or equal to zero")
	ErrStockInReferenceCollision = errors.New("unable to generate unique reference number after retries")
)

type CreateStockInItemInput struct {
	ProductID uint64  `json:"product_id"`
	Quantity  int64   `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type CreateStockInDraftInput struct {
	SupplierID      uint64                   `json:"supplier_id"`
	WarehouseID     uint64                   `json:"warehouse_id"`
	TransactionDate string                   `json:"transaction_date"`
	Note            string                   `json:"note"`
	Items           []CreateStockInItemInput `json:"items"`
}

type StockInDraftItemResponse struct {
	ID        uint64  `json:"id"`
	StockInID uint64  `json:"stock_in_id"`
	ProductID uint64  `json:"product_id"`
	Quantity  int64   `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`
}

type StockInDraftResponse struct {
	ID              uint64                     `json:"id"`
	ReferenceNumber string                     `json:"reference_number"`
	SupplierID      uint64                     `json:"supplier_id"`
	WarehouseID     uint64                     `json:"warehouse_id"`
	TransactionDate string                     `json:"transaction_date"`
	Status          string                     `json:"status"`
	Note            string                     `json:"note"`
	CreatedBy       uint64                     `json:"created_by"`
	PostedAt        *time.Time                 `json:"posted_at"`
	Items           []StockInDraftItemResponse `json:"items"`
}

type StockInService struct {
	db                  contractsorm.Orm
	onHeaderCreatedHook func(tx contractsorm.Query) error
	onItemCreatedHook   func(tx contractsorm.Query) error
}

func NewStockInService(db contractsorm.Orm) *StockInService {
	return &StockInService{db: db}
}

func (s *StockInService) SetOnHeaderCreatedHook(hook func(tx contractsorm.Query) error) {
	s.onHeaderCreatedHook = hook
}

func (s *StockInService) SetOnItemCreatedHook(hook func(tx contractsorm.Query) error) {
	s.onItemCreatedHook = hook
}

func (s *StockInService) Post(ctx context.Context, stockInID, actorID uint64) error {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	return db.Transaction(func(tx contractsorm.Query) error {
		stockInRepository := repositories.NewStockInRepository(tx)
		warehouseStockRepository := repositories.NewWarehouseStockRepository(tx)
		stockMovementRepository := repositories.NewStockMovementRepository(tx)
		auditLogRepository := repositories.NewAuditLogRepository(tx)
		auditLogService := NewAuditLogService(auditLogRepository)

		stockIn, err := stockInRepository.FindByIDForUpdate(stockInID)
		fmt.Printf("[DEBUG SERVICE POST] Panggil Service Post untuk ID: %d, Err: %v\n", stockInID, err)
		if err != nil {
			return err
		}

		if err := validateStockInForPosting(stockIn); err != nil {
			return err
		}

		productIDs := make([]uint64, 0, len(stockIn.Items))
		products := make(map[uint64]struct{}, len(stockIn.Items))
		for _, item := range stockIn.Items {
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
				stockIn.WarehouseID,
				productID,
			)
			if err != nil {
				if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
					stock = &models.WarehouseStock{WarehouseID: stockIn.WarehouseID, ProductID: productID, Quantity: 0}
					if err := warehouseStockRepository.Create(stock); err != nil {
						return err
					}
				} else {
					return err
				}
			}
			lockedStocks[productID] = stock
		}

		for _, item := range stockIn.Items {
			stock := lockedStocks[item.ProductID]

			balanceBefore := stock.Quantity
			balanceAfter := balanceBefore + item.Quantity

			if err := warehouseStockRepository.UpdateQuantity(stock, balanceAfter); err != nil {
				fmt.Printf("[DEBUG EXEC] Step Stock Update Product %d, Err: %v\n", item.ProductID, err)
				return err
			}
			fmt.Printf("[DEBUG EXEC] Step Stock Update Product %d, Err: %v\n", item.ProductID, nil)

			referenceID := uint64(stockIn.ID)
			createdBy := actorID
			movement := &models.StockMovement{
				WarehouseID:   stockIn.WarehouseID,
				ProductID:     item.ProductID,
				Type:          stockMovementIn,
				Quantity:      item.Quantity,
				BalanceBefore: balanceBefore,
				BalanceAfter:  balanceAfter,
				ReferenceType: stockInReference,
				ReferenceID:   &referenceID,
				CreatedBy:     &createdBy,
			}

			if err := stockMovementRepository.Create(movement); err != nil {
				fmt.Printf("[DEBUG EXEC] Step Movement Insert, Err: %v\n", err)
				return err
			}
			fmt.Printf("[DEBUG EXEC] Step Movement Insert, Err: %v\n", nil)
		}

		oldValues := "status=" + stockIn.Status
		newValues := "status=" + StockStatusPosted
		if err := auditLogService.Create(
			&actorID,
			stockInPostAction,
			stockInReference,
			uint64(stockIn.ID),
			oldValues,
			newValues,
			"",
			"",
		); err != nil {
			fmt.Printf("[DEBUG EXEC] Step Audit Log Insert, Err: %v\n", err)
			return err
		}
		fmt.Printf("[DEBUG EXEC] Step Audit Log Insert, Err: %v\n", nil)

		if err := stockInRepository.UpdateStatus(stockIn, StockStatusPosted); err != nil {
			return err
		}

		postedAt := time.Now()
		return stockInRepository.UpdatePostedAt(stockIn, &postedAt)
	})
}

func validateStockInForPosting(stockIn *models.StockIn) error {
	if stockIn == nil {
		return ErrStockInNil
	}

	if stockIn.Status == StockStatusPosted {
		return ErrStockInAlreadyPosted
	}

	fmt.Printf("[DEBUG VALIDATE] Status: %s, Items Count: %d\n", stockIn.Status, len(stockIn.Items))
	if strings.ToUpper(stockIn.Status) != "DRAFT" {
		return fmt.Errorf("%w: %s", ErrStockInInvalidStatus, stockIn.Status)
	}

	if len(stockIn.Items) == 0 {
		return ErrStockInEmptyItems
	}

	products := make(map[uint64]struct{}, len(stockIn.Items))
	for _, item := range stockIn.Items {
		if item.ProductID == 0 {
			return ErrStockInInvalidProduct
		}

		if !(item.Quantity > 0) {
			return fmt.Errorf("%w: product ID %d", ErrStockInInvalidQty, item.ProductID)
		}

		if _, exists := products[item.ProductID]; exists {
			return fmt.Errorf("%w: product ID %d", ErrStockInDuplicateItem, item.ProductID)
		}

		products[item.ProductID] = struct{}{}
	}

	return nil
}

// CreateDraft creates a new stock-in transaction with status DRAFT.
// It performs atomic validation and persistence without modifying warehouse inventory.
func (s *StockInService) CreateDraft(ctx context.Context, input CreateStockInDraftInput, actorID uint64) (*StockInDraftResponse, error) {
	if input.SupplierID == 0 {
		return nil, ErrStockInMissingSupplierID
	}
	if input.WarehouseID == 0 {
		return nil, ErrStockInMissingWarehouseID
	}

	dateStr := strings.TrimSpace(input.TransactionDate)
	if dateStr == "" {
		return nil, ErrStockInInvalidDate
	}
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, ErrStockInInvalidDate
	}

	if len(input.Items) == 0 {
		return nil, ErrStockInEmptyItems
	}

	seenProducts := make(map[uint64]struct{}, len(input.Items))
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return nil, ErrStockInInvalidProduct
		}
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: product ID %d", ErrStockInInvalidQty, item.ProductID)
		}
		if item.UnitPrice < 0 {
			return nil, fmt.Errorf("%w: product ID %d", ErrStockInInvalidUnitPrice, item.ProductID)
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return nil, fmt.Errorf("%w: product ID %d", ErrStockInDuplicateItem, item.ProductID)
		}
		seenProducts[item.ProductID] = struct{}{}
	}

	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	var response *StockInDraftResponse

	err = db.Transaction(func(tx contractsorm.Query) error {
		supplierRepo := repositories.NewSupplierRepository(tx)
		warehouseRepo := repositories.NewWarehouseRepository(tx)
		productRepo := repositories.NewProductRepository(tx)
		stockInRepo := repositories.NewStockInRepository(tx)

		supplier, err := supplierRepo.FindByID(input.SupplierID)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return fmt.Errorf("%w: ID %d", ErrStockInSupplierNotFound, input.SupplierID)
			}
			return err
		}
		if !supplier.IsActive {
			return fmt.Errorf("%w: ID %d", ErrStockInSupplierInactive, input.SupplierID)
		}

		warehouse, err := warehouseRepo.FindByID(input.WarehouseID)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return fmt.Errorf("%w: ID %d", ErrStockInWarehouseNotFound, input.WarehouseID)
			}
			return err
		}
		if !warehouse.IsActive {
			return fmt.Errorf("%w: ID %d", ErrStockInWarehouseInactive, input.WarehouseID)
		}

		for _, item := range input.Items {
			product, err := productRepo.FindByID(item.ProductID)
			if err != nil {
				if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
					return fmt.Errorf("%w: ID %d", ErrStockInProductNotFound, item.ProductID)
				}
				return err
			}
			if !product.IsActive {
				return fmt.Errorf("%w: ID %d", ErrStockInProductInactive, item.ProductID)
			}
		}

		var refNumber string
		var created bool
		var stockIn *models.StockIn

		for attempt := 0; attempt < 3; attempt++ {
			refNumber, err = GenerateStockInReference()
			if err != nil {
				return err
			}

			if existing, findErr := stockInRepo.FindByReferenceNumber(refNumber); findErr == nil && existing != nil {
				continue
			}

			stockIn = &models.StockIn{
				ReferenceNumber: refNumber,
				SupplierID:      input.SupplierID,
				WarehouseID:     input.WarehouseID,
				TransactionDate: parsedDate,
				Status:          StockStatusDraft,
				Note:            input.Note,
				CreatedBy:       actorID,
			}

			if createErr := stockInRepo.Create(stockIn); createErr != nil {
				if isDuplicateEntryError(createErr) {
					continue
				}
				return createErr
			}

			created = true
			break
		}

		if !created {
			return ErrStockInReferenceCollision
		}

		if s.onHeaderCreatedHook != nil {
			if err := s.onHeaderCreatedHook(tx); err != nil {
				return err
			}
		}

		itemResponses := make([]StockInDraftItemResponse, 0, len(input.Items))
		for _, itemInput := range input.Items {
			subtotal := float64(itemInput.Quantity) * itemInput.UnitPrice
			itemModel := &models.StockInItem{
				StockInID: uint64(stockIn.ID),
				ProductID: itemInput.ProductID,
				Quantity:  itemInput.Quantity,
				UnitPrice: itemInput.UnitPrice,
				Subtotal:  subtotal,
			}

			if err := stockInRepo.CreateItem(itemModel); err != nil {
				return err
			}

			itemResponses = append(itemResponses, StockInDraftItemResponse{
				ID:        uint64(itemModel.ID),
				StockInID: itemModel.StockInID,
				ProductID: itemModel.ProductID,
				Quantity:  itemModel.Quantity,
				UnitPrice: itemModel.UnitPrice,
				Subtotal:  itemModel.Subtotal,
			})
		}

		response = &StockInDraftResponse{
			ID:              uint64(stockIn.ID),
			ReferenceNumber: stockIn.ReferenceNumber,
			SupplierID:      stockIn.SupplierID,
			WarehouseID:     stockIn.WarehouseID,
			TransactionDate: parsedDate.Format("2006-01-02"),
			Status:          stockIn.Status,
			Note:            stockIn.Note,
			CreatedBy:       stockIn.CreatedBy,
			PostedAt:        stockIn.PostedAt,
			Items:           itemResponses,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *StockInService) UpdateDraft(ctx context.Context, id uint64, input CreateStockInDraftInput) (*StockInDraftResponse, error) {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	var response *StockInDraftResponse
	err := db.Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewStockInRepository(tx)
		stockIn, err := repo.FindByIDForUpdate(id)
		if err != nil {
			return err
		}
		if stockIn.Status == StockStatusPosted {
			return ErrStockInAlreadyPosted
		}
		if stockIn.Status != StockStatusDraft {
			return fmt.Errorf("%w: %s", ErrStockInInvalidStatus, stockIn.Status)
		}
		date, err := validateStockInDraftUpdateInput(input)
		if err != nil {
			return err
		}
		if err := validateStockInMasters(tx, input); err != nil {
			return err
		}
		if err := repo.DeleteItemsByStockInID(id); err != nil {
			return err
		}
		if err := repo.UpdateHeader(stockIn, repositories.UpdateStockInHeaderData{TransactionDate: date, SupplierID: input.SupplierID, WarehouseID: input.WarehouseID, Note: input.Note}); err != nil {
			return err
		}
		items := make([]StockInDraftItemResponse, 0, len(input.Items))
		for _, in := range input.Items {
			item := &models.StockInItem{StockInID: uint64(stockIn.ID), ProductID: in.ProductID, Quantity: in.Quantity, UnitPrice: in.UnitPrice, Subtotal: float64(in.Quantity) * in.UnitPrice}
			if err := repo.CreateItem(item); err != nil {
				return err
			}
			if s.onItemCreatedHook != nil {
				if err := s.onItemCreatedHook(tx); err != nil {
					return err
				}
			}
			items = append(items, StockInDraftItemResponse{ID: uint64(item.ID), StockInID: item.StockInID, ProductID: item.ProductID, Quantity: item.Quantity, UnitPrice: item.UnitPrice, Subtotal: item.Subtotal})
		}
		response = &StockInDraftResponse{ID: uint64(stockIn.ID), ReferenceNumber: stockIn.ReferenceNumber, SupplierID: stockIn.SupplierID, WarehouseID: stockIn.WarehouseID, TransactionDate: date.Format("2006-01-02"), Status: stockIn.Status, Note: stockIn.Note, CreatedBy: stockIn.CreatedBy, PostedAt: stockIn.PostedAt, Items: items}
		return nil
	})
	return response, err
}

func (s *StockInService) DeleteDraft(ctx context.Context, id uint64) error {
	db := s.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	return db.Transaction(func(tx contractsorm.Query) error {
		repo := repositories.NewStockInRepository(tx)
		stockIn, err := repo.FindByIDForUpdate(id)
		if err != nil {
			return err
		}
		if stockIn.Status == StockStatusPosted {
			return ErrStockInAlreadyPosted
		}
		if stockIn.Status != StockStatusDraft {
			return fmt.Errorf("%w: %s", ErrStockInInvalidStatus, stockIn.Status)
		}
		if err := repo.DeleteItemsByStockInID(id); err != nil {
			return err
		}
		return repo.DeleteByID(id)
	})
}

func validateStockInDraftUpdateInput(input CreateStockInDraftInput) (time.Time, error) {
	if input.SupplierID == 0 {
		return time.Time{}, ErrStockInMissingSupplierID
	}
	if input.WarehouseID == 0 {
		return time.Time{}, ErrStockInMissingWarehouseID
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(input.TransactionDate))
	if err != nil {
		return time.Time{}, ErrStockInInvalidDate
	}
	if len(input.Items) == 0 {
		return time.Time{}, ErrStockInEmptyItems
	}
	seen := map[uint64]bool{}
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return time.Time{}, ErrStockInInvalidProduct
		}
		if item.Quantity <= 0 {
			return time.Time{}, fmt.Errorf("%w: product ID %d", ErrStockInInvalidQty, item.ProductID)
		}
		if item.UnitPrice < 0 {
			return time.Time{}, fmt.Errorf("%w: product ID %d", ErrStockInInvalidUnitPrice, item.ProductID)
		}
		if seen[item.ProductID] {
			return time.Time{}, fmt.Errorf("%w: product ID %d", ErrStockInDuplicateItem, item.ProductID)
		}
		seen[item.ProductID] = true
	}
	return date, nil
}

func validateStockInMasters(tx contractsorm.Query, input CreateStockInDraftInput) error {
	supplier, err := repositories.NewSupplierRepository(tx).FindByID(input.SupplierID)
	if err != nil {
		if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
			return fmt.Errorf("%w: ID %d", ErrStockInSupplierNotFound, input.SupplierID)
		}
		return err
	}
	if !supplier.IsActive {
		return fmt.Errorf("%w: ID %d", ErrStockInSupplierInactive, input.SupplierID)
	}
	warehouse, err := repositories.NewWarehouseRepository(tx).FindByID(input.WarehouseID)
	if err != nil {
		if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
			return fmt.Errorf("%w: ID %d", ErrStockInWarehouseNotFound, input.WarehouseID)
		}
		return err
	}
	if !warehouse.IsActive {
		return fmt.Errorf("%w: ID %d", ErrStockInWarehouseInactive, input.WarehouseID)
	}
	for _, item := range input.Items {
		product, err := repositories.NewProductRepository(tx).FindByID(item.ProductID)
		if err != nil {
			if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
				return fmt.Errorf("%w: ID %d", ErrStockInProductNotFound, item.ProductID)
			}
			return err
		}
		if !product.IsActive {
			return fmt.Errorf("%w: ID %d", ErrStockInProductInactive, item.ProductID)
		}
	}
	return nil
}
