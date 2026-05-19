package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/NemCaBong/executify/internal/domain"
)

type User struct {
	ID           int       `gorm:"primaryKey;column:id;autoIncrement"`
	UUID         string    `gorm:"column:uuid;uniqueIndex;not null"`
	Email        string    `gorm:"column:email;uniqueIndex;not null"`
	Username     string    `gorm:"column:username;uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	IsActive     bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string { return "users" }

// BeforeCreate runs on every GORM INSERT. We mint a UUIDv7 (time-ordered,
// monotonic-ish, indexes well) here so the application layer doesn't depend on
// the MySQL-side default — which is UUIDv4 and not sortable.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID != "" {
		return nil
	}
	v7, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate user uuid v7: %w", err)
	}
	u.UUID = v7.String()
	return nil
}

func (u *User) ToDomain() *domain.User {
	if u == nil {
		return nil
	}
	return &domain.User{
		ID:           u.ID,
		UUID:         u.UUID,
		Email:        u.Email,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func UserFromDomain(d *domain.User) *User {
	if d == nil {
		return nil
	}
	return &User{
		ID:           d.ID,
		UUID:         d.UUID,
		Email:        d.Email,
		Username:     d.Username,
		PasswordHash: d.PasswordHash,
		IsActive:     d.IsActive,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
