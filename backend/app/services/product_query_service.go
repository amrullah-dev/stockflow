package services

import (
	"goravel/app/models"
	"goravel/app/repositories"
)

type ProductCategoryResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type ProductResponse struct {
	ID           uint64                  `json:"id"`
	CategoryID   uint64                  `json:"category_id"`
	Code         string                  `json:"code"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Unit         string                  `json:"unit"`
	MinimumStock int64                   `json:"minimum_stock"`
	IsActive     bool                    `json:"is_active"`
	Category     ProductCategoryResponse `json:"category"`
}

type ProductQueryService struct {
	repository *repositories.ProductRepository
}

func NewProductQueryService(repository *repositories.ProductRepository) *ProductQueryService {
	return &ProductQueryService{repository: repository}
}

func (s *ProductQueryService) FindAll() ([]ProductResponse, error) {
	products, err := s.repository.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, productResponse(product))
	}

	return responses, nil
}

func productResponse(product models.Product) ProductResponse {
	return ProductResponse{
		ID:           uint64(product.ID),
		CategoryID:   product.CategoryID,
		Code:         product.Code,
		Name:         product.Name,
		Description:  product.Description,
		Unit:         product.Unit,
		MinimumStock: product.MinimumStock,
		IsActive:     product.IsActive,
		Category: ProductCategoryResponse{
			ID:          uint64(product.Category.ID),
			Name:        product.Category.Name,
			Description: product.Category.Description,
			IsActive:    product.Category.IsActive,
		},
	}
}
