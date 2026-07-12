package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

type PostgresRoleRepository struct {
	db *sqlx.DB
}

func NewPostgresRoleRepository(db *sqlx.DB) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}

func (r *PostgresRoleRepository) FindAllByAppID(ctx context.Context, appID string) ([]*domain.Role, error) {
	query := `SELECT id, application_id, name FROM roles WHERE application_id = $1 ORDER BY name ASC`
	var roles []*domain.Role
	err := r.db.SelectContext(ctx, &roles, query, appID)
	if roles == nil {
		roles = make([]*domain.Role, 0)
	}
	return roles, err
}

func (r *PostgresRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	query := `INSERT INTO roles (application_id, name) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRowContext(ctx, query, role.ApplicationID, role.Name).Scan(&role.ID)
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
}

// AssignUserToRole usa Transações de Banco para garantir a integridade dos dados!
func (r *PostgresRoleRepository) AssignUserToRole(ctx context.Context, userID string, roleID string) error {
	// Inicia a transação (Se der erro no meio, ele dá rollback em tudo)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { 
		return err 
	}

	// 1. Precisamos descobrir de qual App essa Role é
	var appID string
	err = tx.QueryRow(`SELECT application_id FROM roles WHERE id = $1`, roleID).Scan(&appID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Garante que o usuário está amarrado na tabela de aplicações (user_applications).
	// O 'ON CONFLICT DO NOTHING' é genial aqui: se ele já tem acesso ao app, ignora sem explodir erro.
	_, err = tx.Exec(`INSERT INTO user_applications (user_id, application_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, appID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. Adiciona a Role específica para o usuário
	_, err = tx.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, roleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Se chegou até aqui com sucesso, dá o Commit definitivo no banco!
	return tx.Commit()
}

func (r *PostgresRoleRepository) RemoveUserFromRole(ctx context.Context, userID string, roleID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`, userID, roleID)
	return err
}
