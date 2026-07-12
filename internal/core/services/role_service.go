package services

import (
	"context"
	"errors"

	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

type RoleService struct {
	repo domain.RoleRepository
}

func NewRoleService(repo domain.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) ListByApp(ctx context.Context, appID string) ([]*domain.Role, error) {
	if appID == "" {
		return nil, errors.New("id da aplicação é obrigatório")
	}
	return s.repo.FindAllByAppID(ctx, appID)
}

func (s *RoleService) CreateRole(ctx context.Context, appID, name string) (*domain.Role, error) {
	if appID == "" || name == "" {
		return nil, errors.New("o ID da aplicação e o Nome da Role são obrigatórios")
	}
	role := &domain.Role{
		ApplicationID: appID,
		Name:          name,
	}
	err := s.repo.Create(ctx, role)
	return role, err
}

func (s *RoleService) DeleteRole(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *RoleService) AssignUser(ctx context.Context, userID, roleID string) error {
	if userID == "" || roleID == "" {
		return errors.New("IDs de usuário e role são obrigatórios")
	}
	return s.repo.AssignUserToRole(ctx, userID, roleID)
}

func (s *RoleService) RemoveUser(ctx context.Context, userID, roleID string) error {
	if userID == "" || roleID == "" {
		return errors.New("IDs de usuário e role são obrigatórios")
	}
	return s.repo.RemoveUserFromRole(ctx, userID, roleID)
}
