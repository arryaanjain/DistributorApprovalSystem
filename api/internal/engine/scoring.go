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
	// Requires verified credit bureau report (CIBIL). Unrated / missing report gets 0 pts.
	creditScore := 0
	if vers != nil && vers.CreditReport != nil && vers.CreditReport.BureauScore != nil {
		bs := *vers.CreditReport.BureauScore
		if bs >= 750 {
			creditScore = 30
		} else if bs >= 700 {
			creditScore = 25
		} else if bs >= 660 {
			creditScore = 20
		} else if bs >= 600 {
			creditScore = 12
		} else {
			creditScore = 5
		}
	} else {
		// Unrated / pending bureau report - zero points
		creditScore = 0
	}
	components["credit_risk"] = creditScore

	// 2. Identity / KYC Verification (15 pts max)
	// Requires successful PAN identity verification with tax authority. Unverified gets 0 pts.
	kycScore := 0
	if vers != nil && vers.PAN != nil && string(vers.PAN.Status) == "verified" {
		kycScore = 15
	}
	components["identity_kyc"] = kycScore

	// 3. Business Verification (15 pts max)
	// Requires successful GST verification with GSTN registry. Unverified gets 0 pts.
	bizVerScore := 0
	if docs != nil && docs.HasGST && docs.GSTNumber != nil && *docs.GSTNumber != "" {
		if vers != nil && vers.GST != nil && string(vers.GST.Status) == "verified" {
			bizVerScore = 15
		}
	} else if docs != nil && !docs.HasGST {
		// Non-GST route with alternative business registration evidence
		if docs.FSSAINumber != nil && *docs.FSSAINumber != "" || docs.UdyamNumber != nil && *docs.UdyamNumber != "" || docs.ShopEstNumber != nil && *docs.ShopEstNumber != "" {
			bizVerScore = 10
		}
	}
	components["business_verification"] = bizVerScore

	// 4. Compliance Credentials (FSSAI, Udyam, Shop & Est - 5 pts max)
	complianceScore := 0
	if docs != nil {
		if docs.FSSAINumber != nil && *docs.FSSAINumber != "" {
			complianceScore += 2
		}
		if docs.UdyamNumber != nil && *docs.UdyamNumber != "" {
			complianceScore += 2
		}
		if docs.ShopEstNumber != nil && *docs.ShopEstNumber != "" {
			complianceScore += 1
		}
		if complianceScore > 5 {
			complianceScore = 5
		}
	}
	components["compliance_credentials"] = complianceScore

	// 5. Business Vintage (10 pts max)
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

	// 6. Distribution / FMCG Experience (10 pts max)
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

	// 7. Brand Portfolio Diversity (5 pts max)
	brandScore := 0
	if bp != nil && len(bp.ExistingBrands) > 0 {
		validBrands := 0
		for _, b := range bp.ExistingBrands {
			if len(b) > 0 {
				validBrands++
			}
		}
		if validBrands >= 3 {
			brandScore = 5
		} else if validBrands >= 1 {
			brandScore = 3
		}
	}
	components["brand_portfolio"] = brandScore

	// 8. Business Capacity & Scale (5 pts max)
	scaleScore := 0
	if bp != nil && bp.ApproxMonthlyBusinessPaise != nil {
		inr := *bp.ApproxMonthlyBusinessPaise / 100
		if inr >= 1000000 { // >= 10 Lakhs
			scaleScore = 5
		} else if inr >= 500000 { // 5-10 Lakhs
			scaleScore = 4
		} else if inr >= 200000 { // 2-5 Lakhs
			scaleScore = 3
		} else {
			scaleScore = 1
		}
	}
	components["business_capacity"] = scaleScore

	// 9. Data Consistency & Name Matching (10 pts max)
	// Requires verifications to have run and matched names. Unverified gets 0 pts.
	consistencyScore := 0
	if vers != nil {
		if vers.PAN != nil && string(vers.PAN.Status) == "verified" {
			if vers.PAN.NameMatch == nil || *vers.PAN.NameMatch {
				consistencyScore += 5
			}
		}
		if vers.GST != nil && string(vers.GST.Status) == "verified" {
			if vers.GST.NameMatch == nil || *vers.GST.NameMatch {
				consistencyScore += 5
			}
		} else if docs != nil && !docs.HasGST {
			// Non-GST applicants get 5 pts if PAN name matched
			if vers.PAN != nil && string(vers.PAN.Status) == "verified" && (vers.PAN.NameMatch == nil || *vers.PAN.NameMatch) {
				consistencyScore += 5
			}
		}
	}
	components["data_integrity"] = consistencyScore

	// Sum total score
	total := 0
	for _, s := range components {
		total += s
	}
	if total > 100 {
		total = 100
	}

	// Classify Risk Grade
	riskGrade := "GRADE_HIGH_RISK"
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
