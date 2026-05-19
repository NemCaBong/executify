package domain

import "time"

type User struct {
	ID           int
	UUID         string
	Email        string
	Username     string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
