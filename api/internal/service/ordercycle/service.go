package ordercycle

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	svcagr "github.com/arryaanjain/DistributorApprovalSystem/internal/service/agreement"
	svcrzp "github.com/arryaanjain/DistributorApprovalSystem/internal/service/razorpay"
)

type CreateCycleInput struct {
	OrderID          string `json:"order_id"`
	Frequency        string `json:"frequency"`        // "weekly" | "monthly"
	TotalInstalments int    `json:"total_instalments"` // <= 5 for weekly, <= 2 for monthly
	RedirectURL      string `json:"redirect_url"`
}

type CreateCycleOutput struct {
	Cycle       *repository.OrderCreditCycleRecord `json:"cycle"`
	AgreementID string                             `json:"agreement_id"`
	SigningURL  string                             `json:"signing_url"`
}

type MandateOutput struct {
	Mandate      *repository.UPIMandateRecord      `json:"mandate"`
	ShortURL     string                            `json:"short_url"`
	Subscription string                            `json:"subscription_id"`
	Repayments   []*repository.CycleRepaymentRecord `json:"repayments"`
}

type Service struct {
	cycleRepo  *repository.CreditRepository
	repo       *repository.CreditCycleRepository
	orderRepo  *repository.OrderRepository
	distRepo   *repository.DistributorRepository
	agrService *svcagr.Service
	rzpClient  *svcrzp.Client
}

func NewService(
	cycleRepo *repository.CreditRepository,
	repo *repository.CreditCycleRepository,
	orderRepo *repository.OrderRepository,
	distRepo *repository.DistributorRepository,
	agrService *svcagr.Service,
	rzpClient *svcrzp.Client,
) *Service {
	return &Service{
		cycleRepo:  cycleRepo,
		repo:       repo,
		orderRepo:  orderRepo,
		distRepo:   distRepo,
		agrService: agrService,
		rzpClient:  rzpClient,
	}
}

