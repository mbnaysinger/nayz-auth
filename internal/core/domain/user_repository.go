package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByIdentifier(ctx context.Context, identifier string) (*User, error)
	FindByEmailOrUsername(ctx context.Context, email string, username string) (*User, error)
	FindAllWithPerson(ctx context.Context) ([]*UserWithPerson, error)
	GetUserRoles(ctx context.Context, userID string, appID string) ([]string, error)
	// GetUserPermissions resolve as permissões efetivas (via roles) do usuário na aplicação
	GetUserPermissions(ctx context.Context, userID string, appID string) ([]string, error)
	// FindRolesByUser lista as roles do usuário em todas as aplicações (visão administrativa)
	FindRolesByUser(ctx context.Context, userID string) ([]*Role, error)
	SetActive(ctx context.Context, userID string, isActive bool) error
}
