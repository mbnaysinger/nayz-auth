package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	AppID    string `json:"app_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserHandler struct {
	authService *services.AuthService
}

func NewUserHandler(authService *services.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Formato de JSON inválido"})
		return
	}
	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "E-mail e senha são obrigatórios"})
		return
	}
	user, err := h.authService.RegisterUser(r.Context(), req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login gerencia a requisição de autenticação retornando um JWT válido
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Formato de JSON inválido"})
		return
	}

	if req.AppID == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "app_id, email e password são obrigatórios"})
		return
	}

	// Devolve a chave mágica (JWT) ou o erro formatado da Regra de Negócios!
	token, err := h.authService.Login(r.Context(), req.AppID, req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized) // 401 indica falha de credencial/autorização
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
		"type":  "Bearer",
	})
}
