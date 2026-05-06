package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/NemCaBong/executify/internal/adapter/repository/entity"
	"github.com/NemCaBong/executify/internal/domain"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	e := entity.RefreshTokenFromDomain(token)
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var e entity.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return e.ToDomain(), nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id int) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked_at", now).Error
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}
