package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByIdentifier(ctx context.Context, identifier string) (*User, error)
	FindByEmailOrUsername(ctx context.Context, email string, username string) (*User, error)
	GetUserRoles(ctx context.Context, userID string, appID string) ([]string, error)
}
