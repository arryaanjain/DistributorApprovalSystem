package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OTPRecord is a row from otp_verifications.
type OTPRecord struct {
	ID         string
	Mobile     string
	OTPHash    string
	Purpose    string
	ExpiresAt  time.Time
	Attempts   int
	VerifiedAt *time.Time
	CreatedAt  time.Time
}

// OTPRepository handles all OTP database operations.
type OTPRepository struct{ db *pgxpool.Pool }

func NewOTPRepository(db *pgxpool.Pool) *OTPRepository { return &OTPRepository{db: db} }

// Create inserts a new OTP record and returns the new row's ID.
func (r *OTPRepository) Create(ctx context.Context, mobile, otpHash, purpose string, expiresAt time.Time) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO otp_verifications (mobile, otp_hash, purpose, expires_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		mobile, otpHash, purpose, expiresAt,
	).Scan(&id)
	return id, err
}

// GetLatestPending retrieves the most recent unverified OTP for a mobile+purpose pair.
func (r *OTPRepository) GetLatestPending(ctx context.Context, mobile, purpose string) (*OTPRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, mobile, otp_hash, purpose, expires_at, attempts, verified_at, created_at
		 FROM otp_verifications
		 WHERE mobile = $1 AND purpose = $2 AND verified_at IS NULL
		 ORDER BY created_at DESC LIMIT 1`,
		mobile, purpose,
	)
	return scanOTP(row)
}

// IncrementAttempts increments the attempts counter for an OTP record.
func (r *OTPRepository) IncrementAttempts(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE otp_verifications SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

// MarkVerified sets verified_at to now for an OTP record.
func (r *OTPRepository) MarkVerified(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE otp_verifications SET verified_at = NOW() WHERE id = $1`, id)
	return err
}

func scanOTP(row pgx.Row) (*OTPRecord, error) {
	o := &OTPRecord{}
	err := row.Scan(&o.ID, &o.Mobile, &o.OTPHash, &o.Purpose, &o.ExpiresAt,
		&o.Attempts, &o.VerifiedAt, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return o, err
}
