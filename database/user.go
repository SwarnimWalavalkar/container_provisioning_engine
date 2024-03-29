package database

import (
	"context"

	"github.com/SwarnimWalavalkar/container_provisioning_engine/types"
)

func (d *Database) GetUserByUUID(ctx context.Context, uuid string) (types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE uuid = $1`

	if err := d.Client.GetContext(ctx, &user, query, uuid); err != nil {
		return types.User{}, err
	}

	return user, nil
}

func (d *Database) GetUserByUsername(ctx context.Context, username string) (types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE username = $1`

	if err := d.Client.GetContext(ctx, &user, query, username); err != nil {
		return types.User{}, err
	}

	return user, nil
}

func (d *Database) GetUserByAPIKeyHash(ctx context.Context, apiKeyHash string) (types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE api_key = $1`

	if err := d.Client.GetContext(ctx, &user, query, apiKeyHash); err != nil {
		return types.User{}, err
	}

	return user, nil
}

func (d *Database) CreateUser(ctx context.Context, userAttributes types.CreateUserRequest) (types.User, error) {
	var uuid string
	if err := d.Client.QueryRowxContext(ctx, `INSERT INTO users (username, api_key) VALUES ($1, $2) RETURNING uuid`, userAttributes.Username, userAttributes.ApiKey).Scan(&uuid); err != nil {
		return types.User{}, err
	}

	user, err := d.GetUserByUUID(ctx, uuid)
	if err != nil {
		return types.User{}, err
	}

	return user, nil

}
