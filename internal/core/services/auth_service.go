package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("já existe um usuário cadastrado com este e-mail")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrPersonRequired     = errors.New("esta aplicação exige uma pessoa vinculada ao usuário; contate o administrador")
	ErrNoRolesForApp      = errors.New("acesso negado: você não tem permissões para esta aplicação")
)

type AuthService struct {
	userRepo     domain.UserRepository
	appRepo      domain.ApplicationRepository
	personRepo   domain.PersonRepository
	jwtService   *JWTService
	redisClient  *redis.Client
	emailService *EmailService
}

func NewAuthService(
	userRepo domain.UserRepository,
	appRepo domain.ApplicationRepository,
	personRepo domain.PersonRepository,
	jwtService *JWTService,
	redisClient *redis.Client,
	emailService *EmailService,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		appRepo:      appRepo,
		personRepo:   personRepo,
		jwtService:   jwtService,
		redisClient:  redisClient,
		emailService: emailService,
	}
}

// issueTokens concentra a emissão de sessão: valida roles na aplicação, resolve a
// pessoa vinculada (obrigatória quando a aplicação exige), gera o par de tokens e
// registra o refresh token no Redis. Usado por Login, PasswordlessVerify e Refresh.
func (s *AuthService) issueTokens(ctx context.Context, user *domain.User, app *domain.Application) (string, string, error) {
	roles, err := s.userRepo.GetUserRoles(ctx, user.ID, app.ID)
	if err != nil {
		return "", "", err
	}
	if len(roles) == 0 {
		return "", "", ErrNoRolesForApp
	}

	person, err := s.personRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return "", "", err
	}
	if app.RequirePerson && (person == nil || !person.IsActive) {
		return "", "", ErrPersonRequired
	}

	permissions, err := s.userRepo.GetUserPermissions(ctx, user.ID, app.ID)
	if err != nil {
		return "", "", err
	}

	accessToken, err := s.jwtService.GenerateToken(user, app.ID, roles, permissions, person)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Armazena no Redis associando a aplicação e o usuário por 7 dias
	payload := fmt.Sprintf(`{"user_id":"%s", "app_id":"%s"}`, user.ID, app.ID)
	if err := s.redisClient.Set(ctx, "refresh_token:"+refreshToken, payload, 7*24*time.Hour).Err(); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RegisterUser(ctx context.Context, email, username, password string) (*domain.User, error) {
	existingUser, err := s.userRepo.FindByEmailOrUsername(ctx, email, username)
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
		Username:     username,
		PasswordHash: &hashStr,
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, appID, identifier, password string) (string, string, error) {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return "", "", err
	}
	if app == nil || !app.IsActive {
		return "", "", errors.New("aplicação fornecida não existe ou está inativa")
	}

	allowsPassword := false
	for _, method := range app.AuthMethods {
		if method == "PASSWORD" {
			allowsPassword = true
			break
		}
	}
	if !allowsPassword {
		return "", "", errors.New("esta aplicação não aceita login com senha")
	}

	user, err := s.userRepo.FindByIdentifier(ctx, identifier)
	if err != nil {
		return "", "", err
	}
	if user == nil || !user.IsActive {
		return "", "", ErrInvalidCredentials
	}
	if user.PasswordHash == nil {
		return "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user, app)
}

// GetProfile retorna o usuário autenticado e a pessoa vinculada (nil se não houver)
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*domain.User, *domain.Person, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New("usuário não encontrado")
	}

	person, err := s.personRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	return user, person, nil
}

// ListUsers lista todos os usuários com os dados básicos da pessoa vinculada
func (s *AuthService) ListUsers(ctx context.Context) ([]*domain.UserWithPerson, error) {
	return s.userRepo.FindAllWithPerson(ctx)
}

// SetUserActive ativa/desativa um usuário (desativado não autentica nem renova sessão)
func (s *AuthService) SetUserActive(ctx context.Context, userID string, isActive bool) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("usuário não encontrado")
	}
	return s.userRepo.SetActive(ctx, userID, isActive)
}

// ListUserRoles lista as roles do usuário em todas as aplicações
func (s *AuthService) ListUserRoles(ctx context.Context, userID string) ([]*domain.Role, error) {
	return s.userRepo.FindRolesByUser(ctx, userID)
}

