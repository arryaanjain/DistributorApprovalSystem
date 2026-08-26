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
	IsSample    bool      `json:"is_sample"`
	IsRegular   bool      `json:"is_regular"`
	CreatedAt   time.Time `json:"created_at"`
}

type AddressRecord struct {
	ID            string    `json:"id"`
	DistributorID string    `json:"distributor_id"`
	AddressType   string    `json:"address_type"`
	AddressLine1  string    `json:"address_line1"`
	AddressLine2  *string   `json:"address_line2,omitempty"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	PIN           string    `json:"pin"`
	Phone         *string   `json:"phone,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type SampleOrderRecord struct {
	ID                string         `json:"id"`
	DistributorID     string         `json:"distributor_id"`
	DistributorName   *string        `json:"distributor_name,omitempty"`
	DistributorMobile *string        `json:"distributor_mobile,omitempty"`
	RazorpayOrderID   string         `json:"razorpay_order_id"`
	RazorpayPaymentID *string        `json:"razorpay_payment_id,omitempty"`
	RazorpaySignature *string        `json:"razorpay_signature,omitempty"`
	AmountPaise       int64          `json:"amount_paise"`
	Status            string         `json:"status"`
	ItemsJSON         *string        `json:"items_json,omitempty"`
	AddressID         *string        `json:"address_id,omitempty"`
	ShippingAddress   *AddressRecord `json:"shipping_address,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	ShiprocketOrderID *string        `json:"shiprocket_order_id,omitempty"`
	ShipmentID        *string        `json:"shipment_id,omitempty"`
	AWBCode           *string        `json:"awb_code,omitempty"`
	CourierName       *string        `json:"courier_name,omitempty"`
	LabelURL          *string        `json:"label_url,omitempty"`
	ManifestURL       *string        `json:"manifest_url,omitempty"`
	PickupStatus      *string        `json:"pickup_status,omitempty"`
	PackageWeight     *float64       `json:"package_weight,omitempty"`
	PackageLength     *float64       `json:"package_length,omitempty"`
	PackageBreadth    *float64       `json:"package_breadth,omitempty"`
	PackageHeight     *float64       `json:"package_height,omitempty"`
}

type OrderRecord struct {
	ID                string     `json:"id"`
	OrderNumber       string     `json:"order_number"`
	DistributorID     string     `json:"distributor_id"`
	DistributorName   *string    `json:"distributor_name,omitempty"`
	DistributorMobile *string    `json:"distributor_mobile,omitempty"`
	BusinessName      *string    `json:"business_name,omitempty"`
	TotalAmountPaise  int64      `json:"total_amount_paise"`
	AdvancePaidPaise  int64      `json:"advance_paid_paise"`
	CreditUsedPaise   int64      `json:"credit_used_paise"`
	Status            string     `json:"status"` // PENDING_PAYMENT, PAYMENT_SUBMITTED, PAYMENT_VERIFIED, PENDING_REVIEW, APPROVED, REJECTED, DISPATCHED, CANCELLED
	PaymentProofURL   *string    `json:"payment_proof_url,omitempty"`
	UTRReference      *string    `json:"utr_reference,omitempty"`
	ReviewedBy        *string    `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes       *string    `json:"review_notes,omitempty"`
	DispatchedAt      *time.Time `json:"dispatched_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	ShiprocketOrderID *string    `json:"shiprocket_order_id,omitempty"`
	ShipmentID        *string    `json:"shipment_id,omitempty"`
	AWBCode           *string    `json:"awb_code,omitempty"`
	CourierName       *string    `json:"courier_name,omitempty"`
	LabelURL          *string    `json:"label_url,omitempty"`
	ManifestURL       *string    `json:"manifest_url,omitempty"`
	PickupStatus      *string    `json:"pickup_status,omitempty"`
	PackageWeight     *float64   `json:"package_weight,omitempty"`
	PackageLength     *float64   `json:"package_length,omitempty"`
	PackageBreadth    *float64   `json:"package_breadth,omitempty"`
	PackageHeight     *float64   `json:"package_height,omitempty"`
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
		`SELECT id, sku, name, description, category, price_paise, moq, is_active, is_sample, is_regular, created_at
		 FROM products WHERE is_active = TRUE ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProductRecord
	for rows.Next() {
		p := ProductRecord{}
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.PricePaise, &p.Moq, &p.IsActive, &p.IsSample, &p.IsRegular, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *OrderRepository) ListSampleProducts(ctx context.Context) ([]ProductRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, sku, name, description, category, price_paise, moq, is_active, is_sample, is_regular, created_at
		 FROM products WHERE is_active = TRUE AND is_sample = TRUE ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProductRecord
	for rows.Next() {
		p := ProductRecord{}
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.PricePaise, &p.Moq, &p.IsActive, &p.IsSample, &p.IsRegular, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *OrderRepository) ListAllProductsAdmin(ctx context.Context) ([]ProductRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, sku, name, description, category, price_paise, moq, is_active, is_sample, is_regular, created_at
		 FROM products ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProductRecord
	for rows.Next() {
		p := ProductRecord{}
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.PricePaise, &p.Moq, &p.IsActive, &p.IsSample, &p.IsRegular, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *OrderRepository) CreateProduct(ctx context.Context, p *ProductRecord) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO products (sku, name, description, category, price_paise, moq, is_active, is_sample, is_regular)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		p.SKU, p.Name, p.Description, p.Category, p.PricePaise, p.Moq, p.IsActive, p.IsSample, p.IsRegular,
	).Scan(&id)
	return id, err
}

func (r *OrderRepository) UpdateProduct(ctx context.Context, p *ProductRecord) error {
	_, err := r.db.Exec(ctx,
		`UPDATE products
		 SET name = $1, description = $2, category = $3, price_paise = $4, moq = $5,
		     is_active = $6, is_sample = $7, is_regular = $8
		 WHERE id = $9`,
		p.Name, p.Description, p.Category, p.PricePaise, p.Moq, p.IsActive, p.IsSample, p.IsRegular, p.ID,
	)
	return err
}

func (r *OrderRepository) GetProductByID(ctx context.Context, id string) (*ProductRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, sku, name, description, category, price_paise, moq, is_active, is_sample, is_regular, created_at
		 FROM products WHERE id = $1`, id)
	p := &ProductRecord{}
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category, &p.PricePaise, &p.Moq, &p.IsActive, &p.IsSample, &p.IsRegular, &p.CreatedAt)
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

// ──────────────────────────────── Address Management ────────────────────────
func (r *OrderRepository) CreateAddress(ctx context.Context, addr *AddressRecord) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO addresses (distributor_id, address_type, address_line1, address_line2, city, state, pin, phone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		addr.DistributorID, addr.AddressType, addr.AddressLine1, addr.AddressLine2, addr.City, addr.State, addr.PIN, addr.Phone,
	).Scan(&id)
	return id, err
}

func (r *OrderRepository) GetAddressByID(ctx context.Context, id string) (*AddressRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, address_type, address_line1, address_line2, city, state, pin, phone, created_at
		 FROM addresses WHERE id = $1`, id)
	a := &AddressRecord{}
	err := row.Scan(&a.ID, &a.DistributorID, &a.AddressType, &a.AddressLine1, &a.AddressLine2, &a.City, &a.State, &a.PIN, &a.Phone, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

// ──────────────────────────────── Sample Orders (Razorpay) ───────────────────

func (r *OrderRepository) CreateSampleOrder(ctx context.Context, distributorID, rzpOrderID string, amountPaise int64, itemsJSON string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO sample_orders (distributor_id, razorpay_order_id, amount_paise, status, items)
		 VALUES ($1, $2, $3, 'CREATED', $4::jsonb) RETURNING id`,
		distributorID, rzpOrderID, amountPaise, itemsJSON,
	).Scan(&id)
	return id, err
}

func (r *OrderRepository) CreateSampleOrderWithAddress(ctx context.Context, distributorID, rzpOrderID, addressID string, amountPaise int64, itemsJSON string) (string, error) {
	var id string
	var addrID *string
	if addressID != "" {
		addrID = &addressID
	}
	err := r.db.QueryRow(ctx,
		`INSERT INTO sample_orders (distributor_id, razorpay_order_id, amount_paise, status, items, address_id)
		 VALUES ($1, $2, $3, 'CREATED', $4::jsonb, $5) RETURNING id`,
		distributorID, rzpOrderID, amountPaise, itemsJSON, addrID,
	).Scan(&id)
	return id, err
}

func (r *OrderRepository) VerifySampleOrderPayment(ctx context.Context, rzpOrderID, rzpPaymentID, rzpSignature string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var distID string
	err = tx.QueryRow(ctx,
		`UPDATE sample_orders
		 SET razorpay_payment_id = $1, razorpay_signature = $2, status = 'PAID', updated_at = NOW()
		 WHERE razorpay_order_id = $3
		 RETURNING distributor_id`,
		rzpPaymentID, rzpSignature, rzpOrderID,
	).Scan(&distID)
	if err != nil {
		return err
	}

	// Update active application status to 'trial'
	_, err = tx.Exec(ctx,
		`UPDATE applications
		 SET status = 'trial'::application_status, updated_at = NOW()
		 WHERE distributor_id = $1 AND status NOT IN ('rejected', 'blocked')`,
		distID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) ListAllCatalogOrdersAdmin(ctx context.Context, limit, offset int) ([]OrderRecord, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT o.id, o.order_number, o.distributor_id, d.name, d.mobile, bp.business_name,
		        o.total_amount_paise, o.advance_paid_paise, o.credit_used_paise,
		        o.status, o.payment_proof_url, o.utr_reference, o.reviewed_by, o.reviewed_at, o.review_notes, o.dispatched_at, o.created_at
		 FROM orders o
		 JOIN distributors d ON o.distributor_id = d.id
		 LEFT JOIN business_profiles bp ON bp.distributor_id = d.id
		 ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []OrderRecord
	for rows.Next() {
		o := OrderRecord{}
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.DistributorID, &o.DistributorName, &o.DistributorMobile, &o.BusinessName,
			&o.TotalAmountPaise, &o.AdvancePaidPaise, &o.CreditUsedPaise, &o.Status,
			&o.PaymentProofURL, &o.UTRReference, &o.ReviewedBy, &o.ReviewedAt, &o.ReviewNotes, &o.DispatchedAt, &o.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, o)
	}
	return list, total, nil
}

func (r *OrderRepository) ListSampleOrdersAdmin(ctx context.Context, limit, offset int) ([]SampleOrderRecord, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM sample_orders`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.distributor_id, d.name, d.mobile, s.razorpay_order_id, s.razorpay_payment_id,
		        s.amount_paise, s.status, s.items::TEXT, s.address_id, s.created_at,
		        s.shiprocket_order_id, s.shipment_id, s.awb_code, s.courier_name, s.label_url, s.manifest_url,
		        s.pickup_status, s.package_weight, s.package_length, s.package_breadth, s.package_height,
		        a.address_line1, a.address_line2, a.city, a.state, a.pin, a.phone
		 FROM sample_orders s
		 JOIN distributors d ON s.distributor_id = d.id
		 LEFT JOIN addresses a ON s.address_id = a.id
		 ORDER BY s.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []SampleOrderRecord
	for rows.Next() {
		s := SampleOrderRecord{}
		var addrLine1, addrLine2, city, state, pin, phone *string
		if err := rows.Scan(&s.ID, &s.DistributorID, &s.DistributorName, &s.DistributorMobile,
			&s.RazorpayOrderID, &s.RazorpayPaymentID, &s.AmountPaise, &s.Status, &s.ItemsJSON,
			&s.AddressID, &s.CreatedAt,
			&s.ShiprocketOrderID, &s.ShipmentID, &s.AWBCode, &s.CourierName, &s.LabelURL, &s.ManifestURL,
			&s.PickupStatus, &s.PackageWeight, &s.PackageLength, &s.PackageBreadth, &s.PackageHeight,
			&addrLine1, &addrLine2, &city, &state, &pin, &phone); err != nil {
			return nil, 0, err
		}
		if addrLine1 != nil && s.AddressID != nil {
			s.ShippingAddress = &AddressRecord{
				ID:            *s.AddressID,
				DistributorID: s.DistributorID,
				AddressType:   "shipping",
				AddressLine1:  *addrLine1,
				AddressLine2:  addrLine2,
				City:          *city,
				State:         *state,
				PIN:           *pin,
				Phone:         phone,
			}
		}
		list = append(list, s)
	}
	return list, total, nil
}

func (r *OrderRepository) ListSampleOrdersByDistributor(ctx context.Context, distributorID string) ([]SampleOrderRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.distributor_id, s.razorpay_order_id, s.razorpay_payment_id,
		        s.amount_paise, s.status, s.items::TEXT, s.address_id, s.created_at,
		        s.shiprocket_order_id, s.shipment_id, s.awb_code, s.courier_name, s.label_url, s.manifest_url,
		        s.pickup_status, s.package_weight, s.package_length, s.package_breadth, s.package_height
		 FROM sample_orders s
		 WHERE s.distributor_id = $1
		 ORDER BY s.created_at DESC`, distributorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SampleOrderRecord
	for rows.Next() {
		s := SampleOrderRecord{}
		if err := rows.Scan(&s.ID, &s.DistributorID, &s.RazorpayOrderID, &s.RazorpayPaymentID,
			&s.AmountPaise, &s.Status, &s.ItemsJSON, &s.AddressID, &s.CreatedAt,
			&s.ShiprocketOrderID, &s.ShipmentID, &s.AWBCode, &s.CourierName, &s.LabelURL, &s.ManifestURL,
			&s.PickupStatus, &s.PackageWeight, &s.PackageLength, &s.PackageBreadth, &s.PackageHeight); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *OrderRepository) GetSampleOrderByID(ctx context.Context, id string) (*SampleOrderRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT s.id, s.distributor_id, d.name, d.mobile, s.razorpay_order_id, s.razorpay_payment_id,
		        s.amount_paise, s.status, s.items::TEXT, s.address_id, s.created_at,
		        s.shiprocket_order_id, s.shipment_id, s.awb_code, s.courier_name, s.label_url, s.manifest_url,
		        s.pickup_status, s.package_weight, s.package_length, s.package_breadth, s.package_height,
		        a.address_line1, a.address_line2, a.city, a.state, a.pin, a.phone
		 FROM sample_orders s
		 JOIN distributors d ON s.distributor_id = d.id
		 LEFT JOIN addresses a ON s.address_id = a.id
		 WHERE s.id = $1`, id)
	s := &SampleOrderRecord{}
	var addrLine1, addrLine2, city, state, pin, phone *string
	err := row.Scan(&s.ID, &s.DistributorID, &s.DistributorName, &s.DistributorMobile,
		&s.RazorpayOrderID, &s.RazorpayPaymentID, &s.AmountPaise, &s.Status, &s.ItemsJSON,
		&s.AddressID, &s.CreatedAt,
		&s.ShiprocketOrderID, &s.ShipmentID, &s.AWBCode, &s.CourierName, &s.LabelURL, &s.ManifestURL,
		&s.PickupStatus, &s.PackageWeight, &s.PackageLength, &s.PackageBreadth, &s.PackageHeight,
		&addrLine1, &addrLine2, &city, &state, &pin, &phone)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if addrLine1 != nil && s.AddressID != nil {
		s.ShippingAddress = &AddressRecord{
			ID:            *s.AddressID,
			DistributorID: s.DistributorID,
			AddressType:   "shipping",
			AddressLine1:  *addrLine1,
			AddressLine2:  addrLine2,
			City:          *city,
			State:         *state,
			PIN:           *pin,
			Phone:         phone,
		}
	}
	return s, nil
}

func (r *OrderRepository) UpdateSampleOrderStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE sample_orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *OrderRepository) UpdateSampleOrderShipment(ctx context.Context, id, srOrderID, shipmentID string, weight, length, breadth, height float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sample_orders
		 SET shiprocket_order_id = $1, shipment_id = $2, package_weight = $3,
		     package_length = $4, package_breadth = $5, package_height = $6,
		     status = 'PROCESSING', updated_at = NOW()
		 WHERE id = $7`,
		srOrderID, shipmentID, weight, length, breadth, height, id)
	return err
}

func (r *OrderRepository) UpdateSampleOrderAWB(ctx context.Context, id, awbCode, courierName string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sample_orders
		 SET awb_code = $1, courier_name = $2, status = 'DISPATCHED', updated_at = NOW()
		 WHERE id = $3`,
		awbCode, courierName, id)
	return err
}

