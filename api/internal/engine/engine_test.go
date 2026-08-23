package engine_test

import (
	"testing"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/engine"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

func TestScoringAndDecisionEngine(t *testing.T) {
	vintage := 5.0
	fmcg := 5.0
	monthlyTurnover := int64(100000000) // ₹10 Lakhs in paise
	bureauScore := 760
	hasGST := true

	bp := &repository.BusinessProfileRecord{
		VintageYears:               &vintage,
		FMCGExperienceYears:        &fmcg,
		ApproxMonthlyBusinessPaise: &monthlyTurnover,
	}

	docs := &repository.BusinessDocumentRecord{
		HasGST: hasGST,
	}

	nameMatch := true
	vers := &repository.AllVerifications{
		PAN: &repository.PANVerificationRecord{
			Status:    "verified",
			NameMatch: &nameMatch,
		},
		GST: &repository.GSTVerificationRecord{
			Status:    "verified",
			NameMatch: &nameMatch,
		},
		CreditReport: &repository.CreditReportRecord{
			BureauScore: &bureauScore,
		},
	}

	// 1. Test High Score (Grade A+)
	score := engine.CalculateScore(bp, docs, vers)
	if score.TotalScore < 85 {
		t.Errorf("Expected score >= 85, got %d", score.TotalScore)
	}

	risk := engine.EvaluateHardRisk(nil, docs, vers)
	if risk.HardRiskTriggered {
		t.Errorf("Expected no hard risk, got triggered")
	}

	dec := engine.ComputeDecision(score, risk, docs, "15_days")
	if dec.Decision != "APPROVED" {
		t.Errorf("Expected APPROVED, got %s", dec.Decision)
	}
	if dec.ApprovedLimitPaise != 5000000 { // ₹50,000 in Paise
		t.Errorf("Expected limit 50,000 INR (5,000,000 paise), got %d", dec.ApprovedLimitPaise)
	}

	// 2. Test Non-GST Cap Rule (High score but no GST -> Capped at ₹25,000)
	docsNonGST := &repository.BusinessDocumentRecord{
		HasGST: false,
	}
	decNonGST := engine.ComputeDecision(score, risk, docsNonGST, "15_days")
	if decNonGST.ApprovedLimitPaise != 2500000 { // ₹25,000 in Paise
		t.Errorf("Expected Non-GST cap at 25,000 INR (2,500,000 paise), got %d", decNonGST.ApprovedLimitPaise)
	}
	if !decNonGST.NonGSTCapped {
		t.Errorf("Expected NonGSTCapped flag to be true")
	}

	// 3. Test Hard Risk Override (Writeoff / Default on bureau -> Advance Only)
	hasDefaults := true
	versHardRisk := &repository.AllVerifications{
		CreditReport: &repository.CreditReportRecord{
			BureauScore: &bureauScore,
			HasDefaults: &hasDefaults,
		},
	}
	riskTriggered := engine.EvaluateHardRisk(nil, docs, versHardRisk)
	if !riskTriggered.HardRiskTriggered {
		t.Errorf("Expected hard risk to be triggered due to bureau defaults")
	}
	decHardRisk := engine.ComputeDecision(score, riskTriggered, docs, "15_days")
	if decHardRisk.Decision != "ADVANCE_ONLY" {
		t.Errorf("Expected ADVANCE_ONLY decision for hard risk, got %s", decHardRisk.Decision)
	}
	if decHardRisk.ApprovedLimitPaise != 0 {
		t.Errorf("Expected 0 limit for hard risk, got %d", decHardRisk.ApprovedLimitPaise)
	}
}