// ---------------- FLUXO PASSWORDLESS ---------------- //

func (s *AuthService) PasswordlessStart(ctx context.Context, appID, identifier string) error {
	// 1. Validações da Aplicação
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil || !app.IsActive {
		return errors.New("aplicação fornecida não existe ou está inativa")
	}

	allowsPwdless := false
	for _, method := range app.AuthMethods {
		if method == "PASSWORDLESS" {
			allowsPwdless = true
			break
		}
	}
	if !allowsPwdless {
		return errors.New("aplicação não permite login sem senha")
	}

	// 2. Verifica o usuário (Sem revelar detalhes externamente)
	user, err := s.userRepo.FindByIdentifier(ctx, identifier)
	if err != nil {
		return err
	}
	if user == nil || !user.IsActive {
		// Se não existe, retornamos sucesso falso para evitar ataques de Enumeração de E-mail
		return nil
	}

	// 3. Gera código OTP seguro de 6 dígitos
	otp := generateOTP(6)

	// 4. Salva no Redis (Com expiração automática de 5 minutos!)
	key := fmt.Sprintf("otp:%s:%s", appID, identifier)
	err = s.redisClient.Set(ctx, key, otp, 5*time.Minute).Err()
	if err != nil {
		return err
	}

	// 5. Envia o e-mail em Background (Goroutine) para responder rápido ao Frontend
	go func() {
		_ = s.emailService.SendOTP(user.Email, otp)
	}()

	return nil
}

func (s *AuthService) PasswordlessVerify(ctx context.Context, appID, identifier, code string) (string, string, error) {
	key := fmt.Sprintf("otp:%s:%s", appID, identifier)

	// 1. Busca no Redis
	savedCode, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", "", errors.New("código expirado ou não solicitado")
	} else if err != nil {
		return "", "", err
	}

	// 2. Valida o Código
	if savedCode != code {
		return "", "", errors.New("código inválido")
	}

	// 3. Destrói o código do Redis (Uso único garantido!)
	s.redisClient.Del(ctx, key)

	// 4. Fluxo normal de geração de JWT (igual ao login)
	user, err := s.userRepo.FindByIdentifier(ctx, identifier)
	if err != nil {
		return "", "", err
	}
	if user == nil || !user.IsActive {
		return "", "", ErrInvalidCredentials
	}

	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil {
		return "", "", err
	}
	if app == nil || !app.IsActive {
		return "", "", errors.New("aplicação fornecida não existe ou está inativa")
	}

	return s.issueTokens(ctx, user, app)
}

func generateOTP(length int) string {
	const digits = "0123456789"
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	for i := 0; i < length; i++ {
		bytes[i] = digits[bytes[i]%10] // Sorteia entre 0 e 9
	}
	return string(bytes)
}

// Refresh renova silenciosamente a sessão de um usuário utilizando o token de longa duração
func (s *AuthService) Refresh(ctx context.Context, oldRefreshToken string) (string, string, error) {
	key := "refresh_token:" + oldRefreshToken
	payloadStr, err := s.redisClient.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", "", errors.New("refresh token inválido, revogado ou expirado")
	} else if err != nil {
		return "", "", err
	}

	// Invalida o token antigo imediatamente (Prevenção de roubo / Re-uso)
	s.redisClient.Del(ctx, key)

	var payload struct {
		UserID string `json:"user_id"`
		AppID  string `json:"app_id"`
	}
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return "", "", errors.New("falha ao processar a sessão armazenada")
	}

	// Validações de integridade em Tempo Real
	user, err := s.userRepo.FindByID(ctx, payload.UserID)
	if err != nil || user == nil || !user.IsActive {
		return "", "", errors.New("conta inativa ou não encontrada")
	}

	app, err := s.appRepo.FindByID(ctx, payload.AppID)
	if err != nil {
		return "", "", err
	}
	if app == nil || !app.IsActive {
		return "", "", errors.New("aplicação inativa ou não encontrada")
	}

	// Emissão de Novos Tokens (revalida roles e pessoa vinculada em tempo real)
	return s.issueTokens(ctx, user, app)
}
