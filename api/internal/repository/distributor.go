package repository

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DistributorRecord is a row from the distributors table.
type DistributorRecord struct {
	ID        string    `json:"id"`
	Mobile    string    `json:"mobile"`
	Email     *string   `json:"email,omitempty"`
	Name      *string   `json:"name,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BusinessProfileRecord maps to business_profiles.
type BusinessProfileRecord struct {
	ID                          string   `json:"id"`
	DistributorID               string   `json:"distributor_id"`
	BusinessName                string   `json:"business_name"`
	Constitution                string   `json:"constitution"`
	AddressLine1                string   `json:"address_line1"`
	AddressLine2                *string  `json:"address_line2,omitempty"`
	City                        string   `json:"city"`
	State                       string   `json:"state"`
	PIN                         string   `json:"pin"`
	VintageYears                *float64 `json:"vintage_years,omitempty"`
	FMCGExperienceYears         *float64 `json:"fmcg_experience_years,omitempty"`
	DistributionExperienceYears *float64 `json:"distribution_experience_years,omitempty"`
	ApproxMonthlyBusinessPaise  *int64   `json:"approx_monthly_business_paise,omitempty"`
	RetailerCount               *int     `json:"retailer_count,omitempty"`
	ServicedRetailersWholesalersCount *int `json:"serviced_retailers_wholesalers_count,omitempty"`
	SalespersonCount            *int     `json:"salesperson_count,omitempty"`
	InterestedBusinessRole      *string  `json:"interested_business_role,omitempty"`
	ExistingBrands              []string `json:"existing_brands,omitempty"`
}

// BusinessDocumentRecord maps to business_documents.
type BusinessDocumentRecord struct {
	ID             string  `json:"id"`
	DistributorID  string  `json:"distributor_id"`
	PAN            *string `json:"pan,omitempty"`
	GSTNumber      *string `json:"gst_number,omitempty"`
	FSSAINumber    *string `json:"fssai_number,omitempty"`
	UdyamNumber    *string `json:"udyam_number,omitempty"`
	ShopEstNumber  *string `json:"shop_est_number,omitempty"`
	HasGST         bool    `json:"has_gst"`
}

// BankDetailRecord maps to bank_details.
type BankDetailRecord struct {
	ID             string  `json:"id"`
	DistributorID  string  `json:"distributor_id"`
	AccountNumber  string  `json:"account_number"`
	IFSC           string  `json:"ifsc"`
	AccountHolder  string  `json:"account_holder"`
	BankName       *string `json:"bank_name,omitempty"`
	Branch         *string `json:"branch,omitempty"`
}

// ApplicationRecord maps to the applications table.
type ApplicationRecord struct {
	ID                  string     `json:"id"`
	DistributorID       string     `json:"distributor_id"`
	Status              string     `json:"status"`
	PaymentPreference   *string    `json:"payment_preference,omitempty"`
	ExposureClass       *string    `json:"exposure_class,omitempty"`
	IsDuplicateSuspect  bool       `json:"is_duplicate_suspect"`
	DuplicateReason     *string    `json:"duplicate_reason,omitempty"`
	SubmittedAt         *time.Time `json:"submitted_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// DistributorRepository handles all distributor-related DB operations.
type DistributorRepository struct{ db *pgxpool.Pool }

func NewDistributorRepository(db *pgxpool.Pool) *DistributorRepository {
	return &DistributorRepository{db: db}
}

// ──────────────────────────────── Distributors ────────────────────────────────

func (r *DistributorRepository) GetByMobile(ctx context.Context, mobile string) (*DistributorRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, mobile, email, name, is_active, created_at, updated_at
		 FROM distributors WHERE mobile = $1`, mobile)
	return scanDistributor(row)
}

func (r *DistributorRepository) GetByID(ctx context.Context, id string) (*DistributorRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, mobile, email, name, is_active, created_at, updated_at
		 FROM distributors WHERE id = $1`, id)
	return scanDistributor(row)
}

// Create inserts a new distributor (from mobile verification) and returns the ID.
func (r *DistributorRepository) Create(ctx context.Context, mobile string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO distributors (mobile) VALUES ($1) RETURNING id`, mobile,
	).Scan(&id)
	return id, err
}

// UpdateBasic updates name and email on the distributor record.
func (r *DistributorRepository) UpdateBasic(ctx context.Context, id, name, email string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE distributors SET name = $1, email = $2, updated_at = NOW() WHERE id = $3`,
		name, email, id)
	return err
}

