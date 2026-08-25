// Package auth implements OTP-based authentication for distributors
// and password-based login for internal employees.
package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/config"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/crypto"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/repository"
	"github.com/golang-jwt/jwt/v5"
)

const (
	PurposeOnboarding = "onboarding"
	PurposeLogin      = "login"
)

// Service handles all authentication logic.
type Service struct {
	cfg              *config.Config
	otpRepo          *repository.OTPRepository
	userRepo         *repository.UserRepository
	distRepo         *repository.DistributorRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	msg91            MSG91Client
}

// New creates an Auth service.
func New(cfg *config.Config, otpRepo *repository.OTPRepository,
	userRepo *repository.UserRepository, distRepo *repository.DistributorRepository,
	refreshTokenRepo *repository.RefreshTokenRepository, msg91 MSG91Client) *Service {
	return &Service{
		cfg:              cfg,
		otpRepo:          otpRepo,
		userRepo:         userRepo,
		distRepo:         distRepo,
		refreshTokenRepo: refreshTokenRepo,
		msg91:            msg91,
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Distributor OTP Flow
// ────────────────────────────────────────────────────────────────────────────

// SendOTPResult is returned after generating an OTP.
type SendOTPResult struct {
	OTPID  string // for dev logging reference
	DevOTP string // only populated when OTP_DEV_MODE=true
}

// SendOTP generates an OTP for the given mobile number.
func (s *Service) SendOTP(ctx context.Context, mobile, purpose string) (*SendOTPResult, error) {
	// Validate purpose
	if purpose != PurposeOnboarding && purpose != PurposeLogin {
		return nil, apperrors.Validation("invalid OTP purpose")
	}

	otp, err := crypto.GenerateOTP(s.cfg.OTP.Length)
	if err != nil {
		return nil, apperrors.Internal("generating OTP", err)
	}

	hash, err := crypto.HashPassword(otp)
	if err != nil {
		return nil, apperrors.Internal("hashing OTP", err)
	}

	expiresAt := time.Now().Add(s.cfg.OTP.Expiry)
	id, err := s.otpRepo.Create(ctx, mobile, hash, purpose, expiresAt)
	if err != nil {
		return nil, apperrors.Internal("storing OTP", err)
	}

	result := &SendOTPResult{OTPID: id}

	if s.cfg.OTP.DevMode {
		slog.Warn("OTP_DEV_MODE: OTP not sent via external provider",
			"mobile", mobile, "otp", otp, "purpose", purpose)
		result.DevOTP = otp
	} else {
		if s.msg91 != nil {
			if err := s.msg91.SendOTP(ctx, mobile); err != nil {
				slog.Error("failed to send OTP via MSG91", "error", err, "mobile", mobile)
				return nil, apperrors.Internal("sending OTP via MSG91", err)
			}
		}
		slog.Info("OTP sent via MSG91", "mobile", mobile, "purpose", purpose)
	}

	return result, nil
}

// VerifyOTPResult is returned after a successful OTP verification.
type VerifyOTPResult struct {
	DistributorID string
	Token         string // short-lived access token
	RefreshToken  string // 1-month long-lived refresh token stored in Postgres DB
	IsNewUser     bool
}

// VerifyOTP validates the OTP and issues a distributor JWT and long-lived Postgres DB refresh token.
func (s *Service) VerifyOTP(ctx context.Context, mobile, otp, purpose string) (*VerifyOTPResult, error) {
	record, err := s.otpRepo.GetLatestPending(ctx, mobile, purpose)
	if err != nil {
		return nil, apperrors.Internal("fetching OTP record", err)
	}
	if record == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "no pending OTP found for this mobile number")
	}

	// Check expiry
	if time.Now().After(record.ExpiresAt) {
		return nil, apperrors.Expired("OTP has expired, please request a new one")
	}

	// Check max attempts
	if record.Attempts >= s.cfg.OTP.MaxRetries {
		return nil, apperrors.RateLimited()
	}

	// Validate OTP
	if !s.cfg.OTP.DevMode && s.msg91 != nil {
		ok, err := s.msg91.VerifyOTP(ctx, mobile, otp)
		if err != nil || !ok {
			_ = s.otpRepo.IncrementAttempts(ctx, record.ID)
			return nil, apperrors.Unauthorized("invalid or expired OTP")
		}
	} else {
		if err := crypto.CheckPassword(otp, record.OTPHash); err != nil {
			_ = s.otpRepo.IncrementAttempts(ctx, record.ID)
			return nil, apperrors.Unauthorized("invalid OTP")
		}
	}

	// Mark as verified
	if err := s.otpRepo.MarkVerified(ctx, record.ID); err != nil {
		return nil, apperrors.Internal("marking OTP verified", err)
	}

	// Get or create distributor record
	dist, err := s.distRepo.GetByMobile(ctx, mobile)
	if err != nil {
		return nil, apperrors.Internal("fetching distributor", err)
	}
	isNew := dist == nil
	if isNew {
		distID, err := s.distRepo.Create(ctx, mobile)
		if err != nil {
			return nil, apperrors.Internal("creating distributor", err)
		}
		dist = &repository.DistributorRecord{ID: distID, Mobile: mobile}
	}

	token, err := s.issueDistributorToken(dist.ID, mobile)
	if err != nil {
		return nil, apperrors.Internal("issuing access token", err)
	}

	// Issue long-lived 1-month refresh token in Postgres DB
	refreshTokenStr, err := crypto.GenerateToken(32)
	if err != nil {
		return nil, apperrors.Internal("generating refresh token", err)
	}
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour) // 1 month
	if s.refreshTokenRepo != nil {
		if err := s.refreshTokenRepo.Create(ctx, refreshTokenStr, dist.ID, "distributor", mobile, "", refreshExpiry); err != nil {
			slog.Error("failed to store distributor refresh token in DB", "error", err, "distributor_id", dist.ID)
		}
	}

	return &VerifyOTPResult{
		DistributorID: dist.ID,
		Token:         token,
		RefreshToken:  refreshTokenStr,
		IsNewUser:     isNew,
	}, nil
}

