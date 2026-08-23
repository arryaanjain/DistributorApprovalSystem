package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreditDecisionRecord represents a row in credit_decisions.
type CreditDecisionRecord struct {
	ID                 string
	ApplicationID      string
	DistributorID      string
	PolicyVersionID    *string
	TotalScore         int
	RiskGrade          string
	Decision           string // APPROVED, ADVANCE_ONLY, MANUAL_REVIEW, REJECTED
	ApprovedLimitPaise int64
	ApprovedPeriodDays int
	MaxOutstandingAge  int
	PaymentTerms       string
	NonGSTCapped       bool
	HardRiskTriggered  bool
	DecidedBy          string // SYSTEM or user_id
	DecidedAt          time.Time
}

// CreditOfferRecord represents a row in credit_offers.
type CreditOfferRecord struct {
	ID                 string
	DecisionID         string
	DistributorID      string
	OfferedLimitPaise  int64
	OfferedPeriodDays  int
	PaymentTerms       string
	Status             string // PENDING, ACCEPTED, DECLINED, EXPIRED
	AcceptedAt         *time.Time
	DeclinedAt         *time.Time
	ExpiresAt          time.Time
	CreatedAt          time.Time
}

// AgreementRecord represents a row in distributor_agreements.
type AgreementRecord struct {
	ID                  string
	DistributorID       string
	ApplicationID       string
	AgreementNumber     string
	Version             string
	ApprovedLimitPaise  int64
	ApprovedPeriodDays  int
	Status              string // DRAFT, GENERATED, SENT_FOR_SIGNATURE, SIGNED, EXPIRED
	DocumentURL         *string
	SignedAt            *time.Time
	EsignProviderRef    *string
	CreatedAt           time.Time
}

type CreditRepository struct {
	db *pgxpool.Pool
}

func NewCreditRepository(db *pgxpool.Pool) *CreditRepository {
	return &CreditRepository{db: db}
}

// SaveScore saves the calculated score and score components.
func (r *CreditRepository) SaveScore(ctx context.Context, distributorID, appID string, totalScore int, riskGrade string, components map[string]int) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var scoreID string
	err = tx.QueryRow(ctx,
		`INSERT INTO credit_scores (distributor_id, application_id, total_score, risk_grade)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		distributorID, appID, totalScore, riskGrade,
	).Scan(&scoreID)
	if err != nil {
		return "", err
	}

	for param, val := range components {
		_, err := tx.Exec(ctx,
			`INSERT INTO credit_score_components (credit_score_id, parameter_name, weight, score_earned)
			 VALUES ($1, $2, 0, $3)`,
			scoreID, param, val,
		)
		if err != nil {
			return "", err
		}
	}

	return scoreID, tx.Commit(ctx)
}

// SaveRiskFlags records any hard flags triggered.
func (r *CreditRepository) SaveRiskFlags(ctx context.Context, distributorID, appID string, flags []string) error {
	for _, flag := range flags {
		_, err := r.db.Exec(ctx,
			`INSERT INTO risk_flags (distributor_id, application_id, flag_code, severity, description)
			 VALUES ($1, $2, $3, 'CRITICAL', 'Hard risk rule triggered')
			 ON CONFLICT DO NOTHING`,
			distributorID, appID, flag,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveDecision inserts a formal credit decision.
func (r *CreditRepository) SaveDecision(ctx context.Context, d *CreditDecisionRecord) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO credit_decisions
		 (application_id, distributor_id, policy_version_id, total_score, risk_grade,
		  decision, approved_limit_paise, approved_period_days, max_outstanding_age,
		  payment_terms, non_gst_capped, hard_risk_triggered, decided_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id`,
		d.ApplicationID, d.DistributorID, d.PolicyVersionID, d.TotalScore, d.RiskGrade,
		d.Decision, d.ApprovedLimitPaise, d.ApprovedPeriodDays, d.MaxOutstandingAge,
		d.PaymentTerms, d.NonGSTCapped, d.HardRiskTriggered, d.DecidedBy,
	).Scan(&id)
	return id, err
}

