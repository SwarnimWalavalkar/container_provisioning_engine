package types

import "time"

type User struct {
	ID        *int       `db:"id" json:"-"`
	UUID      *string    `db:"uuid" json:"uuid"`
	Username  *string    `db:"username" json:"username"`
	ApiKey    *string    `db:"api_key" json:"-"`
	CreatedAt *time.Time `db:"created_at" json:"-"`
	UpdatedAt *time.Time `db:"updated_at" json:"-"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required" db:"username"`
	ApiKey   string `json:"apiKey" binding:"required" db:"apiKey"`
}

type AuthTokenRequest struct {
	Username string `json:"username" binding:"required"`
	ApiKey   string `json:"apiKey" binding:"required"`
}
