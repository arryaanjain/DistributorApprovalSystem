package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRecord struct {
	ID            string
	Token         string
	SubjectID     string
	SubjectType   string
	EmailOrMobile string
	Role          string
	ExpiresAt     time.Time
	Revoked       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create inserts a new long-lived refresh token record into Postgres DB.
func (r *RefreshTokenRepository) Create(ctx context.Context, token, subjectID, subjectType, emailOrMobile, role string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO refresh_tokens (token, subject_id, subject_type, email_or_mobile, role, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		token, subjectID, subjectType, emailOrMobile, role, expiresAt,
	)
	return err
}

// GetValid retrieves a non-revoked, non-expired refresh token by its token string.
func (r *RefreshTokenRepository) GetValid(ctx context.Context, token string) (*RefreshTokenRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, token, subject_id, subject_type, email_or_mobile, role, expires_at, revoked, created_at, updated_at
		 FROM refresh_tokens
		 WHERE token = $1 AND revoked = FALSE AND expires_at > NOW()`,
		token,
	)
	rec := &RefreshTokenRecord{}
	err := row.Scan(
		&rec.ID, &rec.Token, &rec.SubjectID, &rec.SubjectType,
		&rec.EmailOrMobile, &rec.Role, &rec.ExpiresAt, &rec.Revoked,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rec, err
}

// Revoke marks a refresh token as revoked.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE, updated_at = NOW() WHERE token = $1`,
		token,
	)
	return err
}

// GetAny retrieves a refresh token record regardless of revoked or expired status (for reuse detection).
func (r *RefreshTokenRepository) GetAny(ctx context.Context, token string) (*RefreshTokenRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, token, subject_id, subject_type, email_or_mobile, role, expires_at, revoked, created_at, updated_at
		 FROM refresh_tokens
		 WHERE token = $1`,
		token,
	)
	rec := &RefreshTokenRecord{}
	err := row.Scan(
		&rec.ID, &rec.Token, &rec.SubjectID, &rec.SubjectType,
		&rec.EmailOrMobile, &rec.Role, &rec.ExpiresAt, &rec.Revoked,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rec, err
}

// RevokeAllForSubject revokes all active refresh tokens for a subject (e.g., on logout or login).
func (r *RefreshTokenRepository) RevokeAllForSubject(ctx context.Context, subjectID, subjectType string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = TRUE, updated_at = NOW() WHERE subject_id = $1 AND subject_type = $2 AND revoked = FALSE`,
		subjectID, subjectType,
	)
	return err
}

// DeleteExpired purges expired and old revoked tokens from the database.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at <= NOW() OR (revoked = TRUE AND updated_at < NOW() - INTERVAL '7 days')`,
	)
	return err
}