func (r *OrderRepository) UpdateSampleOrderLabel(ctx context.Context, id, labelURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sample_orders SET label_url = $1, updated_at = NOW() WHERE id = $2`, labelURL, id)
	return err
}

func (r *OrderRepository) UpdateSampleOrderManifest(ctx context.Context, id, manifestURL string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sample_orders SET manifest_url = $1, updated_at = NOW() WHERE id = $2`, manifestURL, id)
	return err
}

func (r *OrderRepository) UpdateSampleOrderPickupStatus(ctx context.Context, id, pickupStatus string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sample_orders SET pickup_status = $1, updated_at = NOW() WHERE id = $2`, pickupStatus, id)
	return err
}

func (r *OrderRepository) GetSampleOrderByAWB(ctx context.Context, awbCode string) (*SampleOrderRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT s.id, s.distributor_id, d.name, d.mobile, s.razorpay_order_id, s.razorpay_payment_id,
		        s.amount_paise, s.status, s.items::TEXT, s.address_id, s.created_at,
		        s.shiprocket_order_id, s.shipment_id, s.awb_code, s.courier_name, s.label_url, s.manifest_url,
		        s.pickup_status, s.package_weight, s.package_length, s.package_breadth, s.package_height
		 FROM sample_orders s
		 JOIN distributors d ON s.distributor_id = d.id
		 WHERE s.awb_code = $1`, awbCode)
	s := &SampleOrderRecord{}
	err := row.Scan(&s.ID, &s.DistributorID, &s.DistributorName, &s.DistributorMobile,
		&s.RazorpayOrderID, &s.RazorpayPaymentID, &s.AmountPaise, &s.Status, &s.ItemsJSON,
		&s.AddressID, &s.CreatedAt,
		&s.ShiprocketOrderID, &s.ShipmentID, &s.AWBCode, &s.CourierName, &s.LabelURL, &s.ManifestURL,
		&s.PickupStatus, &s.PackageWeight, &s.PackageLength, &s.PackageBreadth, &s.PackageHeight)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}


