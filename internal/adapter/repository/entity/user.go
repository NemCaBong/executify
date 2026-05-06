package entity

import (
	"time"

	"github.com/NemCaBong/executify/internal/domain"
)

type User struct {
	ID           int       `gorm:"primaryKey;column:id;autoIncrement"`
	Email        string    `gorm:"column:email;uniqueIndex;not null"`
	Username     string    `gorm:"column:username;uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	IsActive     bool      `gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string { return "users" }

func (u *User) ToDomain() *domain.User {
	if u == nil {
		return nil
	}
	return &domain.User{
		ID:           u.ID,
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
		Email:        d.Email,
		Username:     d.Username,
		PasswordHash: d.PasswordHash,
		IsActive:     d.IsActive,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