func scanDistributor(row pgx.Row) (*DistributorRecord, error) {
	d := &DistributorRecord{}
	err := row.Scan(&d.ID, &d.Mobile, &d.Email, &d.Name, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return d, err
}

// ──────────────────────────────── Applications ────────────────────────────────

func (r *DistributorRepository) CreateApplication(ctx context.Context, distributorID string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO applications (distributor_id) VALUES ($1) RETURNING id`, distributorID,
	).Scan(&id)
	return id, err
}

func (r *DistributorRepository) GetActiveApplication(ctx context.Context, distributorID string) (*ApplicationRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, status, payment_preference, exposure_class,
		        is_duplicate_suspect, duplicate_reason, submitted_at, created_at, updated_at
		 FROM applications
		 WHERE distributor_id = $1
		   AND status NOT IN ('rejected', 'blocked', 'credit_active', 'advance_only')
		 ORDER BY created_at DESC LIMIT 1`,
		distributorID,
	)
	return scanApplication(row)
}

func (r *DistributorRepository) GetApplicationByID(ctx context.Context, id string) (*ApplicationRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, status, payment_preference, exposure_class,
		        is_duplicate_suspect, duplicate_reason, submitted_at, created_at, updated_at
		 FROM applications WHERE id = $1`, id)
	return scanApplication(row)
}

func (r *DistributorRepository) UpdateApplicationStatus(ctx context.Context, id, toStatus, actorType string, actorID *string, reason *string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get current status for event log
	var fromStatus string
	if err := tx.QueryRow(ctx,
		`SELECT status FROM applications WHERE id = $1`, id,
	).Scan(&fromStatus); err != nil {
		return err
	}

	// Update application
	if _, err := tx.Exec(ctx,
		`UPDATE applications SET status = $1::application_status, updated_at = NOW() WHERE id = $2`,
		toStatus, id,
	); err != nil {
		return err
	}

	// Append event
	if _, err := tx.Exec(ctx,
		`INSERT INTO application_events
		 (application_id, from_status, to_status, actor_type, actor_id, reason)
		 VALUES ($1, $2::application_status, $3::application_status, $4, $5, $6)`,
		id, fromStatus, toStatus, actorType, actorID, reason,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *DistributorRepository) UpdateApplicationPreference(ctx context.Context, appID, preference, exposureClass string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE applications SET payment_preference = $1::payment_preference, exposure_class = $2::exposure_class,
		        status = 'preference_submitted'::application_status, updated_at = NOW()
		 WHERE id = $3`,
		preference, exposureClass, appID)
	return err
}

