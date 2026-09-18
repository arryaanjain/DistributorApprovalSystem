package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	svcauth "github.com/arryaanjain/DistributorApprovalSystem/internal/service/auth"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// AuthHandler handles OTP and employee login endpoints.
type AuthHandler struct {
	svc *svcauth.Service
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(svc *svcauth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// ─── POST /api/v1/auth/otp/send ──────────────────────────────────────────────

type sendOTPRequest struct {
	Mobile  string `json:"mobile"  validate:"required,min=10,max=13"`
	Purpose string `json:"purpose" validate:"required,oneof=onboarding login"`
}

func (h *AuthHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req sendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	result, err := h.svc.SendOTP(r.Context(), req.Mobile, req.Purpose)
	if err != nil {
		writeAppError(w, err)
		return
	}

	resp := map[string]interface{}{"message": "OTP sent successfully"}
	if result.DevOTP != "" {
		// Only included in dev mode — never in production
		resp["dev_otp"] = result.DevOTP
		resp["otp_id"] = result.OTPID
	}

	response.JSON(w, resp)
}

// ─── POST /api/v1/auth/otp/verify ────────────────────────────────────────────

type verifyOTPRequest struct {
	Mobile  string `json:"mobile"  validate:"required"`
	OTP     string `json:"otp"     validate:"required"`
	Purpose string `json:"purpose" validate:"required,oneof=onboarding login"`
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	result, err := h.svc.VerifyOTP(r.Context(), req.Mobile, req.OTP, req.Purpose)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if result.RefreshToken != "" {
		setRefreshTokenCookie(w, r, "kresconet_distributor_refresh_token", result.RefreshToken)
	}

	response.JSON(w, map[string]interface{}{
		"token":          result.Token,
		"refresh_token": result.RefreshToken,
		"distributor_id": result.DistributorID,
		"is_new_user":    result.IsNewUser,
	})
}

// ─── POST /api/v1/auth/employee/login ────────────────────────────────────────

type employeeLoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) EmployeeLogin(w http.ResponseWriter, r *http.Request) {
	var req employeeLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := validate.Struct(req); err != nil {
		response.UnprocessableEntity(w, formatValidationErrors(err))
		return
	}

	result, err := h.svc.EmployeeLogin(r.Context(), req.Email, req.Password)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if result.RefreshToken != "" {
		setRefreshTokenCookie(w, r, "kresconet_admin_refresh_token", result.RefreshToken)
	}

	response.JSON(w, map[string]interface{}{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          result.User,
	})
}

// ─── POST /api/v1/auth/refresh & /api/v1/auth/employee/refresh ────────────────

type refreshRequest struct {
	RefreshToken  string `json:"refresh_token"`
	SubjectID     string `json:"subject_id"`
	EmailOrMobile string `json:"email_or_mobile"`
	Email         string `json:"email"`
	Mobile        string `json:"mobile"`
}

// RefreshToken handles distributor refresh requests using kresconet_distributor_refresh_token.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	refreshTokenStr := req.RefreshToken

	// Fallback to HttpOnly cookie
	if refreshTokenStr == "" {
		if cookie, err := r.Cookie("kresconet_distributor_refresh_token"); err == nil && cookie.Value != "" {
			refreshTokenStr = cookie.Value
		} else if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
			refreshTokenStr = cookie.Value
		}
	}

	distributorID := req.SubjectID
	mobile := req.Mobile
	if mobile == "" {
		mobile = req.EmailOrMobile
	}

	if refreshTokenStr == "" && distributorID == "" && mobile == "" {
		response.Unauthorized(w, "missing refresh token or distributor identification")
		return
	}

	result, err := h.svc.RefreshDistributorToken(r.Context(), refreshTokenStr, distributorID, mobile)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if result.RefreshToken != "" {
		setRefreshTokenCookie(w, r, "kresconet_distributor_refresh_token", result.RefreshToken)
	}

	response.JSON(w, map[string]interface{}{
		"access_token":  result.AccessToken,
		"token":         result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

// EmployeeRefresh handles employee/admin refresh requests using kresconet_admin_refresh_token.
func (h *AuthHandler) EmployeeRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	refreshTokenStr := req.RefreshToken

	// Fallback to HttpOnly cookie
	if refreshTokenStr == "" {
		if cookie, err := r.Cookie("kresconet_admin_refresh_token"); err == nil && cookie.Value != "" {
			refreshTokenStr = cookie.Value
		} else if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
			refreshTokenStr = cookie.Value
		}
	}

	employeeID := req.SubjectID
	email := req.Email
	if email == "" {
		email = req.EmailOrMobile
	}

	if refreshTokenStr == "" && employeeID == "" && email == "" {
		response.Unauthorized(w, "missing admin refresh token or employee identification")
		return
	}

	result, err := h.svc.RefreshEmployeeToken(r.Context(), refreshTokenStr, employeeID, email)
	if err != nil {
		writeAppError(w, err)
		return
	}

	if result.RefreshToken != "" {
		setRefreshTokenCookie(w, r, "kresconet_admin_refresh_token", result.RefreshToken)
	}

	resp := map[string]interface{}{
		"access_token":  result.AccessToken,
		"token":         result.AccessToken,
		"refresh_token": result.RefreshToken,
	}
	if result.User != nil {
		resp["user"] = result.User
	}

	response.JSON(w, resp)
}

