package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VerificationStatus mirrors the DB enum.
type VerificationStatus string

const (
	VerificationPending          VerificationStatus = "pending"
	VerificationVerified         VerificationStatus = "verified"
	VerificationPartiallyVerified VerificationStatus = "partially_verified"
	VerificationMismatch         VerificationStatus = "mismatch"
	VerificationFailed           VerificationStatus = "failed"
	VerificationUnavailable      VerificationStatus = "unavailable"
)

// PANVerificationRecord maps to pan_verifications.
type PANVerificationRecord struct {
	ID            string             `json:"id"`
	DistributorID string             `json:"distributor_id"`
	ApplicationID *string            `json:"application_id,omitempty"`
	PAN           string             `json:"pan"`
	Status        VerificationStatus `json:"status"`
	NameOnPAN     *string            `json:"name_on_pan,omitempty"`
	NameMatch     *bool              `json:"name_match,omitempty"`
	ProviderRef   *string            `json:"provider_ref,omitempty"`
	VerifiedAt    *time.Time         `json:"verified_at,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
}

// GSTVerificationRecord maps to gst_verifications.
type GSTVerificationRecord struct {
	ID               string             `json:"id"`
	DistributorID    string             `json:"distributor_id"`
	ApplicationID    *string            `json:"application_id,omitempty"`
	GSTNumber        string             `json:"gst_number"`
	Status           VerificationStatus `json:"status"`
	LegalName        *string            `json:"legal_name,omitempty"`
	TradeName        *string            `json:"trade_name,omitempty"`
	RegistrationDate *time.Time         `json:"registration_date,omitempty"`
	GSTStatus        *string            `json:"gst_status,omitempty"`
	Address          *string            `json:"address,omitempty"`
	Constitution     *string            `json:"constitution,omitempty"`
	NameMatch        *bool              `json:"name_match,omitempty"`
	ProviderRef      *string            `json:"provider_ref,omitempty"`
	VerifiedAt       *time.Time         `json:"verified_at,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
}

// BankVerificationRecord maps to bank_verifications.
type BankVerificationRecord struct {
	ID            string             `json:"id"`
	DistributorID string             `json:"distributor_id"`
	ApplicationID *string            `json:"application_id,omitempty"`
	AccountNumber string             `json:"account_number"`
	IFSC          string             `json:"ifsc"`
	Status        VerificationStatus `json:"status"`
	AccountHolder *string            `json:"account_holder,omitempty"`
	NameMatch     *bool              `json:"name_match,omitempty"`
	BankName      *string            `json:"bank_name,omitempty"`
	ProviderRef   *string            `json:"provider_ref,omitempty"`
	VerifiedAt    *time.Time         `json:"verified_at,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
}

// CreditReportRecord maps to credit_reports.
type CreditReportRecord struct {
	ID               string     `json:"id"`
	DistributorID    string     `json:"distributor_id"`
	ApplicationID    *string    `json:"application_id,omitempty"`
	PAN              *string    `json:"pan,omitempty"`
	Mobile           *string    `json:"mobile,omitempty"`
	BureauScore      *int       `json:"bureau_score,omitempty"`
	HasDefaults      *bool      `json:"has_defaults,omitempty"`
	HasWriteoffs     *bool      `json:"has_writeoffs,omitempty"`
	HasSettlements   *bool      `json:"has_settlements,omitempty"`
	TotalActiveLoans *int64     `json:"total_active_loans,omitempty"`
	DelinquencyCount *int       `json:"delinquency_count,omitempty"`
	FraudFlag        bool       `json:"fraud_flag"`
	ReportDate       *time.Time `json:"report_date,omitempty"`
	PDFURL           *string    `json:"pdf_url,omitempty"`
	ProviderRef      *string    `json:"provider_ref,omitempty"`
	FetchedAt        *time.Time `json:"fetched_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// VerificationRepository handles all verification table operations.
type VerificationRepository struct{ db *pgxpool.Pool }

func NewVerificationRepository(db *pgxpool.Pool) *VerificationRepository {
	return &VerificationRepository{db: db}
}

// ──────────────────────────────── PAN ─────────────────────────────────────────

func (r *VerificationRepository) CreatePANVerification(ctx context.Context, distributorID string, appID *string, pan string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO pan_verifications (distributor_id, application_id, pan)
		 VALUES ($1, $2, $3) RETURNING id`, distributorID, appID, pan,
	).Scan(&id)
	return id, err
}

func (r *VerificationRepository) UpdatePANVerification(ctx context.Context, id string, status VerificationStatus, nameOnPAN *string, nameMatch *bool, rawResp []byte, providerRef *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE pan_verifications SET
		 status = $1::verification_status, name_on_pan = $2, name_match = $3,
		 raw_response = $4, provider_ref = $5,
		 verified_at = CASE WHEN $1::TEXT = 'verified' THEN NOW() ELSE NULL END
		 WHERE id = $6`,
		status, nameOnPAN, nameMatch, rawResp, providerRef, id)
	return err
}

