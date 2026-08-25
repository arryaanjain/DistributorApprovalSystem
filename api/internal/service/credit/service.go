// Package credit orchestrates application credit scoring, decision making,
// credit offer generation, and agreement creation.
package credit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/engine"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

// Service coordinates scoring, decisions, and credit offer management.
type Service struct {
	creditRepo *repository.CreditRepository
	distRepo   *repository.DistributorRepository
	verRepo    *repository.VerificationRepository
}

func New(
	creditRepo *repository.CreditRepository,
	distRepo *repository.DistributorRepository,
	verRepo *repository.VerificationRepository,
) *Service {
	return &Service{
		creditRepo: creditRepo,
		distRepo:   distRepo,
		verRepo:    verRepo,
	}
}

// EvaluateApplication runs scoring, hard risk evaluation, and decision engine for an application.
func (s *Service) EvaluateApplication(ctx context.Context, appID string) (*repository.CreditDecisionRecord, error) {
	app, err := s.distRepo.GetApplicationByID(ctx, appID)
	if err != nil || app == nil {
		return nil, apperrors.NotFound("application not found")
	}

	distID := app.DistributorID

	bp, bpErr := s.distRepo.GetBusinessProfile(ctx, distID)
	if bpErr != nil {
		slog.Warn("fetching business profile failed during credit eval", "distributor_id", distID, "error", bpErr)
	}

	docs, docErr := s.distRepo.GetBusinessDocuments(ctx, distID)
	if docErr != nil {
		slog.Warn("fetching business documents failed during credit eval", "distributor_id", distID, "error", docErr)
	}

	vers, verErr := s.verRepo.GetAllForApplication(ctx, distID)
	if verErr != nil {
		slog.Warn("fetching verifications failed during credit eval", "distributor_id", distID, "error", verErr)
	}

	// 1. Calculate Score
	scoreResult := engine.CalculateScore(bp, docs, vers)

	// 2. Evaluate Hard Risk
	riskEval := engine.EvaluateHardRisk(app, docs, vers)

	// Save score & components
	scoreID, err := s.creditRepo.SaveScore(ctx, distID, appID, scoreResult.TotalScore, scoreResult.RiskGrade, scoreResult.Components)
	if err != nil {
		slog.Error("saving score failed", "error", err, "distributor_id", distID)
	}

	// Save risk flags (deactivates old flags and saves new ones)
	if err := s.creditRepo.SaveRiskFlags(ctx, distID, appID, riskEval.RiskFlags); err != nil {
		slog.Error("saving risk flags failed", "error", err, "distributor_id", distID)
	}

	// 3. Compute Decision
	pref := "15_days"
	if app.PaymentPreference != nil {
		pref = *app.PaymentPreference
	}
	decOutput := engine.ComputeDecision(scoreResult, riskEval, docs, bp, pref)

	slog.Info("credit evaluation result",
		"distributor_id", distID,
		"application_id", appID,
		"total_score", scoreResult.TotalScore,
		"risk_grade", scoreResult.RiskGrade,
		"hard_risk_triggered", riskEval.HardRiskTriggered,
		"risk_flags", riskEval.RiskFlags,
		"decision", decOutput.Decision,
		"approved_limit_inr", decOutput.ApprovedLimitPaise/100,
	)

	var scoreIDPtr *string
	if scoreID != "" {
		scoreIDPtr = &scoreID
	}

	decisionRecord := &repository.CreditDecisionRecord{
		ApplicationID:      appID,
		DistributorID:      distID,
		CreditScoreID:      scoreIDPtr,
		TotalScore:         scoreResult.TotalScore,
		RiskGrade:          scoreResult.RiskGrade,
		Decision:           decOutput.Decision,
		ApprovedLimitPaise: decOutput.ApprovedLimitPaise,
		ApprovedPeriodDays: decOutput.ApprovedPeriodDays,
		MaxOutstandingAge:  decOutput.MaxOutstandingAge,
		PaymentTerms:       decOutput.PaymentTerms,
		NonGSTCapped:       decOutput.NonGSTCapped,
		HardRiskTriggered:  decOutput.HardRiskTriggered,
		DecidedBy:          "SYSTEM",
	}

	decID, err := s.creditRepo.SaveDecision(ctx, decisionRecord)
	if err != nil {
		return nil, apperrors.Internal("saving credit decision", err)
	}
	decisionRecord.ID = decID

	// Create Offer
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	offer := &repository.CreditOfferRecord{
		ApplicationID:     appID,
		DecisionID:        decID,
		DistributorID:     distID,
		RiskGrade:         scoreResult.RiskGrade,
		OfferedLimitPaise: decOutput.ApprovedLimitPaise,
		OfferedPeriodDays: decOutput.ApprovedPeriodDays,
		MaxOutstandingAge: decOutput.MaxOutstandingAge,
		PaymentTerms:      decOutput.PaymentTerms,
		ExpiresAt:         expiresAt,
	}
	if _, err := s.creditRepo.CreateOffer(ctx, offer); err != nil {
		slog.Error("creating credit offer failed", "error", err, "distributor_id", distID)
	}

	// Update application status
	newStatus := "offer_generated"
	if decOutput.Decision == "ADVANCE_ONLY" {
		newStatus = "advance_only"
	}
	actor := "system"
	reason := fmt.Sprintf("credit evaluation complete (score: %d, decision: %s)", scoreResult.TotalScore, decOutput.Decision)
	_ = s.distRepo.UpdateApplicationStatus(ctx, appID, newStatus, actor, nil, &reason)

	return decisionRecord, nil
}

