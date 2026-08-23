package handler

import (
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcagr "github.com/arryaanjain/DistributorApprovalSystem/internal/service/agreement"
	"github.com/go-chi/chi/v5"
)

type AgreementHandler struct {
	svc *svcagr.Service
}

func NewAgreementHandler(svc *svcagr.Service) *AgreementHandler {
	return &AgreementHandler{svc: svc}
}

// GET /api/v1/agreements/me
func (h *AgreementHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	ag, err := h.svc.GetMine(r.Context(), distID)
	if err != nil {
		writeAppError(w, err)
		return
	}
	response.JSON(w, ag)
}

// POST /api/v1/agreements/init-esign
func (h *AgreementHandler) InitESign(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	redirectURL := r.URL.Query().Get("redirect_url")

	res, err := h.svc.InitESign(r.Context(), distID, redirectURL)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, res)
}

// POST /api/v1/agreements/{id}/complete-esign
func (h *AgreementHandler) CompleteESign(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	agreementID := chi.URLParam(r, "id")
	providerRef := r.URL.Query().Get("provider_ref")

	if err := h.svc.CompleteSigning(r.Context(), distID, agreementID, providerRef); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{
		"message": "agreement successfully signed via Surepass e-Sign",
		"status":  "SIGNED",
	})
}

// GET /api/v1/agreements/esign-callback
func (h *AgreementHandler) ESignCallback(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	agreementID := r.URL.Query().Get("agreement_id")
	_ = token
	_ = agreementID

	// Redirect to distributor catalogue page
	http.Redirect(w, r, "http://localhost:3000/catalogue?esign=completed", http.StatusFound)
}

// POST /api/v1/agreements/{id}/sign (legacy / direct sign fallback)
func (h *AgreementHandler) Sign(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())
	agreementID := chi.URLParam(r, "id")

	if err := h.svc.CompleteSigning(r.Context(), distID, agreementID, ""); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{
		"message": "agreement successfully signed",
		"status":  "SIGNED",
	})
}

// Stubs for legacy registry methods
func (h *AgreementHandler) Generate(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *AgreementHandler) Get(w http.ResponseWriter, r *http.Request)      { stub(w, r) }
