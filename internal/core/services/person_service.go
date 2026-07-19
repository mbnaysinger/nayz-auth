package services

import (
	"context"
	"errors"

	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

var (
	ErrPersonNotFound = errors.New("person not found")
)

type PersonService struct {
	repo domain.PersonRepository
}

func NewPersonService(repo domain.PersonRepository) *PersonService {
	return &PersonService{repo: repo}
}

func (s *PersonService) CreatePerson(ctx context.Context, person *domain.Person) error {
	// Aqui poderíamos ter validações de CPF, formatação de telefone, etc.
	return s.repo.Create(ctx, person)
}

func (s *PersonService) GetPerson(ctx context.Context, id string) (*domain.Person, error) {
	person, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if person == nil {
		return nil, ErrPersonNotFound
	}
	return person, nil
}

func (s *PersonService) ListPersons(ctx context.Context) ([]*domain.Person, error) {
	return s.repo.FindAll(ctx)
}

// GetPersonsByIDs resolve várias pessoas em uma única consulta (uso principal: board do Tallo via gRPC)
func (s *PersonService) GetPersonsByIDs(ctx context.Context, ids []string) ([]*domain.Person, error) {
	if len(ids) == 0 {
		return []*domain.Person{}, nil
	}
	return s.repo.FindByIDs(ctx, ids)
}

func (s *PersonService) DeletePerson(ctx context.Context, id string) error {
	if _, err := s.GetPerson(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *PersonService) UpdatePerson(ctx context.Context, id string, updateData *domain.Person) error {
	person, err := s.GetPerson(ctx, id)
	if err != nil {
		return err
	}

	person.Name = updateData.Name
	person.Identifier = updateData.Identifier
	person.Phone = updateData.Phone
	person.UserID = updateData.UserID
	person.IsActive = updateData.IsActive
	person.BirthDate = updateData.BirthDate

	return s.repo.Update(ctx, person)
}
