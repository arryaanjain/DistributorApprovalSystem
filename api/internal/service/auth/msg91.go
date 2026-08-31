package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// formatMobile ensures standard format with country code (e.g. 919876543210)
func formatMobile(mobile string) string {
	cleaned := strings.TrimSpace(mobile)
	cleaned = strings.TrimPrefix(cleaned, "+")
	cleaned = strings.TrimPrefix(cleaned, "91")
	return "91" + cleaned
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

	// MSG91 API v5 Flow endpoint (same as Laravel): POST https://api.msg91.com/api/v5/flow/
	payload := map[string]interface{}{
		"template_id": c.templateID,
		"mobiles":     formatted,
		"otp":         otp,
		"code":        otp,
		"VAR1":        otp,
		"VAR2":        otp,
		"var":         otp,
	}

	if c.senderID != "" {
		payload["sender"] = c.senderID
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("msg91: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.msg91.com/api/v5/flow/", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("msg91: failed to create request: %w", err)
	}

	req.Header.Set("authkey", c.authKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

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
	// OTP verification for Flow API sent SMS is handled locally in backend DB (crypto.CheckPassword).
	return true, nil
}
