// Package middleware provides chi-compatible HTTP middleware.
package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/config"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/apperrors"
	"github.com/arryaanjain/DistributorApprovalSystem/internal/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey is the type for context keys set by middleware.
type contextKey string

const (
	ContextKeyUserID    contextKey = "user_id"
	ContextKeyUserRole  contextKey = "user_role"
	ContextKeyDistID    contextKey = "distributor_id"
	ContextKeyIsAdmin   contextKey = "is_admin"
	ContextKeyRequestID contextKey = "request_id"
)

// EmployeeClaims are embedded in JWTs for internal users.
type EmployeeClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// DistributorClaims are embedded in JWTs for distributors.
type DistributorClaims struct {
	DistributorID string `json:"distributor_id"`
	Mobile        string `json:"mobile"`
	jwt.RegisteredClaims
}

// RequireEmployee validates an employee JWT and injects user context.
func RequireEmployee(cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				response.Unauthorized(w, "missing or malformed authorization token")
				return
			}

			claims := &EmployeeClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, apperrors.Unauthorized("unexpected signing method")
				}
				return []byte(cfg.AccessSecret), nil
			})
			if err != nil || !token.Valid {
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyUserRole, claims.Role)
			ctx = context.WithValue(ctx, ContextKeyIsAdmin, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireDistributor validates a distributor JWT.
func RequireDistributor(cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				response.Unauthorized(w, "missing or malformed authorization token")
				return
			}

			claims := &DistributorClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, apperrors.Unauthorized("unexpected signing method")
				}
				return []byte(cfg.DistributorSecret), nil
			})
			if err != nil || !token.Valid {
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyDistID, claims.DistributorID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole checks that the authenticated employee has (at least) one of the given roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(ContextKeyUserRole).(string)
			if !ok || !allowed[role] {
				response.Forbidden(w, "insufficient permissions for this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// rateBucket tracks request counts per IP per time window.
type rateBucket struct {
	count   int
	resetAt time.Time
}

// RateLimiter provides a simple per-IP in-memory rate limiter.
// For production, replace with a Redis sliding window implementation.
func RateLimiter(requestsPerMinute int) func(http.Handler) http.Handler {
	var (
		mu      sync.Mutex
		buckets = make(map[string]*rateBucket)
	)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			mu.Lock()
			e, ok := buckets[ip]
			if !ok || time.Now().After(e.resetAt) {
				buckets[ip] = &rateBucket{count: 1, resetAt: time.Now().Add(time.Minute)}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}
			e.count++
			if e.count > requestsPerMinute {
				mu.Unlock()
				response.TooManyRequests(w)
				return
			}
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}

// OTPRateLimiter applies stricter rate limits for OTP endpoints (5 per minute per IP).
func OTPRateLimiter() func(http.Handler) http.Handler {
	return RateLimiter(5)
}

// extractBearerToken pulls the token from the Authorization: Bearer header.
func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return r.RemoteAddr
}

// UserIDFromContext extracts the authenticated employee's user ID.
func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ContextKeyUserID).(string)
	return id
}

// UserRoleFromContext extracts the authenticated employee's role.
func UserRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(ContextKeyUserRole).(string)
	return role
}

// DistributorIDFromContext extracts the authenticated distributor's ID.
func DistributorIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ContextKeyDistID).(string)
	return id
}
