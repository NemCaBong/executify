package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/NemCaBong/executify/internal/adapter/repository/entity"
	"github.com/NemCaBong/executify/internal/domain"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	e := entity.UserFromDomain(user)
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return nil, err
	}
	return e.ToDomain(), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var e entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return e.ToDomain(), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	var e entity.User
	err := r.db.WithContext(ctx).First(&e, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return e.ToDomain(), nil
}
