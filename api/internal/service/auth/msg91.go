package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arryaanjain/DistributorApprovalSystem/internal/config"
)

type MSG91Client interface {
	SendOTP(ctx context.Context, mobile string, otp string) error
	VerifyOTP(ctx context.Context, mobile string, otp string) (bool, error)
}

type msg91Client struct {
	authKey    string
	templateID string
	senderID   string
	httpClient *http.Client
}

func NewMSG91Client(cfg *config.MSG91Config) MSG91Client {
	return &msg91Client{
		authKey:    strings.TrimSpace(cfg.AuthKey),
		templateID: strings.TrimSpace(cfg.TemplateID),
		senderID:   strings.TrimSpace(cfg.SenderID),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// formatMobile ensures standard international format (e.g., 919876543210 for India)
func formatMobile(mobile string) string {
	cleaned := strings.TrimSpace(mobile)
	cleaned = strings.TrimPrefix(cleaned, "+")
	if len(cleaned) == 10 {
		return "91" + cleaned
	}
	return cleaned
}

type msg91Response struct {
	Type    string      `json:"type"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Code    interface{} `json:"code"`
}

func (c *msg91Client) SendOTP(ctx context.Context, mobile string, otp string) error {
	if c.templateID == "" {
		return fmt.Errorf("msg91: MSG91_TEMPLATE_ID is missing in configuration")
	}
	if c.authKey == "" {
		return fmt.Errorf("msg91: MSG91_AUTH_KEY is missing in configuration")
	}

	formatted := formatMobile(mobile)

	// MSG91 API v5 Send OTP endpoint: POST https://control.msg91.com/api/v5/otp
	reqURL, err := url.Parse("https://control.msg91.com/api/v5/otp")
	if err != nil {
		return fmt.Errorf("msg91: failed to parse base URL: %w", err)
	}

	q := reqURL.Query()
	q.Set("template_id", c.templateID)
	q.Set("mobile", formatted)
	q.Set("authkey", c.authKey)
	if otp != "" {
		q.Set("otp", otp)
	}
	// Only set sender if explicitly provided and non-empty
	if c.senderID != "" {
		q.Set("sender", c.senderID)
	}
	reqURL.RawQuery = q.Encode()

	// JSON request body for DLT variable substitution
	payload := map[string]interface{}{
		"template_id": c.templateID,
		"mobile":      formatted,
		"otp":         otp,
		"OTP":         otp,
		"var":         otp,
		"var1":        otp,
	}
	if c.senderID != "" {
		payload["sender"] = c.senderID
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("msg91: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("msg91: failed to create request: %w", err)
	}

	req.Header.Set("authkey", c.authKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("msg91: http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("msg91: send OTP failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var res msg91Response
	if err := json.Unmarshal(respBody, &res); err == nil {
		if strings.EqualFold(res.Type, "error") || strings.EqualFold(res.Status, "error") {
			return fmt.Errorf("msg91: api returned error: %s (code: %v)", res.Message, res.Code)
		}
	}

	return nil
}

func (c *msg91Client) VerifyOTP(ctx context.Context, mobile string, otp string) (bool, error) {
	if c.authKey == "" {
		return false, fmt.Errorf("msg91: MSG91_AUTH_KEY is missing in configuration")
	}

	formatted := formatMobile(mobile)
	verifyURL, err := url.Parse("https://control.msg91.com/api/v5/otp/verify")
	if err != nil {
		return false, fmt.Errorf("msg91: failed to parse verify URL: %w", err)
	}

	q := verifyURL.Query()
	q.Set("mobile", formatted)
	q.Set("otp", otp)
	q.Set("authkey", c.authKey)
	verifyURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, verifyURL.String(), nil)
	if err != nil {
		return false, fmt.Errorf("msg91: failed to create verify request: %w", err)
	}

	req.Header.Set("authkey", c.authKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("msg91: verify http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("msg91: verify failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var res msg91Response
	if err := json.Unmarshal(respBody, &res); err == nil {
		if strings.EqualFold(res.Type, "error") || strings.EqualFold(res.Status, "error") {
			return false, fmt.Errorf("msg91: verify returned error: %s (code: %v)", res.Message, res.Code)
		}
		if strings.EqualFold(res.Type, "success") || strings.EqualFold(res.Status, "success") ||
			strings.EqualFold(res.Message, "OTP verified success") || strings.EqualFold(res.Message, "already_verified") {
			return true, nil
		}
	}

	return true, nil
}
