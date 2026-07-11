package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

// PostgresUserRepository é o Adaptador que implementa a interface domain.UserRepository
type PostgresUserRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create insere um novo usuário no banco e atualiza a struct com o ID e Datas gerados pelo banco
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, status) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`

	// A clausula RETURNING do Postgres combinada com Scan é a forma mais performática no Go 
	err := r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash, user.Status).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	return err
}

// FindByEmail localiza um usuário pelo email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password_hash, status, created_at, updated_at FROM users WHERE email = $1`
	
	// GetContext varre o resultado e já joga certinho nas tags `db` da Struct!
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Não encontrou o usuário, não é um erro fatal
		}
		return nil, err
	}
	return &user, nil
}
