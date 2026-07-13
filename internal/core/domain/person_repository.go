package domain

import "context"

type PersonRepository interface {
	Create(ctx context.Context, person *Person) error
	FindByID(ctx context.Context, id string) (*Person, error)
	FindAll(ctx context.Context) ([]*Person, error)
	Update(ctx context.Context, person *Person) error
}
