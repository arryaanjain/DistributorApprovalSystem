package handler

import (
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcorder "github.com/arryaanjain/DistributorApprovalSystem/internal/service/order"
	"github.com/go-chi/chi/v5"
)

type PaymentHandler struct {
	svc *svcorder.Service
}

func NewPaymentHandler(svc *svcorder.Service) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// POST /api/v1/admin/payments/{orderId}/verify
func (h *PaymentHandler) Verify(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderId")
	verifiedBy := middleware.UserIDFromContext(r.Context())

	if err := h.svc.VerifyPayment(r.Context(), orderID, verifiedBy); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "payment verified successfully"})
}

// Stubs for legacy registry methods
func (h *PaymentHandler) ListPending(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *PaymentHandler) Reject(w http.ResponseWriter, r *http.Request)      { stub(w, r) }
