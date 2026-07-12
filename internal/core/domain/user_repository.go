package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	// GetUserRoles busca todas as roles que este usuário possui para uma aplicação específica
	GetUserRoles(ctx context.Context, userID string, appID string) ([]string, error)
}
