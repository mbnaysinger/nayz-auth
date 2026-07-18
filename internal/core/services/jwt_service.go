package services

import (
	"crypto/rand"
	"encoding/hex"
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
	AppID    string   `json:"app_id"`
	Roles    []string `json:"roles"`
	PersonID string   `json:"person_id,omitempty"`
	Name     string   `json:"name,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken assina o token injetando as permissões mapeadas.
// person é opcional (nil quando o usuário não tem pessoa vinculada).
func (s *JWTService) GenerateToken(user *domain.User, appID string, roles []string, person *domain.Person) (string, error) {
	claims := CustomClaims{
		AppID: appID,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID, // O campo 'sub' (subject) é o padrão do mercado para o ID do usuário
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	if person != nil {
		claims.PersonID = person.ID
		claims.Name = person.Name
	}

	// Gera o token com o algoritmo simétrico HS256 (padrão mais usado)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Assina usando a string ultra secreta do nosso ambiente
	return token.SignedString(s.secretKey)
}

// GenerateRefreshToken cria uma string aleatória forte e segura para servir de Refresh Token opaco
func (s *JWTService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
