package services

import (
	"errors"
	"sort"
	"time"

	"goravel/app/models"
	"goravel/app/repositories"
)

var ErrStockOutQueryInvalidID = errors.New("stock-out ID must be greater than zero")

type StockOutWarehouseResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type StockOutUserResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type StockOutProductResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type StockOutItemResponse struct {
	ID         uint64                  `json:"id"`
	StockOutID uint64                  `json:"stock_out_id"`
	ProductID  uint64                  `json:"product_id"`
	Quantity   int64                   `json:"quantity"`
	Product    StockOutProductResponse `json:"product"`
}

type StockOutListResponse struct {
	ID              uint64                    `json:"id"`
	ReferenceNumber string                    `json:"reference_number"`
	WarehouseID     uint64                    `json:"warehouse_id"`
	TransactionDate time.Time                 `json:"transaction_date"`
	Status          string                    `json:"status"`
	Destination     string                    `json:"destination"`
	Note            string                    `json:"note"`
	CreatedBy       uint64                    `json:"created_by"`
	PostedAt        *time.Time                `json:"posted_at"`
	Warehouse       StockOutWarehouseResponse `json:"warehouse"`
}

type StockOutDetailResponse struct {
	StockOutListResponse
	CreatedByUser StockOutUserResponse   `json:"created_by_user"`
	Items         []StockOutItemResponse `json:"items"`
}

type StockOutQueryService struct {
	repository *repositories.StockOutRepository
}

func NewStockOutQueryService(repository *repositories.StockOutRepository) *StockOutQueryService {
	return &StockOutQueryService{repository: repository}
}

func (s *StockOutQueryService) FindAll() ([]StockOutListResponse, error) {
	stockOuts, err := s.repository.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]StockOutListResponse, 0, len(stockOuts))
	for _, stockOut := range stockOuts {
		responses = append(responses, stockOutListResponse(stockOut))
	}

	return responses, nil
}

func (s *StockOutQueryService) FindByID(id uint64) (*StockOutDetailResponse, error) {
	if id == 0 {
		return nil, ErrStockOutQueryInvalidID
	}

	stockOut, err := s.repository.FindByIDForRead(id)
	if err != nil {
		return nil, err
	}

	sort.Slice(stockOut.Items, func(i, j int) bool {
		return stockOut.Items[i].ID < stockOut.Items[j].ID
	})

	response := stockOutDetailResponse(*stockOut)
	return &response, nil
}

func stockOutListResponse(stockOut models.StockOut) StockOutListResponse {
	return StockOutListResponse{
		ID:              uint64(stockOut.ID),
		ReferenceNumber: stockOut.ReferenceNumber,
		WarehouseID:     stockOut.WarehouseID,
		TransactionDate: stockOut.TransactionDate,
		Status:          stockOut.Status,
		Destination:     stockOut.Destination,
		Note:            stockOut.Note,
		CreatedBy:       stockOut.CreatedBy,
		PostedAt:        stockOut.PostedAt,
		Warehouse: StockOutWarehouseResponse{
			ID:   uint64(stockOut.Warehouse.ID),
			Code: stockOut.Warehouse.Code,
			Name: stockOut.Warehouse.Name,
		},
	}
}

func stockOutDetailResponse(stockOut models.StockOut) StockOutDetailResponse {
	items := make([]StockOutItemResponse, 0, len(stockOut.Items))
	for _, item := range stockOut.Items {
		items = append(items, StockOutItemResponse{
			ID:         uint64(item.ID),
			StockOutID: item.StockOutID,
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Product: StockOutProductResponse{
				ID:   uint64(item.Product.ID),
				Code: item.Product.Code,
				Name: item.Product.Name,
				Unit: item.Product.Unit,
			},
		})
	}

	return StockOutDetailResponse{
		StockOutListResponse: stockOutListResponse(stockOut),
		CreatedByUser: StockOutUserResponse{
			ID:    uint64(stockOut.CreatedByUser.ID),
			Name:  stockOut.CreatedByUser.Name,
			Email: stockOut.CreatedByUser.Email,
		},
		Items: items,
	}
}