func (r *CreditRepository) GetDecisionByAppID(ctx context.Context, appID string) (*CreditDecisionRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, application_id, distributor_id, policy_version_id, total_score, risk_grade,
		        decision, approved_limit_paise, approved_period_days, max_outstanding_age,
		        payment_terms, non_gst_capped, hard_risk_triggered, decided_by, decided_at
		 FROM credit_decisions WHERE application_id = $1 ORDER BY decided_at DESC LIMIT 1`, appID)
	d := &CreditDecisionRecord{}
	err := row.Scan(&d.ID, &d.ApplicationID, &d.DistributorID, &d.PolicyVersionID,
		&d.TotalScore, &d.RiskGrade, &d.Decision, &d.ApprovedLimitPaise, &d.ApprovedPeriodDays,
		&d.MaxOutstandingAge, &d.PaymentTerms, &d.NonGSTCapped, &d.HardRiskTriggered,
		&d.DecidedBy, &d.DecidedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return d, err
}

// ──────────────────────────────── Offers ─────────────────────────────────────

func (r *CreditRepository) CreateOffer(ctx context.Context, o *CreditOfferRecord) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO credit_offers
		 (decision_id, distributor_id, offered_limit_paise, offered_period_days, payment_terms, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		o.DecisionID, o.DistributorID, o.OfferedLimitPaise, o.OfferedPeriodDays, o.PaymentTerms, o.ExpiresAt,
	).Scan(&id)
	return id, err
}

func (r *CreditRepository) GetActiveOfferByDistributor(ctx context.Context, distributorID string) (*CreditOfferRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, decision_id, distributor_id, offered_limit_paise, offered_period_days,
		        payment_terms, status, accepted_at, declined_at, expires_at, created_at
		 FROM credit_offers WHERE distributor_id = $1 ORDER BY created_at DESC LIMIT 1`, distributorID)
	o := &CreditOfferRecord{}
	err := row.Scan(&o.ID, &o.DecisionID, &o.DistributorID, &o.OfferedLimitPaise,
		&o.OfferedPeriodDays, &o.PaymentTerms, &o.Status, &o.AcceptedAt, &o.DeclinedAt,
		&o.ExpiresAt, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return o, err
}

func (r *CreditRepository) UpdateOfferStatus(ctx context.Context, offerID, status string) error {
	var query string
	if status == "ACCEPTED" {
		query = `UPDATE credit_offers SET status = $1, accepted_at = NOW() WHERE id = $2`
	} else if status == "DECLINED" {
		query = `UPDATE credit_offers SET status = $1, declined_at = NOW() WHERE id = $2`
	} else {
		query = `UPDATE credit_offers SET status = $1 WHERE id = $2`
	}
	_, err := r.db.Exec(ctx, query, status, offerID)
	return err
}

// ──────────────────────────────── Agreements ─────────────────────────────────

func (r *CreditRepository) CreateAgreement(ctx context.Context, a *AgreementRecord) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO distributor_agreements
		 (distributor_id, application_id, agreement_number, version, approved_limit_paise, approved_period_days)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		a.DistributorID, a.ApplicationID, a.AgreementNumber, a.Version, a.ApprovedLimitPaise, a.ApprovedPeriodDays,
	).Scan(&id)
	return id, err
}

func (r *CreditRepository) GetAgreementByDistributor(ctx context.Context, distributorID string) (*AgreementRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, distributor_id, application_id, agreement_number, version,
		        approved_limit_paise, approved_period_days, status, document_url, signed_at, esign_provider_ref, created_at
		 FROM distributor_agreements WHERE distributor_id = $1 ORDER BY created_at DESC LIMIT 1`, distributorID)
	a := &AgreementRecord{}
	err := row.Scan(&a.ID, &a.DistributorID, &a.ApplicationID, &a.AgreementNumber,
		&a.Version, &a.ApprovedLimitPaise, &a.ApprovedPeriodDays, &a.Status, &a.DocumentURL,
		&a.SignedAt, &a.EsignProviderRef, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (r *CreditRepository) SignAgreement(ctx context.Context, agreementID, providerRef string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE distributor_agreements
		 SET status = 'SIGNED', signed_at = NOW(), esign_provider_ref = $1
		 WHERE id = $2`,
		providerRef, agreementID)
	return err
}
