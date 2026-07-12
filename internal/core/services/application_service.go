package services

import (
	"context"
	"errors"

	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

type ApplicationService struct {
	repo domain.ApplicationRepository
}

func NewApplicationService(repo domain.ApplicationRepository) *ApplicationService {
	return &ApplicationService{repo: repo}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, name string, authMethods []string) (*domain.Application, error) {
	if name == "" || len(authMethods) == 0 {
		return nil, errors.New("nome e métodos de autenticação são obrigatórios")
	}

	app := &domain.Application{
		Name:        name,
		AuthMethods: authMethods,
		IsActive:    true, // Cria ativado por padrão
	}

	if err := s.repo.Create(ctx, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *ApplicationService) ListApplications(ctx context.Context) ([]*domain.Application, error) {
	return s.repo.FindAll(ctx)
}

func (s *ApplicationService) UpdateApplication(ctx context.Context, id string, name string, authMethods []string, isActive bool) (*domain.Application, error) {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("aplicação não encontrada no banco de dados")
	}

	app.Name = name
	app.AuthMethods = authMethods
	app.IsActive = isActive

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, id string) error {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if app == nil {
		return errors.New("tentativa de excluir uma aplicação inexistente")
	}

	return s.repo.Delete(ctx, id)
}
