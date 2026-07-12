package services

import (
	"context"
	"crypto/rand"
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
)

type AuthService struct {
	userRepo     domain.UserRepository
	appRepo      domain.ApplicationRepository
	jwtService   *JWTService
	redisClient  *redis.Client
	emailService *EmailService
}

func NewAuthService(
	userRepo domain.UserRepository,
	appRepo domain.ApplicationRepository,
	jwtService *JWTService,
	redisClient *redis.Client,
	emailService *EmailService,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		appRepo:      appRepo,
		jwtService:   jwtService,
		redisClient:  redisClient,
		emailService: emailService,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, email, password string) (*domain.User, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil { return nil, err }
	if existingUser != nil { return nil, ErrUserAlreadyExists }

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil { return nil, err }
	hashStr := string(hashedBytes)

	user := &domain.User{
		Email:        email,
		PasswordHash: &hashStr,
		Status:       "ACTIVE",
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil { return nil, err }
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, appID, email, password string) (string, error) {
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil { return "", err }
	if app == nil || !app.IsActive { return "", errors.New("aplicação fornecida não existe ou está inativa") }

	allowsPassword := false
	for _, method := range app.AuthMethods {
		if method == "PASSWORD" {
			allowsPassword = true
			break
		}
	}
	if !allowsPassword { return "", errors.New("esta aplicação não aceita login com senha") }

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil { return "", err }
	if user == nil || user.Status != "ACTIVE" { return "", ErrInvalidCredentials }
	if user.PasswordHash == nil { return "", ErrInvalidCredentials }

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	roles, err := s.userRepo.GetUserRoles(ctx, user.ID, appID)
	if err != nil { return "", err }
	if len(roles) == 0 { return "", errors.New("acesso negado: você não tem permissões para esta aplicação") }

	return s.jwtService.GenerateToken(user, appID, roles)
}

// ---------------- FLUXO PASSWORDLESS ---------------- //

func (s *AuthService) PasswordlessStart(ctx context.Context, appID, email string) error {
	// 1. Validações da Aplicação
	app, err := s.appRepo.FindByID(ctx, appID)
	if err != nil { return err }
	if app == nil || !app.IsActive { return errors.New("aplicação fornecida não existe ou está inativa") }

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
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil { return err }
	if user == nil || user.Status != "ACTIVE" {
		// Se não existe, retornamos sucesso falso para evitar ataques de Enumeração de E-mail
		return nil
	}

	// 3. Gera código OTP seguro de 6 dígitos
	otp := generateOTP(6)
	
	// 4. Salva no Redis (Com expiração automática de 5 minutos!)
	key := fmt.Sprintf("otp:%s:%s", appID, email)
	err = s.redisClient.Set(ctx, key, otp, 5*time.Minute).Err()
	if err != nil { return err }

	// 5. Envia o e-mail em Background (Goroutine) para responder rápido ao Frontend
	go func() {
		_ = s.emailService.SendOTP(email, otp)
	}()

	return nil
}

func (s *AuthService) PasswordlessVerify(ctx context.Context, appID, email, code string) (string, error) {
	key := fmt.Sprintf("otp:%s:%s", appID, email)
	
	// 1. Busca no Redis
	savedCode, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("código expirado ou não solicitado")
	} else if err != nil {
		return "", err
	}

	// 2. Valida o Código
	if savedCode != code {
		return "", errors.New("código inválido")
	}

	// 3. Destrói o código do Redis (Uso único garantido!)
	s.redisClient.Del(ctx, key)

	// 4. Fluxo normal de geração de JWT (igual ao login)
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil { return "", err }
	if user == nil { return "", ErrInvalidCredentials }

	roles, err := s.userRepo.GetUserRoles(ctx, user.ID, appID)
	if err != nil { return "", err }
	if len(roles) == 0 {
		return "", errors.New("acesso negado: você não tem permissões para esta aplicação")
	}

	return s.jwtService.GenerateToken(user, appID, roles)
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
