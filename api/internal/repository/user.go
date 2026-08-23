package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRecord is a row from the users table (internal employees).
type UserRecord struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Role         string
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
}

// UserRepository handles employee user queries.
type UserRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*UserRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role::TEXT, is_active, last_login_at, created_at
		 FROM users WHERE email = $1`, email)
	u := &UserRecord{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role,
		&u.IsActive, &u.LastLoginAt, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*UserRecord, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, email, password_hash, role::TEXT, is_active, last_login_at, created_at
		 FROM users WHERE id = $1`, id)
	u := &UserRecord{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role,
		&u.IsActive, &u.LastLoginAt, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET last_login_at = NOW() WHERE id = $1`, id)
	return err
}