func (r *VerificationRepository) GetLatestPANVerification(ctx context.Context, distributorID string) (*PANVerificationRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, application_id, pan, status::TEXT, name_on_pan, name_match, provider_ref, verified_at, created_at
		 FROM pan_verifications WHERE distributor_id = $1 ORDER BY created_at DESC LIMIT 1`, distributorID)
	v := &PANVerificationRecord{}
	var statusStr string
	err := row.Scan(&v.ID, &v.DistributorID, &v.ApplicationID, &v.PAN, &statusStr,
		&v.NameOnPAN, &v.NameMatch, &v.ProviderRef, &v.VerifiedAt, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	v.Status = VerificationStatus(statusStr)
	return v, err
}

// ──────────────────────────────── GST ─────────────────────────────────────────

func (r *VerificationRepository) CreateGSTVerification(ctx context.Context, distributorID string, appID *string, gst string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO gst_verifications (distributor_id, application_id, gst_number)
		 VALUES ($1, $2, $3) RETURNING id`, distributorID, appID, gst,
	).Scan(&id)
	return id, err
}

func (r *VerificationRepository) UpdateGSTVerification(ctx context.Context, id string, status VerificationStatus,
	legalName, tradeName *string, regDate *time.Time, gstStatus, address, constitution *string,
	nameMatch *bool, rawResp []byte, providerRef *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE gst_verifications SET
		 status = $1::verification_status, legal_name = $2, trade_name = $3, registration_date = $4,
		 gst_status = $5, address = $6, constitution = $7, name_match = $8,
		 raw_response = $9, provider_ref = $10,
		 verified_at = CASE WHEN $1::TEXT = 'verified' THEN NOW() ELSE NULL END
		 WHERE id = $11`,
		status, legalName, tradeName, regDate, gstStatus, address, constitution,
		nameMatch, rawResp, providerRef, id)
	return err
}

func (r *VerificationRepository) GetLatestGSTVerification(ctx context.Context, distributorID string) (*GSTVerificationRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, application_id, gst_number, status::TEXT,
		        legal_name, trade_name, registration_date, gst_status, address, constitution, name_match, provider_ref, verified_at, created_at
		 FROM gst_verifications WHERE distributor_id = $1 ORDER BY created_at DESC LIMIT 1`, distributorID)
	v := &GSTVerificationRecord{}
	var statusStr string
	err := row.Scan(&v.ID, &v.DistributorID, &v.ApplicationID, &v.GSTNumber, &statusStr,
		&v.LegalName, &v.TradeName, &v.RegistrationDate, &v.GSTStatus, &v.Address,
		&v.Constitution, &v.NameMatch, &v.ProviderRef, &v.VerifiedAt, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	v.Status = VerificationStatus(statusStr)
	return v, err
}

// ──────────────────────────────── Bank ────────────────────────────────────────

func (r *VerificationRepository) CreateBankVerification(ctx context.Context, distributorID string, appID *string, accountNumber, ifsc string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO bank_verifications (distributor_id, application_id, account_number, ifsc)
		 VALUES ($1, $2, $3, $4) RETURNING id`, distributorID, appID, accountNumber, ifsc,
	).Scan(&id)
	return id, err
}

func (r *VerificationRepository) UpdateBankVerification(ctx context.Context, id string, status VerificationStatus,
	accountHolder, bankName *string, nameMatch *bool, rawResp []byte, providerRef *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE bank_verifications SET
		 status = $1::verification_status, account_holder = $2, bank_name = $3, name_match = $4,
		 raw_response = $5, provider_ref = $6,
		 verified_at = CASE WHEN $1::TEXT = 'verified' THEN NOW() ELSE NULL END
		 WHERE id = $7`,
		status, accountHolder, bankName, nameMatch, rawResp, providerRef, id)
	return err
}

