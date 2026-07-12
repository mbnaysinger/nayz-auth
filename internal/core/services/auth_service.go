package services

import (
	"context"
	"errors"

	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("já existe um usuário cadastrado com este e-mail")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
)

// AuthService compila todas as dependências do domínio para Auth
type AuthService struct {
	userRepo   domain.UserRepository
	appRepo    domain.ApplicationRepository
	jwtService *JWTService
}

func NewAuthService(userRepo domain.UserRepository, appRepo domain.ApplicationRepository, jwtService *JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		appRepo:    appRepo,
		jwtService: jwtService,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, email, password string) (*domain.User, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashStr := string(hashedBytes)

	user := &domain.User{
		Email:        email,
		PasswordHash: &hashStr,
		Status:       "ACTIVE",
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Login valida o acesso de um usuário para uma aplicação específica e gera o JWT.
func (s *AuthService) Login(ctx context.Context, appID, email, password string) (string, error) {
	// 1. A aplicação tentada existe e está ativa?
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return "", err
	}
	if app == nil || !app.IsActive {
		return "", errors.New("aplicação fornecida não existe ou está inativa")
	}

	// 2. A aplicação permite login por senha? (Aqui brilhou a nossa configuração!)
	allowsPassword := false
	for _, method := range app.AuthMethods {
		if method == "PASSWORD" {
			allowsPassword = true
			break
		}
	}
	if !allowsPassword {
		return "", errors.New("esta aplicação não aceita login com senha (configurada apenas para passwordless)")
	}

	// 3. Usuário existe e está ativo?
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil || user.Status != "ACTIVE" {
		return "", ErrInvalidCredentials
	}
	if user.PasswordHash == nil {
		return "", ErrInvalidCredentials // Usuário é puramente passwordless, não tem hash!
	}

	// 4. A senha bate com o Hash?
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// 5. Ele tem algum acesso (Role) nesta aplicação?
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID, appID)
	if err != nil {
		return "", err
	}
	if len(roles) == 0 {
		return "", errors.New("acesso negado: você não tem permissões para esta aplicação")
	}

	// 6. Tudo perfeito! Gera e devolve o JWT
	return s.jwtService.GenerateToken(user, appID, roles)
}
