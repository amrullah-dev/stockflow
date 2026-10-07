package repositories

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

type CategoryRepository struct{ query contractsorm.Query }

func NewCategoryRepository(query contractsorm.Query) *CategoryRepository {
	return &CategoryRepository{query: query}
}

func (r *CategoryRepository) FindAll() ([]models.Category, error) {
	var categories []models.Category
	if err := r.query.OrderBy("id", "asc").Find(&categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) FindByID(id uint64) (*models.Category, error) {
	category := &models.Category{}
	if err := r.query.Where("id = ?", id).FindOrFail(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (r *CategoryRepository) Create(category *models.Category) error { return r.query.Create(category) }

func (r *CategoryRepository) Update(category *models.Category, values map[string]any) error {
	_, err := r.query.Model(category).Update(values)
	return err
}

func (r *CategoryRepository) Delete(category *models.Category) error {
	_, err := r.query.Where("id = ?", category.ID).Delete(&models.Category{})
	return err
}
