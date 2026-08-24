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
		if _, err := s.triggerPAN(ctx, distributorID, appID, *docs.PAN, expectedName); err != nil {
			slog.Error("PAN verification failed", "error", err, "distributor_id", distributorID)
		}
	}

	// ── GST Verification (if available) ───────────────────────────────────
	if docs != nil && docs.HasGST && docs.GSTNumber != nil {
		bizName := ""
		if bp != nil {
			bizName = bp.BusinessName
		}
		if _, err := s.triggerGST(ctx, distributorID, appID, *docs.GSTNumber, bizName); err != nil {
			slog.Error("GST verification failed", "error", err, "distributor_id", distributorID)
		}
	}

	// ── Bank Verification (Optional) ──────────────────────────────────────
	bankDetails, bErr := s.distRepo.GetBankDetails(ctx, distributorID)
	if bErr == nil && bankDetails != nil && bankDetails.AccountNumber != "" && bankDetails.IFSC != "" {
		if err := s.triggerBank(ctx, distributorID, appID, expectedName, bankDetails); err != nil {
			slog.Warn("bank verification failed", "error", err, "distributor_id", distributorID)
		}
	} else {
		slog.Info("skipping optional bank verification (no bank details provided)", "distributor_id", distributorID)
	}

	// ── Credit Report ─────────────────────────────────────────────────────
	if docs != nil && docs.PAN != nil {
		if err := s.triggerCreditReport(ctx, distributorID, appID, docs.PAN, &dist.Mobile, expectedName); err != nil {
			slog.Error("credit report fetch failed", "error", err, "distributor_id", distributorID)
		}
	}

	return nil
}

func (s *Service) triggerPAN(ctx context.Context, distributorID string, appID *string, pan, expectedName string) (*PANResult, error) {
	recID, err := s.verRepo.CreatePANVerification(ctx, distributorID, appID, pan)
	if err != nil {
		slog.Warn("failed to create initial pan_verifications DB record", "error", err, "distributor_id", distributorID)
	}

	result, err := s.client.VerifyPAN(ctx, pan, expectedName)
	if err != nil {
		return nil, err
	}

	raw := result.RawResponse
	nameMatch := result.NameMatch
	nameOnPAN := &result.NameOnPAN
	provRef := &result.ProviderRef

	if recID != "" {
		if dbErr := s.verRepo.UpdatePANVerification(ctx, recID,
			repository.VerificationStatus(result.Status), nameOnPAN, nameMatch, raw, provRef); dbErr != nil {
			slog.Error("failed to update pan_verifications DB record", "error", dbErr, "distributor_id", distributorID)
		}
	}
	return result, nil
}

func (s *Service) triggerGST(ctx context.Context, distributorID string, appID *string, gst, expectedName string) (*GSTResult, error) {
	recID, err := s.verRepo.CreateGSTVerification(ctx, distributorID, appID, gst)
	if err != nil {
		slog.Warn("failed to create initial gst_verifications DB record", "error", err, "distributor_id", distributorID)
	}

	result, err := s.client.VerifyGST(ctx, gst, expectedName)
	if err != nil {
		return nil, err
	}

	raw := result.RawResponse
	legalName := &result.LegalName
	tradeName := &result.TradeName
	gstStatus := &result.GSTStatus
	address := &result.Address
	constitution := &result.Constitution
	provRef := &result.ProviderRef

	if recID != "" {
		if dbErr := s.verRepo.UpdateGSTVerification(ctx, recID,
			repository.VerificationStatus(result.Status),
			legalName, tradeName, result.RegistrationDate,
			gstStatus, address, constitution, result.NameMatch, raw, provRef); dbErr != nil {
			slog.Error("failed to update gst_verifications DB record", "error", dbErr, "distributor_id", distributorID)
		}
	}
	return result, nil
}

