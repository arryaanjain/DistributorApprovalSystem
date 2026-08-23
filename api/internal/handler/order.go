package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcorder "github.com/arryaanjain/DistributorApprovalSystem/internal/service/order"
	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	svc *svcorder.Service
}

func NewOrderHandler(svc *svcorder.Service) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type createOrderReq struct {
	Items []svcorder.CreateOrderItemInput `json:"items"`
}

// POST /api/v1/orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req createOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	orderRec, err := h.svc.CreateOrder(r.Context(), distID, req.Items)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.Created(w, orderRec)
}

// GET /api/v1/orders/me
func (h *OrderHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	orders, err := h.svc.ListMyOrders(r.Context(), distID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, orders)
}

type paymentProofReq struct {
	ProofURL string `json:"proof_url"`
	UTR      string `json:"utr"`
}

// POST /api/v1/orders/{id}/payment-proof
func (h *OrderHandler) SubmitPaymentProof(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	var req paymentProofReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	if err := h.svc.SubmitPaymentProof(r.Context(), orderID, req.ProofURL, req.UTR); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "payment proof submitted successfully for verification"})
}

// GET /api/v1/admin/orders/pending-review
func (h *OrderHandler) ListPendingReview(w http.ResponseWriter, r *http.Request) {
	orders, err := h.svc.ListPendingReviews(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, orders)
}

type reviewOrderReq struct {
	Action string `json:"action"` // APPROVE or REJECT
	Notes  string `json:"notes"`
}

// POST /api/v1/admin/orders/{id}/review
func (h *OrderHandler) Review(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	reviewerID := middleware.UserIDFromContext(r.Context())

	var req reviewOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON payload")
		return
	}

	if err := h.svc.ReviewOrder(r.Context(), orderID, req.Action, reviewerID, req.Notes); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "order review completed"})
}

// POST /api/v1/admin/orders/{id}/dispatch
func (h *OrderHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	if err := h.svc.DispatchOrder(r.Context(), orderID); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "order dispatched and credit account balance updated"})
}

// Stubs for legacy registry methods
func (h *OrderHandler) GetMine(w http.ResponseWriter, r *http.Request)      { stub(w, r) }
func (h *OrderHandler) ListAll(w http.ResponseWriter, r *http.Request)      { stub(w, r) }
func (h *OrderHandler) AdminGet(w http.ResponseWriter, r *http.Request)     { stub(w, r) }
func (h *OrderHandler) VerifyPayment(w http.ResponseWriter, r *http.Request) { stub(w, r) }
