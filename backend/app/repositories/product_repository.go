package repositories

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

// ProductRepository handles read and write access for products.
type ProductRepository struct {
	query contractsorm.Query
}

func NewProductRepository(query contractsorm.Query) *ProductRepository {
	return &ProductRepository{query: query}
}

func (r *ProductRepository) FindAll() ([]models.Product, error) {
	var products []models.Product

	if err := r.query.
		With("Category").
		OrderBy("id", "asc").
		Find(&products); err != nil {
		return nil, err
	}

	return products, nil
}

// FindByID returns a product by ID. Soft-deleted products are automatically
// excluded by the ORM scope.
func (r *ProductRepository) FindByID(id uint64) (*models.Product, error) {
	product := &models.Product{}

	if err := r.query.
		Where("id = ?", id).
		FindOrFail(product); err != nil {
		return nil, err
	}

	return product, nil
}
