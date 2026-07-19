package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

type PostgresPermissionRepository struct {
	db *sqlx.DB
}

func NewPostgresPermissionRepository(db *sqlx.DB) *PostgresPermissionRepository {
	return &PostgresPermissionRepository{db: db}
}

func (r *PostgresPermissionRepository) Create(ctx context.Context, p *domain.Permission) error {
	query := `INSERT INTO permissions (application_id, name) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRowxContext(ctx, query, p.ApplicationID, p.Name).Scan(&p.ID)
}

func (r *PostgresPermissionRepository) FindByApp(ctx context.Context, appID string) ([]*domain.Permission, error) {
	permissions := []*domain.Permission{}
	query := `SELECT id, application_id, name FROM permissions WHERE application_id = $1 ORDER BY name ASC`
	err := r.db.SelectContext(ctx, &permissions, query, appID)
	return permissions, err
}

func (r *PostgresPermissionRepository) FindByRole(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	permissions := []*domain.Permission{}
	query := `
		SELECT p.id, p.application_id, p.name
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name ASC`
	err := r.db.SelectContext(ctx, &permissions, query, roleID)
	return permissions, err
}

func (r *PostgresPermissionRepository) Delete(ctx context.Context, id string) error {
	// role_permissions cai em cascata (FK ON DELETE CASCADE)
	_, err := r.db.ExecContext(ctx, `DELETE FROM permissions WHERE id = $1`, id)
	return err
}

func (r *PostgresPermissionRepository) AttachToRole(ctx context.Context, roleID string, permissionID string) error {
	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	return err
}

func (r *PostgresPermissionRepository) DetachFromRole(ctx context.Context, roleID string, permissionID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`, roleID, permissionID)
	return err
}