func (s *Service) triggerBank(ctx context.Context, distributorID string, appID *string, expectedName string, bankDetails *repository.BankDetailRecord) error {
	if bankDetails == nil || bankDetails.AccountNumber == "" || bankDetails.IFSC == "" {
		return nil
	}
	accountNumber := bankDetails.AccountNumber
	ifsc := bankDetails.IFSC

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

func (s *Service) triggerCreditReport(ctx context.Context, distributorID string, appID, pan, mobile *string, name string) error {
	recID, err := s.verRepo.CreateCreditReport(ctx, distributorID, appID, pan, mobile)
	if err != nil {
		slog.Warn("failed to create initial credit_reports DB record", "error", err, "distributor_id", distributorID)
	}

	panStr := ""
	if pan != nil {
		panStr = *pan
	}
	mobileStr := ""
	if mobile != nil {
		mobileStr = *mobile
	}

	result, err := s.client.FetchCreditReport(ctx, mobileStr, panStr, name, "")
	if err != nil {
		slog.Error("credit report API call failed", "error", err, "distributor_id", distributorID)
		return nil
	}
	if result == nil {
		return nil
	}

	// Also fetch the PDF link
	pdfURL := ""
	if pdfLink, _, pdfErr := s.client.FetchCreditReportPDF(ctx, mobileStr, panStr, name, ""); pdfErr == nil {
		pdfURL = pdfLink
	}
	result.PDFURL = pdfURL

	raw := result.RawResponse
	pdfPtr := &result.PDFURL
	provRef := &result.ProviderRef

	if recID != "" {
		if dbErr := s.verRepo.UpdateCreditReport(ctx, recID,
			result.BureauScore, &result.HasDefaults, &result.HasWriteoffs, &result.HasSettlements,
			&result.TotalActiveLoans, &result.DelinquencyCount, result.FraudFlag,
			result.ReportDate, pdfPtr, provRef, raw); dbErr != nil {
			slog.Error("failed to update credit_reports DB record", "error", dbErr, "distributor_id", distributorID)
		}
	}
	return nil
}

// GetResults returns all latest verification results for a distributor.
func (s *Service) GetResults(ctx context.Context, distributorID string) (*repository.AllVerifications, error) {
	results, err := s.verRepo.GetAllForApplication(ctx, distributorID)
	if err != nil {
		return nil, apperrors.Internal("fetching verification results", err)
	}
	return results, nil
}

// VerifyPANOnly triggers a PAN verification call (or returns cached result if already verified) and returns the normalized result.
func (s *Service) VerifyPANOnly(ctx context.Context, distributorID, appID, pan, expectedName string) (*PANResult, error) {
	// Re-use existing verification record if already verified for this exact PAN
	latest, err := s.verRepo.GetLatestPANVerification(ctx, distributorID)
	if err == nil && latest != nil && latest.PAN == pan && (latest.Status == repository.VerificationVerified || latest.Status == repository.VerificationMismatch) {
		slog.Info("reusing cached PAN verification result (saving Surepass credits)", "distributor_id", distributorID, "pan", pan, "status", latest.Status)
		nameOnPAN := ""
		if latest.NameOnPAN != nil {
			nameOnPAN = *latest.NameOnPAN
		}
		provRef := ""
		if latest.ProviderRef != nil {
			provRef = *latest.ProviderRef
		}
		return &PANResult{
			Status:      Status(latest.Status),
			NameOnPAN:   nameOnPAN,
			NameMatch:   latest.NameMatch,
			ProviderRef: provRef,
		}, nil
	}

	var appPtr *string
	if appID != "" {
		appPtr = &appID
	}
	return s.triggerPAN(ctx, distributorID, appPtr, pan, expectedName)
}

// VerifyGSTOnly triggers a GST verification call (or returns cached result if already verified) and returns the normalized result.
func (s *Service) VerifyGSTOnly(ctx context.Context, distributorID, appID, gst, expectedName string) (*GSTResult, error) {
	// Re-use existing verification record if already verified for this exact GSTIN
	latest, err := s.verRepo.GetLatestGSTVerification(ctx, distributorID)
	if err == nil && latest != nil && latest.GSTNumber == gst && (latest.Status == repository.VerificationVerified || latest.Status == repository.VerificationMismatch || latest.Status == repository.VerificationPartiallyVerified) {
		slog.Info("reusing cached GST verification result (saving Surepass credits)", "distributor_id", distributorID, "gst", gst, "status", latest.Status)
		legalName := ""
		if latest.LegalName != nil {
			legalName = *latest.LegalName
		}
		tradeName := ""
		if latest.TradeName != nil {
			tradeName = *latest.TradeName
		}
		gstStatus := ""
		if latest.GSTStatus != nil {
			gstStatus = *latest.GSTStatus
		}
		address := ""
		if latest.Address != nil {
			address = *latest.Address
		}
		constitution := ""
		if latest.Constitution != nil {
			constitution = *latest.Constitution
		}
		provRef := ""
		if latest.ProviderRef != nil {
			provRef = *latest.ProviderRef
		}
		return &GSTResult{
			Status:           Status(latest.Status),
			LegalName:        legalName,
			TradeName:        tradeName,
			RegistrationDate: latest.RegistrationDate,
			GSTStatus:        gstStatus,
			Address:          address,
			Constitution:     constitution,
			NameMatch:        latest.NameMatch,
			ProviderRef:      provRef,
		}, nil
	}

	var appPtr *string
	if appID != "" {
		appPtr = &appID
	}
	return s.triggerGST(ctx, distributorID, appPtr, gst, expectedName)
}
