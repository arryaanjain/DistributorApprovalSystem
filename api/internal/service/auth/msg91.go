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
	SendOTP(ctx context.Context, mobile string) error
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
		authKey:    cfg.AuthKey,
		templateID: cfg.TemplateID,
		senderID:   cfg.SenderID,
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

type msg91SendPayload struct {
	TemplateID string `json:"template_id"`
	Mobile     string `json:"mobile"`
	Sender     string `json:"sender,omitempty"`
}

type msg91Response struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *msg91Client) SendOTP(ctx context.Context, mobile string) error {
	formatted := formatMobile(mobile)
	url := "https://control.msg91.com/api/v5/otp"

	payload := msg91SendPayload{
		TemplateID: c.templateID,
		Mobile:     formatted,
		Sender:     c.senderID,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("msg91: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
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
	if err := json.Unmarshal(respBody, &res); err == nil && res.Type == "error" {
		return fmt.Errorf("msg91: api returned error: %s", res.Message)
	}

	return nil
}

func (c *msg91Client) VerifyOTP(ctx context.Context, mobile string, otp string) (bool, error) {
	formatted := formatMobile(mobile)
	url := fmt.Sprintf("https://control.msg91.com/api/v5/otp/verify?mobile=%s&otp=%s", formatted, otp)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return false, nil
	}

	var res msg91Response
	if err := json.Unmarshal(respBody, &res); err == nil {
		if res.Type == "success" || strings.EqualFold(res.Message, "OTP verified success") || strings.EqualFold(res.Message, "already_verified") {
			return true, nil
		}
	}

	return true, nil // HTTP 200 OK from MSG91 verify endpoint indicates valid OTP
}
