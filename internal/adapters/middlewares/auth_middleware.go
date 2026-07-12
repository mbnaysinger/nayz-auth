package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

// Definimos um tipo customizado para a chave do contexto para evitar colisões
type contextKey string

const ClaimsContextKey contextKey = "user_claims"

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: []byte(secret)}
}

// RequireAuth é o middleware que intercepta a requisição, valida o JWT e injeta os dados no contexto
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Pega o cabeçalho "Authorization"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "Token ausente ou em formato inválido"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Transforma a string de volta num objeto JWT usando nosso "CustomClaims"
		claims := &services.CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error": "Token inválido, expirado ou assinatura corrompida"}`, http.StatusUnauthorized)
			return
		}

		// 3. Injeta a sacola de permissões (claims) no "Contexto" da requisição.
		// O Contexto (ctx) trafega por todo o ecossistema Go até chegar no banco de dados!
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		
		// 4. Passa a requisição para a frente (para o próximo Controller) com o novo contexto enriquecido
		next(w, r.WithContext(ctx))
	}
}

// RequireRole é um filtro mais avançado. Além de exigir que esteja logado, exige uma permissão específica.
func (m *AuthMiddleware) RequireRole(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	// Reutilizamos o RequireAuth (composição funcional)
	return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		
		// Recupera os dados que o RequireAuth injetou
		claims, ok := r.Context().Value(ClaimsContextKey).(*services.CustomClaims)
		if !ok {
			http.Error(w, `{"error": "Falha crítica ao ler o contexto da requisição"}`, http.StatusInternalServerError)
			return
		}

		// Verifica se a array de Roles tem a Role exigida
		hasRole := false
		for _, role := range claims.Roles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			http.Error(w, `{"error": "Acesso Negado: Você não possui a permissão necessária"}`, http.StatusForbidden)
			return
		}

		// Tudo certo, libera a passagem!
		next(w, r)
	})
}
