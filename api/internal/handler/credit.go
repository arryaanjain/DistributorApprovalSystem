package handler

import (
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svccredit "github.com/arryaanjain/DistributorApprovalSystem/internal/service/credit"
	"github.com/go-chi/chi/v5"
)

type CreditHandler struct {
	svc *svccredit.Service
}

func NewCreditHandler(svc *svccredit.Service) *CreditHandler {
	return &CreditHandler{svc: svc}
}

// POST /api/v1/credit/evaluate/{applicationId}
func (h *CreditHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "applicationId")
	if appID == "" {
		response.BadRequest(w, "applicationId is required")
		return
	}

	decision, err := h.svc.EvaluateApplication(r.Context(), appID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, decision)
}

// GET /api/v1/credit/decision/{applicationId}
func (h *CreditHandler) GetDecision(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "applicationId")
	if appID == "" {
		response.BadRequest(w, "applicationId is required")
		return
	}

	decision, err := h.svc.GetDecision(r.Context(), appID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, decision)
}

// Stubs for legacy registry methods
func (h *CreditHandler) Score(w http.ResponseWriter, r *http.Request)  { h.Evaluate(w, r) }
func (h *CreditHandler) Decide(w http.ResponseWriter, r *http.Request) { h.Evaluate(w, r) }
