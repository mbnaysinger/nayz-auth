package domain

import "context"

// UserRepository define o contrato (interface) para acesso aos dados do usuário.
// Padrão Inversão de Dependência: A regra de negócio conhece apenas esta interface, não sabe que é Postgres.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
}
