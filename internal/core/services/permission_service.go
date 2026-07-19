package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

// Convenção: permissões nomeiam capacidade no formato recurso:ação (ex.: squads:manage).
// Escopo opcional como terceiro segmento (ex.: activities:update:own).
var permissionNamePattern = regexp.MustCompile(`^[a-z0-9-]+:[a-z0-9-]+(:[a-z0-9-]+)?$`)

type PermissionService struct {
	repo domain.PermissionRepository
}

func NewPermissionService(repo domain.PermissionRepository) *PermissionService {
	return &PermissionService{repo: repo}
}

func (s *PermissionService) CreatePermission(ctx context.Context, appID, name string) (*domain.Permission, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if appID == "" {
		return nil, errors.New("o ID da aplicação é obrigatório")
	}
	if !permissionNamePattern.MatchString(name) {
		return nil, errors.New("nome de permissão inválido: use o formato recurso:acao (ex.: squads:manage)")
	}

	permission := &domain.Permission{ApplicationID: appID, Name: name}
	if err := s.repo.Create(ctx, permission); err != nil {
		return nil, err
	}
	return permission, nil
}

func (s *PermissionService) ListByApp(ctx context.Context, appID string) ([]*domain.Permission, error) {
	if appID == "" {
		return nil, errors.New("o ID da aplicação é obrigatório")
	}
	return s.repo.FindByApp(ctx, appID)
}

func (s *PermissionService) ListByRole(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	if roleID == "" {
		return nil, errors.New("o ID da role é obrigatório")
	}
	return s.repo.FindByRole(ctx, roleID)
}

func (s *PermissionService) DeletePermission(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *PermissionService) AttachToRole(ctx context.Context, roleID, permissionID string) error {
	if roleID == "" || permissionID == "" {
		return errors.New("IDs de role e permissão são obrigatórios")
	}
	return s.repo.AttachToRole(ctx, roleID, permissionID)
}

func (s *PermissionService) DetachFromRole(ctx context.Context, roleID, permissionID string) error {
	return s.repo.DetachFromRole(ctx, roleID, permissionID)
}