func scanApplication(row pgx.Row) (*ApplicationRecord, error) {
	a := &ApplicationRecord{}
	err := row.Scan(&a.ID, &a.DistributorID, &a.Status, &a.PaymentPreference,
		&a.ExposureClass, &a.IsDuplicateSuspect, &a.DuplicateReason,
		&a.SubmittedAt, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

// ──────────────────────────────── Business Profile ────────────────────────────

func (r *DistributorRepository) UpsertBusinessProfile(ctx context.Context, p *BusinessProfileRecord) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO business_profiles
		 (distributor_id, business_name, constitution, address_line1, address_line2,
		  city, state, pin, vintage_years, fmcg_experience_years, distribution_experience_years,
		  approx_monthly_business_paise, retailer_count, serviced_retailers_wholesalers_count,
		  salesperson_count, interested_business_role, existing_brands)
		 VALUES ($1,$2,$3::constitution_type,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		 ON CONFLICT (distributor_id) DO UPDATE SET
		  business_name = EXCLUDED.business_name,
		  constitution = EXCLUDED.constitution,
		  address_line1 = EXCLUDED.address_line1,
		  address_line2 = EXCLUDED.address_line2,
		  city = EXCLUDED.city, state = EXCLUDED.state, pin = EXCLUDED.pin,
		  vintage_years = EXCLUDED.vintage_years,
		  fmcg_experience_years = EXCLUDED.fmcg_experience_years,
		  distribution_experience_years = EXCLUDED.distribution_experience_years,
		  approx_monthly_business_paise = EXCLUDED.approx_monthly_business_paise,
		  retailer_count = EXCLUDED.retailer_count,
		  serviced_retailers_wholesalers_count = EXCLUDED.serviced_retailers_wholesalers_count,
		  salesperson_count = EXCLUDED.salesperson_count,
		  interested_business_role = EXCLUDED.interested_business_role,
		  existing_brands = EXCLUDED.existing_brands,
		  updated_at = NOW()`,
		p.DistributorID, p.BusinessName, p.Constitution, p.AddressLine1, p.AddressLine2,
		p.City, p.State, p.PIN, p.VintageYears, p.FMCGExperienceYears, p.DistributionExperienceYears,
		p.ApproxMonthlyBusinessPaise, p.RetailerCount, p.ServicedRetailersWholesalersCount,
		p.SalespersonCount, p.InterestedBusinessRole, p.ExistingBrands,
	)
	return err
}

func (r *DistributorRepository) GetBusinessProfile(ctx context.Context, distributorID string) (*BusinessProfileRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, business_name, constitution::TEXT, address_line1, address_line2,
		        city, state, pin, vintage_years, fmcg_experience_years, distribution_experience_years,
		        approx_monthly_business_paise, retailer_count, serviced_retailers_wholesalers_count,
		        salesperson_count, interested_business_role, existing_brands
		 FROM business_profiles WHERE distributor_id = $1`, distributorID)
	bp := &BusinessProfileRecord{}
	err := row.Scan(&bp.ID, &bp.DistributorID, &bp.BusinessName, &bp.Constitution,
		&bp.AddressLine1, &bp.AddressLine2, &bp.City, &bp.State, &bp.PIN,
		&bp.VintageYears, &bp.FMCGExperienceYears, &bp.DistributionExperienceYears,
		&bp.ApproxMonthlyBusinessPaise, &bp.RetailerCount, &bp.ServicedRetailersWholesalersCount,
		&bp.SalespersonCount, &bp.InterestedBusinessRole, &bp.ExistingBrands)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return bp, err
}

// ──────────────────────────────── Business Documents ─────────────────────────

func (r *DistributorRepository) UpsertBusinessDocuments(ctx context.Context, d *BusinessDocumentRecord) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO business_documents
		 (distributor_id, pan, gst_number, fssai_number, udyam_number, shop_est_number, has_gst)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (distributor_id) DO UPDATE SET
		  pan = EXCLUDED.pan, gst_number = EXCLUDED.gst_number,
		  fssai_number = EXCLUDED.fssai_number, udyam_number = EXCLUDED.udyam_number,
		  shop_est_number = EXCLUDED.shop_est_number, has_gst = EXCLUDED.has_gst,
		  updated_at = NOW()`,
		d.DistributorID, d.PAN, d.GSTNumber, d.FSSAINumber, d.UdyamNumber, d.ShopEstNumber, d.HasGST,
	)
	return err
}

func (r *DistributorRepository) GetBusinessDocuments(ctx context.Context, distributorID string) (*BusinessDocumentRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, pan, gst_number, fssai_number, udyam_number, shop_est_number, has_gst
		 FROM business_documents WHERE distributor_id = $1`, distributorID)
	bd := &BusinessDocumentRecord{}
	err := row.Scan(&bd.ID, &bd.DistributorID, &bd.PAN, &bd.GSTNumber,
		&bd.FSSAINumber, &bd.UdyamNumber, &bd.ShopEstNumber, &bd.HasGST)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return bd, err
}

// ──────────────────────────────── Bank Details ────────────────────────────────

func (r *DistributorRepository) UpsertBankDetails(ctx context.Context, b *BankDetailRecord) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO bank_details
		 (distributor_id, account_number, ifsc, account_holder, bank_name, branch)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (distributor_id) DO UPDATE SET
		  account_number = EXCLUDED.account_number, ifsc = EXCLUDED.ifsc,
		  account_holder = EXCLUDED.account_holder, bank_name = EXCLUDED.bank_name,
		  branch = EXCLUDED.branch, updated_at = NOW()`,
		b.DistributorID, b.AccountNumber, b.IFSC, b.AccountHolder, b.BankName, b.Branch,
	)
	return err
}

func (r *DistributorRepository) GetBankDetails(ctx context.Context, distributorID string) (*BankDetailRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, account_number, ifsc, account_holder, bank_name, branch
		 FROM bank_details WHERE distributor_id = $1`, distributorID)
	b := &BankDetailRecord{}
	err := row.Scan(&b.ID, &b.DistributorID, &b.AccountNumber, &b.IFSC, &b.AccountHolder, &b.BankName, &b.Branch)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return b, err
}

// ──────────────────────────────── Duplicate Detection ────────────────────────

// FindByPAN returns a distributor ID if a record with the given PAN exists.
func (r *DistributorRepository) FindByPAN(ctx context.Context, pan string) (*string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`SELECT d.id FROM distributors d
		 JOIN business_documents bd ON bd.distributor_id = d.id
		 WHERE bd.pan = $1 LIMIT 1`, pan,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &id, err
}

