// Package onboarding implements the multi-step distributor onboarding flow.
// Each step validates its input, updates the distributor/application record,
// advances the application state machine, and runs duplicate checks.
package onboarding

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	svccredit "github.com/arryaanjain/DistributorApprovalSystem/internal/service/credit"
	svcver "github.com/arryaanjain/DistributorApprovalSystem/internal/service/verification"
)

// Service orchestrates the onboarding multi-step flow.
type Service struct {
	distRepo  *repository.DistributorRepository
	orderRepo *repository.OrderRepository
	verSvc    *svcver.Service
	creditSvc *svccredit.Service
}

func New(distRepo *repository.DistributorRepository, orderRepo *repository.OrderRepository, verSvc *svcver.Service, creditSvc *svccredit.Service) *Service {
	return &Service{
		distRepo:  distRepo,
		orderRepo: orderRepo,
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
	BusinessName                      string   `json:"business_name"   validate:"required,min=2,max=200"`
	Constitution                      string   `json:"constitution"    validate:"required,oneof=proprietorship partnership llp private_limited public_limited huf trust other"`
	AddressLine1                      string   `json:"address_line1"   validate:"required"`
	AddressLine2                      string   `json:"address_line2"`
	City                              string   `json:"city"            validate:"required"`
	State                             string   `json:"state"           validate:"required"`
	PIN                               string   `json:"pin"             validate:"required,len=6"`
	VintageYears                      float64  `json:"vintage_years"`
	FMCGExperienceYears               float64  `json:"fmcg_experience_years"`
	DistributionExperienceYears       float64  `json:"distribution_experience_years"`
	ApproxMonthlyBusinessINR          int64    `json:"approx_monthly_business_inr"`
	RetailerCount                     int      `json:"retailer_count"`
	ServicedRetailersWholesalersCount int      `json:"serviced_retailers_wholesalers_count"`
	SalespersonCount                  int      `json:"salesperson_count"`
	InterestedBusinessRole             string   `json:"interested_business_role"`
	ExistingBrands                    []string `json:"existing_brands"`
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

	distExp := in.DistributionExperienceYears
	if distExp == 0 && in.FMCGExperienceYears > 0 {
		distExp = in.FMCGExperienceYears
	}

	servicedCount := in.ServicedRetailersWholesalersCount
	if servicedCount == 0 && in.RetailerCount > 0 {
		servicedCount = in.RetailerCount
	}

	profile := &repository.BusinessProfileRecord{
		DistributorID:                     distributorID,
		BusinessName:                      in.BusinessName,
		Constitution:                      in.Constitution,
		AddressLine1:                      in.AddressLine1,
		AddressLine2:                      addrLine2,
		City:                              in.City,
		State:                             in.State,
		PIN:                               in.PIN,
		VintageYears:                      nonZeroFloat64(in.VintageYears),
		FMCGExperienceYears:               nonZeroFloat64(in.FMCGExperienceYears),
		DistributionExperienceYears:       nonZeroFloat64(distExp),
		ApproxMonthlyBusinessPaise:        nonZeroInt64(monthlyPaise),
		RetailerCount:                     nonZeroInt(in.RetailerCount),
		ServicedRetailersWholesalersCount: nonZeroInt(servicedCount),
		SalespersonCount:                  nonZeroInt(in.SalespersonCount),
		InterestedBusinessRole:             strPtrIfNotEmpty(in.InterestedBusinessRole),
		ExistingBrands:                    in.ExistingBrands,
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

// StatutoryResult holds duplicate check results, verification statuses, and warnings.
type StatutoryResult struct {
	DuplicateResult *DuplicateCheckResult `json:"duplicate_result,omitempty"`
	Warnings        []string              `json:"warnings,omitempty"`
	PANVerified     bool                  `json:"pan_verified"`
	GSTVerified     bool                  `json:"gst_verified"`
	PANHolderName   string                `json:"pan_holder_name,omitempty"`
	GSTLegalName    string                `json:"gst_legal_name,omitempty"`
}

// DuplicateCheckResult tells the caller whether suspicious matches were found.
type DuplicateCheckResult struct {
	SuspectFound bool     `json:"suspect_found"`
	MatchedOn    []string `json:"matched_on,omitempty"`
}

// SubmitStatutory saves statutory identifiers, performs real-time PAN & GST verifications, and runs duplicate detection.
func (s *Service) SubmitStatutory(ctx context.Context, distributorID string, in *StatutoryInput) (*StatutoryResult, error) {
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

	// Fetch Step 1 details for real-time name / business name cross-validation
	dist, _ := s.distRepo.GetByID(ctx, distributorID)
	step1Name := ""
	if dist != nil && dist.Name != nil {
		step1Name = *dist.Name
	}

	bp, _ := s.distRepo.GetBusinessProfile(ctx, distributorID)
	step1BusinessName := ""
	if bp != nil {
		step1BusinessName = bp.BusinessName
	}

	warnings := []string{}
	panVerified := false
	gstVerified := false
	panHolderName := ""
	gstLegalName := ""

	// ── Real-time PAN Verification ─────────────────────────────────────────
	if s.verSvc != nil && in.PAN != "" {
		panRes, err := s.verSvc.VerifyPANOnly(ctx, distributorID, app.ID, in.PAN, step1Name)
		if err != nil {
			slog.Error("real-time PAN verification error", "error", err, "distributor_id", distributorID)
			warnings = append(warnings, "PAN verification service unavailable; flagged for manual review")
		} else if panRes != nil {
			slog.Info("real-time PAN verification result", "status", panRes.Status, "name", panRes.NameOnPAN, "distributor_id", distributorID)
			panHolderName = panRes.NameOnPAN
			if panRes.Status == "verified" || (step1Name != "" && svcver.NamesMatch(panRes.NameOnPAN, step1Name)) {
				panVerified = true
			} else if panRes.Status == "mismatch" {
				warnings = append(warnings, fmt.Sprintf("PAN holder name '%s' does not match registered name '%s'", panRes.NameOnPAN, step1Name))
			} else if panRes.Status == "failed" || panRes.Status == "unavailable" {
				warnings = append(warnings, "PAN verification failed with tax authority records")
			}
		}
	}

	// ── Real-time GST Verification ─────────────────────────────────────────
	if s.verSvc != nil && in.GSTNumber != "" {
		gstRes, err := s.verSvc.VerifyGSTOnly(ctx, distributorID, app.ID, in.GSTNumber, step1Name, step1BusinessName)
		if err != nil {
			slog.Error("real-time GST verification error", "error", err, "distributor_id", distributorID)
			warnings = append(warnings, "GST verification service unavailable; flagged for manual review")
		} else if gstRes != nil {
			slog.Info("real-time GST verification result", "status", gstRes.Status, "legal_name", gstRes.LegalName, "distributor_id", distributorID)
			gstLegalName = gstRes.LegalName

			matched := (step1Name != "" && (svcver.NamesMatch(gstRes.LegalName, step1Name) || svcver.NamesMatch(gstRes.TradeName, step1Name))) ||
				(step1BusinessName != "" && (svcver.NamesMatch(gstRes.LegalName, step1BusinessName) || svcver.NamesMatch(gstRes.TradeName, step1BusinessName)))

			if matched || gstRes.Status == "verified" || gstRes.Status == "partially_verified" {
				gstVerified = true
			} else if gstRes.Status == "mismatch" {
				warnings = append(warnings, fmt.Sprintf("GST legal name '%s' / trade name '%s' does not match registered name '%s' or business name '%s'", gstRes.LegalName, gstRes.TradeName, step1Name, step1BusinessName))
			} else if gstRes.Status == "failed" || gstRes.Status == "unavailable" {
				warnings = append(warnings, "GST verification failed with GSTIN portal records")
			}

			// Cross-validate address details if available
			if bp != nil && gstRes.Address != "" {
				addrUpper := strings.ToUpper(gstRes.Address)
				cityUpper := strings.ToUpper(bp.City)
				pinStr := bp.PIN
				if cityUpper != "" && !strings.Contains(addrUpper, cityUpper) && pinStr != "" && !strings.Contains(addrUpper, pinStr) {
					warnings = append(warnings, fmt.Sprintf("GST address ('%s') differs from registered location (%s, PIN: %s)", gstRes.Address, bp.City, bp.PIN))
				}
			}
		}
	}

	reason := "statutory details submitted"
	_ = s.distRepo.UpdateApplicationStatus(ctx, app.ID, "statutory_submitted", "distributor", &distributorID, &reason)

	return &StatutoryResult{
		DuplicateResult: dupResult,
		Warnings:        warnings,
		PANVerified:     panVerified,
		GSTVerified:     gstVerified,
		PANHolderName:   panHolderName,
		GSTLegalName:    gstLegalName,
	}, nil
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
	AccountNumber string `json:"account_number"`
	IFSC          string `json:"ifsc"`
	AccountHolder string `json:"account_holder"`
	BankName      string `json:"bank_name"`
	Branch        string `json:"branch"`
}

func (s *Service) SubmitBank(ctx context.Context, distributorID string, in *BankInput) (*DuplicateCheckResult, error) {
	app, err := s.requireActiveApplication(ctx, distributorID)
	if err != nil {
		return nil, err
	}

	result := &DuplicateCheckResult{}

	// If bank details were provided, save them and run duplicate check
	if strings.TrimSpace(in.AccountNumber) != "" && strings.TrimSpace(in.IFSC) != "" {
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

		if existingID, err := s.distRepo.FindByBankAccount(ctx, in.AccountNumber, strings.ToUpper(in.IFSC)); err == nil && existingID != nil && *existingID != distributorID {
			result.SuspectFound = true
			result.MatchedOn = []string{"bank_account"}
			reason := "duplicate bank account detected"
			_ = s.distRepo.MarkApplicationDuplicate(ctx, app.ID, reason)
		}
	}

	reason := "bank details submitted (optional)"
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

	// CIBIL report fetching and credit score evaluation are triggered manually by Admin via the Admin Dashboard.
	slog.Info("Distributor consent recorded; application pending manual admin CIBIL report & score calculation", "application_id", app.ID, "distributor_id", distributorID)

	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// Status / Overview
// ────────────────────────────────────────────────────────────────────────────

// ApplicationStatus is returned by GetStatus.
type ApplicationStatus struct {
	ApplicationID       string  `json:"application_id"`
	Status              string  `json:"status"`
	PaymentPreference   *string `json:"payment_preference,omitempty"`
	ExposureClass       *string `json:"exposure_class,omitempty"`
	IsDuplicateSuspect  bool    `json:"is_duplicate_suspect"`
	AssignedCreditLimit int64   `json:"assigned_credit_limit,omitempty"`
}

func (s *Service) GetStatus(ctx context.Context, distributorID string) (*ApplicationStatus, error) {
	app, err := s.distRepo.GetLatestApplication(ctx, distributorID)
	if err != nil {
		return nil, apperrors.Internal("fetching application", err)
	}
	if app == nil {
		return nil, apperrors.NotFound("no application found")
	}

	var assignedLimit int64
	if s.creditSvc != nil {
		if dec, err := s.creditSvc.GetDecision(ctx, app.ID); err == nil && dec != nil {
			assignedLimit = dec.ApprovedLimitPaise
		}
	}

	return &ApplicationStatus{
		ApplicationID:       app.ID,
		Status:              app.Status,
		PaymentPreference:   app.PaymentPreference,
		ExposureClass:       app.ExposureClass,
		IsDuplicateSuspect:  app.IsDuplicateSuspect,
		AssignedCreditLimit: assignedLimit,
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

func nonEmptyStr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// ────────────────────────────────────────────────────────────────────────────
// Razorpay Sample Orders (Trial Flow)
// ────────────────────────────────────────────────────────────────────────────

type CreateSampleOrderInput struct {
	AmountPaise  int64                        `json:"amount_paise"`
	Items        []repository.OrderItemRecord `json:"items"`
	AddressLine1 string                       `json:"address_line1"`
	AddressLine2 string                       `json:"address_line2"`
	City         string                       `json:"city"`
	State        string                       `json:"state"`
	PIN          string                       `json:"pin"`
	Phone        string                       `json:"phone"`
}

type SampleOrderResult struct {
	SampleOrderID   string `json:"sample_order_id"`
	RazorpayOrderID string `json:"razorpay_order_id"`
	AmountPaise     int64  `json:"amount_paise"`
	Currency        string `json:"currency"`
	KeyID           string `json:"key_id"`
}

func (s *Service) CreateSampleOrder(ctx context.Context, distributorID string, in *CreateSampleOrderInput, rzpKeyID, rzpKeySecret string) (*SampleOrderResult, error) {
	app, err := s.requireActiveApplication(ctx, distributorID)
	if err != nil {
		return nil, err
	}

	amount := in.AmountPaise
	if amount <= 0 {
		amount = 50000 // Default ₹500
	}

	var addressID string
	if in.AddressLine1 != "" && in.City != "" {
		addrRec := &repository.AddressRecord{
			DistributorID: distributorID,
			AddressType:   "shipping",
			AddressLine1:  in.AddressLine1,
			AddressLine2:  nonEmptyStr(in.AddressLine2),
			City:          in.City,
			State:         in.State,
			PIN:           in.PIN,
			Phone:         nonEmptyStr(in.Phone),
		}
		addrID, err := s.orderRepo.CreateAddress(ctx, addrRec)
		if err == nil {
			addressID = addrID
		}
	}

	itemsJSONBytes, _ := json.Marshal(in.Items)

	// Attempt real Razorpay order creation via REST API if real keys are configured
	rzpOrderID := fmt.Sprintf("order_sim_%d", time.Now().UnixNano())
	if rzpKeyID != "" && rzpKeySecret != "" && strings.HasPrefix(rzpKeyID, "rzp_") && rzpKeyID != "rzp_test_kresconet_key" {
		realID, err := createRazorpayOrderAPI(ctx, rzpKeyID, rzpKeySecret, amount)
		if err == nil && realID != "" {
			rzpOrderID = realID
		} else if err != nil {
			slog.Warn("Failed creating real Razorpay API order, using simulated order ID", "error", err)
		}
	}

	sampleID, err := s.orderRepo.CreateSampleOrderWithAddress(ctx, distributorID, rzpOrderID, addressID, amount, string(itemsJSONBytes))
	if err != nil {
		return nil, apperrors.Internal("creating sample order", err)
	}

	reason := "sample order initiated"
	_ = s.distRepo.UpdateApplicationStatus(ctx, app.ID, "business_submitted", "distributor", &distributorID, &reason)

	return &SampleOrderResult{
		SampleOrderID:   sampleID,
		RazorpayOrderID: rzpOrderID,
		AmountPaise:     amount,
		Currency:        "INR",
		KeyID:           rzpKeyID,
	}, nil
}

func createRazorpayOrderAPI(ctx context.Context, keyID, keySecret string, amountPaise int64) (string, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"amount":   amountPaise,
		"currency": "INR",
		"receipt":  fmt.Sprintf("rcpt_sample_%d", time.Now().UnixNano()),
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.razorpay.com/v1/orders", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(keyID, keySecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("razorpay api error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var resStruct struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&resStruct); err != nil {
		return "", err
	}
	return resStruct.ID, nil
}

type VerifySamplePaymentInput struct {
	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
}

func (s *Service) VerifySamplePayment(ctx context.Context, distributorID string, in *VerifySamplePaymentInput, keySecret string) error {
	app, err := s.requireActiveApplication(ctx, distributorID)
	if err != nil {
		return err
	}

	// Verify HMAC-SHA256 signature if key secret is configured and not dummy/simulated
	if keySecret != "" && !strings.HasPrefix(keySecret, "rzp_test_kresconet") && !strings.HasPrefix(in.RazorpayPaymentID, "pay_sim_") {
		mac := hmac.New(sha256.New, []byte(keySecret))
		mac.Write([]byte(in.RazorpayOrderID + "|" + in.RazorpayPaymentID))
		expectedSig := hex.EncodeToString(mac.Sum(nil))

		if expectedSig != in.RazorpaySignature {
			return apperrors.Validation("invalid razorpay payment signature")
		}
	}

	if err := s.orderRepo.VerifySampleOrderPayment(ctx, in.RazorpayOrderID, in.RazorpayPaymentID, in.RazorpaySignature); err != nil {
		return apperrors.Internal("verifying sample payment", err)
	}

	reason := "trial status activated via sample payment"
	return s.distRepo.UpdateApplicationStatus(ctx, app.ID, "trial", "distributor", &distributorID, &reason)
}
