package handler

import (
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcfin "github.com/arryaanjain/DistributorApprovalSystem/internal/service/financial"
)

type CollectionsHandler struct {
	svc *svcfin.Service
}

func NewCollectionsHandler(svc *svcfin.Service) *CollectionsHandler {
	return &CollectionsHandler{svc: svc}
}

// POST /api/v1/admin/collections/evaluate
func (h *CollectionsHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	updated, err := h.svc.EvaluateOverdueCollections(r.Context())
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, map[string]interface{}{
		"message":          "overdue collections evaluation completed",
		"updated_invoices": updated,
	})
}

// Stubs for legacy registry methods
func (h *CollectionsHandler) ListOverdue(w http.ResponseWriter, r *http.Request)  { h.Evaluate(w, r) }
func (h *CollectionsHandler) RecordAction(w http.ResponseWriter, r *http.Request) { stub(w, r) }
