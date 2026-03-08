package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	paymentUC "github.com/vladkonst/mnemonics/internal/usecase/payment"
)

// PaymentHandler handles payment invoice and webhook endpoints.
type PaymentHandler struct {
	uc *paymentUC.UseCase
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(uc *paymentUC.UseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

// createInvoiceRequest is the JSON body for POST /api/v1/users/{user_id}/payment-invoices.
type createInvoiceRequest struct {
	Plan string `json:"plan"`
}

// CreateInvoice handles POST /api/v1/users/{user_id}/payment-invoices.
func (h *PaymentHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	var req createInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}
	if req.Plan == "" {
		req.Plan = "monthly"
	}

	result, err := h.uc.CreateInvoice(r.Context(), userID, req.Plan)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/users/%d/payment-invoices/%s", userID, result.InvoiceID))
	respond.JSON(w, http.StatusCreated, result)
}

// GetPendingInvoice handles GET /api/v1/users/{user_id}/payment-invoices/pending.
// Currently returns a stub — the real implementation would query the payment gateway.
func (h *PaymentHandler) GetPendingInvoice(w http.ResponseWriter, r *http.Request) {
	_, err := parseUserID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	// Stub: return empty pending invoice status.
	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"pending": false,
	})
}

// webhookRequest is the payload sent by the payment gateway.
type webhookRequest struct {
	PaymentID string `json:"payment_id"`
	UserID    int64  `json:"user_id"`
	Plan      string `json:"plan"`
	Status    string `json:"status"`
}

// HandleWebhook handles POST /api/v1/webhooks/payment-gateway.
// Always returns 200 to acknowledge receipt.
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respond.JSON(w, http.StatusOK, map[string]string{"received": "true"})
		return
	}
	defer r.Body.Close()

	var req webhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		// Still return 200 to avoid retries for malformed payloads.
		respond.JSON(w, http.StatusOK, map[string]string{"received": "true"})
		return
	}

	signature := r.Header.Get("X-Payment-Signature")

	event := paymentUC.WebhookEvent{
		PaymentID: req.PaymentID,
		UserID:    req.UserID,
		Plan:      req.Plan,
		Status:    req.Status,
	}

	// Process the webhook; errors are logged but we always return 200.
	_ = h.uc.HandleWebhook(r.Context(), body, signature, event)

	respond.JSON(w, http.StatusOK, map[string]string{"received": "true"})
}
