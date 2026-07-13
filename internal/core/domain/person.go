package domain

import (
	"time"
)

type Person struct {
	ID         string     `json:"id" db:"id"`
	UserID     *string    `json:"user_id" db:"user_id"`
	Identifier string     `json:"identifier" db:"identifier"` // Ex: CPF (alfanumérico)
	Name       string     `json:"name" db:"name"`
	Phone      *string    `json:"phone" db:"phone"` // Ex: +5551999999999
	IsActive   bool       `json:"is_active" db:"is_active"`
	BirthDate  *time.Time `json:"birth_date" db:"birth_date"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}
