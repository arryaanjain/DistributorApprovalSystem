// Package verification orchestrates all verification calls for an application.
// It coordinates the Surepass client with the verification repository.
package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

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
	existingPAN, _ := s.verRepo.GetLatestPANVerification(ctx, distributorID)
	if existingPAN == nil || existingPAN.Status != "verified" {
		if docs != nil && docs.PAN != nil {
			if _, err := s.triggerPAN(ctx, distributorID, appID, *docs.PAN, expectedName); err != nil {
				slog.Error("PAN verification failed", "error", err, "distributor_id", distributorID)
			}
		}
	} else {
		slog.Info("reusing existing verified PAN record", "distributor_id", distributorID)
	}

	// ── GST Verification (if available) ───────────────────────────────────
	existingGST, _ := s.verRepo.GetLatestGSTVerification(ctx, distributorID)
	if existingGST == nil || existingGST.Status != "verified" {
		if docs != nil && docs.HasGST && docs.GSTNumber != nil {
			bizName := ""
			if bp != nil {
				bizName = bp.BusinessName
			}
			if _, err := s.triggerGST(ctx, distributorID, appID, *docs.GSTNumber, expectedName, bizName); err != nil {
				slog.Error("GST verification failed", "error", err, "distributor_id", distributorID)
			}
		}
	} else {
		slog.Info("reusing existing verified GST record", "distributor_id", distributorID)
	}

	// ── Bank Verification (Optional) ──────────────────────────────────────
	existingBank, _ := s.verRepo.GetLatestBankVerification(ctx, distributorID)
	if existingBank == nil || existingBank.Status != "verified" {
		bankDetails, bErr := s.distRepo.GetBankDetails(ctx, distributorID)
		if bErr == nil && bankDetails != nil && bankDetails.AccountNumber != "" && bankDetails.IFSC != "" {
			if err := s.triggerBank(ctx, distributorID, appID, expectedName, bankDetails); err != nil {
				slog.Warn("bank verification failed", "error", err, "distributor_id", distributorID)
			}
		} else {
			slog.Info("skipping optional bank verification (no bank details provided)", "distributor_id", distributorID)
		}
	} else {
		slog.Info("reusing existing verified Bank record", "distributor_id", distributorID)
	}

	// ── Credit Report ─────────────────────────────────────────────────────
	// CIBIL report fetching is handled on-demand via manual admin trigger (TriggerCIBIL)
	existingCredit, _ := s.verRepo.GetLatestCreditReport(ctx, distributorID)
	if existingCredit != nil {
		slog.Info("reusing existing CIBIL credit report record in TriggerAll", "distributor_id", distributorID)
	} else {
		slog.Info("skipping automatic CIBIL fetch in TriggerAll; pending manual admin trigger", "distributor_id", distributorID)
	}

	return nil
}

