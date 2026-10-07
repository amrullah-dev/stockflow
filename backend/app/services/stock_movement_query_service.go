package services

import (
	"github.com/goravel/framework/support/carbon"

	"goravel/app/models"
	"goravel/app/repositories"
)

type StockMovementWarehouseResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type StockMovementProductResponse struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type StockMovementUserResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type StockMovementListResponse struct {
	ID            uint64                         `json:"id"`
	WarehouseID   uint64                         `json:"warehouse_id"`
	Warehouse     StockMovementWarehouseResponse `json:"warehouse"`
	ProductID     uint64                         `json:"product_id"`
	Product       StockMovementProductResponse   `json:"product"`
	Type          string                         `json:"type"`
	Quantity      int64                          `json:"quantity"`
	BalanceBefore int64                          `json:"balance_before"`
	BalanceAfter  int64                          `json:"balance_after"`
	ReferenceType string                         `json:"reference_type"`
	ReferenceID   *uint64                        `json:"reference_id"`
	Note          string                         `json:"note"`
	CreatedBy     *uint64                        `json:"created_by"`
	CreatedByUser *StockMovementUserResponse     `json:"created_by_user"`
	CreatedAt     *carbon.DateTime               `json:"created_at"`
}

type StockMovementQueryService struct {
	repository *repositories.StockMovementRepository
}

func NewStockMovementQueryService(repository *repositories.StockMovementRepository) *StockMovementQueryService {
	return &StockMovementQueryService{repository: repository}
}

func (s *StockMovementQueryService) FindAll() ([]StockMovementListResponse, error) {
	movements, err := s.repository.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]StockMovementListResponse, 0, len(movements))
	for _, movement := range movements {
		responses = append(responses, stockMovementListResponse(movement))
	}

	return responses, nil
}

func stockMovementListResponse(movement models.StockMovement) StockMovementListResponse {
	var createdByUser *StockMovementUserResponse
	if movement.CreatedByUser != nil {
		createdByUser = &StockMovementUserResponse{
			ID:    uint64(movement.CreatedByUser.ID),
			Name:  movement.CreatedByUser.Name,
			Email: movement.CreatedByUser.Email,
		}
	}

	return StockMovementListResponse{
		ID:            uint64(movement.ID),
		WarehouseID:   movement.WarehouseID,
		Warehouse:     StockMovementWarehouseResponse{ID: uint64(movement.Warehouse.ID), Code: movement.Warehouse.Code, Name: movement.Warehouse.Name},
		ProductID:     movement.ProductID,
		Product:       StockMovementProductResponse{ID: uint64(movement.Product.ID), Code: movement.Product.Code, Name: movement.Product.Name, Unit: movement.Product.Unit},
		Type:          movement.Type,
		Quantity:      movement.Quantity,
		BalanceBefore: movement.BalanceBefore,
		BalanceAfter:  movement.BalanceAfter,
		ReferenceType: movement.ReferenceType,
		ReferenceID:   movement.ReferenceID,
		Note:          movement.Note,
		CreatedBy:     movement.CreatedBy,
		CreatedByUser: createdByUser,
		CreatedAt:     movement.CreatedAt,
	}
}
