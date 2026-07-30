package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zippyra/backend/shared/errors"
)

func (h *ExitHandler) HandleGetRecentExitAttempts(w http.ResponseWriter, r *http.Request) {
	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		errors.WriteError(w, http.StatusBadRequest, errors.CodeInvalidRequest, "store_id query parameter is required", nil)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 100 {
		limit = 100
	}

	attempts, err := h.repo.GetRecentExitAttempts(r.Context(), storeID, limit)
	if err != nil {
		attempts = []*ExitAttempt{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"attempts": attempts,
		"store_id": storeID,
	})
}
