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
	ID                 string    `json:"id"`
	ApplicationID      string    `json:"application_id"`
	DistributorID      string    `json:"distributor_id"`
	CreditScoreID      *string   `json:"credit_score_id,omitempty"`
	PolicyVersion      string    `json:"policy_version"`
	TotalScore         int       `json:"total_score"`
	RiskGrade          string    `json:"risk_grade"`
	Decision           string    `json:"decision"`
	ApprovedLimitPaise int64     `json:"approved_limit_paise"`
	ApprovedPeriodDays int       `json:"approved_period_days"`
	MaxOutstandingAge  int       `json:"max_outstanding_age"`
	PaymentTerms       string    `json:"payment_terms"`
	NonGSTCapped       bool      `json:"non_gst_capped"`
	HardRiskTriggered  bool      `json:"hard_risk_triggered"`
	DecidedBy          string    `json:"decided_by"`
	DecidedAt          time.Time `json:"decided_at"`
}

// CreditOfferRecord represents a row in credit_offers.
type CreditOfferRecord struct {
	ID                 string     `json:"id"`
	ApplicationID      string     `json:"application_id"`
	DecisionID         string     `json:"decision_id"`
	DistributorID      string     `json:"distributor_id"`
	RiskGrade          string     `json:"risk_grade"`
	OfferedLimitPaise  int64      `json:"offered_limit_paise"`
	OfferedPeriodDays  int        `json:"offered_period_days"`
	MaxOutstandingAge  int        `json:"max_outstanding_age"`
	PaymentTerms       string     `json:"payment_terms"`
	Status             string     `json:"status"`
	AcceptedAt         *time.Time `json:"accepted_at,omitempty"`
	DeclinedAt         *time.Time `json:"declined_at,omitempty"`
	ExpiresAt          time.Time  `json:"expires_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

// AgreementRecord represents a row in distributor_agreements.
type AgreementRecord struct {
	ID                 string     `json:"id"`
	DistributorID      string     `json:"distributor_id"`
	ApplicationID      string     `json:"application_id"`
	AgreementNumber    string     `json:"agreement_number"`
	Version            string     `json:"version"`
	ApprovedLimitPaise int64      `json:"approved_limit_paise"`
	ApprovedPeriodDays int        `json:"approved_period_days"`
	Status             string     `json:"status"`
	DocumentURL        *string    `json:"document_url,omitempty"`
	SignedAt           *time.Time `json:"signed_at,omitempty"`
	EsignProviderRef   *string    `json:"esign_provider_ref,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
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

	rg := "A"
	switch riskGrade {
	case "GRADE_A_PLUS", "GRADE_A":
		rg = "A"
	case "GRADE_B":
		rg = "B"
	case "GRADE_C":
		rg = "C"
	case "GRADE_HIGH_RISK":
		rg = "D"
	default:
		rg = "A"
	}

	var scoreID string
	err = tx.QueryRow(ctx,
		`INSERT INTO credit_scores (distributor_id, application_id, policy_version, total_score, risk_grade, inputs)
		 VALUES ($1, $2, 'v1.0', $3, $4::risk_grade, '{}'::jsonb) RETURNING id`,
		distributorID, appID, totalScore, rg,
	).Scan(&scoreID)
	if err != nil {
		return "", err
	}

	for param, val := range components {
		_, err := tx.Exec(ctx,
			`INSERT INTO credit_score_components (credit_score_id, component_name, weight, weighted_score)
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
	// Deactivate existing active risk flags for this distributor to support dynamic re-evaluation
	_, _ = r.db.Exec(ctx, `UPDATE risk_flags SET is_active = false WHERE distributor_id = $1`, distributorID)

	for _, flag := range flags {
		_, err := r.db.Exec(ctx,
			`INSERT INTO risk_flags (distributor_id, application_id, flag_code, flag_description, severity, triggered_by)
			 VALUES ($1, $2, $3, 'Hard risk rule triggered', 'hard', 'system')`,
			distributorID, appID, flag,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveDecision inserts a formal credit decision matching PostgreSQL schema.
func (r *CreditRepository) SaveDecision(ctx context.Context, d *CreditDecisionRecord) (string, error) {
	eligibility := "credit"
	if d.Decision == "ADVANCE_ONLY" {
		eligibility = "advance_only"
	} else if d.Decision == "REFER_MANUAL" {
		eligibility = "hold"
	} else if d.Decision == "REJECT" {
		eligibility = "blocked"
	}

	approvedPeriod := "15_days"
	if d.ApprovedPeriodDays == 30 {
		approvedPeriod = "30_days"
	} else if d.ApprovedPeriodDays == 0 {
		approvedPeriod = "cod"
	}

	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO credit_decisions
		 (application_id, distributor_id, credit_score_id, policy_version, eligibility,
		  approved_limit_paise, approved_period, max_outstanding_days, hard_flags_present, decision_source)
		 VALUES ($1, $2, $3, 'v1.0', $4::eligibility_decision, $5, $6::credit_period_code, $7, $8, 'auto')
		 RETURNING id`,
		d.ApplicationID, d.DistributorID, d.CreditScoreID, eligibility,
		d.ApprovedLimitPaise, approvedPeriod, d.MaxOutstandingAge, d.HardRiskTriggered,
	).Scan(&id)
	return id, err
}

func (r *CreditRepository) GetDecisionByAppID(ctx context.Context, appID string) (*CreditDecisionRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT d.id, d.application_id, d.distributor_id, d.credit_score_id, d.policy_version, d.eligibility,
		        d.approved_limit_paise, d.approved_period, d.max_outstanding_days, d.hard_flags_present, d.decided_at,
		        COALESCE(s.total_score, 0), COALESCE(s.risk_grade::TEXT, '')
		 FROM credit_decisions d
		 LEFT JOIN credit_scores s ON d.credit_score_id = s.id
		 WHERE d.application_id = $1 ORDER BY d.decided_at DESC LIMIT 1`, appID)
	d := &CreditDecisionRecord{}
	var eligibilityStr, periodStr string
	err := row.Scan(&d.ID, &d.ApplicationID, &d.DistributorID, &d.CreditScoreID, &d.PolicyVersion,
		&eligibilityStr, &d.ApprovedLimitPaise, &periodStr,
		&d.MaxOutstandingAge, &d.HardRiskTriggered, &d.DecidedAt,
		&d.TotalScore, &d.RiskGrade)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	d.Decision = eligibilityStr
	d.PaymentTerms = periodStr
	return d, err
}

// ──────────────────────────────── Offers ─────────────────────────────────────

func (r *CreditRepository) CreateOffer(ctx context.Context, o *CreditOfferRecord) (string, error) {
	var id string
	riskGrade := o.RiskGrade
	if riskGrade != "A" && riskGrade != "B" && riskGrade != "C" && riskGrade != "D" && riskGrade != "E" {
		riskGrade = "A"
	}
	periodCode := "15_days"
	if o.OfferedPeriodDays == 30 {
		periodCode = "30_days"
	} else if o.OfferedPeriodDays == 0 {
		periodCode = "cod"
	}

	err := r.db.QueryRow(ctx,
		`INSERT INTO credit_offers
		 (application_id, distributor_id, credit_decision_id, risk_grade, offered_limit_paise, offered_period, max_outstanding_days, expires_at)
		 VALUES ($1, $2, $3, $4::risk_grade, $5, $6::credit_period_code, $7, $8) RETURNING id`,
		o.ApplicationID, o.DistributorID, o.DecisionID, riskGrade, o.OfferedLimitPaise, periodCode, o.MaxOutstandingAge, o.ExpiresAt,
	).Scan(&id)
	return id, err
}

func (r *CreditRepository) GetActiveOfferByDistributor(ctx context.Context, distributorID string) (*CreditOfferRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, credit_decision_id, distributor_id, risk_grade, offered_limit_paise,
		        offered_period, decision, accepted_at, declined_at, expires_at, created_at
		 FROM credit_offers WHERE distributor_id = $1 ORDER BY created_at DESC LIMIT 1`, distributorID)
	o := &CreditOfferRecord{}
	var riskGradeStr, periodStr string
	err := row.Scan(&o.ID, &o.DecisionID, &o.DistributorID, &riskGradeStr, &o.OfferedLimitPaise,
		&periodStr, &o.Status, &o.AcceptedAt, &o.DeclinedAt,
		&o.ExpiresAt, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.RiskGrade = riskGradeStr
	o.PaymentTerms = periodStr
	return o, nil
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

func (r *CreditRepository) GetActiveRiskFlags(ctx context.Context, distributorID string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT flag_code FROM risk_flags WHERE distributor_id = $1 AND is_active = true`, distributorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var flags []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err == nil {
			flags = append(flags, code)
		}
	}
	return flags, nil
}

func (r *CreditRepository) GetScoreComponents(ctx context.Context, appID string) (map[string]int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT c.component_name, c.weighted_score 
		 FROM credit_score_components c
		 JOIN credit_scores s ON c.credit_score_id = s.id
		 WHERE s.application_id = $1`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comps := make(map[string]int)
	for rows.Next() {
		var name string
		var score int
		if err := rows.Scan(&name, &score); err == nil {
			comps[name] = score
		}
	}
	return comps, nil
}

func (r *CreditRepository) SignAgreement(ctx context.Context, agreementID, providerRef string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE distributor_agreements
		 SET status = 'SIGNED', signed_at = NOW(), esign_provider_ref = $1
		 WHERE id = $2`,
		providerRef, agreementID)
	return err
}

type DashboardStats struct {
	TotalApplications     int   `json:"total_applications"`
	PendingVerifications  int   `json:"pending_verifications"`
	TotalDistributors     int   `json:"total_distributors"`
	SanctionedCreditPaise int64 `json:"sanctioned_credit_paise"`
	UtilizedCreditPaise   int64 `json:"utilized_credit_paise"`
	AvailableCreditPaise  int64 `json:"available_credit_paise"`
}

func (r *CreditRepository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM applications`).Scan(&stats.TotalApplications)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM applications WHERE status IN ('submitted', 'basic_submitted', 'business_submitted', 'preference_submitted', 'statutory_submitted', 'consent_given', 'under_review', 'hold')`).Scan(&stats.PendingVerifications)
	
	// Count active onboarding-completed distributors
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM distributors WHERE is_active = true`).Scan(&stats.TotalDistributors)
	if stats.TotalDistributors == 0 {
		_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM distributors`).Scan(&stats.TotalDistributors)
	}

	// Calculate credit account stats if accounts exist
	_ = r.db.QueryRow(ctx, `SELECT COALESCE(SUM(approved_limit_paise), 0), COALESCE(SUM(current_credit_paise), 0), COALESCE(SUM(available_credit_paise), 0) FROM credit_accounts`).Scan(&stats.SanctionedCreditPaise, &stats.UtilizedCreditPaise, &stats.AvailableCreditPaise)

	// Fallback to summing only the latest decision per distributor (preventing historical duplicates)
	if stats.SanctionedCreditPaise == 0 {
		query := `
			SELECT COALESCE(SUM(approved_limit_paise), 0)
			FROM (
				SELECT DISTINCT ON (distributor_id) approved_limit_paise
				FROM credit_decisions
				WHERE eligibility = 'credit'
				ORDER BY distributor_id, decided_at DESC
			) latest_decisions
		`
		_ = r.db.QueryRow(ctx, query).Scan(&stats.SanctionedCreditPaise)
	}

	return stats, nil
}

// GetCreditDecisionTrail returns all historical credit decisions for a distributor ordered by decision date DESC.
func (r *CreditRepository) GetCreditDecisionTrail(ctx context.Context, distributorID string) ([]CreditDecisionRecord, error) {
	rows, err := r.db.Query(ctx,
		`SELECT d.id, d.application_id, d.distributor_id, d.credit_score_id, d.policy_version, d.eligibility,
		        d.approved_limit_paise, d.approved_period, d.max_outstanding_days, d.hard_flags_present, d.decided_at,
		        COALESCE(s.total_score, 0), COALESCE(s.risk_grade::TEXT, '')
		 FROM credit_decisions d
		 LEFT JOIN credit_scores s ON d.credit_score_id = s.id
		 WHERE d.distributor_id = $1
		 ORDER BY d.decided_at DESC`, distributorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trail []CreditDecisionRecord
	for rows.Next() {
		d := CreditDecisionRecord{}
		var eligibilityStr, periodStr string
		err := rows.Scan(&d.ID, &d.ApplicationID, &d.DistributorID, &d.CreditScoreID, &d.PolicyVersion,
			&eligibilityStr, &d.ApprovedLimitPaise, &periodStr,
			&d.MaxOutstandingAge, &d.HardRiskTriggered, &d.DecidedAt,
			&d.TotalScore, &d.RiskGrade)
		if err != nil {
			return nil, err
		}
		d.Decision = eligibilityStr
		d.PaymentTerms = periodStr
		trail = append(trail, d)
	}
	return trail, nil
}

