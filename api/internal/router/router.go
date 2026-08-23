// Package router wires the chi router with all versioned API routes.
package router

import (
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/config"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/handler"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
)

// Dependencies bundles all injected services that handlers need.
type Dependencies struct {
	Cfg     *config.Config
	Handlers *handler.Registry
}

// New builds and returns the fully wired HTTP router.
func New(deps *Dependencies) http.Handler {
	r := chi.NewRouter()

	// --- Global middleware (applied to all routes) ---
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.CleanPath)
	r.Use(chimiddleware.StripSlashes)

	// CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   deps.Cfg.Server.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID", "Idempotency-Key"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(corsHandler.Handler)

	h := deps.Handlers

	// --- Versioned API ---
	r.Route("/api/v1", func(r chi.Router) {

		// ── Health ────────────────────────────────────────────────────────────
		r.Get("/health", h.Health.Check)

		// ── Auth (distributor OTP + employee login) ───────────────────────────
		r.Route("/auth", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.OTPRateLimiter())
				r.Post("/otp/send", h.Auth.SendOTP)
				r.Post("/otp/verify", h.Auth.VerifyOTP)
			})
			r.Post("/employee/login", h.Auth.EmployeeLogin)
			r.Post("/employee/refresh", h.Auth.EmployeeRefresh)
		})

		// ── Onboarding (distributor-facing, requires distributor JWT) ─────────
		r.Route("/onboarding", func(r chi.Router) {
			r.Use(middleware.RequireDistributor(&deps.Cfg.JWT))
			r.Post("/basic",                  h.Onboarding.SubmitBasic)
			r.Post("/business",               h.Onboarding.SubmitBusiness)
			r.Post("/statutory",              h.Onboarding.SubmitStatutory)
			r.Post("/bank",                   h.Onboarding.SubmitBank)
			r.Post("/preference",             h.Onboarding.SubmitPreference)
			r.Post("/consent",                h.Onboarding.SubmitConsent)
			r.Post("/sample-order",           h.Onboarding.CreateSampleOrder)
			r.Post("/sample-payment/verify",  h.Onboarding.VerifySamplePayment)
			r.Get("/status",                  h.Onboarding.GetStatus)
		})

		// ── Distributors ───────────────────────────────────────────────────────
		r.Route("/distributors", func(r chi.Router) {
			// Self-service (distributor)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireDistributor(&deps.Cfg.JWT))
				r.Get("/me", h.Distributor.GetMe)
			})
			// Employee access
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
				r.Get("/", h.Distributor.List)
				r.Get("/{id}", h.Distributor.Get)
				r.Get("/{id}/summary", h.Distributor.Summary)
			})
		})

		// ── Verification (internal, runs async via queue) ─────────────────────
		r.Route("/verification", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Post("/{applicationId}/trigger", h.Verification.TriggerAll)
			r.Get("/{applicationId}/results",  h.Verification.GetResults)
		})

		// ── Credit ─────────────────────────────────────────────────────────────
		r.Route("/credit", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Post("/{applicationId}/score",    h.Credit.Score)
			r.Post("/{applicationId}/decide",   h.Credit.Decide)
			r.Get("/{applicationId}/decision",  h.Credit.GetDecision)
		})

		// ── Risk ───────────────────────────────────────────────────────────────
		r.Route("/risk", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Get("/{applicationId}/flags", h.Risk.GetFlags)
			r.Post("/flags/{id}/resolve",   h.Risk.ResolveFlag)
		})

		// ── Offers ─────────────────────────────────────────────────────────────
		r.Route("/offers", func(r chi.Router) {
			// Distributor views and accepts their own offer
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireDistributor(&deps.Cfg.JWT))
				r.Get("/mine",          h.Offer.GetMyOffer)
				r.Post("/{id}/accept",  h.Offer.Accept)
				r.Post("/{id}/decline", h.Offer.Decline)
			})
			// Employee creates / views offers
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
				r.Post("/", h.Offer.Create)
				r.Get("/{id}", h.Offer.Get)
			})
		})

		// ── Agreements ─────────────────────────────────────────────────────────
		r.Route("/agreements", func(r chi.Router) {
			r.Get("/esign-callback", h.Agreement.ESignCallback)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireDistributor(&deps.Cfg.JWT))
				r.Get("/mine",                 h.Agreement.GetMine)
				r.Post("/init-esign",          h.Agreement.InitESign)
				r.Post("/{id}/complete-esign", h.Agreement.CompleteESign)
				r.Post("/{id}/sign",           h.Agreement.Sign)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
				r.Post("/generate/{applicationId}", h.Agreement.Generate)
				r.Get("/{id}", h.Agreement.Get)
			})
		})

		// ── Catalogue ──────────────────────────────────────────────────────────
		r.Route("/catalogue", func(r chi.Router) {
			// Distributors browse catalogue & sample products
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireDistributor(&deps.Cfg.JWT))
				r.Get("/", h.Catalogue.List)
				r.Get("/samples", h.Catalogue.ListSamples)
				r.Get("/{id}", h.Catalogue.Get)
			})
			// Admin manages catalogue
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
				r.Get("/admin", h.Catalogue.ListAdmin)
				r.Post("/", h.Catalogue.Create)
				r.Put("/{id}", h.Catalogue.Update)
				r.Delete("/{id}", h.Catalogue.Delete)
			})
		})

		// ── Orders ─────────────────────────────────────────────────────────────
		r.Route("/orders", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireDistributor(&deps.Cfg.JWT))
				r.Post("/", h.Order.Create)
				r.Get("/", h.Order.ListMine)
				r.Get("/{id}", h.Order.GetMine)
				// Payment proof upload (Phase 12)
				r.Post("/{id}/payment-proof", h.Order.SubmitPaymentProof)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
				r.Get("/pending-review", h.Order.ListPendingReview)
				r.Get("/all", h.Order.ListAll)
				r.Get("/{id}/admin", h.Order.AdminGet)
				// First human review (Phase 13)
				r.Post("/{id}/review", h.Order.Review)
				// Payment verification (Phase 12)
				r.Post("/{id}/verify-payment", h.Order.VerifyPayment)
				// Dispatch
				r.Post("/{id}/dispatch", h.Order.Dispatch)
			})
		})

		// ── Payments ───────────────────────────────────────────────────────────
		r.Route("/payments", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Get("/pending-verification", h.Payment.ListPending)
			r.Post("/{id}/verify", h.Payment.Verify)
			r.Post("/{id}/reject", h.Payment.Reject)
		})

		// ── Invoices ───────────────────────────────────────────────────────────
		r.Route("/invoices", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireDistributor(&deps.Cfg.JWT))
				r.Get("/mine", h.Invoice.ListMine)
				r.Get("/{id}/mine", h.Invoice.GetMine)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
				r.Get("/", h.Invoice.List)
				r.Get("/{id}", h.Invoice.Get)
				r.Post("/", h.Invoice.Create)
				r.Post("/{id}/record-payment", h.Invoice.RecordPayment)
			})
		})

		// ── Outstanding ────────────────────────────────────────────────────────
		r.Route("/outstanding", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireDistributor(&deps.Cfg.JWT))
				r.Get("/mine", h.Outstanding.GetMine)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
				r.Get("/{distributorId}", h.Outstanding.Get)
				r.Get("/", h.Outstanding.List)
			})
		})

		// ── Collections ────────────────────────────────────────────────────────
		r.Route("/collections", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Get("/overdue", h.Collections.ListOverdue)
			r.Post("/{distributorId}/action", h.Collections.RecordAction)
		})

		// ── Behaviour ──────────────────────────────────────────────────────────
		r.Route("/behaviour", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Get("/{distributorId}", h.Behaviour.Get)
			r.Post("/{distributorId}/recalculate", h.Behaviour.Recalculate)
		})

		// ── Credit Enhancement ─────────────────────────────────────────────────
		r.Route("/credit-enhancement", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Get("/pending", h.Enhancement.ListPending)
			r.Get("/{distributorId}/evaluate", h.Enhancement.Evaluate)
			r.Post("/{id}/approve", h.Enhancement.Approve)
			r.Post("/{id}/reject",  h.Enhancement.Reject)
		})

		// ── Credit Control ─────────────────────────────────────────────────────
		r.Route("/credit-control", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Use(middleware.RequireRole("credit_manager", "accounts", "super_admin"))
			r.Get("/{distributorId}", h.CreditControl.Get)
			r.Post("/{distributorId}/restrict", h.CreditControl.Restrict)
			r.Post("/{distributorId}/hold",     h.CreditControl.Hold)
			r.Post("/{distributorId}/unhold",   h.CreditControl.Unhold)
			r.Post("/{distributorId}/block",    h.CreditControl.Block)
			r.Post("/{distributorId}/activate", h.CreditControl.Activate)
		})

		// ── Admin (internal management) ────────────────────────────────────────
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))

			// Application pipeline
			r.Get("/applications", h.Admin.ListApplications)
			r.Get("/applications/{id}", h.Admin.GetApplication)
			r.Post("/applications/{id}/approve", h.Admin.ApproveApplication)
			r.Post("/applications/{id}/reject",  h.Admin.RejectApplication)
			r.Post("/applications/{id}/hold",    h.Admin.HoldApplication)

			// Distributor management
			r.Get("/distributors",     h.Admin.ListDistributors)
			r.Get("/distributors/{id}", h.Admin.GetDistributor)

			// Policy management (super_admin only)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("super_admin"))
				r.Get("/policy", h.Admin.GetPolicy)
				r.Post("/policy/reload", h.Admin.ReloadPolicy)
			})

			// Duplicate suspects
			r.Get("/duplicates", h.Admin.ListDuplicates)
			r.Post("/duplicates/{id}/resolve", h.Admin.ResolveDuplicate)
		})

		// ── Audit ──────────────────────────────────────────────────────────────
		r.Route("/audit", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Use(middleware.RequireRole("super_admin", "credit_manager", "accounts"))
			r.Get("/", h.Audit.List)
			r.Get("/{entityType}/{entityId}", h.Audit.GetForEntity)
		})

		// ── Notifications ──────────────────────────────────────────────────────
		r.Route("/notifications", func(r chi.Router) {
			r.Use(middleware.RequireEmployee(&deps.Cfg.JWT))
			r.Get("/", h.Notification.List)
			r.Post("/{id}/retry", h.Notification.Retry)
		})
	})

	return r
}
