package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/policy"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	svccredit "github.com/arryaanjain/DistributorApprovalSystem/internal/service/credit"
	svcver "github.com/arryaanjain/DistributorApprovalSystem/internal/service/verification"
)

type AdminHandler struct {
	distRepo     *repository.DistributorRepository
	verRepo      *repository.VerificationRepository
	creditRepo   *repository.CreditRepository
	verSvc       *svcver.Service
	creditSvc    *svccredit.Service
	policyLoader *policy.Loader
}

func NewAdminHandler(
	distRepo *repository.DistributorRepository,
	verRepo *repository.VerificationRepository,
	creditRepo *repository.CreditRepository,
	verSvc *svcver.Service,
	creditSvc *svccredit.Service,
	policyLoader *policy.Loader,
) *AdminHandler {
	return &AdminHandler{
		distRepo:     distRepo,
		verRepo:      verRepo,
		creditRepo:   creditRepo,
		verSvc:       verSvc,
		creditSvc:    creditSvc,
		policyLoader: policyLoader,
	}
}

func (h *AdminHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	apps, total, err := h.distRepo.ListApplications(r.Context(), status, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, map[string]interface{}{
		"applications": apps,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *AdminHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	app, err := h.distRepo.GetApplicationByID(r.Context(), appID)
	if err != nil || app == nil {
		app, err = h.distRepo.GetActiveApplication(r.Context(), appID)
		if err != nil || app == nil {
			response.NotFound(w, "application not found")
			return
		}
	}

	dist, _ := h.distRepo.GetByID(r.Context(), app.DistributorID)
	profile, _ := h.distRepo.GetBusinessProfile(r.Context(), app.DistributorID)
	docs, _ := h.distRepo.GetBusinessDocuments(r.Context(), app.DistributorID)
	bank, _ := h.distRepo.GetBankDetails(r.Context(), app.DistributorID)

	verifications, _ := h.verRepo.GetAllForApplication(r.Context(), app.DistributorID)
	decision, _ := h.creditRepo.GetDecisionByAppID(r.Context(), appID)
	riskFlags, _ := h.creditRepo.GetActiveRiskFlags(r.Context(), app.DistributorID)
	scoreComponents, _ := h.creditRepo.GetScoreComponents(r.Context(), appID)

	response.JSON(w, map[string]interface{}{
		"application":      app,
		"distributor":      dist,
		"profile":          profile,
		"documents":        docs,
		"bank_details":     bank,
		"verifications":    verifications,
		"decision":         decision,
		"risk_flags":       riskFlags,
		"score_components": scoreComponents,
	})
}

type ActionReasonInput struct {
	Reason             string `json:"reason"`
	ApprovedLimitPaise *int64 `json:"approved_limit_paise"`
	ApprovedPeriodDays *int   `json:"approved_period_days"`
}

func (h *AdminHandler) ApproveApplication(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	userID := middleware.UserIDFromContext(r.Context())

	var input ActionReasonInput
	_ = json.NewDecoder(r.Body).Decode(&input)

	app, err := h.distRepo.GetApplicationByID(r.Context(), appID)
	if err != nil || app == nil {
		response.NotFound(w, "application not found")
		return
	}

	// 1. Run verifications if not already done
	if h.verSvc != nil {
		_ = h.verSvc.TriggerAll(r.Context(), app.ID, app.DistributorID)
	}

	// 2. Evaluate credit decision & generate offer
	var decRecord *repository.CreditDecisionRecord
	if h.creditSvc != nil {
		decRecord, _ = h.creditSvc.EvaluateApplication(r.Context(), app.ID)
	}

	// Manual Admin Override if limit specified
	if input.ApprovedLimitPaise != nil && h.creditRepo != nil {
		limit := *input.ApprovedLimitPaise
		period := 15
		if input.ApprovedPeriodDays != nil && *input.ApprovedPeriodDays > 0 {
			period = *input.ApprovedPeriodDays
		}
		eligibility := "APPROVED"
		if limit == 0 {
			eligibility = "ADVANCE_ONLY"
		}
		overrideRec := &repository.CreditDecisionRecord{
			ApplicationID:      app.ID,
			DistributorID:      app.DistributorID,
			PolicyVersion:      "v1.0-override",
			ApprovedLimitPaise: limit,
			ApprovedPeriodDays: period,
			Decision:           eligibility,
			DecidedBy:          userID,
			DecidedAt:          time.Now(),
		}
		if decRecord != nil && decRecord.CreditScoreID != nil {
			overrideRec.CreditScoreID = decRecord.CreditScoreID
		}
		_, _ = h.creditRepo.SaveDecision(r.Context(), overrideRec)
		decRecord = overrideRec
	}

	// 3. Mark application status as 'credit_active' once credit is calculated & approved
	targetStatus := "credit_active"
	if decRecord != nil && decRecord.Decision == "ADVANCE_ONLY" && (input.ApprovedLimitPaise == nil || *input.ApprovedLimitPaise == 0) {
		targetStatus = "advance_only"
	} else if decRecord != nil && decRecord.Decision == "REJECT" {
		targetStatus = "rejected"
	}

	reason := "Approved by employee"
	if input.Reason != "" {
		reason = input.Reason
	}

	var actorID *string
	if userID != "" {
		actorID = &userID
	}

	if err := h.distRepo.UpdateApplicationStatus(r.Context(), app.ID, targetStatus, "employee", actorID, &reason); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, map[string]interface{}{
		"status":         targetStatus,
		"application_id": app.ID,
		"message":        "Application approved and completed successfully",
	})
}

func (h *AdminHandler) RejectApplication(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	userID := middleware.UserIDFromContext(r.Context())

	var input ActionReasonInput
	_ = json.NewDecoder(r.Body).Decode(&input)

	app, err := h.distRepo.GetApplicationByID(r.Context(), appID)
	if err != nil || app == nil {
		response.NotFound(w, "application not found")
		return
	}

	reason := "Rejected by employee"
	if input.Reason != "" {
		reason = input.Reason
	}

	var actorID *string
	if userID != "" {
		actorID = &userID
	}

	if err := h.distRepo.UpdateApplicationStatus(r.Context(), app.ID, "rejected", "employee", actorID, &reason); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, map[string]interface{}{
		"status":         "rejected",
		"application_id": app.ID,
		"message":        "Application rejected",
	})
}