// CreateCycle initializes an Order Credit Cycle, validates limits, and generates Surepass eSign link
func (s *Service) CreateCycle(ctx context.Context, distributorID string, input CreateCycleInput) (*CreateCycleOutput, error) {
	// 1. Validate frequency & instalments
	if input.Frequency != "weekly" && input.Frequency != "monthly" {
		return nil, apperrors.Validation("frequency must be 'weekly' or 'monthly'")
	}
	if input.Frequency == "weekly" && (input.TotalInstalments < 1 || input.TotalInstalments > 5) {
		return nil, apperrors.Validation("weekly instalments must be between 1 and 5")
	}
	if input.Frequency == "monthly" && (input.TotalInstalments < 1 || input.TotalInstalments > 2) {
		return nil, apperrors.Validation("monthly instalments must be between 1 and 2")
	}

	// 2. Fetch order
	order, err := s.orderRepo.GetOrderByID(ctx, input.OrderID)
	if err != nil || order == nil {
		return nil, apperrors.NotFound("order not found")
	}
	if order.DistributorID != distributorID {
		return nil, apperrors.Forbidden("order does not belong to distributor")
	}

	// 3. Check existing cycle for this order
	existingCycle, _ := s.repo.GetCycleByOrderID(ctx, input.OrderID)
	if existingCycle != nil {
		// If cycle already exists, return current cycle details & new eSign link if pending
		esignRes, err := s.agrService.InitOrderCreditESign(
			ctx, distributorID, order.ID, order.OrderNumber,
			existingCycle.TotalAmountPaise, existingCycle.AdvancePaidPaise, existingCycle.CreditAmountPaise,
			existingCycle.Frequency, existingCycle.TotalInstalments, existingCycle.InstalmentAmountPaise,
			existingCycle.GracePeriodDays, input.RedirectURL,
		)
		if err != nil {
			return nil, err
		}
		return &CreateCycleOutput{
			Cycle:       existingCycle,
			AgreementID: esignRes.AgreementID,
			SigningURL:  esignRes.SigningURL,
		}, nil
	}

	// 4. Validate Credit Account and Available Limit
	acc, err := s.orderRepo.GetCreditAccount(ctx, distributorID)
	if err != nil || acc == nil {
		return nil, apperrors.Validation("distributor has no active credit account")
	}
	if acc.Status != "active" && acc.Status != "ACTIVE" {
		return nil, apperrors.Validation(fmt.Sprintf("credit account is currently on %s status", acc.Status))
	}

	creditAmountPaise := order.CreditUsedPaise
	if creditAmountPaise <= 0 {
		creditAmountPaise = order.TotalAmountPaise - order.AdvancePaidPaise
	}
	if creditAmountPaise <= 0 && order.TotalAmountPaise > 0 {
		creditAmountPaise = order.TotalAmountPaise
	}

	if creditAmountPaise <= 0 {
		return nil, apperrors.Validation("order has no credit portion to finance")
	}

	availableLimit := acc.AvailableCreditPaise
	if availableLimit <= 0 && acc.ApprovedLimitPaise > 0 {
		availableLimit = acc.ApprovedLimitPaise
	}
	if availableLimit <= 0 {
		availableLimit = 50000000
	}

	if creditAmountPaise > availableLimit {
		return nil, apperrors.Validation(fmt.Sprintf("insufficient available credit limit. Required: ₹%.2f, Available: ₹%.2f", float64(creditAmountPaise)/100.0, float64(availableLimit)/100.0))
	}

	// 5. Calculate instalment breakdown
	instalmentAmountPaise := int64(math.Ceil(float64(creditAmountPaise) / float64(input.TotalInstalments)))

	// 6. Save order credit cycle
	cycle := &repository.OrderCreditCycleRecord{
		OrderID:               order.ID,
		DistributorID:         distributorID,
		AccountID:             &acc.ID,
		TotalAmountPaise:      order.TotalAmountPaise,
		AdvancePaidPaise:      order.AdvancePaidPaise,
		CreditAmountPaise:     creditAmountPaise,
		InstalmentAmountPaise: instalmentAmountPaise,
		Frequency:             input.Frequency,
		TotalInstalments:      input.TotalInstalments,
		CompletedInstalments:  0,
		Status:                "esign_pending",
		GracePeriodDays:       7,
	}

	if err := s.repo.CreateCycle(ctx, cycle); err != nil {
		return nil, apperrors.Internal("failed to create order credit cycle", err)
	}

	// 7. Initiate Surepass eSign
	esignRes, err := s.agrService.InitOrderCreditESign(
		ctx, distributorID, order.ID, order.OrderNumber,
		cycle.TotalAmountPaise, cycle.AdvancePaidPaise, cycle.CreditAmountPaise,
		cycle.Frequency, cycle.TotalInstalments, cycle.InstalmentAmountPaise,
		cycle.GracePeriodDays, input.RedirectURL,
	)
	if err != nil {
		return nil, fmt.Errorf("initiating esign: %w", err)
	}

	_ = s.repo.UpdateCycleAgreement(ctx, cycle.ID, esignRes.AgreementID)
	cycle.AgreementID = &esignRes.AgreementID

	return &CreateCycleOutput{
		Cycle:       cycle,
		AgreementID: esignRes.AgreementID,
		SigningURL:  esignRes.SigningURL,
	}, nil
}

// CompleteESign handles completion callback from Surepass eSign
func (s *Service) CompleteESign(ctx context.Context, orderID string) (*MandateOutput, error) {
	cycle, err := s.repo.GetCycleByOrderID(ctx, orderID)
	if err != nil || cycle == nil {
		return nil, apperrors.NotFound("order credit cycle not found")
	}

	if cycle.Status == "draft" || cycle.Status == "esign_pending" {
		if err := s.repo.UpdateCycleStatus(ctx, cycle.ID, "esign_completed"); err != nil {
			return nil, apperrors.Internal("updating cycle status", err)
		}
		cycle.Status = "esign_completed"
	}

	// Automatically set up UPI Mandate via Razorpay Subscriptions
	return s.SetupMandate(ctx, orderID)
}

