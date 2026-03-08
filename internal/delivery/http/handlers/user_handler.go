// Package handlers contains HTTP handler types for the delivery layer.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	userUC "github.com/vladkonst/mnemonics/internal/usecase/user"
	"github.com/vladkonst/mnemonics/internal/domain/user"
)

// UserHandler handles user-related HTTP endpoints.
type UserHandler struct {
	uc *userUC.UseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(uc *userUC.UseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

// registerUserRequest is the JSON body for POST /api/v1/users.
type registerUserRequest struct {
	TelegramID int64  `json:"telegram_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Username   string `json:"username"`
}

// RegisterUser handles POST /api/v1/users.
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.TelegramID <= 0 {
		respond.Error(w, http.StatusBadRequest, "bad_request", "telegram_id is required")
		return
	}
	if req.FirstName == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "first_name is required")
		return
	}

	u, err := h.uc.Register(r.Context(), req.TelegramID, req.FirstName, req.LastName, req.Username)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/users/%d", u.TelegramID))
	respond.JSON(w, http.StatusCreated, u)
}

// updateUserRequest is the JSON body for PATCH /api/v1/users/{user_id}.
type updateUserRequest struct {
	Role                 *string `json:"role"`
	Language             *string `json:"language"`
	NotificationsEnabled *bool   `json:"notifications_enabled"`
}

// UpdateUser handles PATCH /api/v1/users/{user_id}.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	var u interface{}

	if req.Role != nil {
		u, err = h.uc.UpdateRole(r.Context(), userID, user.Role(*req.Role))
		if err != nil {
			respond.ErrorFrom(w, err)
			return
		}
	}

	if req.Language != nil || req.NotificationsEnabled != nil {
		u, err = h.uc.UpdateSettings(r.Context(), userID, req.Language, req.NotificationsEnabled)
		if err != nil {
			respond.ErrorFrom(w, err)
			return
		}
	}

	if u == nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "no updatable fields provided")
		return
	}

	respond.JSON(w, http.StatusOK, u)
}

// GetSubscription handles GET /api/v1/users/{user_id}/subscription.
func (h *UserHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	sub, err := h.uc.GetSubscription(r.Context(), userID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, sub)
}

// parseUserID extracts and validates the user_id path parameter.
func parseUserID(r *http.Request) (int64, error) {
	raw := r.PathValue("user_id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("user_id must be a valid positive integer")
	}
	return id, nil
}
