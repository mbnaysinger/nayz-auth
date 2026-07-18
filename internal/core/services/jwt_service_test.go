package services

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
)

const testSecret = "segredo-de-teste"

func parseClaims(t *testing.T, tokenStr string) *CustomClaims {
	t.Helper()
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("token inválido: %v", err)
	}
	return claims
}

func TestGenerateTokenComPessoaVinculada(t *testing.T) {
	svc := NewJWTService(testSecret)
	user := &domain.User{ID: "user-123", Email: "a@b.com", Username: "alexandre"}
	person := &domain.Person{ID: "person-456", Name: "Alexandre Riff", IsActive: true}

	tokenStr, err := svc.GenerateToken(user, "app-tallo", []string{"ADMIN"}, person)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	claims := parseClaims(t, tokenStr)
	if claims.Subject != "user-123" {
		t.Errorf("sub esperado user-123, veio %s", claims.Subject)
	}
	if claims.AppID != "app-tallo" {
		t.Errorf("app_id esperado app-tallo, veio %s", claims.AppID)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "ADMIN" {
		t.Errorf("roles esperadas [ADMIN], vieram %v", claims.Roles)
	}
	if claims.PersonID != "person-456" {
		t.Errorf("person_id esperado person-456, veio %s", claims.PersonID)
	}
	if claims.Name != "Alexandre Riff" {
		t.Errorf("name esperado 'Alexandre Riff', veio %s", claims.Name)
	}
}

func TestGenerateTokenSemPessoaOmiteClaims(t *testing.T) {
	svc := NewJWTService(testSecret)
	user := &domain.User{ID: "user-123"}

	tokenStr, err := svc.GenerateToken(user, "app-x", []string{"USER"}, nil)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	claims := parseClaims(t, tokenStr)
	if claims.PersonID != "" || claims.Name != "" {
		t.Errorf("claims de pessoa deveriam estar vazios, vieram person_id=%q name=%q", claims.PersonID, claims.Name)
	}
}
