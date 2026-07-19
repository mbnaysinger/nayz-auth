package domain

import "context"

// Permission modela uma capacidade no formato recurso:ação (ex.: "squads:manage").
// Permissões compõem roles (role_permissions) e chegam às aplicações via claim
// "permissions" no JWT — a aplicação cliente decide onde aplicar o gate.
type Permission struct {
	ID            string `json:"id" db:"id"`
	ApplicationID string `json:"application_id" db:"application_id"`
	Name          string `json:"name" db:"name"`
}

type PermissionRepository interface {
	Create(ctx context.Context, permission *Permission) error
	FindByApp(ctx context.Context, appID string) ([]*Permission, error)
	FindByRole(ctx context.Context, roleID string) ([]*Permission, error)
	Delete(ctx context.Context, id string) error
	AttachToRole(ctx context.Context, roleID string, permissionID string) error
	DetachFromRole(ctx context.Context, roleID string, permissionID string) error
}
