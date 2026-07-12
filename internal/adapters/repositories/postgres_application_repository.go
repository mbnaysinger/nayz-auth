package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

type PostgresApplicationRepository struct {
	db *sqlx.DB
}

func NewPostgresApplicationRepository(db *sqlx.DB) *PostgresApplicationRepository {
	return &PostgresApplicationRepository{db: db}
}

func (r *PostgresApplicationRepository) FindByID(ctx context.Context, id string) (*domain.Application, error) {
	var app domain.Application
	var authMethods pq.StringArray // Necessário porque o driver do postgres lê arrays como um tipo específico

	query := `SELECT id, name, auth_methods, is_active FROM applications WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&app.ID, &app.Name, &authMethods, &app.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	
	// Transforma o array nativo do Postgres de volta pro Slice padrão do Go []string
	app.AuthMethods = authMethods
	return &app, nil
}

func (r *PostgresApplicationRepository) FindAll(ctx context.Context) ([]*domain.Application, error) {
	query := `SELECT id, name, auth_methods, is_active FROM applications ORDER BY name ASC`
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	var apps []*domain.Application
	for rows.Next() {
		var app domain.Application
		var authMethods pq.StringArray
		if err := rows.Scan(&app.ID, &app.Name, &authMethods, &app.IsActive); err != nil {
			return nil, err
		}
		app.AuthMethods = authMethods
		apps = append(apps, &app)
	}
	return apps, nil
}

func (r *PostgresApplicationRepository) Create(ctx context.Context, app *domain.Application) error {
	query := `INSERT INTO applications (name, auth_methods, is_active) VALUES ($1, $2, $3) RETURNING id`
	// pq.Array() encapsula nosso Slice do Go e converte nativamente para o Array do Postgres
	err := r.db.QueryRowContext(ctx, query, app.Name, pq.Array(app.AuthMethods), app.IsActive).Scan(&app.ID)
	return err
}

func (r *PostgresApplicationRepository) Update(ctx context.Context, app *domain.Application) error {
	query := `UPDATE applications SET name = $1, auth_methods = $2, is_active = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, app.Name, pq.Array(app.AuthMethods), app.IsActive, app.ID)
	return err
}

func (r *PostgresApplicationRepository) Delete(ctx context.Context, id string) error {
	// A deleção em cascata garantirá que user_applications e roles associadas a essa aplicação também sejam destruídas
	query := `DELETE FROM applications WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
