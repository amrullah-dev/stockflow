package repositories

import (
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"

	"goravel/app/models"
)

type UserRepository struct{ query contractsorm.Query }

func NewUserRepository(query contractsorm.Query) *UserRepository {
	return &UserRepository{query: query}
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	if err := r.query.With("Role").Where("email = ?", email).FindOrFail(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateLastLoginAt(user *models.User, value time.Time) error {
	_, err := r.query.Model(user).Update("last_login_at", value)
	return err
}
