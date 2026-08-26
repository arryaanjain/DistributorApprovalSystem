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

	// Issue long-lived 1-month refresh token in Postgres DB after revoking old sessions
	if s.refreshTokenRepo != nil {
		_ = s.refreshTokenRepo.RevokeAllForSubject(ctx, dist.ID, "distributor")
	}
	refreshTokenStr, err := crypto.GenerateToken(32)
	if err != nil {
		return nil, apperrors.Internal("generating refresh token", err)
	}
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour) // 1 month
	if s.refreshTokenRepo != nil {
		if err := s.refreshTokenRepo.Create(ctx, refreshTokenStr, dist.ID, "distributor", mobile, "", refreshExpiry); err != nil {
			slog.Error("failed to store distributor refresh token in DB", "error", err, "distributor_id", dist.ID)
		}
		go func() { _ = s.refreshTokenRepo.DeleteExpired(context.Background()) }()
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
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	User         EmployeeProfile `json:"user"`
}

// RefreshResult holds rotated access and refresh tokens.
type RefreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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

	// Revoke previous active sessions for employee & issue fresh 1-month refresh token
	if s.refreshTokenRepo != nil {
		_ = s.refreshTokenRepo.RevokeAllForSubject(ctx, user.ID, "employee")
	}
	refreshTokenStr, err := crypto.GenerateToken(32)
	if err != nil {
		return nil, apperrors.Internal("generating refresh token", err)
	}
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour) // 1 month
	if s.refreshTokenRepo != nil {
		if err := s.refreshTokenRepo.Create(ctx, refreshTokenStr, user.ID, "employee", user.Email, user.Role, refreshExpiry); err != nil {
			slog.Error("failed to store employee refresh token in DB", "error", err, "user_id", user.ID)
		}
		go func() { _ = s.refreshTokenRepo.DeleteExpired(context.Background()) }()
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &EmployeeLoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         EmployeeProfile{ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role},
	}, nil
}

// RefreshToken validates a long-lived Postgres DB refresh token, rotates it (issues a NEW refresh token + access token), and revokes the old refresh token.
func (s *Service) RefreshToken(ctx context.Context, refreshTokenStr string) (*RefreshResult, error) {
	if refreshTokenStr == "" {
		return nil, apperrors.Unauthorized("missing refresh token")
	}

	if s.refreshTokenRepo == nil {
		return nil, apperrors.Unauthorized("refresh token repo unavailable")
	}

	rec, err := s.refreshTokenRepo.GetValid(ctx, refreshTokenStr)
	if err != nil {
		return nil, apperrors.Internal("verifying refresh token from DB", err)
	}

	// Reuse detection: if token is not valid, check if it was previously revoked (potential theft attack)
	if rec == nil {
		anyRec, _ := s.refreshTokenRepo.GetAny(ctx, refreshTokenStr)
		if anyRec != nil && anyRec.Revoked {
			slog.Warn("REVOKED REFRESH TOKEN REUSE DETECTED! Revoking all sessions for subject",
				"subject_id", anyRec.SubjectID, "subject_type", anyRec.SubjectType)
			_ = s.refreshTokenRepo.RevokeAllForSubject(ctx, anyRec.SubjectID, anyRec.SubjectType)
		}
		return nil, apperrors.Unauthorized("refresh token expired or invalid, please login again")
	}

	var newAccessToken string
	var newRefreshTokenStr string
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour)

	// Revoke the used refresh token (Token Rotation)
	_ = s.refreshTokenRepo.Revoke(ctx, refreshTokenStr)

	newRefreshTokenStr, err = crypto.GenerateToken(32)
	if err != nil {
		return nil, apperrors.Internal("generating rotated refresh token", err)
	}

	if rec.SubjectType == "employee" {
		user, err := s.userRepo.GetByID(ctx, rec.SubjectID)
		if err != nil || user == nil || !user.IsActive {
			return nil, apperrors.Unauthorized("employee account is inactive or not found")
		}
		newAccessToken, err = s.issueEmployeeToken(user.ID, user.Email, user.Role, s.cfg.JWT.AccessExpiry)
		if err != nil {
			return nil, apperrors.Internal("issuing employee access token", err)
		}
		_ = s.refreshTokenRepo.Create(ctx, newRefreshTokenStr, user.ID, "employee", user.Email, user.Role, refreshExpiry)
	} else if rec.SubjectType == "distributor" {
		dist, err := s.distRepo.GetByID(ctx, rec.SubjectID)
		if err != nil || dist == nil {
			return nil, apperrors.Unauthorized("distributor account not found")
		}
		newAccessToken, err = s.issueDistributorToken(dist.ID, dist.Mobile)
		if err != nil {
			return nil, apperrors.Internal("issuing distributor access token", err)
		}
		_ = s.refreshTokenRepo.Create(ctx, newRefreshTokenStr, dist.ID, "distributor", dist.Mobile, "", refreshExpiry)
	} else {
		return nil, apperrors.Unauthorized("invalid subject type for refresh token")
	}

	return &RefreshResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshTokenStr,
	}, nil
}

// RefreshEmployeeToken is a wrapper for employee refresh calls.
func (s *Service) RefreshEmployeeToken(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	return s.RefreshToken(ctx, refreshToken)
}

// Logout revokes all sessions for the subject and/or the specific refresh token in the database.
func (s *Service) Logout(ctx context.Context, refreshTokenStr, subjectID, subjectType string) error {
	if s.refreshTokenRepo == nil {
		return nil
	}
	if subjectID != "" && subjectType != "" {
		_ = s.refreshTokenRepo.RevokeAllForSubject(ctx, subjectID, subjectType)
	}
	if refreshTokenStr != "" {
		_ = s.refreshTokenRepo.Revoke(ctx, refreshTokenStr)
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
