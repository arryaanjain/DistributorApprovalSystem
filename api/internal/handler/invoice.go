package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcfin "github.com/arryaanjain/DistributorApprovalSystem/internal/service/financial"
	"github.com/go-chi/chi/v5"
)

type InvoiceHandler struct {
	svc *svcfin.Service
}

func NewInvoiceHandler(svc *svcfin.Service) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

// GET /api/v1/invoices/me
func (h *InvoiceHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	invoices, err := h.svc.ListDistributorInvoices(r.Context(), distID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, invoices)
}

type recordPaymentReq struct {
	PaymentMode string `json:"payment_mode"`
	UTR         string `json:"utr"`
	AmountPaise int64  `json:"amount_paise"`
}

// POST /api/v1/invoices/{id}/payments
func (h *InvoiceHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	invoiceID := chi.URLParam(r, "id")
	recordedBy := middleware.UserIDFromContext(r.Context())

	var req recordPaymentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	var recByID *string
	if recordedBy != "" {
		recByID = &recordedBy
	}

	if err := h.svc.RecordPayment(r.Context(), invoiceID, distID, req.PaymentMode, req.UTR, req.AmountPaise, recByID); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "payment recorded and credit limit replenished"})
}

// Stubs for legacy registry methods
func (h *InvoiceHandler) GetMine(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request)    { stub(w, r) }
func (h *InvoiceHandler) Get(w http.ResponseWriter, r *http.Request)     { stub(w, r) }
func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request)  { stub(w, r) }
