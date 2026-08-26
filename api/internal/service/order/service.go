// Package order manages product catalogue, order placement, credit/advance splits,
// payment proof verification, and dispatch credit guard invariants.
package order

import (
	"context"
	"fmt"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
)

type CreateOrderItemInput struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Service struct {
	orderRepo  *repository.OrderRepository
	creditRepo *repository.CreditRepository
	distRepo   *repository.DistributorRepository
}

func New(orderRepo *repository.OrderRepository, creditRepo *repository.CreditRepository, distRepo *repository.DistributorRepository) *Service {
	return &Service{orderRepo: orderRepo, creditRepo: creditRepo, distRepo: distRepo}
}

func (s *Service) ListCatalogue(ctx context.Context) ([]repository.ProductRecord, error) {
	return s.orderRepo.ListProducts(ctx)
}

func (s *Service) ListSampleCatalogue(ctx context.Context) ([]repository.ProductRecord, error) {
	return s.orderRepo.ListSampleProducts(ctx)
}

func (s *Service) ListAllProductsAdmin(ctx context.Context) ([]repository.ProductRecord, error) {
	return s.orderRepo.ListAllProductsAdmin(ctx)
}

func (s *Service) CreateProduct(ctx context.Context, p *repository.ProductRecord) (string, error) {
	if p.Name == "" || p.SKU == "" {
		return "", apperrors.Validation("name and SKU are required")
	}
	return s.orderRepo.CreateProduct(ctx, p)
}

func (s *Service) UpdateProduct(ctx context.Context, p *repository.ProductRecord) error {
	if p.ID == "" {
		return apperrors.Validation("product ID is required")
	}
	return s.orderRepo.UpdateProduct(ctx, p)
}

func (s *Service) CreateOrder(ctx context.Context, distributorID string, items []CreateOrderItemInput) (*repository.OrderRecord, error) {
	if len(items) == 0 {
		return nil, apperrors.Validation("order must contain at least 1 item")
	}

	var totalPaise int64
	var orderItems []repository.OrderItemRecord

	for _, item := range items {
		p, err := s.orderRepo.GetProductByID(ctx, item.ProductID)
		if err != nil || p == nil {
			return nil, apperrors.NotFound(fmt.Sprintf("product not found: %s", item.ProductID))
		}
		if item.Quantity < p.Moq {
			return nil, apperrors.Validation(fmt.Sprintf("minimum order quantity for %s is %d", p.Name, p.Moq))
		}
		itemTotal := p.PricePaise * int64(item.Quantity)
		totalPaise += itemTotal

		orderItems = append(orderItems, repository.OrderItemRecord{
			ProductID:   p.ID,
			ProductName: p.Name,
			Quantity:    item.Quantity,
			UnitPrice:   p.PricePaise,
			TotalPrice:  itemTotal,
		})
	}

	// Fetch distributor credit account
	offer, _ := s.creditRepo.GetActiveOfferByDistributor(ctx, distributorID)
	approvedLimit := int64(0)
	if offer != nil && offer.Status == "ACCEPTED" {
		approvedLimit = offer.OfferedLimitPaise
	}

	acc, err := s.orderRepo.GetOrCreateCreditAccount(ctx, distributorID, approvedLimit)
	if err != nil {
		return nil, apperrors.Internal("fetching credit account", err)
	}

	// Calculate Credit vs Advance Split:
	// Order value IS NOT capped by credit limit. If order > available credit, excess is paid in advance!
	availableCredit := acc.AvailableCreditPaise
	var creditUsed, advancePaid int64
	status := "PENDING_REVIEW"

	if totalPaise <= availableCredit {
		creditUsed = totalPaise
		advancePaid = 0
		status = "PENDING_REVIEW"
	} else {
		creditUsed = availableCredit
		advancePaid = totalPaise - availableCredit
		status = "PENDING_PAYMENT"
	}

	orderNum := fmt.Sprintf("KCN-ORD-%d", time.Now().Unix())
	orderRec := &repository.OrderRecord{
		OrderNumber:      orderNum,
		DistributorID:    distributorID,
		TotalAmountPaise: totalPaise,
		AdvancePaidPaise: advancePaid,
		CreditUsedPaise:  creditUsed,
		Status:           status,
	}

	orderID, err := s.orderRepo.CreateOrder(ctx, orderRec, orderItems)
	if err != nil {
		return nil, apperrors.Internal("creating order", err)
	}
	orderRec.ID = orderID

	return orderRec, nil
}

func (s *Service) SubmitPaymentProof(ctx context.Context, orderID, proofURL, utr string) error {
	return s.orderRepo.SubmitPaymentProof(ctx, orderID, proofURL, utr)
}

func (s *Service) VerifyPayment(ctx context.Context, orderID, verifiedBy string) error {
	return s.orderRepo.VerifyPayment(ctx, orderID, verifiedBy)
}

func (s *Service) ReviewOrder(ctx context.Context, orderID, action, reviewedBy, notes string) error {
	return s.orderRepo.ReviewOrder(ctx, orderID, action, reviewedBy, notes)
}

func (s *Service) DispatchOrder(ctx context.Context, orderID string) error {
	err := s.orderRepo.DispatchOrder(ctx, orderID)
	if err != nil {
		return apperrors.Conflict(err.Error())
	}
	return nil
}

func (s *Service) ListMyOrders(ctx context.Context, distributorID string) ([]repository.OrderRecord, error) {
	return s.orderRepo.ListOrdersByDistributor(ctx, distributorID)
}

func (s *Service) ListMySampleOrders(ctx context.Context, distributorID string) ([]repository.SampleOrderRecord, error) {
	return s.orderRepo.ListSampleOrdersByDistributor(ctx, distributorID)
}

func (s *Service) ListPendingReviews(ctx context.Context) ([]repository.OrderRecord, error) {
	return s.orderRepo.ListOrdersForReview(ctx)
}

func (s *Service) ListAllCatalogOrders(ctx context.Context, limit, offset int) ([]repository.OrderRecord, int, error) {
	return s.orderRepo.ListAllCatalogOrdersAdmin(ctx, limit, offset)
}

func (s *Service) ListAllSampleOrders(ctx context.Context, limit, offset int) ([]repository.SampleOrderRecord, int, error) {
	return s.orderRepo.ListSampleOrdersAdmin(ctx, limit, offset)
}

