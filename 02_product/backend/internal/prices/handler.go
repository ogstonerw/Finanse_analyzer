package prices

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

func (h *Handler) ListByTicker(w http.ResponseWriter, r *http.Request) {
	ticker := r.PathValue("ticker")

	items, err := h.service.ListByTicker(r.Context(), ticker)
	if err != nil {
		switch {
		case errors.Is(err, ErrAssetNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ticker":    strings.ToUpper(strings.TrimSpace(ticker)),
		"timeframe": dailyTimeframe,
		"items":     items,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
