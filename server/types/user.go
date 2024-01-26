package types

import "time"

type User struct {
	ID        *int       `db:"id" json:"-"`
	UUID      *string    `db:"uuid" json:"uuid"`
	Name      *string    `db:"name" json:"name"`
	ApiKey    *string    `db:"api_key" json:"apiKey"`
	CreatedAt *time.Time `db:"created_at" json:"-"`
	UpdatedAt *time.Time `db:"updated_at" json:"-"`
}

type CreateUserRequest struct {
	Name   string `json:"name" binding:"required" db:"name"`
	ApiKey string `json:"apiKey" binding:"required" db:"apiKey"`
}

type AuthTokenRequest struct {
	ApiKey string `json:"apiKey" binding:"required"`
}
