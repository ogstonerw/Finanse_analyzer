package forecasts

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var request GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(request.Ticker) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ticker is required"})
		return
	}

	item, err := h.service.Generate(r.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, ErrAssetNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
		case errors.Is(err, ErrEventNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "event not found"})
		case errors.Is(err, ErrEventAssetMismatch):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "event does not match the requested asset"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) Latest(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Latest(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, ErrForecastNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "forecast not found"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
