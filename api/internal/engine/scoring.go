package engine

import (
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

// ScoreResult details total score and param breakdown.
type ScoreResult struct {
	TotalScore int            `json:"total_score"`
	RiskGrade  string         `json:"risk_grade"`
	Components map[string]int `json:"components"`
}

// CalculateScore computes a deterministic 0–100 score based on active policy weights.
func CalculateScore(
	bp *repository.BusinessProfileRecord,
	docs *repository.BusinessDocumentRecord,
	vers *repository.AllVerifications,
) *ScoreResult {
	components := make(map[string]int)

	// 1. Credit / Repayment Risk (30 pts max)
	creditScore := 0
	if vers != nil && vers.CreditReport != nil && vers.CreditReport.BureauScore != nil {
		bs := *vers.CreditReport.BureauScore
		if bs >= 750 {
			creditScore = 30
		} else if bs >= 700 {
			creditScore = 25
		} else if bs >= 650 {
			creditScore = 18
		} else if bs >= 600 {
			creditScore = 10
		} else {
			creditScore = 5
		}
	} else {
		// Default neutral score if no credit bureau report found
		creditScore = 15
	}
	components["credit_risk"] = creditScore

	// 2. Identity / KYC Verification (15 pts max)
	kycScore := 0
	if vers != nil && vers.PAN != nil && string(vers.PAN.Status) == "verified" {
		kycScore = 15
	} else if docs != nil && docs.PAN != nil {
		kycScore = 10
	}
	components["identity_kyc"] = kycScore

	// 3. Business Verification (15 pts max)
	bizVerScore := 0
	if docs != nil && docs.HasGST && docs.GSTNumber != nil {
		if vers != nil && vers.GST != nil && string(vers.GST.Status) == "verified" {
			bizVerScore = 15
		} else {
			bizVerScore = 10
		}
	} else {
		// Non-GST route with alternative evidence
		if docs != nil && (docs.FSSAINumber != nil || docs.UdyamNumber != nil || docs.ShopEstNumber != nil) {
			bizVerScore = 10
		} else {
			bizVerScore = 5
		}
	}
	components["business_verification"] = bizVerScore

	// 4. Business Vintage (10 pts max)
	vintageScore := 0
	if bp != nil && bp.VintageYears != nil {
		v := *bp.VintageYears
		if v >= 5 {
			vintageScore = 10
		} else if v >= 3 {
			vintageScore = 8
		} else if v >= 1 {
			vintageScore = 5
		} else {
			vintageScore = 2
		}
	}
	components["business_vintage"] = vintageScore

	// 5. Distribution / FMCG Experience (10 pts max)
	fmcgScore := 0
	if bp != nil && bp.FMCGExperienceYears != nil {
		e := *bp.FMCGExperienceYears
		if e >= 5 {
			fmcgScore = 10
		} else if e >= 3 {
			fmcgScore = 8
		} else if e >= 1 {
			fmcgScore = 5
		} else {
			fmcgScore = 2
		}
	}
	components["fmcg_experience"] = fmcgScore

	// 6. Business Capacity & Scale (10 pts max)
	scaleScore := 0
	if bp != nil && bp.ApproxMonthlyBusinessPaise != nil {
		inr := *bp.ApproxMonthlyBusinessPaise / 100
		if inr >= 1000000 { // >= 10 Lakhs
			scaleScore = 10
		} else if inr >= 500000 { // 5-10 Lakhs
			scaleScore = 8
		} else if inr >= 200000 { // 2-5 Lakhs
			scaleScore = 6
		} else {
			scaleScore = 3
		}
	}
	components["business_capacity"] = scaleScore

	// 7. Data Consistency / Integrity (10 pts max)
	consistencyScore := 10
	if vers != nil && vers.PAN != nil && vers.PAN.NameMatch != nil && !*vers.PAN.NameMatch {
		consistencyScore -= 5
	}
	if vers != nil && vers.GST != nil && vers.GST.NameMatch != nil && !*vers.GST.NameMatch {
		consistencyScore -= 5
	}
	if consistencyScore < 0 {
		consistencyScore = 0
	}
	components["data_integrity"] = consistencyScore

	// Sum total score
	total := 0
	for _, s := range components {
		total += s
	}

	// Classify Risk Grade
	riskGrade := "GRADE_C"
	if total >= 85 {
		riskGrade = "GRADE_A_PLUS"
	} else if total >= 75 {
		riskGrade = "GRADE_A"
	} else if total >= 65 {
		riskGrade = "GRADE_B"
	} else if total >= 55 {
		riskGrade = "GRADE_C"
	} else {
		riskGrade = "GRADE_HIGH_RISK"
	}

	return &ScoreResult{
		TotalScore: total,
		RiskGrade:  riskGrade,
		Components: components,
	}
}
