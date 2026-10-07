package services

import (
	"errors"
	"sort"
	"time"

	"goravel/app/models"
	"goravel/app/repositories"
)

type StockInSupplierResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type StockInWarehouseResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type StockInUserResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type StockInProductResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type StockInItemResponse struct {
	ID        uint64                 `json:"id"`
	StockInID uint64                 `json:"stock_in_id"`
	ProductID uint64                 `json:"product_id"`
	Quantity  int64                  `json:"quantity"`
	UnitPrice float64                `json:"unit_price"`
	Subtotal  float64                `json:"subtotal"`
	Product   StockInProductResponse `json:"product"`
}

type StockInListResponse struct {
	ID              uint64                   `json:"id"`
	ReferenceNumber string                   `json:"reference_number"`
	SupplierID      uint64                   `json:"supplier_id"`
	WarehouseID     uint64                   `json:"warehouse_id"`
	TransactionDate time.Time                `json:"transaction_date"`
	Status          string                   `json:"status"`
	Note            string                   `json:"note"`
	CreatedBy       uint64                   `json:"created_by"`
	PostedAt        *time.Time               `json:"posted_at"`
	Supplier        StockInSupplierResponse  `json:"supplier"`
	Warehouse       StockInWarehouseResponse `json:"warehouse"`
}

type StockInDetailResponse struct {
	StockInListResponse
	CreatedByUser StockInUserResponse   `json:"created_by_user"`
	Items         []StockInItemResponse `json:"items"`
}

type StockInQueryService struct {
	repository *repositories.StockInRepository
}

func NewStockInQueryService(repository *repositories.StockInRepository) *StockInQueryService {
	return &StockInQueryService{repository: repository}
}

func (s *StockInQueryService) FindAll() ([]StockInListResponse, error) {
	stockIns, err := s.repository.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]StockInListResponse, 0, len(stockIns))
	for _, stockIn := range stockIns {
		responses = append(responses, stockInListResponse(stockIn))
	}

	return responses, nil
}

func (s *StockInQueryService) FindByID(id uint64) (*StockInDetailResponse, error) {
	if id == 0 {
		return nil, ErrStockInQueryInvalidID
	}

	stockIn, err := s.repository.FindByIDForRead(id)
	if err != nil {
		return nil, err
	}

	sort.Slice(stockIn.Items, func(i, j int) bool {
		return stockIn.Items[i].ID < stockIn.Items[j].ID
	})

	response := stockInDetailResponse(*stockIn)
	return &response, nil
}

var ErrStockInQueryInvalidID = errors.New("stock-in ID must be greater than zero")

func stockInListResponse(stockIn models.StockIn) StockInListResponse {
	return StockInListResponse{
		ID:              uint64(stockIn.ID),
		ReferenceNumber: stockIn.ReferenceNumber,
		SupplierID:      stockIn.SupplierID,
		WarehouseID:     stockIn.WarehouseID,
		TransactionDate: stockIn.TransactionDate,
		Status:          stockIn.Status,
		Note:            stockIn.Note,
		CreatedBy:       stockIn.CreatedBy,
		PostedAt:        stockIn.PostedAt,
		Supplier:        stockInSupplierResponse(stockIn.Supplier),
		Warehouse:       stockInWarehouseResponse(stockIn.Warehouse),
	}
}

func stockInDetailResponse(stockIn models.StockIn) StockInDetailResponse {
	items := make([]StockInItemResponse, 0, len(stockIn.Items))
	for _, item := range stockIn.Items {
		items = append(items, StockInItemResponse{
			ID:        uint64(item.ID),
			StockInID: item.StockInID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal:  item.Subtotal,
			Product: StockInProductResponse{
				ID:   uint64(item.Product.ID),
				Code: item.Product.Code,
				Name: item.Product.Name,
				Unit: item.Product.Unit,
			},
		})
	}

	return StockInDetailResponse{
		StockInListResponse: stockInListResponse(stockIn),
		CreatedByUser: StockInUserResponse{
			ID:    uint64(stockIn.CreatedByUser.ID),
			Name:  stockIn.CreatedByUser.Name,
			Email: stockIn.CreatedByUser.Email,
		},
		Items: items,
	}
}

func stockInSupplierResponse(supplier models.Supplier) StockInSupplierResponse {
	return StockInSupplierResponse{
		ID:   uint64(supplier.ID),
		Code: supplier.Code,
		Name: supplier.Name,
	}
}

func stockInWarehouseResponse(warehouse models.Warehouse) StockInWarehouseResponse {
	return StockInWarehouseResponse{
		ID:   uint64(warehouse.ID),
		Code: warehouse.Code,
		Name: warehouse.Name,
	}
}
