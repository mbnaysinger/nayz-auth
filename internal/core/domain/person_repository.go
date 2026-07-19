package domain

import "context"

type PersonRepository interface {
	Create(ctx context.Context, person *Person) error
	FindByID(ctx context.Context, id string) (*Person, error)
	FindByUserID(ctx context.Context, userID string) (*Person, error)
	FindByIDs(ctx context.Context, ids []string) ([]*Person, error)
	FindAll(ctx context.Context) ([]*Person, error)
	Update(ctx context.Context, person *Person) error
	Delete(ctx context.Context, id string) error
}
