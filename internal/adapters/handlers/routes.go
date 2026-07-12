package handlers

import (
	"fmt"
	"net/http"

	"github.com/mbnaysinger/nayz-auth/internal/adapters/middlewares"
	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

// SetupRoutes centraliza todo o mapeamento de rotas da aplicação
func SetupRoutes(
	userHandler *UserHandler,
	appHandler *ApplicationHandler,
	roleHandler *RoleHandler,
	authMiddleware *middlewares.AuthMiddleware,
) *http.ServeMux {
	mux := http.NewServeMux()

	// ---- Rotas Públicas ----
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP", "message": "Serviço nayz-auth operando normalmente"}`))
	})
	
	// Autenticação Clássica
	mux.HandleFunc("POST /api/v1/users/register", userHandler.Register)
	mux.HandleFunc("POST /api/v1/users/login", userHandler.Login)
	
	// Autenticação Passwordless
	mux.HandleFunc("POST /api/v1/users/passwordless/start", userHandler.PasswordlessStart)
	mux.HandleFunc("POST /api/v1/users/passwordless/verify", userHandler.PasswordlessVerify)

	// ---- Rotas Privadas (Admin Console) ----
	
	// Dashboard genérico
	mux.HandleFunc("GET /api/v1/admin/dashboard", authMiddleware.RequireRole("SUPER_ADMIN", func(w http.ResponseWriter, r *http.Request) {
		claims, _ := r.Context().Value(middlewares.ClaimsContextKey).(*services.CustomClaims)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonStr := fmt.Sprintf(`{"message": "Bem vindo ao painel de controle!", "user_id": "%s", "app_id": "%s"}`, claims.Subject, claims.AppID)
		w.Write([]byte(jsonStr))
	}))

	// CRUD de Aplicações
	mux.HandleFunc("POST /api/v1/admin/applications", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Create))
	mux.HandleFunc("GET /api/v1/admin/applications", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.List))
	mux.HandleFunc("PUT /api/v1/admin/applications/{id}", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Update))
	mux.HandleFunc("DELETE /api/v1/admin/applications/{id}", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Delete))

	// CRUD de Roles
	mux.HandleFunc("POST /api/v1/admin/applications/{app_id}/roles", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.Create))
	mux.HandleFunc("GET /api/v1/admin/applications/{app_id}/roles", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.ListByApp))
	mux.HandleFunc("DELETE /api/v1/admin/roles/{id}", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.Delete))
	
	// Atribuição de Acessos
	mux.HandleFunc("POST /api/v1/admin/users/{user_id}/roles/{role_id}", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.AssignUser))
	mux.HandleFunc("DELETE /api/v1/admin/users/{user_id}/roles/{role_id}", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.RemoveUser))

	return mux
}
