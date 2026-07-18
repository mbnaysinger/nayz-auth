package domain

import (
	"time"
)

// User representa a entidade central de um usuário no sistema.
// Os campos com `db` mapeiam diretamente para as colunas do PostgreSQL usando o sqlx.
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash *string   `json:"-" db:"password_hash"` // Ponteiro porque pode ser Nulo se for 100% Passwordless
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserWithPerson é a projeção de listagem: usuário + dados básicos da pessoa vinculada (se houver).
type UserWithPerson struct {
	User
	PersonID   *string `json:"person_id" db:"person_id"`
	PersonName *string `json:"person_name" db:"person_name"`
}
