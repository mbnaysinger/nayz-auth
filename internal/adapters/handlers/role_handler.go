package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

type RoleHandler struct {
	roleService *services.RoleService
}

func NewRoleHandler(s *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: s}
}

// DTO de Criação
type CreateRoleRequest struct {
	Name string `json:"name"`
}

func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Formato de JSON inválido"}`, http.StatusBadRequest)
		return
	}

	role, err := h.roleService.CreateRole(r.Context(), r.PathValue("app_id"), req.Name)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	slog.Info("Role criada com sucesso", "role", role)
	json.NewEncoder(w).Encode(role)
}

func (h *RoleHandler) ListByApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	appID := r.PathValue("app_id") // Feature do Go 1.22

	roles, err := h.roleService.ListByApp(r.Context(), appID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(roles)
}

func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.roleService.DeleteRole(r.Context(), id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RoleHandler) AssignUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	roleID := r.PathValue("role_id")

	if err := h.roleService.AssignUser(r.Context(), userID, roleID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RoleHandler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	roleID := r.PathValue("role_id")

	if err := h.roleService.RemoveUser(r.Context(), userID, roleID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
