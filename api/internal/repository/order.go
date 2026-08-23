package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRecord struct {
	ID          string    `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Category    string    `json:"category"`
	PricePaise  int64     `json:"price_paise"`
	Moq         int       `json:"moq"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrderRecord struct {
	ID                string     `json:"id"`
	OrderNumber       string     `json:"order_number"`
	DistributorID     string     `json:"distributor_id"`
	TotalAmountPaise  int64      `json:"total_amount_paise"`
	AdvancePaidPaise  int64      `json:"advance_paid_paise"`
	CreditUsedPaise   int64      `json:"credit_used_paise"`
	Status            string     `json:"status"` // PENDING_PAYMENT, PAYMENT_SUBMITTED, PAYMENT_VERIFIED, PENDING_REVIEW, APPROVED, REJECTED, DISPATCHED, CANCELLED
	PaymentProofURL   *string    `json:"payment_proof_url"`
	UTRReference      *string    `json:"utr_reference"`
	ReviewedBy        *string    `json:"reviewed_by"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	ReviewNotes       *string    `json:"review_notes"`
	DispatchedAt      *time.Time `json:"dispatched_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

type OrderItemRecord struct {
	ID           string `json:"id"`
	OrderID      string `json:"order_id"`
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	Quantity     int    `json:"quantity"`
	UnitPrice    int64  `json:"unit_price_paise"`
	TotalPrice   int64  `json:"total_price_paise"`
}

type CreditAccountRecord struct {
	ID                  string    `json:"id"`
	DistributorID       string    `json:"distributor_id"`
	ApprovedLimitPaise  int64     `json:"approved_limit_paise"`
	CurrentCreditPaise  int64     `json:"current_credit_paise"`
	AvailableCreditPaise int64    `json:"available_credit_paise"`
	Status              string    `json:"status"` // ACTIVE, RESTRICTED, HOLD, BLOCKED
	UpdatedAt           time.Time `json:"updated_at"`
}

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// ──────────────────────────────── Catalogue ───────────────────────────────────

func (r *OrderRepository) ListProducts(ctx context.Context) ([]ProductRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, sku, name, description, category, price_paise, moq, is_active, created_at
		 FROM products WHERE is_active = TRUE ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProductRecord
	for rows.Next() {
		p := ProductRecord{}
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.PricePaise, &p.Moq, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *OrderRepository) GetProductByID(ctx context.Context, id string) (*ProductRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, sku, name, description, category, price_paise, moq, is_active, created_at
		 FROM products WHERE id = $1`, id)
	p := &ProductRecord{}
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.PricePaise, &p.Moq, &p.IsActive, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

// ──────────────────────────────── Credit Account ─────────────────────────────

func (r *OrderRepository) GetCreditAccount(ctx context.Context, distributorID string) (*CreditAccountRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, approved_limit_paise, current_credit_paise, available_credit_paise, status, updated_at
		 FROM credit_accounts WHERE distributor_id = $1`, distributorID)
	c := &CreditAccountRecord{}
	err := row.Scan(&c.ID, &c.DistributorID, &c.ApprovedLimitPaise, &c.CurrentCreditPaise, &c.AvailableCreditPaise, &c.Status, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *OrderRepository) GetOrCreateCreditAccount(ctx context.Context, distributorID string, defaultLimit int64) (*CreditAccountRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, approved_limit_paise, current_credit_paise, available_credit_paise, status, updated_at
		 FROM credit_accounts WHERE distributor_id = $1`, distributorID)
	c := &CreditAccountRecord{}
	err := row.Scan(&c.ID, &c.DistributorID, &c.ApprovedLimitPaise, &c.CurrentCreditPaise, &c.AvailableCreditPaise, &c.Status, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// Initialize credit account from latest approved decision
		var accID string
		err := r.db.QueryRow(ctx,
			`INSERT INTO credit_accounts (distributor_id, approved_limit_paise, available_credit_paise)
			 VALUES ($1, $2, $2) RETURNING id`, distributorID, defaultLimit,
		).Scan(&accID)
		if err != nil {
			return nil, err
		}
		return &CreditAccountRecord{
			ID:                   accID,
			DistributorID:        distributorID,
			ApprovedLimitPaise:   defaultLimit,
			CurrentCreditPaise:   0,
			AvailableCreditPaise: defaultLimit,
			Status:               "ACTIVE",
			UpdatedAt:            time.Now(),
		}, nil
	}
	return c, err
}

// ──────────────────────────────── Orders ─────────────────────────────────────

func (r *OrderRepository) CreateOrder(ctx context.Context, o *OrderRecord, items []OrderItemRecord) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var orderID string
	err = tx.QueryRow(ctx,
		`INSERT INTO orders
		 (order_number, distributor_id, total_amount_paise, advance_paid_paise, credit_used_paise, status)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		o.OrderNumber, o.DistributorID, o.TotalAmountPaise, o.AdvancePaidPaise, o.CreditUsedPaise, o.Status,
	).Scan(&orderID)
	if err != nil {
		return "", err
	}

	for _, item := range items {
		_, err := tx.Exec(ctx,
			`INSERT INTO order_items
			 (order_id, product_id, product_name, quantity, unit_price_paise, total_price_paise)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			orderID, item.ProductID, item.ProductName, item.Quantity, item.UnitPrice, item.TotalPrice,
		)
		if err != nil {
			return "", err
		}
	}

	return orderID, tx.Commit(ctx)
}

func (r *OrderRepository) SubmitPaymentProof(ctx context.Context, orderID, proofURL, utr string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders
		 SET payment_proof_url = $1, utr_reference = $2, status = 'PAYMENT_SUBMITTED', updated_at = NOW()
		 WHERE id = $3`,
		proofURL, utr, orderID)
	return err
}

func (r *OrderRepository) VerifyPayment(ctx context.Context, orderID, verifiedBy string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders
		 SET status = 'PENDING_REVIEW', reviewed_by = $1, reviewed_at = NOW(), updated_at = NOW()
		 WHERE id = $2`,
		verifiedBy, orderID)
	return err
}

func (r *OrderRepository) ReviewOrder(ctx context.Context, orderID, action, reviewedBy, notes string) error {
	status := "APPROVED"
	if action == "REJECT" {
		status = "REJECTED"
	}
	_, err := r.db.Exec(ctx,
		`UPDATE orders
		 SET status = $1, reviewed_by = $2, review_notes = $3, reviewed_at = NOW(), updated_at = NOW()
		 WHERE id = $4`,
		status, reviewedBy, notes, orderID)
	return err
}

func (r *OrderRepository) DispatchOrder(ctx context.Context, orderID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch order
	var o OrderRecord
	err = tx.QueryRow(ctx,
		`SELECT id, distributor_id, total_amount_paise, credit_used_paise, status
		 FROM orders WHERE id = $1 FOR UPDATE`, orderID,
	).Scan(&o.ID, &o.DistributorID, &o.TotalAmountPaise, &o.CreditUsedPaise, &o.Status)
	if err != nil {
		return err
	}

	// Dispatch Credit Guard Invariant: Available Credit = Approved Limit - Outstanding
	var acc CreditAccountRecord
	err = tx.QueryRow(ctx,
		`SELECT id, approved_limit_paise, current_credit_paise, available_credit_paise, status
		 FROM credit_accounts WHERE distributor_id = $1 FOR UPDATE`, o.DistributorID,
	).Scan(&acc.ID, &acc.ApprovedLimitPaise, &acc.CurrentCreditPaise, &acc.AvailableCreditPaise, &acc.Status)
	if err != nil {
		return err
	}

	if acc.Status == "HOLD" || acc.Status == "BLOCKED" {
		return errors.New("cannot dispatch: distributor credit account is placed on HOLD or BLOCKED")
	}

	if o.CreditUsedPaise > acc.AvailableCreditPaise {
		return errors.New("dispatch credit guard violation: order credit amount exceeds available credit limit")
	}

	// Consume credit
	newCurrent := acc.CurrentCreditPaise + o.CreditUsedPaise
	newAvailable := acc.ApprovedLimitPaise - newCurrent

	_, err = tx.Exec(ctx,
		`UPDATE credit_accounts
		 SET current_credit_paise = $1, available_credit_paise = $2, updated_at = NOW()
		 WHERE id = $3`,
		newCurrent, newAvailable, acc.ID,
	)
	if err != nil {
		return err
	}

	// Update order status
	_, err = tx.Exec(ctx,
		`UPDATE orders SET status = 'DISPATCHED', dispatched_at = NOW(), updated_at = NOW() WHERE id = $1`,
		orderID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) ListOrdersByDistributor(ctx context.Context, distributorID string) ([]OrderRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, order_number, distributor_id, total_amount_paise, advance_paid_paise, credit_used_paise,
		        status, payment_proof_url, utr_reference, reviewed_by, reviewed_at, review_notes, dispatched_at, created_at
		 FROM orders WHERE distributor_id = $1 ORDER BY created_at DESC`, distributorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []OrderRecord
	for rows.Next() {
		o := OrderRecord{}
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.DistributorID, &o.TotalAmountPaise, &o.AdvancePaidPaise,
			&o.CreditUsedPaise, &o.Status, &o.PaymentProofURL, &o.UTRReference, &o.ReviewedBy, &o.ReviewedAt, &o.ReviewNotes, &o.DispatchedAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, nil
}

func (r *OrderRepository) ListOrdersForReview(ctx context.Context) ([]OrderRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, order_number, distributor_id, total_amount_paise, advance_paid_paise, credit_used_paise,
		        status, payment_proof_url, utr_reference, reviewed_by, reviewed_at, review_notes, dispatched_at, created_at
		 FROM orders WHERE status IN ('PAYMENT_SUBMITTED', 'PENDING_REVIEW') ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []OrderRecord
	for rows.Next() {
		o := OrderRecord{}
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.DistributorID, &o.TotalAmountPaise, &o.AdvancePaidPaise,
			&o.CreditUsedPaise, &o.Status, &o.PaymentProofURL, &o.UTRReference, &o.ReviewedBy, &o.ReviewedAt, &o.ReviewNotes, &o.DispatchedAt, &o.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, nil
}
