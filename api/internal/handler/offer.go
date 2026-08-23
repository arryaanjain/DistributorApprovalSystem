package handler

import (
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svccredit "github.com/arryaanjain/DistributorApprovalSystem/internal/service/credit"
	"github.com/go-chi/chi/v5"
)

type OfferHandler struct {
	svc *svccredit.Service
}

func NewOfferHandler(svc *svccredit.Service) *OfferHandler {
	return &OfferHandler{svc: svc}
}

// GET /api/v1/offers/me
func (h *OfferHandler) GetMyOffer(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	offer, err := h.svc.GetOffer(r.Context(), distID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, offer)
}

// POST /api/v1/offers/{id}/accept
func (h *OfferHandler) Accept(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	offerID := chi.URLParam(r, "id")

	ag, err := h.svc.AcceptOffer(r.Context(), distID, offerID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]interface{}{
		"message":   "credit offer accepted",
		"agreement": ag,
	})
}

// POST /api/v1/offers/{id}/decline
func (h *OfferHandler) Decline(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	offerID := chi.URLParam(r, "id")

	if err := h.svc.DeclineOffer(r.Context(), distID, offerID); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "credit offer declined"})
}

// Stubs for legacy registry methods
func (h *OfferHandler) Create(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *OfferHandler) Get(w http.ResponseWriter, r *http.Request)    { stub(w, r) }
