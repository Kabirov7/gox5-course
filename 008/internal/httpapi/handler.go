package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"lesson8/httpserver/internal/domain"
	"lesson8/httpserver/internal/uuidv7"
)

type UserService interface {
	Get(context.Context) ([]domain.User, error)
	GetById(context.Context, string) (domain.User, error)
	Create(context.Context, string, domain.CreateUser) (domain.User, error)
	Update(context.Context, string, domain.User) error
	Delete(context.Context, string) error
}

type handler struct {
	users UserService
}

func NewHandler(users UserService, logger *slog.Logger) http.Handler {
	h := &handler{users: users}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /users", h.create)
	mux.HandleFunc("GET /users", h.list)
	mux.HandleFunc("GET /users/{user_id}", h.get)
	mux.HandleFunc("PUT /users/{user_id}", h.update)
	mux.HandleFunc("DELETE /users/{user_id}", h.delete)
	return requestID(logger, logRequest(recoverPanic(mux)))
}

func (h *handler) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	input, err := decodeUser(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := uuidv7.New()
	if err != nil {
		serviceError(w, r, err)
		return
	}
	user, err := h.users.Create(r.Context(), id, input)
	if err != nil {
		serviceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		ID string `json:"id"`
	}{ID: user.ID})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.Get(r.Context())
	if err != nil {
		serviceError(w, r, err)
		return
	}
	if users == nil {
		users = []domain.User{}
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	user, err := h.users.GetById(r.Context(), r.PathValue("user_id"))
	if err != nil {
		serviceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	input, err := decodeUser(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := r.PathValue("user_id")
	user := domain.User{ID: id, Name: input.Name, Age: input.Age}
	if err := h.users.Update(r.Context(), id, user); err != nil {
		serviceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.users.Delete(r.Context(), r.PathValue("user_id")); err != nil {
		serviceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeUser(w http.ResponseWriter, r *http.Request) (domain.CreateUser, error) {
	var input struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return domain.CreateUser{}, errors.New("body must be a JSON object containing name (string) and age (integer): " + err.Error())
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return domain.CreateUser{}, errors.New("body must contain exactly one JSON object")
	}
	if input.Name == nil || strings.TrimSpace(*input.Name) == "" {
		return domain.CreateUser{}, errors.New("name is required and must not be blank")
	}
	if input.Age == nil {
		return domain.CreateUser{}, errors.New("age is required")
	}
	if *input.Age < 0 {
		return domain.CreateUser{}, errors.New("age must be non-negative")
	}
	return domain.CreateUser{Name: strings.TrimSpace(*input.Name), Age: *input.Age}, nil
}

func serviceError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	Logger(r.Context()).ErrorContext(r.Context(), "request failed", "error", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, struct {
		Error string `json:"error"`
	}{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