// SetupMandate creates Razorpay Plan + Subscription and populates repayment schedule
func (s *Service) SetupMandate(ctx context.Context, orderID string) (*MandateOutput, error) {
	cycle, err := s.repo.GetCycleByOrderID(ctx, orderID)
	if err != nil || cycle == nil {
		return nil, apperrors.NotFound("order credit cycle not found")
	}

	// Check if mandate already created
	existingMandate, _ := s.repo.GetMandateByCycleID(ctx, cycle.ID)
	if existingMandate != nil {
		repayments, _ := s.repo.ListRepaymentsByCycleID(ctx, cycle.ID)
		shortURL := ""
		if existingMandate.RazorpayShortURL != nil {
			shortURL = *existingMandate.RazorpayShortURL
		}
		return &MandateOutput{
			Mandate:      existingMandate,
			ShortURL:     shortURL,
			Subscription: existingMandate.RazorpaySubscriptionID,
			Repayments:   repayments,
		}, nil
	}

	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil || order == nil {
		return nil, apperrors.NotFound("order not found")
	}

	// 1. Create Razorpay Plan
	planReq := svcrzp.CreatePlanRequest{
		Period:   cycle.Frequency,
		Interval: 1,
		Item: svcrzp.CreatePlanRequestItem{
			Name:        fmt.Sprintf("Order Credit Repayment - %s", order.OrderNumber),
			Amount:      cycle.InstalmentAmountPaise,
			Currency:    "INR",
			Description: fmt.Sprintf("%s instalment for order %s", cycle.Frequency, order.OrderNumber),
		},
	}
	plan, err := s.rzpClient.CreatePlan(ctx, planReq)
	if err != nil {
		return nil, apperrors.Internal("failed to create razorpay plan", err)
	}

	// 2. Create Razorpay Subscription (UPI Mandate)
	subReq := svcrzp.CreateSubscriptionRequest{
		PlanID:         plan.ID,
		TotalCount:     cycle.TotalInstalments,
		Quantity:       1,
		CustomerNotify: 1,
		Notes: map[string]string{
			"order_id":       order.ID,
			"cycle_id":       cycle.ID,
			"distributor_id": cycle.DistributorID,
			"order_number":   order.OrderNumber,
		},
	}
	sub, err := s.rzpClient.CreateSubscription(ctx, subReq)
	if err != nil {
		return nil, apperrors.Internal("failed to create razorpay subscription", err)
	}

	// 3. Save UPI Mandate Record
	shortURL := sub.ShortURL
	mandate := &repository.UPIMandateRecord{
		OrderID:                  order.ID,
		CycleID:                  cycle.ID,
		DistributorID:            cycle.DistributorID,
		RazorpayPlanID:           plan.ID,
		RazorpaySubscriptionID:   sub.ID,
		RazorpayShortURL:         &shortURL,
		MandateStatus:            "created",
		Frequency:                cycle.Frequency,
		TotalCount:               cycle.TotalInstalments,
		PaidCount:                0,
		AmountPerInstalmentPaise: cycle.InstalmentAmountPaise,
	}

	if err := s.repo.CreateMandate(ctx, mandate); err != nil {
		return nil, apperrors.Internal("failed to save mandate record", err)
	}

	// 4. Generate scheduled repayments if not already present
	repayments, err := s.repo.ListRepaymentsByCycleID(ctx, cycle.ID)
	if err != nil || len(repayments) == 0 {
		repayments = nil
		now := time.Now()
		for i := 1; i <= cycle.TotalInstalments; i++ {
			var dueDate time.Time
			if cycle.Frequency == "weekly" {
				dueDate = now.AddDate(0, 0, 7*i)
			} else {
				dueDate = now.AddDate(0, i, 0)
			}

			repayments = append(repayments, &repository.CycleRepaymentRecord{
				CycleID:          cycle.ID,
				MandateID:        &mandate.ID,
				InstalmentNumber: i,
				DueDate:          dueDate,
				AmountPaise:      cycle.InstalmentAmountPaise,
				Status:           "scheduled",
			})
		}

		if err := s.repo.CreateRepayments(ctx, repayments); err != nil {
			return nil, apperrors.Internal("failed to save scheduled repayments", err)
		}
	}

	// Update cycle status to mandate_pending
	_ = s.repo.UpdateCycleStatus(ctx, cycle.ID, "mandate_pending")

	return &MandateOutput{
		Mandate:      mandate,
		ShortURL:     shortURL,
		Subscription: sub.ID,
		Repayments:   repayments,
	}, nil
}

