// Package handler defines all HTTP handlers and a Registry that the router
// uses for dependency injection. Each domain gets its own handler struct.
// Handlers are thin — they parse input, call service layer, and write response.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
)

// Registry holds all domain handlers.
type Registry struct {
	Health        *HealthHandler
	Auth          *AuthHandler
	Onboarding    *OnboardingHandler
	Distributor   *DistributorHandler
	Verification  *VerificationHandler
	Credit        *CreditHandler
	Risk          *RiskHandler
	Offer         *OfferHandler
	Agreement     *AgreementHandler
	Catalogue     *CatalogueHandler
	Order         *OrderHandler
	Payment       *PaymentHandler
	Invoice       *InvoiceHandler
	Outstanding   *OutstandingHandler
	Collections   *CollectionsHandler
	Behaviour     *BehaviourHandler
	Enhancement   *EnhancementHandler
	CreditControl *CreditControlHandler
	Admin         *AdminHandler
	Audit         *AuditHandler
	Notification  *NotificationHandler
}

// ────────────────────────────────────────────────────────────────────────────
// Health Handler
// ────────────────────────────────────────────────────────────────────────────

type HealthHandler struct{}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, map[string]string{"status": "ok", "version": "v1"})
}

// ────────────────────────────────────────────────────────────────────────────
// Stub handlers — to be implemented in subsequent milestones.
// ────────────────────────────────────────────────────────────────────────────

type RiskHandler struct{}
type OutstandingHandler struct{}
type BehaviourHandler struct{}
type CreditControlHandler struct{}
type AuditHandler struct{}
type NotificationHandler struct{}

// stub marks an endpoint as not yet implemented.
func stub(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "not_implemented",
		"message": "this endpoint will be implemented in an upcoming milestone",
	})
}

// Risk
func (h *RiskHandler) GetFlags(w http.ResponseWriter, r *http.Request)    { stub(w, r) }
func (h *RiskHandler) ResolveFlag(w http.ResponseWriter, r *http.Request) { stub(w, r) }

// Outstanding
func (h *OutstandingHandler) GetMine(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *OutstandingHandler) Get(w http.ResponseWriter, r *http.Request)     { stub(w, r) }
func (h *OutstandingHandler) List(w http.ResponseWriter, r *http.Request)    { stub(w, r) }

// Behaviour
func (h *BehaviourHandler) Get(w http.ResponseWriter, r *http.Request)         { stub(w, r) }
func (h *BehaviourHandler) Recalculate(w http.ResponseWriter, r *http.Request) { stub(w, r) }

// Credit Control
func (h *CreditControlHandler) Get(w http.ResponseWriter, r *http.Request)      { stub(w, r) }
func (h *CreditControlHandler) Restrict(w http.ResponseWriter, r *http.Request) { stub(w, r) }
func (h *CreditControlHandler) Hold(w http.ResponseWriter, r *http.Request)     { stub(w, r) }
func (h *CreditControlHandler) Unhold(w http.ResponseWriter, r *http.Request)   { stub(w, r) }
func (h *CreditControlHandler) Block(w http.ResponseWriter, r *http.Request)    { stub(w, r) }
func (h *CreditControlHandler) Activate(w http.ResponseWriter, r *http.Request) { stub(w, r) }

// Audit
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request)         { stub(w, r) }
func (h *AuditHandler) GetForEntity(w http.ResponseWriter, r *http.Request) { stub(w, r) }

// Notification
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request)  { stub(w, r) }
func (h *NotificationHandler) Retry(w http.ResponseWriter, r *http.Request) { stub(w, r) }