func (r *VerificationRepository) GetLatestBankVerification(ctx context.Context, distributorID string) (*BankVerificationRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, application_id, account_number, ifsc,
		        status::TEXT, account_holder, name_match, bank_name, provider_ref, verified_at, created_at
		 FROM bank_verifications WHERE distributor_id = $1 ORDER BY created_at DESC LIMIT 1`, distributorID)
	v := &BankVerificationRecord{}
	var statusStr string
	err := row.Scan(&v.ID, &v.DistributorID, &v.ApplicationID, &v.AccountNumber, &v.IFSC,
		&statusStr, &v.AccountHolder, &v.NameMatch, &v.BankName, &v.ProviderRef, &v.VerifiedAt, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	v.Status = VerificationStatus(statusStr)
	return v, err
}

// ──────────────────────────────── Credit Report ───────────────────────────────

func (r *VerificationRepository) CreateCreditReport(ctx context.Context, distributorID string, appID *string, pan, mobile *string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO credit_reports (distributor_id, application_id, pan, mobile)
		 VALUES ($1, $2, $3, $4) RETURNING id`, distributorID, appID, pan, mobile,
	).Scan(&id)
	return id, err
}

func (r *VerificationRepository) UpdateCreditReport(ctx context.Context, id string, score *int, hasDefaults, hasWriteoffs, hasSettlements *bool,
	totalLoans *int64, delinquencyCount *int, fraudFlag bool, reportDate *time.Time, pdfURL, providerRef *string, rawResp []byte) error {
	_, err := r.db.Exec(ctx,
		`UPDATE credit_reports SET
		 bureau_score = $1, has_defaults = $2, has_writeoffs = $3, has_settlements = $4,
		 total_active_loans = $5, delinquency_count = $6, fraud_flag = $7,
		 report_date = $8, pdf_url = $9, provider_ref = $10, raw_response = $11,
		 fetched_at = NOW()
		 WHERE id = $12`,
		score, hasDefaults, hasWriteoffs, hasSettlements, totalLoans, delinquencyCount,
		fraudFlag, reportDate, pdfURL, providerRef, rawResp, id)
	return err
}

func (r *VerificationRepository) GetLatestCreditReport(ctx context.Context, distributorID string) (*CreditReportRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, application_id, pan, mobile, bureau_score,
		        has_defaults, has_writeoffs, has_settlements, total_active_loans,
		        delinquency_count, fraud_flag, report_date, pdf_url, provider_ref, fetched_at, created_at
		 FROM credit_reports WHERE distributor_id = $1
		 ORDER BY bureau_score IS NOT NULL DESC, created_at DESC LIMIT 1`, distributorID)
	v := &CreditReportRecord{}
	err := row.Scan(&v.ID, &v.DistributorID, &v.ApplicationID, &v.PAN, &v.Mobile,
		&v.BureauScore, &v.HasDefaults, &v.HasWriteoffs, &v.HasSettlements,
		&v.TotalActiveLoans, &v.DelinquencyCount, &v.FraudFlag, &v.ReportDate,
		&v.PDFURL, &v.ProviderRef, &v.FetchedAt, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return v, err
}

// GetAllForApplication returns the latest verification of each type for an application.
type AllVerifications struct {
	PAN          *PANVerificationRecord  `json:"pan,omitempty"`
	GST          *GSTVerificationRecord  `json:"gst,omitempty"`
	Bank         *BankVerificationRecord `json:"bank,omitempty"`
	CreditReport *CreditReportRecord     `json:"credit_report,omitempty"`
}

func (r *VerificationRepository) GetAllForApplication(ctx context.Context, distributorID string) (*AllVerifications, error) {
	all := &AllVerifications{}
	var err error
	all.PAN, err = r.GetLatestPANVerification(ctx, distributorID)
	if err != nil {
		return nil, err
	}
	all.GST, err = r.GetLatestGSTVerification(ctx, distributorID)
	if err != nil {
		return nil, err
	}
	all.Bank, err = r.GetLatestBankVerification(ctx, distributorID)
	if err != nil {
		return nil, err
	}
	all.CreditReport, err = r.GetLatestCreditReport(ctx, distributorID)
	if err != nil {
		return nil, err
	}
	return all, nil
}