// GetDecision returns the decision record for an application.
func (s *Service) GetDecision(ctx context.Context, appID string) (*repository.CreditDecisionRecord, error) {
	d, err := s.creditRepo.GetDecisionByAppID(ctx, appID)
	if err != nil {
		return nil, apperrors.Internal("fetching decision", err)
	}
	if d == nil {
		return nil, apperrors.NotFound("decision not found")
	}
	return d, nil
}

// GetOffer returns the active offer for a distributor.
func (s *Service) GetOffer(ctx context.Context, distributorID string) (*repository.CreditOfferRecord, error) {
	o, err := s.creditRepo.GetActiveOfferByDistributor(ctx, distributorID)
	if err != nil {
		return nil, apperrors.Internal("fetching offer", err)
	}
	if o == nil {
		return nil, apperrors.NotFound("no active credit offer found")
	}
	return o, nil
}

// AcceptOffer marks the offer as accepted and creates the formal agreement.
func (s *Service) AcceptOffer(ctx context.Context, distributorID, offerID string) (*repository.AgreementRecord, error) {
	offer, err := s.creditRepo.GetActiveOfferByDistributor(ctx, distributorID)
	if err != nil || offer == nil || offer.ID != offerID {
		return nil, apperrors.NotFound("offer not found or invalid")
	}

	if err := s.creditRepo.UpdateOfferStatus(ctx, offerID, "ACCEPTED"); err != nil {
		return nil, apperrors.Internal("updating offer status", err)
	}

	app, err := s.distRepo.GetActiveApplication(ctx, distributorID)
	appID := ""
	if app != nil {
		appID = app.ID
	}

	agNumber := fmt.Sprintf("KCN-AGR-%d", time.Now().Unix())
	ag := &repository.AgreementRecord{
		DistributorID:      distributorID,
		ApplicationID:      appID,
		AgreementNumber:    agNumber,
		Version:            "v1.0",
		ApprovedLimitPaise: offer.OfferedLimitPaise,
		ApprovedPeriodDays: offer.OfferedPeriodDays,
	}

	agID, err := s.creditRepo.CreateAgreement(ctx, ag)
	if err != nil {
		return nil, apperrors.Internal("creating agreement", err)
	}
	ag.ID = agID

	return ag, nil
}

// DeclineOffer marks the offer as declined.
func (s *Service) DeclineOffer(ctx context.Context, distributorID, offerID string) error {
	return s.creditRepo.UpdateOfferStatus(ctx, offerID, "DECLINED")
}
