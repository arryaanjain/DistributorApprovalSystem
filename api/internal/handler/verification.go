package handler

import (
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcver "github.com/arryaanjain/DistributorApprovalSystem/internal/service/verification"
	"github.com/go-chi/chi/v5"
)

// VerificationHandler handles verification trigger and result endpoints.
type VerificationHandler struct {
	svc *svcver.Service
}

func NewVerificationHandler(svc *svcver.Service) *VerificationHandler {
	return &VerificationHandler{svc: svc}
}

// POST /api/v1/verification/{applicationId}/trigger
func (h *VerificationHandler) TriggerAll(w http.ResponseWriter, r *http.Request) {
	applicationID := chi.URLParam(r, "applicationId")
	if applicationID == "" {
		response.BadRequest(w, "applicationId is required")
		return
	}

	// distributorID needs to come from the application record
	// For now accept it as a query param during development
	distributorID := r.URL.Query().Get("distributor_id")
	if distributorID == "" {
		response.BadRequest(w, "distributor_id query param required")
		return
	}

	if err := h.svc.TriggerAll(r.Context(), applicationID, distributorID); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{
		"message": "verification triggered — results will be available shortly",
		"status":  "pending",
	})
}

// GET /api/v1/verification/{applicationId}/results
func (h *VerificationHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	distributorID := r.URL.Query().Get("distributor_id")
	if distributorID == "" {
		response.BadRequest(w, "distributor_id query param required")
		return
	}

	results, err := h.svc.GetResults(r.Context(), distributorID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, results)
}
