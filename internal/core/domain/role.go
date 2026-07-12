package domain

import "context"

type Role struct {
	ID            string `json:"id" db:"id"`
	ApplicationID string `json:"application_id" db:"application_id"`
	Name          string `json:"name" db:"name"`
}

type RoleRepository interface {
	FindAllByAppID(ctx context.Context, appID string) ([]*Role, error)
	Create(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
	
	// Gerenciamento de vínculo de usuários com roles
	AssignUserToRole(ctx context.Context, userID string, roleID string) error
	RemoveUserFromRole(ctx context.Context, userID string, roleID string) error
}
