package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

// JWTService encapsula toda a lógica de criação e assinatura dos tokens.
type JWTService struct {
	secretKey []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secretKey: []byte(secret)}
}

// CustomClaims extende os Claims padrão do JWT (como subject e expiração)
// injetando os nossos dados customizados: a Aplicação logada e as Roles do usuário
type CustomClaims struct {
	AppID string   `json:"app_id"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateToken assina o token injetando as permissões mapeadas
func (s *JWTService) GenerateToken(user *domain.User, appID string, roles []string) (string, error) {
	claims := CustomClaims{
		AppID: appID,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID, // O campo 'sub' (subject) é o padrão do mercado para o ID do usuário
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Gera o token com o algoritmo simétrico HS256 (padrão mais usado)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Assina usando a string ultra secreta do nosso ambiente
	return token.SignedString(s.secretKey)
}
