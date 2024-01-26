package database

import (
	"context"

	"github.com/SwarnimWalavalkar/aether/types"
)

func (d *Database) GetUserByUUID(ctx context.Context, uuid string) (types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE uuid = $1`

	if err := d.Client.GetContext(ctx, &user, query, uuid); err != nil {
		return types.User{}, err
	}

	return user, nil
}

func (d *Database) GetUserByAPIKey(ctx context.Context, apiKey string) (types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE api_key = $1`

	if err := d.Client.GetContext(ctx, &user, query, apiKey); err != nil {
		return types.User{}, err
	}

	return user, nil
}

func (d *Database) CreateUser(ctx context.Context, userAttributes types.CreateUserRequest) (types.User, error) {
	var uuid string
	if err := d.Client.QueryRowxContext(ctx, `INSERT INTO users (name, api_key) VALUES ($1, $2) RETURNING uuid`, userAttributes.Name, userAttributes.ApiKey).Scan(&uuid); err != nil {
		return types.User{}, err
	}

	user, err := d.GetUserByUUID(ctx, uuid)
	if err != nil {
		return types.User{}, err
	}

	return user, nil

}
