package agreement

import (
	"context"
	"fmt"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	svcver "github.com/arryaanjain/DistributorApprovalSystem/internal/service/verification"
)

type ESignInitResult struct {
	AgreementID string `json:"agreement_id"`
	SigningURL  string `json:"signing_url"`
	Token       string `json:"token"`
}

type Service struct {
	creditRepo *repository.CreditRepository
	distRepo   *repository.DistributorRepository
	orderRepo  *repository.OrderRepository
	esign      svcver.ESignClient
}

func New(
	creditRepo *repository.CreditRepository,
	distRepo *repository.DistributorRepository,
	orderRepo *repository.OrderRepository,
	esign svcver.ESignClient,
) *Service {
	return &Service{
		creditRepo: creditRepo,
		distRepo:   distRepo,
		orderRepo:  orderRepo,
		esign:      esign,
	}
}

// GetMine returns the current agreement for a distributor.
func (s *Service) GetMine(ctx context.Context, distributorID string) (*repository.AgreementRecord, error) {
	ag, err := s.creditRepo.GetAgreementByDistributor(ctx, distributorID)
	if err != nil {
		return nil, apperrors.Internal("fetching agreement", err)
	}
	if ag == nil {
		return nil, apperrors.NotFound("no agreement found for distributor")
	}
	return ag, nil
}

// InitESign initializes agreement PDF generation and Surepass e-Sign session.
func (s *Service) InitESign(ctx context.Context, distributorID string, redirectURL string) (*ESignInitResult, error) {
	offer, err := s.creditRepo.GetActiveOfferByDistributor(ctx, distributorID)
	if err != nil || offer == nil {
		return nil, apperrors.Validation("no active credit offer found to sign")
	}

	dist, err := s.distRepo.GetByID(ctx, distributorID)
	if err != nil || dist == nil {
		return nil, apperrors.NotFound("distributor not found")
	}

	profile, _ := s.distRepo.GetBusinessProfile(ctx, distributorID)
	docs, _ := s.distRepo.GetBusinessDocuments(ctx, distributorID)
	app, _ := s.distRepo.GetActiveApplication(ctx, distributorID)

	appID := ""
	if app != nil {
		appID = app.ID
	}

	// Check if agreement already exists or create new
	ag, _ := s.creditRepo.GetAgreementByDistributor(ctx, distributorID)
	agID := ""
	if ag != nil {
		agID = ag.ID
	} else {
		agNumber := fmt.Sprintf("KRESCO-AGR-%d", time.Now().Unix())
		newAg := &repository.AgreementRecord{
			DistributorID:      distributorID,
			ApplicationID:      appID,
			AgreementNumber:    agNumber,
			Version:            "1.0",
			ApprovedLimitPaise: offer.OfferedLimitPaise,
			ApprovedPeriodDays: offer.OfferedPeriodDays,
		}
		id, err := s.creditRepo.CreateAgreement(ctx, newAg)
		if err != nil {
			return nil, apperrors.Internal("creating agreement record", err)
		}
		agID = id
	}

	busName := ""
	constitution := "PROP"
	addr := ""
	cityState := ""
	if profile != nil {
		busName = profile.BusinessName
		constitution = profile.Constitution
		addr = profile.AddressLine1
		cityState = fmt.Sprintf("%s, %s - %s", profile.City, profile.State, profile.PIN)
	}

	pan := ""
	gst := ""
	if docs != nil {
		if docs.PAN != nil {
			pan = *docs.PAN
		}
		if docs.GSTNumber != nil {
			gst = *docs.GSTNumber
		}
	}

	distName := "Authorized Officer"
	if dist.Name != nil {
		distName = *dist.Name
	}
	email := ""
	if dist.Email != nil {
		email = *dist.Email
	}

	pdfData := &AgreementPDFData{
		AgreementID:      agID,
		DistributorName:  distName,
		BusinessName:     busName,
		Constitution:     constitution,
		PAN:              pan,
		GST:              gst,
		Address:          addr,
		CityStatePIN:     cityState,
		Mobile:           dist.Mobile,
		Email:            email,
		CreditLimitPaise: offer.OfferedLimitPaise,
		PaymentTermsDays: offer.OfferedPeriodDays,
		InterestRatePct:  2.0,
		EffectiveDate:    time.Now(),
	}

	genPDF, err := GenerateAgreementPDF(pdfData)
	if err != nil {
		return nil, apperrors.Internal("generating agreement pdf", err)
	}

	defaultRedirect := "http://localhost:8081/api/v1/agreements/esign-callback"
	if redirectURL != "" {
		defaultRedirect = redirectURL
	}

	eSignReq := &svcver.ESignInitRequest{
		FullName:     distName,
		UserEmail:    email,
		MobileNumber: dist.Mobile,
		PageNum:      genPDF.PageNum,
		SignX:        genPDF.SignX,
		SignY:        genPDF.SignY,
		RedirectURL:  defaultRedirect,
	}

	var esignResp *svcver.ESignInitResponse
	if s.esign != nil {
		esignResp, err = s.esign.InitializeESignSession(ctx, eSignReq)
		if err != nil {
			return nil, apperrors.Internal("initializing esign session", err)
		}
	} else {
		fallbackToken := fmt.Sprintf("ESIGN-DEMO-TOKEN-%d", time.Now().Unix())
		esignResp = &svcver.ESignInitResponse{
			Token:    fallbackToken,
			URL:      fmt.Sprintf("https://esign-client.surepass.io/?token=%s", fallbackToken),
			ClientID: "DEMO-CLIENT-ID",
		}
	}

	return &ESignInitResult{
		AgreementID: agID,
		SigningURL:  esignResp.URL,
		Token:       esignResp.Token,
	}, nil
}

// CompleteSigning executes completion of digital signing (e-Sign) for a distributor agreement.
func (s *Service) CompleteSigning(ctx context.Context, distributorID, agreementID, providerRef string) error {
	ag, err := s.creditRepo.GetAgreementByDistributor(ctx, distributorID)
	if err != nil || ag == nil {
		return apperrors.NotFound("agreement not found")
	}

	if ag.Status == "SIGNED" {
		return nil // Already signed idempotently
	}

	ref := providerRef
	if ref == "" {
		ref = fmt.Sprintf("SUREPASS-ESIGN-%d", time.Now().Unix())
	}

	if err := s.creditRepo.SignAgreement(ctx, ag.ID, ref); err != nil {
		return apperrors.Internal("signing agreement", err)
	}

	// Fetch active credit offer limit to initialize credit account
	limitPaise := ag.ApprovedLimitPaise
	if limitPaise <= 0 {
		offer, err := s.creditRepo.GetActiveOfferByDistributor(ctx, distributorID)
		if err == nil && offer != nil {
			limitPaise = offer.OfferedLimitPaise
		}
	}

	if s.orderRepo != nil {
		if _, err := s.orderRepo.GetOrCreateCreditAccount(ctx, distributorID, limitPaise); err != nil {
			return apperrors.Internal("initializing credit account", err)
		}
	}

	// Update active application to 'credit_active'
	if app, err := s.distRepo.GetActiveApplication(ctx, distributorID); err == nil && app != nil {
		reason := "agreement signed digitally via Surepass SureSign"
		_ = s.distRepo.UpdateApplicationStatus(ctx, app.ID, "credit_active", "distributor", &distributorID, &reason)
	}

	return nil
}
