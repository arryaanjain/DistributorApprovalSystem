package financial_test

import (
	"testing"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

func TestCreditLadderStepUp(t *testing.T) {
	finRepo := &repository.FinancialRepository{}

	// Current limit ₹50,000 (5,000,000 paise) -> Next step should be ₹1,00,000 (10,000,000 paise)
	nextStep, canEnhance := finRepo.GetNextLadderStep(5000000)
	if !canEnhance {
		t.Fatalf("Expected credit ladder enhancement to be available")
	}
	if nextStep != 10000000 {
		t.Errorf("Expected next ladder step to be ₹1,00,000 (10,000,000 paise), got %d", nextStep)
	}

	// Max limit ₹3,00,000 (30,000,000 paise) -> No higher step available
	_, canEnhanceMax := finRepo.GetNextLadderStep(30000000)
	if canEnhanceMax {
		t.Errorf("Expected no higher step for maximum credit limit ₹3,00,000")
	}
}
