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

// activatePromoCodeRequest is the JSON body for POST /api/v1/teachers/{teacher_id}/promo-codes.
type activatePromoCodeRequest struct {
	Code string `json:"code"`
}

// ActivatePromoCode handles POST /api/v1/teachers/{teacher_id}/promo-codes.
func (h *SubscriptionHandler) ActivatePromoCode(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if !middleware.RequireOwner(w, r, teacherID) {
		return
	}

	var req activatePromoCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Code == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "code is required")
		return
	}

	promo, err := h.uc.ActivatePromoCode(r.Context(), teacherID, req.Code)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, promo)
}

// GetTeacherPromoCodes handles GET /api/v1/teachers/{teacher_id}/promo-codes.
func (h *SubscriptionHandler) GetTeacherPromoCodes(w http.ResponseWriter, r *http.Request) {
	teacherID, err := parseTeacherID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	promoCodes, err := h.uc.GetTeacherPromoCodes(r.Context(), teacherID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"promo_codes": promoCodes,
	})
}

// createSubscriptionRequest is the JSON body for POST /api/v1/users/{user_id}/subscriptions.
type createSubscriptionRequest struct {
	Type      string `json:"type"` // "promo" or "payment"
	PromoCode string `json:"promo_code"`
	PaymentID string `json:"payment_id"`
	Plan      string `json:"plan"`
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
	case "promo":
		if req.PromoCode == "" {
			respond.Error(w, http.StatusBadRequest, "bad_request", "promo_code is required for type=promo")
			return
		}
		sub, err := h.uc.CreatePromoSubscription(r.Context(), userID, req.PromoCode)
		if err != nil {
			respond.ErrorFrom(w, err)
			return
		}
		w.Header().Set("Location", fmt.Sprintf("/api/v1/users/%d/subscriptions/%s", userID, sub.PaymentID))
		respond.JSON(w, http.StatusCreated, sub)

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

	default:
		respond.Error(w, http.StatusBadRequest, "bad_request", "type must be 'promo' or 'payment'")
	}
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
