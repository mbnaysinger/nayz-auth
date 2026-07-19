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
	personHandler *PersonHandler,
	permissionHandler *PermissionHandler,
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
	
	// Rota de Renovação de Sessão (Transparente)
	mux.HandleFunc("POST /api/v1/users/refresh", userHandler.Refresh)

	// Perfil do usuário autenticado (usuário + pessoa vinculada)
	mux.HandleFunc("GET /api/v1/me", authMiddleware.RequireAuth(userHandler.Me))

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

	// Gestão de Usuários
	mux.HandleFunc("GET /api/v1/admin/users", authMiddleware.RequireRole("SUPER_ADMIN", userHandler.List))
	mux.HandleFunc("PATCH /api/v1/admin/users/{id}/active", authMiddleware.RequireRole("SUPER_ADMIN", userHandler.SetActive))
	mux.HandleFunc("GET /api/v1/admin/users/{id}/roles", authMiddleware.RequireRole("SUPER_ADMIN", userHandler.ListRoles))

	// CRUD de Permissões + composição de Roles
	mux.HandleFunc("POST /api/v1/admin/applications/{app_id}/permissions", authMiddleware.RequireRole("SUPER_ADMIN", permissionHandler.Create))
	mux.HandleFunc("GET /api/v1/admin/applications/{app_id}/permissions", authMiddleware.RequireRole("SUPER_ADMIN", permissionHandler.ListByApp))
	mux.HandleFunc("DELETE /api/v1/admin/permissions/{id}", authMiddleware.RequireRole("SUPER_ADMIN", permissionHandler.Delete))
	mux.HandleFunc("GET /api/v1/admin/roles/{role_id}/permissions", authMiddleware.RequireRole("SUPER_ADMIN", permissionHandler.ListByRole))
	mux.HandleFunc("POST /api/v1/admin/roles/{role_id}/permissions/{permission_id}", authMiddleware.RequireRole("SUPER_ADMIN", permissionHandler.Attach))
	mux.HandleFunc("DELETE /api/v1/admin/roles/{role_id}/permissions/{permission_id}", authMiddleware.RequireRole("SUPER_ADMIN", permissionHandler.Detach))

	// Atribuição de Acessos
	mux.HandleFunc("POST /api/v1/admin/users/{user_id}/roles/{role_id}", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.AssignUser))
	mux.HandleFunc("DELETE /api/v1/admin/users/{user_id}/roles/{role_id}", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.RemoveUser))

	// CRUD de Pessoas
	mux.HandleFunc("POST /api/v1/admin/persons", authMiddleware.RequireRole("SUPER_ADMIN", personHandler.Create))
	mux.HandleFunc("GET /api/v1/admin/persons", authMiddleware.RequireRole("SUPER_ADMIN", personHandler.List))
	mux.HandleFunc("GET /api/v1/admin/persons/{id}", authMiddleware.RequireRole("SUPER_ADMIN", personHandler.Get))
	mux.HandleFunc("PUT /api/v1/admin/persons/{id}", authMiddleware.RequireRole("SUPER_ADMIN", personHandler.Update))
	mux.HandleFunc("DELETE /api/v1/admin/persons/{id}", authMiddleware.RequireRole("SUPER_ADMIN", personHandler.Delete))

	return mux
}