// ─── POST /api/v1/auth/logout ─────────────────────────────────────────────────

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr := ""

	// Check all cookie variants
	for _, c := range r.Cookies() {
		if (c.Name == "kresconet_admin_refresh_token" || c.Name == "kresconet_distributor_refresh_token" || c.Name == "refresh_token") && c.Value != "" {
			refreshTokenStr = c.Value
			break
		}
	}

	if refreshTokenStr == "" && r.Body != nil {
		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.RefreshToken != "" {
			refreshTokenStr = req.RefreshToken
		}
	}

	subjectID := ""
	subjectType := ""
	if empID, ok := r.Context().Value("user_id").(string); ok && empID != "" {
		subjectID = empID
		subjectType = "employee"
	} else if distID, ok := r.Context().Value("distributor_id").(string); ok && distID != "" {
		subjectID = distID
		subjectType = "distributor"
	}

	_ = h.svc.Logout(r.Context(), refreshTokenStr, subjectID, subjectType)

	// Purge all cookies
	cookieNames := []string{"kresconet_admin_refresh_token", "kresconet_distributor_refresh_token", "refresh_token"}
	for _, name := range cookieNames {
		clearRefreshTokenCookie(w, r, name)
	}

	response.JSON(w, map[string]string{"message": "logged out successfully"})
}

func setRefreshTokenCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	secure := isHTTPS

	origin := r.Header.Get("Origin")
	if origin != "" && isHTTPS {
		sameSite = http.SameSiteNoneMode
		secure = true
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   secure,
		MaxAge:   30 * 86400,
	})
}

func clearRefreshTokenCookie(w http.ResponseWriter, r *http.Request, name string) {
	isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	secure := isHTTPS

	origin := r.Header.Get("Origin")
	if origin != "" && isHTTPS {
		sameSite = http.SameSiteNoneMode
		secure = true
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   secure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// ─── Shared helpers ───────────────────────────────────────────────────────────

// writeAppError maps an AppError to the appropriate HTTP status.
func writeAppError(w http.ResponseWriter, err error) {
	if ae, ok := err.(*apperrors.AppError); ok {
		switch ae.Code {
		case apperrors.CodeNotFound:
			response.NotFound(w, ae.Message)
		case apperrors.CodeUnauthorized:
			response.Unauthorized(w, ae.Message)
		case apperrors.CodeForbidden:
			response.Forbidden(w, ae.Message)
		case apperrors.CodeConflict, apperrors.CodeDuplicate:
			response.Conflict(w, ae.Message)
		case apperrors.CodeValidation:
			response.UnprocessableEntity(w, ae.Message)
		case apperrors.CodeExpired:
			response.Error(w, http.StatusGone, string(ae.Code), ae.Message)
		case apperrors.CodeRateLimited:
			response.TooManyRequests(w)
		case apperrors.CodeCreditBlocked, apperrors.CodeHardFlag:
			response.Error(w, http.StatusForbidden, string(ae.Code), ae.Message)
		case apperrors.CodeInsufficientCredit:
			response.Error(w, http.StatusPaymentRequired, string(ae.Code), ae.Message)
		default:
			log.Printf("[ERROR] internal app error: %v", ae)
			response.InternalError(w)
		}
		return
	}
	log.Printf("[ERROR] unexpected server error: %v", err)
	response.InternalError(w)
}

// formatValidationErrors converts validator errors to a map for the response.
func formatValidationErrors(err error) map[string]string {
	errs := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			errs[e.Field()] = e.Tag()
		}
	}
	return errs
}
