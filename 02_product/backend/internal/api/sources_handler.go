package api

import (
	"net/http"

	"diploma-market-ai/02_product/backend/internal/storage"
)

type SourcesHandler struct {
	repository *storage.SourcesRepository
}

func NewSourcesHandler(repository *storage.SourcesRepository) *SourcesHandler {
	return &SourcesHandler{repository: repository}
}

func (h *SourcesHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.repository.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
	})
}
