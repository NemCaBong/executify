package entity

import (
	"time"

	"github.com/NemCaBong/executify/internal/domain"
)

type RefreshToken struct {
	ID        int        `gorm:"primaryKey;column:id;autoIncrement"`
	UserID    int        `gorm:"column:user_id;not null;index"`
	TokenHash string     `gorm:"column:token_hash;uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

func (rt *RefreshToken) ToDomain() *domain.RefreshToken {
	if rt == nil {
		return nil
	}
	return &domain.RefreshToken{
		ID:        rt.ID,
		UserID:    rt.UserID,
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
		CreatedAt: rt.CreatedAt,
		RevokedAt: rt.RevokedAt,
	}
}

func RefreshTokenFromDomain(d *domain.RefreshToken) *RefreshToken {
	if d == nil {
		return nil
	}
	return &RefreshToken{
		ID:        d.ID,
		UserID:    d.UserID,
		TokenHash: d.TokenHash,
		ExpiresAt: d.ExpiresAt,
		CreatedAt: d.CreatedAt,
		RevokedAt: d.RevokedAt,
	}
}
