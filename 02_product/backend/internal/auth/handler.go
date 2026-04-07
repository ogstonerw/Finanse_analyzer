package auth

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.Register(r.Context(), input)
	writeJSON(w, http.StatusAccepted, map[string]any{
		"message": "register endpoint placeholder",
		"status":  statusFromErr(err),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.service.Login(r.Context(), input)
	writeJSON(w, http.StatusAccepted, map[string]any{
		"message": "login endpoint placeholder",
		"status":  statusFromErr(err),
	})
}

func statusFromErr(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
