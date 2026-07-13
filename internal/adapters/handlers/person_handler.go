package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mbnaysinger/nayz-auth/internal/core/domain"
	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

type PersonHandler struct {
	personService *services.PersonService
}

func NewPersonHandler(personService *services.PersonService) *PersonHandler {
	return &PersonHandler{personService: personService}
}

// DTOs
type PersonRequest struct {
	UserID     *string    `json:"user_id"`
	Identifier string     `json:"identifier"`
	Name       string     `json:"name"`
	Phone      *string    `json:"phone"`
	IsActive   bool       `json:"is_active"`
	BirthDate  *time.Time `json:"birth_date"`
}

func (h *PersonHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req PersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Formato de JSON inválido"}`, http.StatusBadRequest)
		return
	}

	person := &domain.Person{
		UserID:     req.UserID,
		Identifier: req.Identifier,
		Name:       req.Name,
		Phone:      req.Phone,
		IsActive:   req.IsActive,
		BirthDate:  req.BirthDate,
	}

	if err := h.personService.CreatePerson(r.Context(), person); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(person)
}

func (h *PersonHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	persons, err := h.personService.ListPersons(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Erro interno ao buscar pessoas"}`, http.StatusInternalServerError)
		return
	}
	
	if persons == nil {
		persons = make([]*domain.Person, 0)
	}

	json.NewEncoder(w).Encode(persons)
}

func (h *PersonHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"ID da pessoa é obrigatório"}`, http.StatusBadRequest)
		return
	}

	person, err := h.personService.GetPerson(r.Context(), id)
	if err != nil {
		if err == services.ErrPersonNotFound {
			http.Error(w, `{"error":"Pessoa não encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(person)
}

func (h *PersonHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"ID da pessoa é obrigatório"}`, http.StatusBadRequest)
		return
	}

	var req PersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Formato de JSON inválido"}`, http.StatusBadRequest)
		return
	}

	personData := &domain.Person{
		UserID:     req.UserID,
		Identifier: req.Identifier,
		Name:       req.Name,
		Phone:      req.Phone,
		IsActive:   req.IsActive,
		BirthDate:  req.BirthDate,
	}

	err := h.personService.UpdatePerson(r.Context(), id, personData)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Fetch updated person to return it
	updatedPerson, _ := h.personService.GetPerson(r.Context(), id)
	json.NewEncoder(w).Encode(updatedPerson)
}
