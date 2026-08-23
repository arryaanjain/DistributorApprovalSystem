// Package financial handles invoice generation, payment allocation,
// collections escalation, overdue tracking, and credit ladder progression.
package financial

import (
	"context"
	"fmt"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

type Service struct {
	finRepo   *repository.FinancialRepository
	orderRepo *repository.OrderRepository
}

func New(finRepo *repository.FinancialRepository, orderRepo *repository.OrderRepository) *Service {
	return &Service{finRepo: finRepo, orderRepo: orderRepo}
}

func (s *Service) CreateInvoiceForOrder(ctx context.Context, order *repository.OrderRecord) (*repository.InvoiceRecord, error) {
	acc, err := s.orderRepo.GetOrCreateCreditAccount(ctx, order.DistributorID, 0)
	if err != nil {
		return nil, apperrors.Internal("fetching credit account for invoice", err)
	}

	invNum := fmt.Sprintf("KCN-INV-%d", time.Now().Unix())
	dueDate := time.Now().AddDate(0, 0, 15) // Default 15-day term

	inv := &repository.InvoiceRecord{
		InvoiceNumber:    invNum,
		DistributorID:    order.DistributorID,
		AccountID:        acc.ID,
		OrderID:          order.ID,
		InvoiceDate:      time.Now(),
		DueDate:          dueDate,
		TotalPaise:       order.TotalAmountPaise,
		CreditPaise:      order.CreditUsedPaise,
		AdvancePaise:     order.AdvancePaidPaise,
		OutstandingPaise: order.CreditUsedPaise,
		Status:           "open",
		OverdueTier:      "current",
	}

	invID, err := s.finRepo.CreateInvoice(ctx, inv)
	if err != nil {
		return nil, apperrors.Internal("creating invoice record", err)
	}
	inv.ID = invID

	return inv, nil
}

func (s *Service) ListDistributorInvoices(ctx context.Context, distributorID string) ([]repository.InvoiceRecord, error) {
	return s.finRepo.ListInvoicesByDistributor(ctx, distributorID)
}

func (s *Service) RecordPayment(ctx context.Context, invoiceID, distributorID, paymentMode, utr string, amountPaise int64, recordedBy *string) error {
	if amountPaise <= 0 {
		return apperrors.Validation("payment amount must be greater than 0")
	}

	err := s.finRepo.RecordInvoicePayment(ctx, invoiceID, distributorID, paymentMode, utr, amountPaise, recordedBy)
	if err != nil {
		return apperrors.Internal("recording invoice payment", err)
	}

	// Trigger credit ladder evaluation post payment
	_, _ = s.EvaluateCreditEnhancement(ctx, distributorID)

	return nil
}

func (s *Service) EvaluateOverdueCollections(ctx context.Context) ([]string, error) {
	return s.finRepo.EvaluateOverdueTiers(ctx)
}

func (s *Service) EvaluateCreditEnhancement(ctx context.Context, distributorID string) (int64, error) {
	invoices, err := s.finRepo.ListInvoicesByDistributor(ctx, distributorID)
	if err != nil {
		return 0, err
	}

	// Count fully paid invoices with 0 overdue days
	paidCleanInvoices := 0
	for _, inv := range invoices {
		if inv.Status == "paid" && inv.DaysOutstanding <= 0 {
			paidCleanInvoices++
		}
	}

	// Rule: Every 3 clean paid invoices unlocks the next tier on the Credit Ladder
	if paidCleanInvoices > 0 && paidCleanInvoices%3 == 0 {
		acc, err := s.orderRepo.GetOrCreateCreditAccount(ctx, distributorID, 0)
		if err != nil {
			return 0, err
		}

		nextLimit, canEnhance := s.finRepo.GetNextLadderStep(acc.ApprovedLimitPaise)
		if canEnhance {
			if err := s.finRepo.EnhanceCreditLimit(ctx, distributorID, nextLimit); err != nil {
				return 0, err
			}
			return nextLimit, nil
		}
	}

	return 0, nil
}
