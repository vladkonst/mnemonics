package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/vladkonst/mnemonics/internal/delivery/http/respond"
	managerUC "github.com/vladkonst/mnemonics/internal/usecase/manager"
)

// ManagerHandler handles corporate manager HTTP endpoints.
type ManagerHandler struct {
	uc *managerUC.UseCase
}

// NewManagerHandler creates a new ManagerHandler.
func NewManagerHandler(uc *managerUC.UseCase) *ManagerHandler {
	return &ManagerHandler{uc: uc}
}

// GetPurchases handles GET /api/v1/managers/{manager_id}/purchases.
func (h *ManagerHandler) GetPurchases(w http.ResponseWriter, r *http.Request) {
	managerID, err := parseManagerID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	purchases, err := h.uc.GetPurchases(r.Context(), managerID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"purchases": purchases,
		"total":     len(purchases),
	})
}

// GetPurchasePDF handles GET /api/v1/managers/{manager_id}/purchases/{payment_id}/pdf.
func (h *ManagerHandler) GetPurchasePDF(w http.ResponseWriter, r *http.Request) {
	managerID, err := parseManagerID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	paymentID := r.PathValue("payment_id")
	if paymentID == "" {
		respond.Error(w, http.StatusBadRequest, "bad_request", "payment_id is required")
		return
	}

	data, err := h.uc.GeneratePurchasePDF(r.Context(), managerID, paymentID)
	if err != nil {
		respond.ErrorFrom(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="mnemo_corporate_%s.pdf"`, paymentID))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// parseManagerID extracts and validates the manager_id path parameter.
func parseManagerID(r *http.Request) (int64, error) {
	raw := r.PathValue("manager_id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("manager_id must be a valid positive integer")
	}
	return id, nil
}
