package handler

import (
	"encoding/json"
	"net/http"

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

	response.JSON(w, map[string]interface{}{
		"token":          result.Token,
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

	response.JSON(w, map[string]interface{}{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          result.User,
	})
}

// ─── POST /api/v1/auth/employee/refresh ──────────────────────────────────────

type employeeRefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) EmployeeRefresh(w http.ResponseWriter, r *http.Request) {
	var req employeeRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	accessToken, err := h.svc.RefreshEmployeeToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeAppError(w, err)
		return
	}

	response.JSON(w, map[string]string{"access_token": accessToken})
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
			response.InternalError(w)
		}
		return
	}
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
