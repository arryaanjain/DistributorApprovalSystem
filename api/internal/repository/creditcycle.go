package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderCreditCycleRecord struct {
	ID                    string     `json:"id"`
	OrderID               string     `json:"order_id"`
	DistributorID         string     `json:"distributor_id"`
	AccountID             *string    `json:"account_id,omitempty"`
	AgreementID           *string    `json:"agreement_id,omitempty"`
	TotalAmountPaise      int64      `json:"total_amount_paise"`
	AdvancePaidPaise      int64      `json:"advance_paid_paise"`
	CreditAmountPaise     int64      `json:"credit_amount_paise"`
	InstalmentAmountPaise int64      `json:"instalment_amount_paise"`
	Frequency             string     `json:"frequency"` // 'weekly' | 'monthly'
	TotalInstalments      int        `json:"total_instalments"`
	CompletedInstalments  int        `json:"completed_instalments"`
	Status                string     `json:"status"` // 'draft', 'esign_pending', 'esign_completed', 'mandate_pending', 'mandate_active', 'repaying', 'completed', 'defaulted'
	GracePeriodDays       int        `json:"grace_period_days"`
	GracePeriodEndsAt     *time.Time `json:"grace_period_ends_at,omitempty"`
	HoldTriggeredAt       *time.Time `json:"hold_triggered_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type UPIMandateRecord struct {
	ID                        string     `json:"id"`
	OrderID                   string     `json:"order_id"`
	CycleID                   string     `json:"cycle_id"`
	DistributorID             string     `json:"distributor_id"`
	RazorpayPlanID            string     `json:"razorpay_plan_id"`
	RazorpaySubscriptionID    string     `json:"razorpay_subscription_id"`
	RazorpayShortURL          *string    `json:"razorpay_short_url,omitempty"`
	MandateStatus             string     `json:"mandate_status"` // 'created', 'authenticated', 'active', 'paused', 'failed', 'cancelled', 'completed'
	Frequency                 string     `json:"frequency"`
	TotalCount                int        `json:"total_count"`
	PaidCount                 int        `json:"paid_count"`
	AmountPerInstalmentPaise int64      `json:"amount_per_instalment_paise"`
	StartAt                   *time.Time `json:"start_at,omitempty"`
	NextChargeAt              *time.Time `json:"next_charge_at,omitempty"`
	AuthenticatedAt           *time.Time `json:"authenticated_at,omitempty"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

type CycleRepaymentRecord struct {
	ID                string     `json:"id"`
	CycleID           string     `json:"cycle_id"`
	MandateID         *string    `json:"mandate_id,omitempty"`
	InstalmentNumber  int        `json:"instalment_number"`
	DueDate           time.Time  `json:"due_date"`
	AmountPaise       int64      `json:"amount_paise"`
	Status            string     `json:"status"` // 'scheduled', 'pending', 'paid', 'failed', 'bounced'
	RazorpayPaymentID *string    `json:"razorpay_payment_id,omitempty"`
	RazorpayInvoiceID *string    `json:"razorpay_invoice_id,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	FailureReason     *string    `json:"failure_reason,omitempty"`
	RetryCount        int        `json:"retry_count"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreditCycleRepository struct {
	db *pgxpool.Pool
}

func NewCreditCycleRepository(db *pgxpool.Pool) *CreditCycleRepository {
	return &CreditCycleRepository{db: db}
}

// CreateCycle inserts a new order_credit_cycle record
func (r *CreditCycleRepository) CreateCycle(ctx context.Context, cycle *OrderCreditCycleRecord) error {
	query := `
		INSERT INTO order_credit_cycles (
			order_id, distributor_id, account_id, agreement_id,
			total_amount_paise, advance_paid_paise, credit_amount_paise,
			instalment_amount_paise, frequency, total_instalments,
			completed_instalments, status, grace_period_days
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		cycle.OrderID, cycle.DistributorID, cycle.AccountID, cycle.AgreementID,
		cycle.TotalAmountPaise, cycle.AdvancePaidPaise, cycle.CreditAmountPaise,
		cycle.InstalmentAmountPaise, cycle.Frequency, cycle.TotalInstalments,
		cycle.CompletedInstalments, cycle.Status, cycle.GracePeriodDays,
	).Scan(&cycle.ID, &cycle.CreatedAt, &cycle.UpdatedAt)
}

// GetCycleByID fetches a cycle by ID
func (r *CreditCycleRepository) GetCycleByID(ctx context.Context, id string) (*OrderCreditCycleRecord, error) {
	query := `
		SELECT id, order_id, distributor_id, account_id, agreement_id,
		       total_amount_paise, advance_paid_paise, credit_amount_paise,
		       instalment_amount_paise, frequency, total_instalments,
		       completed_instalments, status, grace_period_days,
		       grace_period_ends_at, hold_triggered_at, created_at, updated_at
		FROM order_credit_cycles
		WHERE id = $1
	`
	c := &OrderCreditCycleRecord{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.OrderID, &c.DistributorID, &c.AccountID, &c.AgreementID,
		&c.TotalAmountPaise, &c.AdvancePaidPaise, &c.CreditAmountPaise,
		&c.InstalmentAmountPaise, &c.Frequency, &c.TotalInstalments,
		&c.CompletedInstalments, &c.Status, &c.GracePeriodDays,
		&c.GracePeriodEndsAt, &c.HoldTriggeredAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting cycle by id: %w", err)
	}
	return c, nil
}

// GetCycleByOrderID fetches a cycle by order_id
func (r *CreditCycleRepository) GetCycleByOrderID(ctx context.Context, orderID string) (*OrderCreditCycleRecord, error) {
	query := `
		SELECT id, order_id, distributor_id, account_id, agreement_id,
		       total_amount_paise, advance_paid_paise, credit_amount_paise,
		       instalment_amount_paise, frequency, total_instalments,
		       completed_instalments, status, grace_period_days,
		       grace_period_ends_at, hold_triggered_at, created_at, updated_at
		FROM order_credit_cycles
		WHERE order_id = $1
	`
	c := &OrderCreditCycleRecord{}
	err := r.db.QueryRow(ctx, query, orderID).Scan(
		&c.ID, &c.OrderID, &c.DistributorID, &c.AccountID, &c.AgreementID,
		&c.TotalAmountPaise, &c.AdvancePaidPaise, &c.CreditAmountPaise,
		&c.InstalmentAmountPaise, &c.Frequency, &c.TotalInstalments,
		&c.CompletedInstalments, &c.Status, &c.GracePeriodDays,
		&c.GracePeriodEndsAt, &c.HoldTriggeredAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting cycle by order id: %w", err)
	}
	return c, nil
}

// ListCyclesByDistributor fetches cycles for a distributor
func (r *CreditCycleRepository) ListCyclesByDistributor(ctx context.Context, distributorID string) ([]*OrderCreditCycleRecord, error) {
	query := `
		SELECT id, order_id, distributor_id, account_id, agreement_id,
		       total_amount_paise, advance_paid_paise, credit_amount_paise,
		       instalment_amount_paise, frequency, total_instalments,
		       completed_instalments, status, grace_period_days,
		       grace_period_ends_at, hold_triggered_at, created_at, updated_at
		FROM order_credit_cycles
		WHERE distributor_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, distributorID)
	if err != nil {
		return nil, fmt.Errorf("listing cycles by distributor: %w", err)
	}
	defer rows.Close()

	var cycles []*OrderCreditCycleRecord
	for rows.Next() {
		c := &OrderCreditCycleRecord{}
		err := rows.Scan(
			&c.ID, &c.OrderID, &c.DistributorID, &c.AccountID, &c.AgreementID,
			&c.TotalAmountPaise, &c.AdvancePaidPaise, &c.CreditAmountPaise,
			&c.InstalmentAmountPaise, &c.Frequency, &c.TotalInstalments,
			&c.CompletedInstalments, &c.Status, &c.GracePeriodDays,
			&c.GracePeriodEndsAt, &c.HoldTriggeredAt, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning cycle row: %w", err)
		}
		cycles = append(cycles, c)
	}
	return cycles, nil
}

// UpdateCycleStatus updates the status of a cycle
func (r *CreditCycleRepository) UpdateCycleStatus(ctx context.Context, cycleID, status string) error {
	query := `
		UPDATE order_credit_cycles
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, cycleID, status)
	return err
}

// UpdateCycleAgreement attaches agreement_id to cycle
func (r *CreditCycleRepository) UpdateCycleAgreement(ctx context.Context, cycleID, agreementID string) error {
	query := `
		UPDATE order_credit_cycles
		SET agreement_id = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, cycleID, agreementID)
	return err
}

