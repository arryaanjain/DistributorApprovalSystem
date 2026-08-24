package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/config"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svconboarding "github.com/arryaanjain/DistributorApprovalSystem/internal/service/onboarding"
)

// OnboardingHandler handles multi-step distributor onboarding.
type OnboardingHandler struct {
	svc *svconboarding.Service
	cfg *config.Config
}

func NewOnboardingHandler(svc *svconboarding.Service, cfg *config.Config) *OnboardingHandler {
	return &OnboardingHandler{svc: svc, cfg: cfg}
}

// POST /api/v1/onboarding/basic
func (h *OnboardingHandler) SubmitBasic(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req svconboarding.BasicInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	if err := h.svc.SubmitBasic(r.Context(), distID, &req); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "basic details saved", "next_step": "business"})
}

// POST /api/v1/onboarding/business
func (h *OnboardingHandler) SubmitBusiness(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req svconboarding.BusinessInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	if err := h.svc.SubmitBusiness(r.Context(), distID, &req); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "business details saved", "next_step": "statutory"})
}

// POST /api/v1/onboarding/statutory
func (h *OnboardingHandler) SubmitStatutory(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req svconboarding.StatutoryInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	result, err := h.svc.SubmitStatutory(r.Context(), distID, &req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	resp := map[string]interface{}{
		"message":         "statutory details saved",
		"next_step":       "bank",
		"pan_verified":    result.PANVerified,
		"gst_verified":    result.GSTVerified,
		"pan_holder_name": result.PANHolderName,
		"gst_legal_name":  result.GSTLegalName,
		"warnings":        result.Warnings,
	}
	if result.DuplicateResult != nil && result.DuplicateResult.SuspectFound {
		resp["warning"] = "your application contains identifiers that match an existing record and has been flagged for review"
		resp["matched_on"] = result.DuplicateResult.MatchedOn
	}

	response.JSON(w, resp)
}

// POST /api/v1/onboarding/bank
func (h *OnboardingHandler) SubmitBank(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req svconboarding.BankInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	dupResult, err := h.svc.SubmitBank(r.Context(), distID, &req)
	if err != nil {
		writeAppError(w, err)
		return
	}

	resp := map[string]interface{}{
		"message":   "bank details saved",
		"next_step": "preference",
	}
	if dupResult != nil && dupResult.SuspectFound {
		resp["warning"] = "this bank account is already associated with another application"
		resp["matched_on"] = dupResult.MatchedOn
	}

	response.JSON(w, resp)
}

// POST /api/v1/onboarding/preference
func (h *OnboardingHandler) SubmitPreference(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req svconboarding.PreferenceInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	if err := h.svc.SubmitPreference(r.Context(), distID, &req); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{
		"message":    "payment preference saved",
		"next_step":  "consent",
		"note":       "your preference is noted but Kresconet will make the final credit determination independently",
	})
}

// POST /api/v1/onboarding/consent
func (h *OnboardingHandler) SubmitConsent(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req svconboarding.ConsentInput
	// Capture IP and User-Agent from request
	req.IPAddress = r.Header.Get("X-Real-IP")
	if req.IPAddress == "" {
		req.IPAddress = r.RemoteAddr
	}
	req.UserAgent = r.Header.Get("User-Agent")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	// Extract mobile from context (we'd need to add this to the distributor JWT)
	// For now pass empty — TODO: add mobile to context in RequireDistributor middleware
	if err := h.svc.SubmitConsent(r.Context(), distID, "", &req); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{
		"message": "consent recorded, your application is now under review",
		"status":  "consent_given",
	})
}

// GET /api/v1/onboarding/status
func (h *OnboardingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	status, err := h.svc.GetStatus(r.Context(), distID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, status)
}

// POST /api/v1/onboarding/sample-order
func (h *OnboardingHandler) CreateSampleOrder(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req svconboarding.CreateSampleOrderInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	keyID := "rzp_test_kresconet_key"
	if h.cfg != nil && h.cfg.Razorpay.KeyID != "" {
		keyID = h.cfg.Razorpay.KeyID
	}

	res, err := h.svc.CreateSampleOrder(r.Context(), distID, &req, keyID)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, res)
}

// POST /api/v1/onboarding/sample-payment/verify
func (h *OnboardingHandler) VerifySamplePayment(w http.ResponseWriter, r *http.Request) {
	distID := middleware.DistributorIDFromContext(r.Context())

	var req svconboarding.VerifySamplePaymentInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	keySecret := ""
	if h.cfg != nil {
		keySecret = h.cfg.Razorpay.KeySecret
	}

	if err := h.svc.VerifySamplePayment(r.Context(), distID, &req, keySecret); err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{
		"message": "payment verified successfully, trial status activated",
		"status":  "trial",
	})
}
