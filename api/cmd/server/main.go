// Kresconet Distributor Credit Platform — API Server
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/config"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/database"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/handler"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/policy"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/router"

	svcagr "github.com/arryaanjain/DistributorApprovalSystem/internal/service/agreement"
	svcauth "github.com/arryaanjain/DistributorApprovalSystem/internal/service/auth"
	svccredit "github.com/arryaanjain/DistributorApprovalSystem/internal/service/credit"
	svcfin "github.com/arryaanjain/DistributorApprovalSystem/internal/service/financial"
	svconboarding "github.com/arryaanjain/DistributorApprovalSystem/internal/service/onboarding"
	svcorder "github.com/arryaanjain/DistributorApprovalSystem/internal/service/order"
	svcver "github.com/arryaanjain/DistributorApprovalSystem/internal/service/verification"
)

func main() {
	// ── Structured logging ─────────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	slog.Info("Kresconet Distributor Credit Platform — starting")

	// ── Configuration ──────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	slog.Info("config loaded", "env", cfg.App.Env, "port", cfg.Server.Port)

	// ── Database ───────────────────────────────────────────────────────────
	ctx := context.Background()
	pool, err := database.Connect(ctx, &cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database connected")

	// ── Migrations ─────────────────────────────────────────────────────────
	if err := database.Migrate(pool, cfg.Database.MigrationsDir); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations applied")

	// ── Credit Policy ──────────────────────────────────────────────────────
	policyLoader := policy.NewLoader(pool)
	activePolicy, err := policyLoader.Active(ctx)
	if err != nil {
		slog.Error("failed to load credit policy", "error", err)
		os.Exit(1)
	}
	slog.Info("credit policy loaded", "version", activePolicy.Version)

	// ── Repositories ───────────────────────────────────────────────────────
	otpRepo    := repository.NewOTPRepository(pool)
	userRepo   := repository.NewUserRepository(pool)
	distRepo   := repository.NewDistributorRepository(pool)
	verRepo    := repository.NewVerificationRepository(pool)
	creditRepo := repository.NewCreditRepository(pool)
	orderRepo  := repository.NewOrderRepository(pool)
	finRepo    := repository.NewFinancialRepository(pool)

	// ── Services ───────────────────────────────────────────────────────────
	msg91     := svcauth.NewMSG91Client(&cfg.MSG91)
	authSvc   := svcauth.New(cfg, otpRepo, userRepo, distRepo, msg91)
	surepass  := svcver.NewSurepassClient(cfg.Surepass.BaseURL, cfg.Surepass.Token)
	verSvc    := svcver.New(verRepo, distRepo, surepass)
	creditSvc := svccredit.New(creditRepo, distRepo, verRepo)
	onbSvc    := svconboarding.New(distRepo, orderRepo, verSvc, creditSvc)
	agrSvc    := svcagr.New(creditRepo, distRepo, orderRepo, surepass)
	orderSvc  := svcorder.New(orderRepo, creditRepo, distRepo)
	finSvc    := svcfin.New(finRepo, orderRepo)

	// ── Handlers ───────────────────────────────────────────────────────────
	handlers := &handler.Registry{
		Health:        &handler.HealthHandler{},
		Auth:          handler.NewAuthHandler(authSvc),
		Onboarding:    handler.NewOnboardingHandler(onbSvc, cfg),
		Distributor:   handler.NewDistributorHandler(distRepo),
		Verification:  handler.NewVerificationHandler(verSvc),
		Credit:        handler.NewCreditHandler(creditSvc),
		Offer:         handler.NewOfferHandler(creditSvc),
		Agreement:     handler.NewAgreementHandler(agrSvc),
		Catalogue:     handler.NewCatalogueHandler(orderSvc),
		Order:         handler.NewOrderHandler(orderSvc),
		Payment:       handler.NewPaymentHandler(orderSvc),
		Invoice:       handler.NewInvoiceHandler(finSvc),
		Collections:   handler.NewCollectionsHandler(finSvc),
		Enhancement:   handler.NewEnhancementHandler(finSvc),
		Admin:         handler.NewAdminHandler(distRepo, verRepo, creditRepo, verSvc, creditSvc, policyLoader),
		// Remaining handlers — stubs until implemented
		Risk:          &handler.RiskHandler{},
		Outstanding:   &handler.OutstandingHandler{},
		Behaviour:     &handler.BehaviourHandler{},
		CreditControl: &handler.CreditControlHandler{},
		Audit:         &handler.AuditHandler{},
		Notification:  &handler.NotificationHandler{},
	}

	// ── Router ─────────────────────────────────────────────────────────────
	httpRouter := router.New(&router.Dependencies{
		Cfg:      cfg,
		Handlers: handlers,
	})

	// ── HTTP Server ────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      httpRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// ── Graceful Shutdown ──────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		slog.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}