// IncrementCompletedInstalments increments completed instalments and completes cycle if done
func (r *CreditCycleRepository) IncrementCompletedInstalments(ctx context.Context, cycleID string) error {
	query := `
		UPDATE order_credit_cycles
		SET completed_instalments = completed_instalments + 1,
		    status = CASE 
		        WHEN completed_instalments + 1 >= total_instalments THEN 'completed'
		        ELSE status
		    END,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, cycleID)
	return err
}

// SetCycleGracePeriod sets grace period details
func (r *CreditCycleRepository) SetCycleGracePeriod(ctx context.Context, cycleID string, endsAt time.Time) error {
	query := `
		UPDATE order_credit_cycles
		SET grace_period_ends_at = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, cycleID, endsAt)
	return err
}

// SetCycleHoldTriggered records hold trigger timestamp and sets status to defaulted
func (r *CreditCycleRepository) SetCycleHoldTriggered(ctx context.Context, cycleID string) error {
	query := `
		UPDATE order_credit_cycles
		SET hold_triggered_at = NOW(), status = 'defaulted', updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, cycleID)
	return err
}

// CreateMandate inserts a upi_mandate record
func (r *CreditCycleRepository) CreateMandate(ctx context.Context, m *UPIMandateRecord) error {
	query := `
		INSERT INTO upi_mandates (
			order_id, cycle_id, distributor_id, razorpay_plan_id,
			razorpay_subscription_id, razorpay_short_url, mandate_status,
			frequency, total_count, paid_count, amount_per_instalment_paise
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		m.OrderID, m.CycleID, m.DistributorID, m.RazorpayPlanID,
		m.RazorpaySubscriptionID, m.RazorpayShortURL, m.MandateStatus,
		m.Frequency, m.TotalCount, m.PaidCount, m.AmountPerInstalmentPaise,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

// GetMandateByCycleID fetches mandate for a cycle
func (r *CreditCycleRepository) GetMandateByCycleID(ctx context.Context, cycleID string) (*UPIMandateRecord, error) {
	query := `
		SELECT id, order_id, cycle_id, distributor_id, razorpay_plan_id,
		       razorpay_subscription_id, razorpay_short_url, mandate_status,
		       frequency, total_count, paid_count, amount_per_instalment_paise,
		       start_at, next_charge_at, authenticated_at, created_at, updated_at
		FROM upi_mandates
		WHERE cycle_id = $1
	`
	m := &UPIMandateRecord{}
	err := r.db.QueryRow(ctx, query, cycleID).Scan(
		&m.ID, &m.OrderID, &m.CycleID, &m.DistributorID, &m.RazorpayPlanID,
		&m.RazorpaySubscriptionID, &m.RazorpayShortURL, &m.MandateStatus,
		&m.Frequency, &m.TotalCount, &m.PaidCount, &m.AmountPerInstalmentPaise,
		&m.StartAt, &m.NextChargeAt, &m.AuthenticatedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting mandate by cycle id: %w", err)
	}
	return m, nil
}

// GetMandateBySubscriptionID fetches mandate by razorpay subscription id
func (r *CreditCycleRepository) GetMandateBySubscriptionID(ctx context.Context, subID string) (*UPIMandateRecord, error) {
	query := `
		SELECT id, order_id, cycle_id, distributor_id, razorpay_plan_id,
		       razorpay_subscription_id, razorpay_short_url, mandate_status,
		       frequency, total_count, paid_count, amount_per_instalment_paise,
		       start_at, next_charge_at, authenticated_at, created_at, updated_at
		FROM upi_mandates
		WHERE razorpay_subscription_id = $1
	`
	m := &UPIMandateRecord{}
	err := r.db.QueryRow(ctx, query, subID).Scan(
		&m.ID, &m.OrderID, &m.CycleID, &m.DistributorID, &m.RazorpayPlanID,
		&m.RazorpaySubscriptionID, &m.RazorpayShortURL, &m.MandateStatus,
		&m.Frequency, &m.TotalCount, &m.PaidCount, &m.AmountPerInstalmentPaise,
		&m.StartAt, &m.NextChargeAt, &m.AuthenticatedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting mandate by subscription id: %w", err)
	}
	return m, nil
}

// UpdateMandateStatus updates mandate status and counts
func (r *CreditCycleRepository) UpdateMandateStatus(ctx context.Context, subID, status string, paidCount int) error {
	query := `
		UPDATE upi_mandates
		SET mandate_status = $2, paid_count = GREATEST(paid_count, $3),
		    authenticated_at = CASE WHEN $2 = 'authenticated' OR $2 = 'active' THEN COALESCE(authenticated_at, NOW()) ELSE authenticated_at END,
		    updated_at = NOW()
		WHERE razorpay_subscription_id = $1
	`
	_, err := r.db.Exec(ctx, query, subID, status, paidCount)
	return err
}

// CreateRepayments batch inserts cycle_repayments
func (r *CreditCycleRepository) CreateRepayments(ctx context.Context, repayments []*CycleRepaymentRecord) error {
	for _, rep := range repayments {
		query := `
			INSERT INTO cycle_repayments (
				cycle_id, mandate_id, instalment_number, due_date, amount_paise, status
			) VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (cycle_id, instalment_number) DO NOTHING
			RETURNING id, created_at, updated_at
		`
		err := r.db.QueryRow(ctx, query,
			rep.CycleID, rep.MandateID, rep.InstalmentNumber, rep.DueDate, rep.AmountPaise, rep.Status,
		).Scan(&rep.ID, &rep.CreatedAt, &rep.UpdatedAt)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("creating repayment: %w", err)
		}
	}
	return nil
}

// ListRepaymentsByCycleID fetches all repayments for a cycle
func (r *CreditCycleRepository) ListRepaymentsByCycleID(ctx context.Context, cycleID string) ([]*CycleRepaymentRecord, error) {
	query := `
		SELECT DISTINCT ON (instalment_number)
		       id, cycle_id, mandate_id, instalment_number, due_date, amount_paise,
		       status, razorpay_payment_id, razorpay_invoice_id, paid_at, failure_reason, retry_count,
		       created_at, updated_at
		FROM cycle_repayments
		WHERE cycle_id = $1
		ORDER BY instalment_number ASC, created_at ASC
	`
	rows, err := r.db.Query(ctx, query, cycleID)
	if err != nil {
		return nil, fmt.Errorf("listing repayments: %w", err)
	}
	defer rows.Close()

	var repayments []*CycleRepaymentRecord
	for rows.Next() {
		rep := &CycleRepaymentRecord{}
		err := rows.Scan(
			&rep.ID, &rep.CycleID, &rep.MandateID, &rep.InstalmentNumber, &rep.DueDate, &rep.AmountPaise,
			&rep.Status, &rep.RazorpayPaymentID, &rep.RazorpayInvoiceID, &rep.PaidAt, &rep.FailureReason, &rep.RetryCount,
			&rep.CreatedAt, &rep.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning repayment row: %w", err)
		}
		repayments = append(repayments, rep)
	}
	return repayments, nil
}

// UpdateRepaymentStatus updates repayment record status
func (r *CreditCycleRepository) UpdateRepaymentStatus(ctx context.Context, id, status string, paymentID, failureReason *string) error {
	query := `
		UPDATE cycle_repayments
		SET status = $2,
		    razorpay_payment_id = COALESCE($3, razorpay_payment_id),
		    failure_reason = COALESCE($4, failure_reason),
		    paid_at = CASE WHEN $2 = 'paid' THEN NOW() ELSE paid_at END,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, status, paymentID, failureReason)
	return err
}

// GetOverdueCyclesWithExpiredGrace fetches cycles where mandate/repayment failed and grace period has passed
func (r *CreditCycleRepository) GetOverdueCyclesWithExpiredGrace(ctx context.Context) ([]*OrderCreditCycleRecord, error) {
	query := `
		SELECT id, order_id, distributor_id, account_id, agreement_id,
		       total_amount_paise, advance_paid_paise, credit_amount_paise,
		       instalment_amount_paise, frequency, total_instalments,
		       completed_instalments, status, grace_period_days,
		       grace_period_ends_at, hold_triggered_at, created_at, updated_at
		FROM order_credit_cycles
		WHERE status IN ('mandate_pending', 'repaying', 'defaulted')
		  AND grace_period_ends_at IS NOT NULL
		  AND grace_period_ends_at <= NOW()
		  AND hold_triggered_at IS NULL
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying overdue cycles: %w", err)
	}
	defer rows.Close()

	var cycles []*OrderCreditCycleRecord
	for rows.Next() {
		c := &OrderCreditCycleRecord{}
		err := rows.Scan(
			&c.ID, &c.OrderID, &c.DistributorID, &c.AccountID, &c.AgreementID,
			&c.TotalAmountPaise, &c.AdvancePaidPaise, &c.CreditAmountPaise,
			&c.InstalmentAmountPaise, &c.Frequency, &c.TotalInstalments,
			&c.CompletedInstalments, &c.Status, &c.GracePeriodDays,
			&c.GracePeriodEndsAt, &c.HoldTriggeredAt, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning overdue cycle row: %w", err)
		}
		cycles = append(cycles, c)
	}
	return cycles, nil
}

// UpdateAccountStatus sets distributor credit_account status (e.g. 'hold', 'active')
func (r *CreditCycleRepository) UpdateAccountStatus(ctx context.Context, distributorID, status string) error {
	query := `
		UPDATE credit_accounts
		SET status = $2, updated_at = NOW()
		WHERE distributor_id = $1
	`
	_, err := r.db.Exec(ctx, query, distributorID, status)
	return err
}