// FindByGST returns a distributor ID if a record with the given GST exists.
func (r *DistributorRepository) FindByGST(ctx context.Context, gst string) (*string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`SELECT d.id FROM distributors d
		 JOIN business_documents bd ON bd.distributor_id = d.id
		 WHERE bd.gst_number = $1 LIMIT 1`, gst,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &id, err
}

// FindByBankAccount returns a distributor ID if a matching bank account exists.
func (r *DistributorRepository) FindByBankAccount(ctx context.Context, accountNumber, ifsc string) (*string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`SELECT distributor_id FROM bank_details
		 WHERE account_number = $1 AND ifsc = $2 LIMIT 1`, accountNumber, ifsc,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &id, err
}

// RecordDuplicateSuspect inserts a duplicate suspect pair.
func (r *DistributorRepository) RecordDuplicateSuspect(ctx context.Context, appA, appB string, matchFields []string, confidence string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO duplicate_suspects (application_id_a, application_id_b, match_fields, confidence)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT DO NOTHING`,
		appA, appB, matchFields, confidence)
	return err
}

// MarkApplicationDuplicate flags an application as a duplicate suspect.
func (r *DistributorRepository) MarkApplicationDuplicate(ctx context.Context, appID, reason string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE applications SET is_duplicate_suspect = TRUE, duplicate_reason = $1 WHERE id = $2`,
		reason, appID)
	return err
}

// ──────────────────────────────── Consents ───────────────────────────────────

func (r *DistributorRepository) RecordConsent(ctx context.Context, distributorID, mobile, consentType, consentText, consentVersion, ip, userAgent string) error {
	var ipPtr *string
	if ip != "" {
		host, _, err := net.SplitHostPort(ip)
		if err != nil {
			host = ip
		}
		if parsed := net.ParseIP(host); parsed != nil {
			ipPtr = &host
		}
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO consents
		 (distributor_id, mobile, consent_type, consent_text, consent_version, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		distributorID, mobile, consentType, consentText, consentVersion, ipPtr, userAgent)
	return err
}

// ListAll returns distributors with pagination for admin views.
func (r *DistributorRepository) ListAll(ctx context.Context, limit, offset int) ([]DistributorRecord, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM distributors`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, mobile, email, name, is_active, created_at, updated_at
		 FROM distributors ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []DistributorRecord
	for rows.Next() {
		d := DistributorRecord{}
		if err := rows.Scan(&d.ID, &d.Mobile, &d.Email, &d.Name, &d.IsActive,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, d)
	}
	return list, total, nil
}

type ApplicationSummaryRecord struct {
	ID                 string     `json:"id"`
	DistributorID      string     `json:"distributor_id"`
	DistributorName    *string    `json:"distributor_name"`
	DistributorMobile  string     `json:"distributor_mobile"`
	BusinessName       *string    `json:"business_name"`
	Status             string     `json:"status"`
	IsDuplicateSuspect bool       `json:"is_duplicate_suspect"`
	SubmittedAt        *time.Time `json:"submitted_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

func (r *DistributorRepository) ListApplications(ctx context.Context, statusFilter string, limit, offset int) ([]ApplicationSummaryRecord, int, error) {
	queryCount := `SELECT COUNT(*) FROM applications`
	querySelect := `SELECT a.id, a.distributor_id, d.name, d.mobile, bp.business_name,
	                       a.status, a.is_duplicate_suspect, a.submitted_at, a.created_at
	                FROM applications a
	                JOIN distributors d ON a.distributor_id = d.id
	                LEFT JOIN business_profiles bp ON bp.distributor_id = d.id`
	args := []interface{}{}

	if statusFilter != "" && statusFilter != "all" {
		queryCount += ` WHERE status = $1`
		querySelect += ` WHERE a.status = $1`
		args = append(args, statusFilter)
	}

	var total int
	if err := r.db.QueryRow(ctx, queryCount, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySelect += ` ORDER BY a.created_at DESC`
	if len(args) > 0 {
		querySelect += ` LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		querySelect += ` LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, querySelect, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []ApplicationSummaryRecord
	for rows.Next() {
		app := ApplicationSummaryRecord{}
		if err := rows.Scan(&app.ID, &app.DistributorID, &app.DistributorName, &app.DistributorMobile,
			&app.BusinessName, &app.Status, &app.IsDuplicateSuspect, &app.SubmittedAt, &app.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, app)
	}
	return list, total, nil
}
