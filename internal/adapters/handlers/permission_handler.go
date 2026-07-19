package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

type PermissionHandler struct {
	permissionService *services.PermissionService
}

func NewPermissionHandler(permissionService *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

type CreatePermissionRequest struct {
	Name string `json:"name"`
}

// Create registra uma permissão na aplicação (formato recurso:acao)
func (h *PermissionHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	appID := r.PathValue("app_id")

	var req CreatePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Formato de JSON inválido"}`, http.StatusBadRequest)
		return
	}

	permission, err := h.permissionService.CreatePermission(r.Context(), appID, req.Name)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(permission)
}

func (h *PermissionHandler) ListByApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	permissions, err := h.permissionService.ListByApp(r.Context(), r.PathValue("app_id"))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(permissions)
}

func (h *PermissionHandler) ListByRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	permissions, err := h.permissionService.ListByRole(r.Context(), r.PathValue("role_id"))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(permissions)
}

func (h *PermissionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := h.permissionService.DeletePermission(r.Context(), r.PathValue("id")); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Attach vincula uma permissão a uma role (composição da role)
func (h *PermissionHandler) Attach(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := h.permissionService.AttachToRole(r.Context(), r.PathValue("role_id"), r.PathValue("permission_id"))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Permissão vinculada à role"}`))
}

func (h *PermissionHandler) Detach(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := h.permissionService.DetachFromRole(r.Context(), r.PathValue("role_id"), r.PathValue("permission_id"))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
