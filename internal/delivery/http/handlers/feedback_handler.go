package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/vladkonst/mnemonics/internal/delivery/http/middleware"
	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	"github.com/vladkonst/mnemonics/internal/domain/feedback"
	"github.com/vladkonst/mnemonics/internal/domain/interfaces"
)

// FeedbackHandler handles feedback submission from bot users.
type FeedbackHandler struct {
	repo interfaces.FeedbackRepository
}

// NewFeedbackHandler creates a new FeedbackHandler.
func NewFeedbackHandler(repo interfaces.FeedbackRepository) *FeedbackHandler {
	return &FeedbackHandler{repo: repo}
}

type submitFeedbackRequest struct {
	Text string `json:"text"`
}

// SubmitFeedback handles POST /api/v1/users/{user_id}/feedback.
func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if !middleware.RequireOwner(w, r, userID) {
		return
	}

	var req submitFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Text == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "text is required")
		return
	}

	f := &feedback.Feedback{
		UserID:    userID,
		Text:      req.Text,
		CreatedAt: time.Now().UTC(),
	}
	if err := h.repo.Create(r.Context(), f); err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, f)
}
