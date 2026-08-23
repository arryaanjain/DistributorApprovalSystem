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

// ComputeDecision evaluates total score, hard risk flags, and GST status to make a credit decision.
func ComputeDecision(
	scoreResult *ScoreResult,
	riskEval *RiskEvaluation,
	docs *repository.BusinessDocumentRecord,
	preference string,
) *DecisionOutput {
	// If hard risk is triggered -> Block credit, require Advance Only
	if riskEval != nil && riskEval.HardRiskTriggered {
		return &DecisionOutput{
			Decision:           "ADVANCE_ONLY",
			ApprovedLimitPaise: 0,
			ApprovedPeriodDays: 0,
			MaxOutstandingAge:  0,
			PaymentTerms       : "ADVANCE_100",
			NonGSTCapped:       false,
			HardRiskTriggered:  true,
		}
	}

	score := scoreResult.TotalScore
	var rawLimitINR int64
	var decision string

	// Score Bands Prequalification
	if score >= 85 {
		rawLimitINR = 50000
		decision = "APPROVED"
	} else if score >= 75 {
		rawLimitINR = 35000
		decision = "APPROVED"
	} else if score >= 65 {
		rawLimitINR = 25000
		decision = "APPROVED"
	} else if score >= 55 {
		rawLimitINR = 15000
		decision = "APPROVED"
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