// TriggerCIBIL manually executes CIBIL report fetch for an application.
// If force is false and an existing credit report record exists in DB, it reuses the existing report without touching the paid API.
func (s *Service) TriggerCIBIL(ctx context.Context, applicationID, distributorID string, force bool) (*CreditReportResult, error) {
	if !force {
		existingCredit, err := s.verRepo.GetLatestCreditReport(ctx, distributorID)
		if err == nil && existingCredit != nil && existingCredit.BureauScore != nil {
			slog.Info("reusing existing cached CIBIL credit report record", "distributor_id", distributorID, "bureau_score", *existingCredit.BureauScore)
			pdfURL := ""
			if existingCredit.PDFURL != nil {
				pdfURL = *existingCredit.PDFURL
			}
			provRef := ""
			if existingCredit.ProviderRef != nil {
				provRef = *existingCredit.ProviderRef
			}
			return &CreditReportResult{
				BureauScore:      existingCredit.BureauScore,
				HasDefaults:      existingCredit.HasDefaults != nil && *existingCredit.HasDefaults,
				HasWriteoffs:     existingCredit.HasWriteoffs != nil && *existingCredit.HasWriteoffs,
				HasSettlements:   existingCredit.HasSettlements != nil && *existingCredit.HasSettlements,
				TotalActiveLoans: func() int64 { if existingCredit.TotalActiveLoans != nil { return *existingCredit.TotalActiveLoans }; return 0 }(),
				DelinquencyCount: func() int { if existingCredit.DelinquencyCount != nil { return *existingCredit.DelinquencyCount }; return 0 }(),
				FraudFlag:        existingCredit.FraudFlag,
				ReportDate:       existingCredit.ReportDate,
				PDFURL:           pdfURL,
				ProviderRef:      provRef,
			}, nil
		}
	}

	docs, err := s.distRepo.GetBusinessDocuments(ctx, distributorID)
	if err != nil {
		return nil, apperrors.Internal("fetching business documents", err)
	}
	if docs == nil || docs.PAN == nil || strings.TrimSpace(*docs.PAN) == "" {
		return nil, apperrors.Validation("cannot fetch CIBIL report: PAN document is missing")
	}

	dist, err := s.distRepo.GetByID(ctx, distributorID)
	if err != nil || dist == nil {
		return nil, apperrors.NotFound("distributor not found")
	}

	expectedName := ""
	if dist.Name != nil {
		expectedName = *dist.Name
	}

	return s.triggerCreditReport(ctx, distributorID, &applicationID, docs.PAN, &dist.Mobile, expectedName)
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

func (s *Service) triggerGST(ctx context.Context, distributorID string, appID *string, gst, expectedUserName, expectedBizName string) (*GSTResult, error) {
	recID, err := s.verRepo.CreateGSTVerification(ctx, distributorID, appID, gst)
	if err != nil {
		slog.Warn("failed to create initial gst_verifications DB record", "error", err, "distributor_id", distributorID)
	}

	result, err := s.client.VerifyGST(ctx, gst, expectedUserName, expectedBizName)
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

func (s *Service) resolveApplicantName(ctx context.Context, distributorID, fallbackName string) string {
	if strings.TrimSpace(fallbackName) != "" {
		return strings.TrimSpace(fallbackName)
	}

	if panRec, err := s.verRepo.GetLatestPANVerification(ctx, distributorID); err == nil && panRec != nil && panRec.NameOnPAN != nil && strings.TrimSpace(*panRec.NameOnPAN) != "" {
		return strings.TrimSpace(*panRec.NameOnPAN)
	}

	if gstRec, err := s.verRepo.GetLatestGSTVerification(ctx, distributorID); err == nil && gstRec != nil {
		if gstRec.LegalName != nil && strings.TrimSpace(*gstRec.LegalName) != "" {
			return strings.TrimSpace(*gstRec.LegalName)
		}
		if gstRec.TradeName != nil && strings.TrimSpace(*gstRec.TradeName) != "" {
			return strings.TrimSpace(*gstRec.TradeName)
		}
	}

	if bp, err := s.distRepo.GetBusinessProfile(ctx, distributorID); err == nil && bp != nil && strings.TrimSpace(bp.BusinessName) != "" {
		return strings.TrimSpace(bp.BusinessName)
	}

	if dist, err := s.distRepo.GetByID(ctx, distributorID); err == nil && dist != nil && dist.Name != nil && strings.TrimSpace(*dist.Name) != "" {
		return strings.TrimSpace(*dist.Name)
	}

	return ""
}

func (s *Service) triggerCreditReport(ctx context.Context, distributorID string, appID, pan, mobile *string, expectedName string) (*CreditReportResult, error) {
	panStr := ""
	if pan != nil {
		panStr = strings.ToUpper(strings.TrimSpace(*pan))
	}
	mobileStr := ""
	if mobile != nil {
		mobileStr = strings.TrimSpace(*mobile)
	}

	// ── Pre-flight Input Validation ──────────────────────────────────────
	panRegex := regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)
	if !panRegex.MatchString(panStr) {
		return nil, apperrors.Validation(fmt.Sprintf("invalid PAN number format '%s' (expected 10-char PAN e.g. ABCDE1234F)", panStr))
	}

	cleanMobile := regexp.MustCompile(`\D`).ReplaceAllString(mobileStr, "")
	if len(cleanMobile) == 12 && strings.HasPrefix(cleanMobile, "91") {
		cleanMobile = cleanMobile[2:]
	}
	mobileRegex := regexp.MustCompile(`^[6-9]\d{9}$`)
	if !mobileRegex.MatchString(cleanMobile) {
		return nil, apperrors.Validation(fmt.Sprintf("invalid mobile number format '%s' (expected 10-digit Indian phone number)", mobileStr))
	}

	resolvedName := s.resolveApplicantName(ctx, distributorID, expectedName)
	if resolvedName == "" {
		return nil, apperrors.Validation("cannot fetch CIBIL report: applicant name could not be resolved from PAN, GST, Business Profile, or account records")
	}

	recID, err := s.verRepo.CreateCreditReport(ctx, distributorID, appID, &panStr, &cleanMobile)
	if err != nil {
		slog.Warn("failed to create initial credit_reports DB record", "error", err, "distributor_id", distributorID)
	}

	// Make single-token CIBIL JSON fetch call (DO NOT call PDF endpoint automatically)
	result, apiErr := s.client.FetchCreditReport(ctx, cleanMobile, panStr, resolvedName, "")

	var raw []byte
	var provRef *string
	if result != nil {
		raw = result.RawResponse
		if result.ProviderRef != "" {
			provRef = &result.ProviderRef
		}
	}

	if apiErr != nil {
		slog.Error("CIBIL report API call failed", "error", apiErr, "distributor_id", distributorID)
		if recID != "" && len(raw) > 0 {
			_ = s.verRepo.UpdateCreditReport(ctx, recID, nil, nil, nil, nil, nil, nil, false, nil, nil, provRef, raw)
		}
		return result, apiErr
	}

	if result == nil {
		return nil, apperrors.Internal("CIBIL API returned empty result", nil)
	}

	pdfPtr := &result.PDFURL
	if result.PDFURL == "" {
		pdfPtr = nil
	}

	if recID != "" {
		if dbErr := s.verRepo.UpdateCreditReport(ctx, recID,
			result.BureauScore, &result.HasDefaults, &result.HasWriteoffs, &result.HasSettlements,
			&result.TotalActiveLoans, &result.DelinquencyCount, result.FraudFlag,
			result.ReportDate, pdfPtr, provRef, raw); dbErr != nil {
			slog.Error("failed to update credit_reports DB record", "error", dbErr, "distributor_id", distributorID)
		}
	}

	return result, nil
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
		nameOnPAN := ""
		if latest.NameOnPAN != nil {
			nameOnPAN = *latest.NameOnPAN
		}
		provRef := ""
		if latest.ProviderRef != nil {
			provRef = *latest.ProviderRef
		}

		status := Status(latest.Status)
		isMatch := false
		if expectedName != "" && nameOnPAN != "" && NamesMatch(nameOnPAN, expectedName) {
			status = StatusVerified
			isMatch = true
		} else if latest.NameMatch != nil {
			isMatch = *latest.NameMatch
		}

		slog.Info("reusing cached PAN verification result", "distributor_id", distributorID, "pan", pan, "status", status)
		return &PANResult{
			Status:      status,
			NameOnPAN:   nameOnPAN,
			NameMatch:   &isMatch,
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
func (s *Service) VerifyGSTOnly(ctx context.Context, distributorID, appID, gst, expectedUserName, expectedBizName string) (*GSTResult, error) {
	// Re-use existing verification record if already verified for this exact GSTIN
	latest, err := s.verRepo.GetLatestGSTVerification(ctx, distributorID)
	if err == nil && latest != nil && latest.GSTNumber == gst && (latest.Status == repository.VerificationVerified || latest.Status == repository.VerificationMismatch || latest.Status == repository.VerificationPartiallyVerified) {
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

		status := Status(latest.Status)
		isMatch := false
		matched := (expectedUserName != "" && (NamesMatch(legalName, expectedUserName) || NamesMatch(tradeName, expectedUserName))) ||
			(expectedBizName != "" && (NamesMatch(legalName, expectedBizName) || NamesMatch(tradeName, expectedBizName)))

		if matched {
			if gstStatus != "" && gstStatus != "Active" {
				status = StatusPartiallyVerified
			} else {
				status = StatusVerified
			}
			isMatch = true
		} else if latest.NameMatch != nil {
			isMatch = *latest.NameMatch
		}

		slog.Info("reusing cached GST verification result", "distributor_id", distributorID, "gst", gst, "status", status)
		return &GSTResult{
			Status:           status,
			LegalName:        legalName,
			TradeName:        tradeName,
			RegistrationDate: latest.RegistrationDate,
			GSTStatus:        gstStatus,
			Address:          address,
			Constitution:     constitution,
			NameMatch:        &isMatch,
			ProviderRef:      provRef,
		}, nil
	}

	var appPtr *string
	if appID != "" {
		appPtr = &appID
	}
	return s.triggerGST(ctx, distributorID, appPtr, gst, expectedUserName, expectedBizName)
}
