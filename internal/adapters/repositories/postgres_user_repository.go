package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

type PostgresUserRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, status) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash, user.Status).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password_hash, status, created_at, updated_at FROM users WHERE email = $1`
	
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil 
		}
		return nil, err
	}
	return &user, nil
}

// GetUserRoles executa um JOIN clássico para descobrir quais os papéis daquele usuário naquela aplicação
func (r *PostgresUserRepository) GetUserRoles(ctx context.Context, userID string, appID string) ([]string, error) {
	query := `
		SELECT r.name 
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.application_id = $2
	`
	var roles []string
	// SelectContext varre todas as linhas retornadas e joga dentro do array (slice)
	err := r.db.SelectContext(ctx, &roles, query, userID, appID)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
