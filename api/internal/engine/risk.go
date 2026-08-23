package engine

import (
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

// RiskEvaluation holds hard risk results.
type RiskEvaluation struct {
	HardRiskTriggered bool     `json:"hard_risk_triggered"`
	RiskFlags         []string `json:"risk_flags"`
}

// EvaluateHardRisk checks hard override rules before final decision.
// If any hard flag is triggered, credit is blocked regardless of total score.
func EvaluateHardRisk(
	app *repository.ApplicationRecord,
	docs *repository.BusinessDocumentRecord,
	vers *repository.AllVerifications,
) *RiskEvaluation {
	var flags []string

	// 1. Invalid PAN or identity verification failure
	if vers != nil && vers.PAN != nil && string(vers.PAN.Status) == "failed" {
		flags = append(flags, "INVALID_PAN_IDENTITY")
	}

	// 2. Bank account mismatch / verification failed
	if vers != nil && vers.Bank != nil && string(vers.Bank.Status) == "failed" {
		flags = append(flags, "BANK_VERIFICATION_FAILED")
	}

	// 3. Active defaults / write-offs on CIBIL credit report
	if vers != nil && vers.CreditReport != nil {
		if vers.CreditReport.HasDefaults != nil && *vers.CreditReport.HasDefaults {
			flags = append(flags, "CREDIT_BUREAU_DEFAULT")
		}
		if vers.CreditReport.HasWriteoffs != nil && *vers.CreditReport.HasWriteoffs {
			flags = append(flags, "CREDIT_BUREAU_WRITEOFF")
		}
		if vers.CreditReport.FraudFlag {
			flags = append(flags, "BUREAU_FRAUD_FLAG")
		}
	}

	// 4. Duplicate suspect application flag
	if app != nil && app.IsDuplicateSuspect {
		flags = append(flags, "DUPLICATE_APPLICATION_SUSPECT")
	}

	return &RiskEvaluation{
		HardRiskTriggered: len(flags) > 0,
		RiskFlags:         flags,
	}
}
