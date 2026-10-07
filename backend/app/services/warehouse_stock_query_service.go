package services

import (
	"errors"

	"goravel/app/models"
	"goravel/app/repositories"
)

var ErrWarehouseStockInvalidWarehouseID = errors.New("warehouse ID must be greater than zero")

type WarehouseStockProductResponse struct {
	ID           uint64 `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Unit         string `json:"unit"`
	MinimumStock int64  `json:"minimum_stock"`
	IsActive     bool   `json:"is_active"`
}

type WarehouseStockResponse struct {
	ID          uint64                        `json:"id"`
	WarehouseID uint64                        `json:"warehouse_id"`
	ProductID   uint64                        `json:"product_id"`
	Quantity    int64                         `json:"quantity"`
	Product     WarehouseStockProductResponse `json:"product"`
}

type WarehouseStockQueryService struct {
	warehouseRepository      *repositories.WarehouseRepository
	warehouseStockRepository *repositories.WarehouseStockRepository
}

func NewWarehouseStockQueryService(
	warehouseRepository *repositories.WarehouseRepository,
	warehouseStockRepository *repositories.WarehouseStockRepository,
) *WarehouseStockQueryService {
	return &WarehouseStockQueryService{
		warehouseRepository:      warehouseRepository,
		warehouseStockRepository: warehouseStockRepository,
	}
}

func (s *WarehouseStockQueryService) FindByWarehouse(warehouseID uint64) ([]WarehouseStockResponse, error) {
	if warehouseID == 0 {
		return nil, ErrWarehouseStockInvalidWarehouseID
	}

	if _, err := s.warehouseRepository.FindByID(warehouseID); err != nil {
		return nil, err
	}

	stocks, err := s.warehouseStockRepository.FindByWarehouse(warehouseID)
	if err != nil {
		return nil, err
	}

	responses := make([]WarehouseStockResponse, 0, len(stocks))
	for _, stock := range stocks {
		responses = append(responses, warehouseStockResponse(stock))
	}

	return responses, nil
}

func warehouseStockResponse(stock models.WarehouseStock) WarehouseStockResponse {
	return WarehouseStockResponse{
		ID:          uint64(stock.ID),
		WarehouseID: stock.WarehouseID,
		ProductID:   stock.ProductID,
		Quantity:    stock.Quantity,
		Product: WarehouseStockProductResponse{
			ID:           uint64(stock.Product.ID),
			Code:         stock.Product.Code,
			Name:         stock.Product.Name,
			Unit:         stock.Product.Unit,
			MinimumStock: stock.Product.MinimumStock,
			IsActive:     stock.Product.IsActive,
		},
	}
}