// ProcessMandateWebhook handles incoming Razorpay webhook events
func (s *Service) ProcessMandateWebhook(ctx context.Context, eventType string, payload map[string]interface{}) error {
	subObj, ok := payload["subscription"].(map[string]interface{})
	if !ok {
		return nil // Not a subscription event
	}

	subID, _ := subObj["id"].(string)
	if subID == "" {
		return nil
	}

	mandate, err := s.repo.GetMandateBySubscriptionID(ctx, subID)
	if err != nil || mandate == nil {
		return nil // Mandate not found in our DB
	}

	cycle, err := s.repo.GetCycleByID(ctx, mandate.CycleID)
	if err != nil || cycle == nil {
		return nil
	}

	switch eventType {
	case "subscription.authenticated", "subscription.activated":
		_ = s.repo.UpdateMandateStatus(ctx, subID, "active", mandate.PaidCount)
		_ = s.repo.UpdateCycleStatus(ctx, cycle.ID, "repaying")

	case "subscription.charged":
		paymentObj, _ := payload["payment"].(map[string]interface{})
		paymentID := ""
		if paymentObj != nil {
			paymentID, _ = paymentObj["id"].(string)
		}

		paidCountFloat, _ := subObj["paid_count"].(float64)
		paidCount := int(paidCountFloat)
		if paidCount == 0 {
			paidCount = mandate.PaidCount + 1
		}

		_ = s.repo.UpdateMandateStatus(ctx, subID, "active", paidCount)

		// Mark matching scheduled repayment as paid
		repayments, _ := s.repo.ListRepaymentsByCycleID(ctx, cycle.ID)
		for _, rep := range repayments {
			if rep.Status == "scheduled" || rep.Status == "pending" || rep.Status == "failed" {
				pID := paymentID
				_ = s.repo.UpdateRepaymentStatus(ctx, rep.ID, "paid", &pID, nil)
				break
			}
		}

		_ = s.repo.IncrementCompletedInstalments(ctx, cycle.ID)

	case "subscription.halted", "subscription.pending":
		_ = s.repo.UpdateMandateStatus(ctx, subID, "failed", mandate.PaidCount)
		
		// Mark next scheduled repayment as failed
		repayments, _ := s.repo.ListRepaymentsByCycleID(ctx, cycle.ID)
		for _, rep := range repayments {
			if rep.Status == "scheduled" || rep.Status == "pending" {
				failReason := fmt.Sprintf("Razorpay Mandate Charge Failed (%s)", eventType)
				_ = s.repo.UpdateRepaymentStatus(ctx, rep.ID, "failed", nil, &failReason)
				break
			}
		}

		// Trigger Grace Period of 7 days
		graceEnd := time.Now().AddDate(0, 0, 7)
		_ = s.repo.SetCycleGracePeriod(ctx, cycle.ID, graceEnd)
	}

	return nil
}

// GetCycleDetails returns complete details of a cycle, mandate, and repayment tracker
func (s *Service) GetCycleDetails(ctx context.Context, distributorID, orderID string) (map[string]interface{}, error) {
	cycle, err := s.repo.GetCycleByOrderID(ctx, orderID)
	if err != nil || cycle == nil {
		return nil, apperrors.NotFound("no credit cycle found for this order")
	}

	if cycle.DistributorID != distributorID {
		return nil, apperrors.Forbidden("access denied")
	}

	mandate, _ := s.repo.GetMandateByCycleID(ctx, cycle.ID)
	repayments, _ := s.repo.ListRepaymentsByCycleID(ctx, cycle.ID)

	isGraceActive := false
	graceRemainingDays := 0
	if cycle.GracePeriodEndsAt != nil && time.Now().Before(*cycle.GracePeriodEndsAt) {
		isGraceActive = true
		graceRemainingDays = int(time.Until(*cycle.GracePeriodEndsAt).Hours() / 24)
		if graceRemainingDays < 1 {
			graceRemainingDays = 1
		}
	}

	return map[string]interface{}{
		"cycle":                cycle,
		"mandate":              mandate,
		"repayments":           repayments,
		"is_grace_active":      isGraceActive,
		"grace_remaining_days": graceRemainingDays,
	}, nil
}

// GetDistributorCycles lists all order credit cycles for a distributor
func (s *Service) GetDistributorCycles(ctx context.Context, distributorID string) ([]*repository.OrderCreditCycleRecord, error) {
	return s.repo.ListCyclesByDistributor(ctx, distributorID)
}
