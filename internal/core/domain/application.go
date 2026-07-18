package domain

import "context"

type Application struct {
	ID            string   `json:"id" db:"id"`
	Name          string   `json:"name" db:"name"`
	AuthMethods   []string `json:"auth_methods"`
	IsActive      bool     `json:"is_active" db:"is_active"`
	RequirePerson bool     `json:"require_person" db:"require_person"`
}

// ApplicationRepository é a Porta para gerenciar Aplicações
type ApplicationRepository interface {
	FindByID(ctx context.Context, id string) (*Application, error)
	FindAll(ctx context.Context) ([]*Application, error)
	Create(ctx context.Context, app *Application) error
	Update(ctx context.Context, app *Application) error
	Delete(ctx context.Context, id string) error
}
