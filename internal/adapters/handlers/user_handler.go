package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

type UserHandler struct {
	authService *services.AuthService
}

func NewUserHandler(authService *services.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

// DTOs
type RegisterUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	AppID      string `json:"app_id"`
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type PwdlessStartRequest struct {
	AppID      string `json:"app_id"`
	Identifier string `json:"identifier"`
}

type PwdlessVerifyRequest struct {
	AppID      string `json:"app_id"`
	Identifier string `json:"identifier"`
	Code       string `json:"code"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Rotas Clássicas (Registro e Login com Senha)
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Formato de JSON inválido"}`, http.StatusBadRequest)
		return
	}
	user, err := h.authService.RegisterUser(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Formato de JSON inválido"}`, http.StatusBadRequest)
		return
	}
	accessToken, refreshToken, err := h.authService.Login(r.Context(), req.AppID, req.Identifier, req.Password)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"type":          "Bearer",
	})
}

// Rotas do Passwordless (Redis + E-mail)
func (h *UserHandler) PasswordlessStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req PwdlessStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"json invalido"}`, http.StatusBadRequest)
		return
	}

	err := h.authService.PasswordlessStart(r.Context(), req.AppID, req.Identifier)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Retornamos sucesso indepentente do e-mail existir no banco (Segurança)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Se o e-mail estiver cadastrado, um código foi enviado para sua caixa de entrada."}`))
}

func (h *UserHandler) PasswordlessVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req PwdlessVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"json invalido"}`, http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authService.PasswordlessVerify(r.Context(), req.AppID, req.Identifier, req.Code)
	if err != nil {
		// Se o código for inválido, retornamos 401
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"type":          "Bearer",
	})
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Formato JSON inválido"}`, http.StatusBadRequest)
		return
	}

	accessToken, newRefreshToken, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token":         accessToken,
		"refresh_token": newRefreshToken,
		"type":          "Bearer",
	})
}
