package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

type PostgresPersonRepository struct {
	db *sqlx.DB
}

func NewPostgresPersonRepository(db *sqlx.DB) *PostgresPersonRepository {
	return &PostgresPersonRepository{db: db}
}

func (r *PostgresPersonRepository) Create(ctx context.Context, person *domain.Person) error {
	query := `
		INSERT INTO persons (user_id, identifier, name, phone, is_active, birth_date) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, person.UserID, person.Identifier, person.Name, person.Phone, person.IsActive, person.BirthDate).
		Scan(&person.ID, &person.CreatedAt, &person.UpdatedAt)

	return err
}

func (r *PostgresPersonRepository) FindByID(ctx context.Context, id string) (*domain.Person, error) {
	var person domain.Person
	query := `SELECT id, user_id, identifier, name, phone, is_active, birth_date, created_at, updated_at FROM persons WHERE id = $1`

	err := r.db.GetContext(ctx, &person, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &person, nil
}

func (r *PostgresPersonRepository) FindAll(ctx context.Context) ([]*domain.Person, error) {
	var persons []*domain.Person
	query := `SELECT id, user_id, identifier, name, phone, is_active, birth_date, created_at, updated_at FROM persons ORDER BY name ASC`

	err := r.db.SelectContext(ctx, &persons, query)
	if err != nil {
		return nil, err
	}
	return persons, nil
}

func (r *PostgresPersonRepository) Update(ctx context.Context, person *domain.Person) error {
	query := `
		UPDATE persons 
		SET user_id = $1, identifier = $2, name = $3, phone = $4, is_active = $5, birth_date = $6, updated_at = $7
		WHERE id = $8
	`
	person.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, person.UserID, person.Identifier, person.Name, person.Phone, person.IsActive, person.BirthDate, person.UpdatedAt, person.ID)
	return err
}
