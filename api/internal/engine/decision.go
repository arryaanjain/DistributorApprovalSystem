package engine

import (
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

// DecisionOutput is the 3-part credit determination result.
type DecisionOutput struct {
	Decision           string `json:"decision"` // APPROVED, ADVANCE_ONLY, MANUAL_REVIEW, REJECTED
	ApprovedLimitPaise int64  `json:"approved_limit_paise"`
	ApprovedPeriodDays int    `json:"approved_period_days"`
	MaxOutstandingAge  int    `json:"max_outstanding_age"`
	PaymentTerms       string `json:"payment_terms"`
	NonGSTCapped       bool   `json:"non_gst_capped"`
	HardRiskTriggered  bool   `json:"hard_risk_triggered"`
}

// ComputeDecision evaluates total score, hard risk flags, business turnover, and GST status to make a credit decision.
func ComputeDecision(
	scoreResult *ScoreResult,
	riskEval *RiskEvaluation,
	docs *repository.BusinessDocumentRecord,
	bp *repository.BusinessProfileRecord,
	preference string,
) *DecisionOutput {
	// If hard risk is triggered -> Block credit, require Advance Only
	if riskEval != nil && riskEval.HardRiskTriggered {
		return &DecisionOutput{
			Decision:           "ADVANCE_ONLY",
			ApprovedLimitPaise: 0,
			ApprovedPeriodDays: 0,
			MaxOutstandingAge:  0,
			PaymentTerms:       "ADVANCE_100",
			NonGSTCapped:       false,
			HardRiskTriggered:  true,
		}
	}

	score := scoreResult.TotalScore
	var rawLimitINR int64
	var decision string

	// Base limit derived from monthly business turnover (realistic 5% - 10% credit exposure)
	var monthlyTurnoverINR int64 = 500000 // default 5L if not provided
	if bp != nil && bp.ApproxMonthlyBusinessPaise != nil && *bp.ApproxMonthlyBusinessPaise > 0 {
		monthlyTurnoverINR = *bp.ApproxMonthlyBusinessPaise / 100
	}

	// Initial onboarding credit limit policy: Capped up to ₹50,000 max for initial orders.
	// Limits increase dynamically during subsequent order performance history.
	if score >= 85 {
		// Grade A+: 10% of monthly turnover, capped at ₹50,000 max for initial onboarding
		decision = "APPROVED"
		calc := int64(float64(monthlyTurnoverINR) * 0.10)
		if calc < 35000 {
			calc = 35000
		}
		if calc > 50000 {
			calc = 50000
		}
		rawLimitINR = calc
	} else if score >= 75 {
		// Grade A: 7.5% of monthly turnover, capped at ₹45,000 max
		decision = "APPROVED"
		calc := int64(float64(monthlyTurnoverINR) * 0.075)
		if calc < 30000 {
			calc = 30000
		}
		if calc > 45000 {
			calc = 45000
		}
		rawLimitINR = calc
	} else if score >= 65 {
		// Grade B: 5% of monthly turnover, capped at ₹35,000 max
		decision = "APPROVED"
		calc := int64(float64(monthlyTurnoverINR) * 0.05)
		if calc < 20000 {
			calc = 20000
		}
		if calc > 35000 {
			calc = 35000
		}
		rawLimitINR = calc
	} else if score >= 55 {
		// Grade C: 2.5% of monthly turnover, capped at ₹20,000 max
		decision = "APPROVED"
		calc := int64(float64(monthlyTurnoverINR) * 0.025)
		if calc < 10000 {
			calc = 10000
		}
		if calc > 20000 {
			calc = 20000
		}
		rawLimitINR = calc
	} else {
		rawLimitINR = 0
		decision = "ADVANCE_ONLY"
	}

	// Non-GST Cap Enforcer: Cap at ₹25,000 max if no GSTIN
	nonGSTCapped := false
	if docs != nil && !docs.HasGST {
		if rawLimitINR > 25000 {
			rawLimitINR = 25000
			nonGSTCapped = true
		}
	}

	// Determine Approved Period & Payment Terms from preference
	periodDays := 15
	terms := "15_DAYS_CREDIT"

	switch preference {
	case "30_days":
		periodDays = 30
		terms = "30_DAYS_CREDIT"
	case "15_days":
		periodDays = 15
		terms = "15_DAYS_CREDIT"
	case "cod":
		periodDays = 0
		terms = "CASH_ON_DELIVERY"
	case "advance_100":
		periodDays = 0
		terms = "ADVANCE_100"
		decision = "ADVANCE_ONLY"
		rawLimitINR = 0
	}

	maxAge := periodDays + 7 // grace period of 7 days before critical block

	return &DecisionOutput{
		Decision:           decision,
		ApprovedLimitPaise: rawLimitINR * 100, // convert INR to Paise
		ApprovedPeriodDays: periodDays,
		MaxOutstandingAge:  maxAge,
		PaymentTerms:       terms,
		NonGSTCapped:       nonGSTCapped,
		HardRiskTriggered:  false,
	}
}
