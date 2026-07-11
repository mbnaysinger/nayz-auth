package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

// RegisterUserRequest é o nosso DTO de entrada.
// As tags `json:"email"` ensinam o Go como mapear o JSON recebido para essa Struct.
type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserHandler atua como o Controller HTTP
type UserHandler struct {
	authService *services.AuthService
}

// NewUserHandler é o "construtor" que recebe a injeção do serviço
func NewUserHandler(authService *services.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

// Register gerencia a rota POST para cadastrar usuários
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Definimos que a resposta será sempre em JSON
	w.Header().Set("Content-Type", "application/json")

	// 1. Receber e decodificar o JSON (Deserialization)
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Formato de JSON inválido"})
		return
	}

	// Validação super básica (poderíamos usar pacotes como o validator/v10 depois)
	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "E-mail e senha são obrigatórios"})
		return
	}

	// 2. Chamar a regra de negócios passando o Contexto da requisição original
	user, err := h.authService.RegisterUser(r.Context(), req.Email, req.Password)
	if err != nil {
		// Retornamos 409 Conflict se for duplicação de e-mail
		w.WriteHeader(http.StatusConflict) 
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 3. Sucesso! Retornar HTTP 201 e o objeto do usuário cadastrado (com ID e Datas)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
