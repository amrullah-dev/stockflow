package services

import (
	"goravel/app/models"
	"goravel/app/repositories"
)

type WarehouseResponse struct {
	ID          uint64 `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type WarehouseQueryService struct {
	repository *repositories.WarehouseRepository
}

func NewWarehouseQueryService(repository *repositories.WarehouseRepository) *WarehouseQueryService {
	return &WarehouseQueryService{repository: repository}
}

func (s *WarehouseQueryService) FindAll() ([]WarehouseResponse, error) {
	warehouses, err := s.repository.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]WarehouseResponse, 0, len(warehouses))
	for _, warehouse := range warehouses {
		responses = append(responses, warehouseResponse(warehouse))
	}

	return responses, nil
}

func warehouseResponse(warehouse models.Warehouse) WarehouseResponse {
	return WarehouseResponse{
		ID:          uint64(warehouse.ID),
		Code:        warehouse.Code,
		Name:        warehouse.Name,
		Location:    warehouse.Location,
		Description: warehouse.Description,
		IsActive:    warehouse.IsActive,
	}
}
