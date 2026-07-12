package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

type ApplicationHandler struct {
	appService *services.ApplicationService
}

func NewApplicationHandler(appService *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: appService}
}

// DTOs
type CreateAppRequest struct {
	Name        string   `json:"name"`
	AuthMethods []string `json:"auth_methods"`
}

type UpdateAppRequest struct {
	Name        string   `json:"name"`
	AuthMethods []string `json:"auth_methods"`
	IsActive    bool     `json:"is_active"`
}

func (h *ApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Formato de JSON inválido"}`, http.StatusBadRequest)
		return
	}

	app, err := h.appService.CreateApplication(r.Context(), req.Name, req.AuthMethods)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

func (h *ApplicationHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	apps, err := h.appService.ListApplications(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Erro interno ao buscar aplicações"}`, http.StatusInternalServerError)
		return
	}
	// Mesmo vazio, retorna [] ao invés de null para facilitar o FrontEnd
	if apps == nil {
		apps = make([]*domain.Application, 0) // Omitido temporariamente, o json trata slices vazios.
	}

	json.NewEncoder(w).Encode(apps)
}

func (h *ApplicationHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// A maravilha do Go 1.22: extrair variáveis do Path nativamente!
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"ID da aplicação é obrigatório"}`, http.StatusBadRequest)
		return
	}

	var req UpdateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Formato de JSON inválido"}`, http.StatusBadRequest)
		return
	}

	app, err := h.appService.UpdateApplication(r.Context(), id, req.Name, req.AuthMethods, req.IsActive)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(app)
}

func (h *ApplicationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"ID da aplicação é obrigatório"}`, http.StatusBadRequest)
		return
	}

	err := h.appService.DeleteApplication(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Retorna HTTP 204 (No Content) indicando sucesso sem corpo na resposta
	w.WriteHeader(http.StatusNoContent)
}