func (h *AdminHandler) HoldApplication(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	userID := middleware.UserIDFromContext(r.Context())

	var input ActionReasonInput
	_ = json.NewDecoder(r.Body).Decode(&input)

	app, err := h.distRepo.GetApplicationByID(r.Context(), appID)
	if err != nil || app == nil {
		response.NotFound(w, "application not found")
		return
	}

	reason := "Put on hold by employee"
	if input.Reason != "" {
		reason = input.Reason
	}

	var actorID *string
	if userID != "" {
		actorID = &userID
	}

	if err := h.distRepo.UpdateApplicationStatus(r.Context(), app.ID, "hold", "employee", actorID, &reason); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, map[string]interface{}{
		"status":         "hold",
		"application_id": app.ID,
		"message":        "Application put on hold",
	})
}

func (h *AdminHandler) ListDistributors(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	dists, total, err := h.distRepo.ListAll(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, map[string]interface{}{
		"distributors": dists,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *AdminHandler) GetDistributor(w http.ResponseWriter, r *http.Request) {
	distID := chi.URLParam(r, "id")
	dist, err := h.distRepo.GetByID(r.Context(), distID)
	if err != nil || dist == nil {
		response.NotFound(w, "distributor not found")
		return
	}

	profile, _ := h.distRepo.GetBusinessProfile(r.Context(), distID)
	docs, _ := h.distRepo.GetBusinessDocuments(r.Context(), distID)
	bank, _ := h.distRepo.GetBankDetails(r.Context(), distID)
	app, _ := h.distRepo.GetActiveApplication(r.Context(), distID)
	offer, _ := h.creditRepo.GetActiveOfferByDistributor(r.Context(), distID)

	response.JSON(w, map[string]interface{}{
		"distributor": dist,
		"profile":     profile,
		"documents":   docs,
		"bank":        bank,
		"application": app,
		"offer":       offer,
	})
}

func (h *AdminHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	if h.policyLoader == nil {
		response.JSON(w, map[string]interface{}{"status": "policy loader not initialized"})
		return
	}
	p, err := h.policyLoader.Active(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	response.JSON(w, p)
}

func (h *AdminHandler) ReloadPolicy(w http.ResponseWriter, r *http.Request) {
	if h.policyLoader == nil {
		response.JSON(w, map[string]interface{}{"status": "policy loader not initialized"})
		return
	}
	p, err := h.policyLoader.Reload(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	response.JSON(w, map[string]interface{}{
		"status": "reloaded",
		"policy": p,
	})
}

func (r *AdminHandler) GetDashboardStats(w http.ResponseWriter, req *http.Request) {
	if r.creditRepo == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "credit repo not initialized")
		return
	}
	stats, err := r.creditRepo.GetDashboardStats(req.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	response.JSON(w, stats)
}

func (h *AdminHandler) ListDuplicates(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, map[string]interface{}{
		"suspects": []interface{}{},
		"total":    0,
	})
}

func (h *AdminHandler) ResolveDuplicate(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, map[string]interface{}{
		"status": "resolved",
	})
}

