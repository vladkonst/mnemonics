package handlers

import (
	"net/http"
	"strconv"
	"fmt"

	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	progressUC "github.com/vladkonst/mnemonics/internal/usecase/progress"
)

// ProgressHandler handles progress-related HTTP endpoints.
type ProgressHandler struct {
	uc *progressUC.UseCase
}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler(uc *progressUC.UseCase) *ProgressHandler {
	return &ProgressHandler{uc: uc}
}

// GetUserProgress handles GET /api/v1/users/{user_id}/progress.
func (h *ProgressHandler) GetUserProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.uc.GetUserProgress(r.Context(), userID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// GetModuleProgress handles GET /api/v1/users/{user_id}/progress/modules/{module_id}.
func (h *ProgressHandler) GetModuleProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	moduleID, err := parseProgressModuleID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.uc.GetModuleProgress(r.Context(), userID, moduleID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, result)
}

// parseProgressModuleID extracts and validates module_id from the path.
func parseProgressModuleID(r *http.Request) (int, error) {
	raw := r.PathValue("module_id")
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("module_id must be a valid positive integer")
	}
	return id, nil
}
