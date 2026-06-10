package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vladkonst/mnemonics/internal/delivery/http/middleware"
	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	subscriptionUC "github.com/vladkonst/mnemonics/internal/usecase/subscription"
)

// SubscriptionHandler handles subscription and promo code endpoints.
type SubscriptionHandler struct {
	uc *subscriptionUC.UseCase
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(uc *subscriptionUC.UseCase) *SubscriptionHandler {
	return &SubscriptionHandler{uc: uc}
}

// createSubscriptionRequest is the JSON body for POST /api/v1/users/{user_id}/subscriptions.
type createSubscriptionRequest struct {
	Type         string `json:"type"` // "invite" or "payment"
	InviteLinkID string `json:"invite_link_id"`
	PaymentID    string `json:"payment_id"`
	Plan         string `json:"plan"`
}

// CreateSubscription handles POST /api/v1/users/{user_id}/subscriptions.
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if !middleware.RequireOwner(w, r, userID) {
		return
	}

	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	switch req.Type {
	case "payment":
		if req.PaymentID == "" {
			respond.Error(w, http.StatusBadRequest, "bad_request", "payment_id is required for type=payment")
			return
		}
		plan := req.Plan
		if plan == "" {
			plan = "monthly"
		}
		sub, err := h.uc.CreatePaymentSubscription(r.Context(), userID, req.PaymentID, plan)
		if err != nil {
			respond.ErrorFrom(w, err)
			return
		}
		w.Header().Set("Location", fmt.Sprintf("/api/v1/users/%d/subscriptions/%s", userID, sub.PaymentID))
		respond.JSON(w, http.StatusCreated, sub)

	case "invite":
		if req.InviteLinkID == "" {
			respond.Error(w, http.StatusBadRequest, "bad_request", "invite_link_id is required for type=invite")
			return
		}
		sub, err := h.uc.CreateInviteSubscription(r.Context(), userID, req.InviteLinkID)
		if err != nil {
			respond.ErrorFrom(w, err)
			return
		}
		w.Header().Set("Location", fmt.Sprintf("/api/v1/users/%d/subscriptions/%s", userID, sub.PaymentID))
		respond.JSON(w, http.StatusCreated, sub)

	default:
		respond.Error(w, http.StatusBadRequest, "bad_request", "type must be 'invite' or 'payment'")
	}
}

// GetTeacherInviteLinks handles GET /api/v1/teachers/{teacher_id}/invite-links.
func (h *SubscriptionHandler) GetTeacherInviteLinks(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if !middleware.RequireOwner(w, r, teacherID) {
		return
	}

	links, err := h.uc.GetTeacherInviteLinks(r.Context(), teacherID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{"invite_links": links})
}

// parseTeacherID extracts and validates the teacher_id path parameter.
func parseTeacherID(r *http.Request) (int64, error) {
	raw := r.PathValue("teacher_id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("teacher_id must be a valid positive integer")
	}
	return id, nil
}
