package handler

import (
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcfin "github.com/arryaanjain/DistributorApprovalSystem/internal/service/financial"
)

type EnhancementHandler struct {
	svc *svcfin.Service
}

func NewEnhancementHandler(svc *svcfin.Service) *EnhancementHandler {
	return &EnhancementHandler{svc: svc}
}

// POST /api/v1/credit/enhance/me
func (h *EnhancementHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	newLimit, err := h.svc.EvaluateCreditEnhancement(r.Context(), distID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if newLimit > 0 {
		response.JSON(w, map[string]interface{}{
			"status":            "ENHANCED",
			"message":           "Congratulations! You have unlocked a credit limit enhancement on the Credit Ladder.",
			"new_limit_paise":   newLimit,
			"new_limit_rupees":  newLimit / 100,
		})
		return
	}

	response.JSON(w, map[string]interface{}{
		"status":  "CURRENT",
		"message": "Continue your clean repayment history to unlock the next credit limit level.",
	})
}

// Stubs for legacy registry methods
func (h *EnhancementHandler) ListPending(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *EnhancementHandler) Approve(w http.ResponseWriter, r *http.Request)     { stub(w, r) }
func (h *EnhancementHandler) Reject(w http.ResponseWriter, r *http.Request)      { stub(w, r) }
