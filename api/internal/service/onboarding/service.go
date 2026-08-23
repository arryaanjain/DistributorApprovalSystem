// Package onboarding implements the multi-step distributor onboarding flow.
// Each step validates its input, updates the distributor/application record,
// advances the application state machine, and runs duplicate checks.
package onboarding

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	svccredit "github.com/arryaanjain/DistributorApprovalSystem/internal/service/credit"
	svcver "github.com/arryaanjain/DistributorApprovalSystem/internal/service/verification"
)

// Service orchestrates the onboarding multi-step flow.
type Service struct {
	distRepo  *repository.DistributorRepository
	verSvc    *svcver.Service
	creditSvc *svccredit.Service
}

func New(distRepo *repository.DistributorRepository, verSvc *svcver.Service, creditSvc *svccredit.Service) *Service {
	return &Service{
		distRepo:  distRepo,
		verSvc:    verSvc,
		creditSvc: creditSvc,
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Step 1 – Basic Details
// ────────────────────────────────────────────────────────────────────────────

// BasicInput is what the distributor submits in step 1.
type BasicInput struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// SubmitBasic saves name + email and advances the application to 'basic_submitted'.
func (s *Service) SubmitBasic(ctx context.Context, distributorID string, in *BasicInput) error {
	// Ensure an active application exists — if not, create one
	app, err := s.distRepo.GetActiveApplication(ctx, distributorID)
	if err != nil {
		return apperrors.Internal("fetching application", err)
	}
	if app == nil {
		appID, err := s.distRepo.CreateApplication(ctx, distributorID)
		if err != nil {
			return apperrors.Internal("creating application", err)
		}
		// Set app ID for status update below
		_ = appID
	}

	if err := s.distRepo.UpdateBasic(ctx, distributorID, in.Name, in.Email); err != nil {
		return apperrors.Internal("saving basic details", err)
	}

	// Re-fetch to get app ID
	app, err = s.distRepo.GetActiveApplication(ctx, distributorID)
	if err != nil || app == nil {
		return apperrors.Internal("fetching application after creation", err)
	}

	actorType := "distributor"
	reason := "basic details submitted"
	return s.distRepo.UpdateApplicationStatus(ctx, app.ID, "basic_submitted", actorType, &distributorID, &reason)
}

// ────────────────────────────────────────────────────────────────────────────
// Step 2 – Business Details
// ────────────────────────────────────────────────────────────────────────────

type BusinessInput struct {
	BusinessName               string   `json:"business_name"   validate:"required,min=2,max=200"`
	Constitution               string   `json:"constitution"    validate:"required,oneof=proprietorship partnership llp private_limited public_limited huf trust other"`
	AddressLine1               string   `json:"address_line1"   validate:"required"`
	AddressLine2               string   `json:"address_line2"`
	City                       string   `json:"city"            validate:"required"`
	State                      string   `json:"state"           validate:"required"`
	PIN                        string   `json:"pin"             validate:"required,len=6"`
	VintageYears               float64  `json:"vintage_years"`
	FMCGExperienceYears        float64  `json:"fmcg_experience_years"`
	ApproxMonthlyBusinessINR   int64    `json:"approx_monthly_business_inr"`
	RetailerCount              int      `json:"retailer_count"`
	SalespersonCount           int      `json:"salesperson_count"`
	ExistingBrands             []string `json:"existing_brands"`
}

func (s *Service) SubmitBusiness(ctx context.Context, distributorID string, in *BusinessInput) error {
	app, err := s.requireActiveApplication(ctx, distributorID)
	if err != nil {
		return err
	}

	var addrLine2 *string
	if in.AddressLine2 != "" {
		addrLine2 = &in.AddressLine2
	}

	// Convert INR → paise for storage
	monthlyPaise := in.ApproxMonthlyBusinessINR * 100

	profile := &repository.BusinessProfileRecord{
		DistributorID:              distributorID,
		BusinessName:               in.BusinessName,
		Constitution:               in.Constitution,
		AddressLine1:               in.AddressLine1,
		AddressLine2:               addrLine2,
		City:                       in.City,
		State:                      in.State,
		PIN:                        in.PIN,
		VintageYears:               nonZeroFloat64(in.VintageYears),
		FMCGExperienceYears:        nonZeroFloat64(in.FMCGExperienceYears),
		ApproxMonthlyBusinessPaise: nonZeroInt64(monthlyPaise),
		RetailerCount:              nonZeroInt(in.RetailerCount),
		SalespersonCount:           nonZeroInt(in.SalespersonCount),
		ExistingBrands:             in.ExistingBrands,
	}

	if err := s.distRepo.UpsertBusinessProfile(ctx, profile); err != nil {
		return apperrors.Internal("saving business profile", err)
	}

	reason := "business details submitted"
	return s.distRepo.UpdateApplicationStatus(ctx, app.ID, "business_submitted", "distributor", &distributorID, &reason)
}

// ────────────────────────────────────────────────────────────────────────────
// Step 3 – Statutory Details
// ────────────────────────────────────────────────────────────────────────────

type StatutoryInput struct {
	PAN           string `json:"pan"            validate:"required,len=10"`
	GSTNumber     string `json:"gst_number"`    // optional
	FSSAINumber   string `json:"fssai_number"`
	UdyamNumber   string `json:"udyam_number"`
	ShopEstNumber string `json:"shop_est_number"`
}

// SubmitStatutory saves statutory identifiers and runs duplicate detection.
func (s *Service) SubmitStatutory(ctx context.Context, distributorID string, in *StatutoryInput) (*DuplicateCheckResult, error) {
	app, err := s.requireActiveApplication(ctx, distributorID)
	if err != nil {
		return nil, err
	}

	hasGST := strings.TrimSpace(in.GSTNumber) != ""

	doc := &repository.BusinessDocumentRecord{
		DistributorID: distributorID,
		PAN:           strPtr(strings.ToUpper(in.PAN)),
		GSTNumber:     strPtrIfNotEmpty(in.GSTNumber),
		FSSAINumber:   strPtrIfNotEmpty(in.FSSAINumber),
		UdyamNumber:   strPtrIfNotEmpty(in.UdyamNumber),
		ShopEstNumber: strPtrIfNotEmpty(in.ShopEstNumber),
		HasGST:        hasGST,
	}

	if err := s.distRepo.UpsertBusinessDocuments(ctx, doc); err != nil {
		return nil, apperrors.Internal("saving statutory details", err)
	}

	// ── Duplicate Detection (Phase 1) ──────────────────────────────────────
	dupResult := s.runStatutoryDuplicateCheck(ctx, distributorID, app.ID, in)

	reason := "statutory details submitted"
	_ = s.distRepo.UpdateApplicationStatus(ctx, app.ID, "statutory_submitted", "distributor", &distributorID, &reason)

	return dupResult, nil
}

// DuplicateCheckResult tells the caller whether suspicious matches were found.
type DuplicateCheckResult struct {
	SuspectFound bool     `json:"suspect_found"`
	MatchedOn    []string `json:"matched_on,omitempty"`
}

func (s *Service) runStatutoryDuplicateCheck(ctx context.Context, distributorID, appID string, in *StatutoryInput) *DuplicateCheckResult {
	result := &DuplicateCheckResult{}

	// Check PAN
	if existingID, err := s.distRepo.FindByPAN(ctx, strings.ToUpper(in.PAN)); err == nil && existingID != nil && *existingID != distributorID {
		result.SuspectFound = true
		result.MatchedOn = append(result.MatchedOn, "PAN")
	}

	// Check GST if provided
	if in.GSTNumber != "" {
		if existingID, err := s.distRepo.FindByGST(ctx, in.GSTNumber); err == nil && existingID != nil && *existingID != distributorID {
			result.SuspectFound = true
			result.MatchedOn = append(result.MatchedOn, "GST")
		}
	}

	if result.SuspectFound {
		reason := fmt.Sprintf("duplicate match on: %s", strings.Join(result.MatchedOn, ", "))
		_ = s.distRepo.MarkApplicationDuplicate(ctx, appID, reason)
	}

	return result
}

// ────────────────────────────────────────────────────────────────────────────
// Step 4 – Bank Details
// ────────────────────────────────────────────────────────────────────────────

type BankInput struct {
	AccountNumber string `json:"account_number" validate:"required"`
	IFSC          string `json:"ifsc"           validate:"required,len=11"`
	AccountHolder string `json:"account_holder" validate:"required"`
	BankName      string `json:"bank_name"`
	Branch        string `json:"branch"`
}

func (s *Service) SubmitBank(ctx context.Context, distributorID string, in *BankInput) (*DuplicateCheckResult, error) {
	app, err := s.requireActiveApplication(ctx, distributorID)
	if err != nil {
		return nil, err
	}

	bank := &repository.BankDetailRecord{
		DistributorID: distributorID,
		AccountNumber: in.AccountNumber,
		IFSC:          strings.ToUpper(in.IFSC),
		AccountHolder: in.AccountHolder,
		BankName:      strPtrIfNotEmpty(in.BankName),
		Branch:        strPtrIfNotEmpty(in.Branch),
	}

	if err := s.distRepo.UpsertBankDetails(ctx, bank); err != nil {
		return nil, apperrors.Internal("saving bank details", err)
	}

	// Duplicate check on bank account
	result := &DuplicateCheckResult{}
	if existingID, err := s.distRepo.FindByBankAccount(ctx, in.AccountNumber, strings.ToUpper(in.IFSC)); err == nil && existingID != nil && *existingID != distributorID {
		result.SuspectFound = true
		result.MatchedOn = []string{"bank_account"}
		reason := "duplicate bank account detected"
		_ = s.distRepo.MarkApplicationDuplicate(ctx, app.ID, reason)
	}

	reason := "bank details submitted"
	_ = s.distRepo.UpdateApplicationStatus(ctx, app.ID, "bank_submitted", "distributor", &distributorID, &reason)

	return result, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Step 5 – Payment Preference
// ────────────────────────────────────────────────────────────────────────────

type PreferenceInput struct {
	Preference string `json:"preference" validate:"required,oneof=advance_100 partial_delivery partial_receipt partial_15d cod 15_days 30_days bill_to_bill"`
}

// classifyExposure maps payment preference to risk exposure class.
func classifyExposure(preference string) string {
	switch preference {
	case "advance_100":
		return "low_no"
	case "partial_delivery", "partial_receipt", "cod":
		return "short"
	case "partial_15d", "15_days":
		return "standard"
	case "30_days", "bill_to_bill":
		return "extended"
	default:
		return "standard"
	}
}

func (s *Service) SubmitPreference(ctx context.Context, distributorID string, in *PreferenceInput) error {
	app, err := s.requireActiveApplication(ctx, distributorID)
	if err != nil {
		return err
	}

	exposureClass := classifyExposure(in.Preference)
	return s.distRepo.UpdateApplicationPreference(ctx, app.ID, in.Preference, exposureClass)
}

// ────────────────────────────────────────────────────────────────────────────
// Step 6 – Consent
// ────────────────────────────────────────────────────────────────────────────

type ConsentInput struct {
	ConsentType    string `json:"consent_type"    validate:"required,oneof=credit_assessment data_processing agreement"`
	ConsentText    string `json:"consent_text"    validate:"required"`
	ConsentVersion string `json:"consent_version" validate:"required"`
	IPAddress      string `json:"ip_address"`
	UserAgent      string `json:"user_agent"`
}

func (s *Service) SubmitConsent(ctx context.Context, distributorID, mobile string, in *ConsentInput) error {
	app, err := s.requireActiveApplication(ctx, distributorID)
	if err != nil {
		return err
	}

	if mobile == "" {
		dist, err := s.distRepo.GetByID(ctx, distributorID)
		if err == nil && dist != nil {
			mobile = dist.Mobile
		}
	}

	if err := s.distRepo.RecordConsent(ctx, distributorID, mobile,
		in.ConsentType, in.ConsentText, in.ConsentVersion, in.IPAddress, in.UserAgent); err != nil {
		return apperrors.Internal("recording consent", err)
	}

	reason := "consent given"
	if err := s.distRepo.UpdateApplicationStatus(ctx, app.ID, "consent_given", "distributor", &distributorID, &reason); err != nil {
		return err
	}

	// Auto-trigger verifications and credit evaluation asynchronously
	if s.verSvc != nil {
		go func(appID, distID string) {
			bgCtx := context.Background()
			slog.Info("Auto-triggering verifications after consent", "application_id", appID)
			if err := s.verSvc.TriggerAll(bgCtx, appID, distID); err != nil {
				slog.Error("failed auto-triggering verifications", "error", err, "application_id", appID)
			}

			if s.creditSvc != nil {
				slog.Info("Auto-evaluating credit decision", "application_id", appID)
				if _, err := s.creditSvc.EvaluateApplication(bgCtx, appID); err != nil {
					slog.Error("failed auto-evaluating credit application", "error", err, "application_id", appID)
				}
			}
		}(app.ID, distributorID)
	}

	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// Status / Overview
// ────────────────────────────────────────────────────────────────────────────

// ApplicationStatus is returned by GetStatus.
type ApplicationStatus struct {
	ApplicationID      string  `json:"application_id"`
	Status             string  `json:"status"`
	PaymentPreference  *string `json:"payment_preference,omitempty"`
	ExposureClass      *string `json:"exposure_class,omitempty"`
	IsDuplicateSuspect bool    `json:"is_duplicate_suspect"`
}

func (s *Service) GetStatus(ctx context.Context, distributorID string) (*ApplicationStatus, error) {
	app, err := s.distRepo.GetActiveApplication(ctx, distributorID)
	if err != nil {
		return nil, apperrors.Internal("fetching application", err)
	}
	if app == nil {
		return nil, apperrors.NotFound("no active application found")
	}
	return &ApplicationStatus{
		ApplicationID:      app.ID,
		Status:             app.Status,
		PaymentPreference:  app.PaymentPreference,
		ExposureClass:      app.ExposureClass,
		IsDuplicateSuspect: app.IsDuplicateSuspect,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

func (s *Service) requireActiveApplication(ctx context.Context, distributorID string) (*repository.ApplicationRecord, error) {
	app, err := s.distRepo.GetActiveApplication(ctx, distributorID)
	if err != nil {
		return nil, apperrors.Internal("fetching application", err)
	}
	if app == nil {
		appID, err := s.distRepo.CreateApplication(ctx, distributorID)
		if err != nil {
			return nil, apperrors.Internal("creating application", err)
		}
		app, err = s.distRepo.GetApplicationByID(ctx, appID)
		if err != nil || app == nil {
			return nil, apperrors.Internal("fetching application after creation", err)
		}
	}
	return app, nil
}

func strPtr(s string) *string { return &s }

func strPtrIfNotEmpty(s string) *string {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil
	}
	return &t
}

func nonZeroFloat64(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}

func nonZeroInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func nonZeroInt(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}
