package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceRecord struct {
	ID               string    `json:"id"`
	InvoiceNumber    string    `json:"invoice_number"`
	DistributorID    string    `json:"distributor_id"`
	AccountID        string    `json:"account_id"`
	OrderID          string    `json:"order_id"`
	InvoiceDate      time.Time `json:"invoice_date"`
	DueDate          time.Time `json:"due_date"`
	TotalPaise       int64     `json:"total_paise"`
	CreditPaise      int64     `json:"credit_paise"`
	AdvancePaise     int64     `json:"advance_paise"`
	OutstandingPaise int64     `json:"outstanding_paise"`
	Status           string    `json:"status"` // open, partially_paid, paid, overdue, written_off
	DaysOutstanding  int       `json:"days_outstanding"`
	OverdueTier      string    `json:"overdue_tier"` // current, overdue_1_3, overdue_4_7, overdue_8_15, serious
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type FinancialRepository struct {
	db *pgxpool.Pool
}

func NewFinancialRepository(db *pgxpool.Pool) *FinancialRepository {
	return &FinancialRepository{db: db}
}

// ──────────────────────────────── Invoices ───────────────────────────────────

func (r *FinancialRepository) CreateInvoice(ctx context.Context, inv *InvoiceRecord) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO invoices
		 (invoice_number, distributor_id, account_id, order_id, invoice_date, due_date,
		  total_paise, credit_paise, advance_paise, outstanding_paise, status, overdue_tier)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`,
		inv.InvoiceNumber, inv.DistributorID, inv.AccountID, inv.OrderID, inv.InvoiceDate, inv.DueDate,
		inv.TotalPaise, inv.CreditPaise, inv.AdvancePaise, inv.OutstandingPaise, inv.Status, inv.OverdueTier,
	).Scan(&id)
	return id, err
}

func (r *FinancialRepository) ListInvoicesByDistributor(ctx context.Context, distributorID string) ([]InvoiceRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, invoice_number, distributor_id, account_id, order_id, invoice_date, due_date,
		        total_paise, credit_paise, advance_paise, outstanding_paise, status, days_outstanding, overdue_tier, created_at, updated_at
		 FROM invoices WHERE distributor_id = $1 ORDER BY due_date ASC`, distributorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []InvoiceRecord
	for rows.Next() {
		inv := InvoiceRecord{}
		if err := rows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.DistributorID, &inv.AccountID, &inv.OrderID,
			&inv.InvoiceDate, &inv.DueDate, &inv.TotalPaise, &inv.CreditPaise, &inv.AdvancePaise,
			&inv.OutstandingPaise, &inv.Status, &inv.DaysOutstanding, &inv.OverdueTier, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, nil
}

func (r *FinancialRepository) RecordInvoicePayment(ctx context.Context, invoiceID, distributorID, paymentMode, utr string, amountPaise int64, recordedBy *string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert invoice_payments
	var paymentID string
	err = tx.QueryRow(ctx,
		`INSERT INTO invoice_payments
		 (invoice_id, distributor_id, amount_paise, payment_mode, utr_reference, payment_date, recorded_by)
		 VALUES ($1, $2, $3, $4, $5, CURRENT_DATE, $6) RETURNING id`,
		invoiceID, distributorID, amountPaise, paymentMode, utr, recordedBy,
	).Scan(&paymentID)
	if err != nil {
		return err
	}

	// Update invoice outstanding and status
	var currentOutstanding int64
	var creditPaise int64
	var accountID string
	err = tx.QueryRow(ctx,
		`SELECT outstanding_paise, credit_paise, account_id FROM invoices WHERE id = $1 FOR UPDATE`, invoiceID,
	).Scan(&currentOutstanding, &creditPaise, &accountID)
	if err != nil {
		return err
	}

	newOutstanding := currentOutstanding - amountPaise
	if newOutstanding < 0 {
		newOutstanding = 0
	}

	newStatus := "partially_paid"
	if newOutstanding == 0 {
		newStatus = "paid"
	}

	_, err = tx.Exec(ctx,
		`UPDATE invoices SET outstanding_paise = $1, status = $2, updated_at = NOW() WHERE id = $3`,
		newOutstanding, newStatus, invoiceID,
	)
	if err != nil {
		return err
	}

	// Replenish credit account available limit
	var acc CreditAccountRecord
	err = tx.QueryRow(ctx,
		`SELECT id, approved_limit_paise, current_credit_paise FROM credit_accounts WHERE id = $1 FOR UPDATE`, accountID,
	).Scan(&acc.ID, &acc.ApprovedLimitPaise, &acc.CurrentCreditPaise)
	if err != nil {
		return err
	}

	newCurrentCredit := acc.CurrentCreditPaise - amountPaise
	if newCurrentCredit < 0 {
		newCurrentCredit = 0
	}
	newAvailableCredit := acc.ApprovedLimitPaise - newCurrentCredit

	_, err = tx.Exec(ctx,
		`UPDATE credit_accounts SET current_credit_paise = $1, available_credit_paise = $2, updated_at = NOW() WHERE id = $3`,
		newCurrentCredit, newAvailableCredit, acc.ID,
	)
	if err != nil {
		return err
	}

	// Log transaction
	_, err = tx.Exec(ctx,
		`INSERT INTO credit_transactions
		 (distributor_id, account_id, transaction_type, amount_paise, reference_id, reference_type, balance_after_paise)
		 VALUES ($1, $2, 'PAYMENT_RECEIVED', $3, $4, 'invoice', $5)`,
		distributorID, accountID, amountPaise, invoiceID, newAvailableCredit,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ──────────────────────────────── Collections ────────────────────────────────

func (r *FinancialRepository) EvaluateOverdueTiers(ctx context.Context) ([]string, error) {
	// Updates overdue status for all open invoices where CURRENT_DATE > due_date
	rows, err := r.db.Query(ctx,
		`SELECT id, distributor_id, due_date FROM invoices WHERE status IN ('open', 'partially_paid') AND CURRENT_DATE > due_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updatedInvoices []string
	type invMeta struct {
		id     string
		distID string
		due    time.Time
	}
	var overdueList []invMeta

	for rows.Next() {
		var meta invMeta
		if err := rows.Scan(&meta.id, &meta.distID, &meta.due); err == nil {
			overdueList = append(overdueList, meta)
		}
	}

	now := time.Now()
	for _, inv := range overdueList {
		daysOverdue := int(now.Sub(inv.due).Hours() / 24)
		tier := "overdue_1_3"
		if daysOverdue >= 15 {
			tier = "serious"
		} else if daysOverdue >= 8 {
			tier = "overdue_8_15"
		} else if daysOverdue >= 4 {
			tier = "overdue_4_7"
		}

		_, _ = r.db.Exec(ctx,
			`UPDATE invoices SET status = 'overdue', days_outstanding = $1, overdue_tier = $2, updated_at = NOW() WHERE id = $3`,
			daysOverdue, tier, inv.id,
		)

		// Auto-restrict credit account if overdue >= 15 days
		if daysOverdue >= 15 {
			_, _ = r.db.Exec(ctx,
				`UPDATE credit_accounts SET status = 'RESTRICTED', updated_at = NOW() WHERE distributor_id = $1`,
				inv.distID,
			)
		}
		updatedInvoices = append(updatedInvoices, inv.id)
	}

	return updatedInvoices, nil
}

// ──────────────────────────────── Credit Enhancement / Ladder ─────────────────

// Step levels for the credit ladder: ₹15k, ₹25k, ₹35k, ₹50k, ₹1L, ₹1.5L, ₹2L, ₹3L (in paise)
var CreditLadderSteps = []int64{
	1500000,  // ₹15,000
	2500000,  // ₹25,000
	3500000,  // ₹35,000
	5000000,  // ₹50,000
	10000000, // ₹1,00,000
	15000000, // ₹1,50,000
	20000000, // ₹2,00,000
	30000000, // ₹3,00,000
}

func (r *FinancialRepository) GetNextLadderStep(currentLimitPaise int64) (int64, bool) {
	for _, step := range CreditLadderSteps {
		if step > currentLimitPaise {
			return step, true
		}
	}
	return currentLimitPaise, false
}

func (r *FinancialRepository) EnhanceCreditLimit(ctx context.Context, distributorID string, newLimitPaise int64) error {
	var acc CreditAccountRecord
	err := r.db.QueryRow(ctx,
		`SELECT id, approved_limit_paise, current_credit_paise FROM credit_accounts WHERE distributor_id = $1`, distributorID,
	).Scan(&acc.ID, &acc.ApprovedLimitPaise, &acc.CurrentCreditPaise)
	if err != nil {
		return err
	}

	newAvailable := newLimitPaise - acc.CurrentCreditPaise
	if newAvailable < 0 {
		newAvailable = 0
	}

	_, err = r.db.Exec(ctx,
		`UPDATE credit_accounts SET approved_limit_paise = $1, available_credit_paise = $2, updated_at = NOW() WHERE id = $3`,
		newLimitPaise, newAvailable, acc.ID,
	)
	return err
}
