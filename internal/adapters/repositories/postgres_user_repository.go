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
		INSERT INTO users (username, email, password_hash, is_active) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.PasswordHash, user.IsActive).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, username, email, password_hash, is_active, created_at, updated_at FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) FindByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, username, email, password_hash, is_active, created_at, updated_at FROM users WHERE email = $1 OR username = $1`

	err := r.db.GetContext(ctx, &user, query, identifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) FindByEmailOrUsername(ctx context.Context, email string, username string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, username, email, password_hash, is_active, created_at, updated_at FROM users WHERE email = $1 OR username = $2`

	err := r.db.GetContext(ctx, &user, query, email, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindAllWithPerson lista os usuários com os dados básicos da pessoa vinculada (LEFT JOIN: pessoa é opcional)
func (r *PostgresUserRepository) FindAllWithPerson(ctx context.Context) ([]*domain.UserWithPerson, error) {
	var users []*domain.UserWithPerson
	query := `
		SELECT u.id, u.username, u.email, u.is_active, u.created_at, u.updated_at,
		       p.id AS person_id, p.name AS person_name
		FROM users u
		LEFT JOIN persons p ON p.user_id = u.id
		ORDER BY u.username ASC`

	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserPermissions resolve as permissões efetivas do usuário na aplicação (união das roles)
func (r *PostgresUserRepository) GetUserPermissions(ctx context.Context, userID string, appID string) ([]string, error) {
	permissions := []string{}
	query := `
		SELECT DISTINCT p.name
		FROM user_roles ur
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE ur.user_id = $1 AND p.application_id = $2
		ORDER BY p.name`
	err := r.db.SelectContext(ctx, &permissions, query, userID, appID)
	return permissions, err
}

// FindRolesByUser lista as roles do usuário em todas as aplicações
func (r *PostgresUserRepository) FindRolesByUser(ctx context.Context, userID string) ([]*domain.Role, error) {
	roles := []*domain.Role{}
	query := `
		SELECT r.id, r.application_id, r.name
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name`
	err := r.db.SelectContext(ctx, &roles, query, userID)
	return roles, err
}

func (r *PostgresUserRepository) SetActive(ctx context.Context, userID string, isActive bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2`, isActive, userID)
	return err
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
