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

// AuthService contém a verdadeira Regra de Negócios (Business Logic) da aplicação
type AuthService struct {
	userRepo domain.UserRepository
}

func NewAuthService(repo domain.UserRepository) *AuthService {
	return &AuthService{userRepo: repo}
}

// RegisterUser cria um novo usuário usando senha forte.
func (s *AuthService) RegisterUser(ctx context.Context, email, password string) (*domain.User, error) {
	
	// 1. Validar duplicidade
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err // Um erro real de banco de dados
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// 2. Hash da Senha (sal e hash automáticos pelo bcrypt)
	// DefaultCost é ideal para equilíbrio de segurança e performance
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashStr := string(hashedBytes)

	// 3. Montar a entidade de Domínio
	user := &domain.User{
		Email:        email,
		PasswordHash: &hashStr,
		Status:       "ACTIVE",
	}

	// 4. Salvar via contrato (sem importar driver do postgres aqui!)
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