func (s *Service) issueDistributorToken(distributorID, mobile string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"distributor_id": distributorID,
		"mobile":         mobile,
		"iat":            now.Unix(),
		"exp":            now.Add(s.cfg.JWT.AccessExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.DistributorSecret))
}

// ────────────────────────────────────────────────────────────────────────────
// Employee Login
// ────────────────────────────────────────────────────────────────────────────

// EmployeeLoginResult holds tokens for an authenticated employee.
type EmployeeLoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         EmployeeProfile `json:"user"`
}

// EmployeeProfile is safe to return to the frontend (no password hash).
type EmployeeProfile struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// EmployeeLogin validates email + password and issues access + 1-month refresh tokens in Postgres DB.
func (s *Service) EmployeeLogin(ctx context.Context, email, password string) (*EmployeeLoginResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.Internal("fetching user", err)
	}
	if user == nil || !user.IsActive {
		return nil, apperrors.Unauthorized("invalid credentials")
	}

	if err := crypto.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, apperrors.Unauthorized("invalid credentials")
	}

	accessToken, err := s.issueEmployeeToken(user.ID, user.Email, user.Role, s.cfg.JWT.AccessExpiry)
	if err != nil {
		return nil, apperrors.Internal("issuing access token", err)
	}

	// Generate 1-month long-lived refresh token in Postgres DB
	refreshTokenStr, err := crypto.GenerateToken(32)
	if err != nil {
		return nil, apperrors.Internal("generating refresh token", err)
	}
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour) // 1 month
	if s.refreshTokenRepo != nil {
		if err := s.refreshTokenRepo.Create(ctx, refreshTokenStr, user.ID, "employee", user.Email, user.Role, refreshExpiry); err != nil {
			slog.Error("failed to store employee refresh token in DB", "error", err, "user_id", user.ID)
		}
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &EmployeeLoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         EmployeeProfile{ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role},
	}, nil
}

// RefreshToken validates a long-lived Postgres DB refresh token and issues a new access token.
// The refresh token itself is NOT rotated — it gracefully expires after 1 month.
func (s *Service) RefreshToken(ctx context.Context, refreshTokenStr string) (string, error) {
	if refreshTokenStr == "" {
		return "", apperrors.Unauthorized("missing refresh token")
	}

	if s.refreshTokenRepo == nil {
		return "", apperrors.Unauthorized("refresh token repo unavailable")
	}

	rec, err := s.refreshTokenRepo.GetValid(ctx, refreshTokenStr)
	if err != nil {
		return "", apperrors.Internal("verifying refresh token from DB", err)
	}
	if rec == nil {
		return "", apperrors.Unauthorized("refresh token expired or invalid, please login again")
	}

	if rec.SubjectType == "employee" {
		user, err := s.userRepo.GetByID(ctx, rec.SubjectID)
		if err != nil || user == nil || !user.IsActive {
			return "", apperrors.Unauthorized("employee account is inactive or not found")
		}
		return s.issueEmployeeToken(user.ID, user.Email, user.Role, s.cfg.JWT.AccessExpiry)
	} else if rec.SubjectType == "distributor" {
		dist, err := s.distRepo.GetByID(ctx, rec.SubjectID)
		if err != nil || dist == nil {
			return "", apperrors.Unauthorized("distributor account not found")
		}
		return s.issueDistributorToken(dist.ID, dist.Mobile)
	}

	return "", apperrors.Unauthorized("invalid subject type for refresh token")
}

// RefreshEmployeeToken is a wrapper for employee refresh calls.
func (s *Service) RefreshEmployeeToken(ctx context.Context, refreshToken string) (string, error) {
	return s.RefreshToken(ctx, refreshToken)
}

// Logout revokes a refresh token in the database.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	if s.refreshTokenRepo != nil {
		return s.refreshTokenRepo.Revoke(ctx, refreshToken)
	}
	return nil
}

func (s *Service) issueEmployeeToken(userID, email, role string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"iat":     now.Unix(),
		"exp":     now.Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.AccessSecret))
}

func (s *Service) issueEmployeeRefreshToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"iat":     now.Unix(),
		"exp":     now.Add(s.cfg.JWT.RefreshExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.RefreshSecret))
}
