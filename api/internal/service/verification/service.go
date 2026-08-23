// Package verification orchestrates all verification calls for an application.
// It coordinates the Surepass client with the verification repository.
package verification

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

// Service orchestrates verifications for a distributor application.
type Service struct {
	verRepo  *repository.VerificationRepository
	distRepo *repository.DistributorRepository
	client   *SurepassClient
}

// New creates a Verification service.
func New(verRepo *repository.VerificationRepository, distRepo *repository.DistributorRepository, client *SurepassClient) *Service {
	return &Service{verRepo: verRepo, distRepo: distRepo, client: client}
}

// TriggerAll launches all applicable verifications for an application.
// This is called either directly by an employee or by the scoring pipeline.
// Each verification is idempotent — it creates a new verification record per call.
func (s *Service) TriggerAll(ctx context.Context, applicationID, distributorID string) error {
	// Fetch the documents for this distributor
	docs, err := s.distRepo.GetBusinessDocuments(ctx, distributorID)
	if err != nil {
		return apperrors.Internal("fetching business documents", err)
	}

	dist, err := s.distRepo.GetByID(ctx, distributorID)
	if err != nil || dist == nil {
		return apperrors.NotFound("distributor not found")
	}

	bp, err := s.distRepo.GetBusinessProfile(ctx, distributorID)
	if err != nil {
		return apperrors.Internal("fetching business profile", err)
	}

	appID := &applicationID
	expectedName := ""
	if dist.Name != nil {
		expectedName = *dist.Name
	}

	// ── PAN Verification ──────────────────────────────────────────────────
	if docs != nil && docs.PAN != nil {
		if err := s.triggerPAN(ctx, distributorID, appID, *docs.PAN, expectedName); err != nil {
			slog.Error("PAN verification failed", "error", err, "distributor_id", distributorID)
		}
	}

	// ── GST Verification (if available) ───────────────────────────────────
	if docs != nil && docs.HasGST && docs.GSTNumber != nil {
		bizName := ""
		if bp != nil {
			bizName = bp.BusinessName
		}
		if err := s.triggerGST(ctx, distributorID, appID, *docs.GSTNumber, bizName); err != nil {
			slog.Error("GST verification failed", "error", err, "distributor_id", distributorID)
		}
	}

	// ── Bank Verification ─────────────────────────────────────────────────
	bank, err := s.distRepo.GetBusinessProfile(ctx, distributorID)
	if err == nil && bank != nil {
		// Bank details are in bank_details table — re-query
		if err := s.triggerBank(ctx, distributorID, appID, expectedName); err != nil {
			slog.Error("bank verification failed", "error", err, "distributor_id", distributorID)
		}
	}

	// ── Credit Report ─────────────────────────────────────────────────────
	if docs != nil && docs.PAN != nil {
		if err := s.triggerCreditReport(ctx, distributorID, appID, docs.PAN, &dist.Mobile); err != nil {
			slog.Error("credit report fetch failed", "error", err, "distributor_id", distributorID)
		}
	}

	return nil
}

func (s *Service) triggerPAN(ctx context.Context, distributorID string, appID *string, pan, expectedName string) error {
	recID, err := s.verRepo.CreatePANVerification(ctx, distributorID, appID, pan)
	if err != nil {
		return err
	}

	result, err := s.client.VerifyPAN(ctx, pan, expectedName)
	if err != nil {
		return err
	}

	raw, _ := json.Marshal(result.RawResponse)
	nameMatch := result.NameMatch
	nameOnPAN := &result.NameOnPAN
	provRef := &result.ProviderRef

	return s.verRepo.UpdatePANVerification(ctx, recID,
		repository.VerificationStatus(result.Status), nameOnPAN, nameMatch, raw, provRef)
}

func (s *Service) triggerGST(ctx context.Context, distributorID string, appID *string, gst, expectedName string) error {
	recID, err := s.verRepo.CreateGSTVerification(ctx, distributorID, appID, gst)
	if err != nil {
		return err
	}

	result, err := s.client.VerifyGST(ctx, gst, expectedName)
	if err != nil {
		return err
	}

	raw, _ := json.Marshal(result.RawResponse)
	legalName := &result.LegalName
	tradeName := &result.TradeName
	gstStatus := &result.GSTStatus
	address := &result.Address
	constitution := &result.Constitution
	provRef := &result.ProviderRef

	return s.verRepo.UpdateGSTVerification(ctx, recID,
		repository.VerificationStatus(result.Status),
		legalName, tradeName, result.RegistrationDate,
		gstStatus, address, constitution, result.NameMatch, raw, provRef)
}

func (s *Service) triggerBank(ctx context.Context, distributorID string, appID *string, expectedName string) error {
	bankDetails, err := s.distRepo.GetBankDetails(ctx, distributorID)
	var accountNumber, ifsc string
	if err == nil && bankDetails != nil {
		accountNumber = bankDetails.AccountNumber
		ifsc = bankDetails.IFSC
	}

	recID, err := s.verRepo.CreateBankVerification(ctx, distributorID, appID, accountNumber, ifsc)
	if err != nil {
		return err
	}

	result, err := s.client.VerifyBankAccount(ctx, accountNumber, ifsc, expectedName)
	if err != nil {
		return err
	}

	raw, _ := json.Marshal(result.RawResponse)
	holder := &result.AccountHolder
	bankName := &result.BankName
	provRef := &result.ProviderRef

	return s.verRepo.UpdateBankVerification(ctx, recID,
		repository.VerificationStatus(result.Status),
		holder, bankName, result.NameMatch, raw, provRef)
}

func (s *Service) triggerCreditReport(ctx context.Context, distributorID string, appID, pan, mobile *string) error {
	recID, err := s.verRepo.CreateCreditReport(ctx, distributorID, appID, pan, mobile)
	if err != nil {
		return err
	}

	panStr := ""
	if pan != nil {
		panStr = *pan
	}
	mobileStr := ""
	if mobile != nil {
		mobileStr = *mobile
	}

	result, err := s.client.FetchCreditReport(ctx, mobileStr, panStr)
	if err != nil {
		return err
	}

	raw, _ := json.Marshal(result.RawResponse)
	pdfURL := &result.PDFURL
	provRef := &result.ProviderRef

	return s.verRepo.UpdateCreditReport(ctx, recID,
		result.BureauScore, &result.HasDefaults, &result.HasWriteoffs, &result.HasSettlements,
		&result.TotalActiveLoans, &result.DelinquencyCount, result.FraudFlag,
		result.ReportDate, pdfURL, provRef, raw)
}

// GetResults returns all latest verification results for a distributor.
func (s *Service) GetResults(ctx context.Context, distributorID string) (*repository.AllVerifications, error) {
	results, err := s.verRepo.GetAllForApplication(ctx, distributorID)
	if err != nil {
		return nil, apperrors.Internal("fetching verification results", err)
	}
	return results, nil
}
