// Package verification implements the VerificationService abstraction.
// All external providers (Surepass PAN, GST, Bank, CIBIL) return a normalized
// VerificationResult — the rest of the system never sees provider-specific data.
package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Status is the normalized outcome of any verification call.
type Status string

const (
	StatusPending           Status = "pending"
	StatusVerified          Status = "verified"
	StatusPartiallyVerified Status = "partially_verified"
	StatusMismatch          Status = "mismatch"
	StatusFailed            Status = "failed"
	StatusUnavailable       Status = "unavailable"
)

// PANResult is the normalized output of a PAN verification.
type PANResult struct {
	Status      Status
	NameOnPAN   string
	NameMatch   *bool
	ProviderRef string
	RawResponse json.RawMessage
}

// GSTResult is the normalized output of a GST verification.
type GSTResult struct {
	Status           Status
	LegalName        string
	TradeName        string
	RegistrationDate *time.Time
	GSTStatus        string
	Address          string
	Constitution     string
	NameMatch        *bool
	ProviderRef      string
	RawResponse      json.RawMessage
}

// BankResult is the normalized output of a bank verification.
type BankResult struct {
	Status        Status
	AccountHolder string
	BankName      string
	NameMatch     *bool
	ProviderRef   string
	RawResponse   json.RawMessage
}

// CreditReportResult is the normalized output of a CIBIL credit report fetch.
type CreditReportResult struct {
	BureauScore      *int
	HasDefaults      bool
	HasWriteoffs     bool
	HasSettlements   bool
	TotalActiveLoans int64
	DelinquencyCount int
	FraudFlag        bool
	ReportDate       *time.Time
	PDFURL           string
	ProviderRef      string
	RawResponse      json.RawMessage
}

// ────────────────────────────────────────────────────────────────────────────
// Surepass HTTP Client
// ────────────────────────────────────────────────────────────────────────────

// SurepassClient is the HTTP client for the Surepass API.
type SurepassClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewSurepassClient creates a Surepass API client.
func NewSurepassClient(baseURL, token string) *SurepassClient {
	return &SurepassClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *SurepassClient) post(ctx context.Context, path string, body interface{}) (json.RawMessage, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("surepass server error %d: %s", resp.StatusCode, string(raw))
	}

	return raw, nil
}

// ────────────────────────────────────────────────────────────────────────────
// PAN Comprehensive (pan-comprehensive)
// ────────────────────────────────────────────────────────────────────────────

type panRequest struct {
	IDNUMBER string `json:"id_number"`
}

type surepassPANResponse struct {
	Data struct {
		ClientID   string `json:"client_id"`
		IDNumber   string `json:"id_number"`
		Name       string `json:"name"`
		Status     string `json:"status"`
	} `json:"data"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Success    bool   `json:"success"`
}

// VerifyPAN calls Surepass pan-comprehensive and returns a normalized PANResult.
func (c *SurepassClient) VerifyPAN(ctx context.Context, pan, expectedName string) (*PANResult, error) {
	raw, err := c.post(ctx, "/pan-comprehensive", panRequest{IDNUMBER: pan})
	if err != nil {
		return &PANResult{Status: StatusUnavailable, RawResponse: raw}, nil
	}

	var resp surepassPANResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return &PANResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	if !resp.Success || resp.Data.Status == "" {
		return &PANResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	status := StatusVerified
	nameMatch := false
	if resp.Data.Status != "VALID" {
		status = StatusFailed
	} else if expectedName != "" {
		nameMatch = namesMatch(resp.Data.Name, expectedName)
		if !nameMatch {
			status = StatusMismatch
		}
	}

	return &PANResult{
		Status:      status,
		NameOnPAN:   resp.Data.Name,
		NameMatch:   &nameMatch,
		ProviderRef: resp.Data.ClientID,
		RawResponse: raw,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// GST (corporate-gstin)
// ────────────────────────────────────────────────────────────────────────────

type gstRequest struct {
	GSTIN string `json:"gstin"`
}

type surepassGSTResponse struct {
	Data struct {
		ClientID         string `json:"client_id"`
		GSTINStatus      string `json:"gstin_status"`
		LegalName        string `json:"legal_name_of_business"`
		TradeName        string `json:"trade_name"`
		DateOfReg        string `json:"date_of_registration"`
		PrincipalAddress string `json:"principal_place_address"`
		ConstitutionType string `json:"constitution_of_business"`
	} `json:"data"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Success    bool   `json:"success"`
}

// VerifyGST calls Surepass corporate-gstin and returns a normalized GSTResult.
func (c *SurepassClient) VerifyGST(ctx context.Context, gstin, expectedName string) (*GSTResult, error) {
	raw, err := c.post(ctx, "/corporate-gstin", gstRequest{GSTIN: gstin})
	if err != nil {
		return &GSTResult{Status: StatusUnavailable, RawResponse: raw}, nil
	}

	var resp surepassGSTResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return &GSTResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	if !resp.Success {
		return &GSTResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	status := StatusVerified
	if resp.Data.GSTINStatus != "Active" {
		status = StatusPartiallyVerified
	}

	nameMatch := false
	if expectedName != "" {
		nameMatch = namesMatch(resp.Data.LegalName, expectedName) ||
			namesMatch(resp.Data.TradeName, expectedName)
		if !nameMatch {
			status = StatusMismatch
		}
	}

	var regDate *time.Time
	if resp.Data.DateOfReg != "" {
		if t, err := time.Parse("02/01/2006", resp.Data.DateOfReg); err == nil {
			regDate = &t
		}
	}

	return &GSTResult{
		Status:           status,
		LegalName:        resp.Data.LegalName,
		TradeName:        resp.Data.TradeName,
		RegistrationDate: regDate,
		GSTStatus:        resp.Data.GSTINStatus,
		Address:          resp.Data.PrincipalAddress,
		Constitution:     resp.Data.ConstitutionType,
		NameMatch:        &nameMatch,
		ProviderRef:      resp.Data.ClientID,
		RawResponse:      raw,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Bank Account Verification
// ────────────────────────────────────────────────────────────────────────────

type bankRequest struct {
	IDNumber    string `json:"id_number"`
	IFSC        string `json:"ifsc"`
	AccountType string `json:"account_type"`
	Name        string `json:"name"`
}

type surepassBankResponse struct {
	Data struct {
		ClientID      string `json:"client_id"`
		AccountExists bool   `json:"account_exists"`
		NameAtBank    string `json:"name_at_bank"`
		BankName      string `json:"bank_name"`
		NameMatchScore float64 `json:"name_match_score"`
	} `json:"data"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Success    bool   `json:"success"`
}

// VerifyBankAccount calls Surepass bank verification and returns a normalized BankResult.
func (c *SurepassClient) VerifyBankAccount(ctx context.Context, accountNumber, ifsc, expectedName string) (*BankResult, error) {
	raw, err := c.post(ctx, "/bank-verification", bankRequest{
		IDNumber:    accountNumber,
		IFSC:        ifsc,
		AccountType: "savings",
		Name:        expectedName,
	})
	if err != nil {
		return &BankResult{Status: StatusUnavailable, RawResponse: raw}, nil
	}

	var resp surepassBankResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return &BankResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	if !resp.Success || !resp.Data.AccountExists {
		return &BankResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	nameMatch := resp.Data.NameMatchScore >= 0.75
	status := StatusVerified
	if !nameMatch {
		status = StatusMismatch
	}

	return &BankResult{
		Status:        status,
		AccountHolder: resp.Data.NameAtBank,
		BankName:      resp.Data.BankName,
		NameMatch:     &nameMatch,
		ProviderRef:   resp.Data.ClientID,
		RawResponse:   raw,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// CIBIL Credit Report (credit-report-cibil)
// ────────────────────────────────────────────────────────────────────────────

type cibilRequest struct {
	MobileNumber string `json:"mobile_number"`
	PAN          string `json:"pan"`
}

type surepassCIBILResponse struct {
	Data struct {
		ClientID    string `json:"client_id"`
		CreditScore int    `json:"credit_score"`
		ReportURL   string `json:"report_url"`
		Accounts    []struct {
			AccountStatus string  `json:"account_status"`
			CurrentBalance float64 `json:"current_balance"`
			OverdueAmount  float64 `json:"overdue_amount"`
		} `json:"accounts"`
	} `json:"data"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Success    bool   `json:"success"`
}

// FetchCreditReport calls Surepass credit-report-cibil.
func (c *SurepassClient) FetchCreditReport(ctx context.Context, mobile, pan string) (*CreditReportResult, error) {
	raw, err := c.post(ctx, "/credit-report-cibil", cibilRequest{
		MobileNumber: mobile,
		PAN:          pan,
	})
	if err != nil {
		return &CreditReportResult{RawResponse: raw}, nil
	}

	var resp surepassCIBILResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return &CreditReportResult{RawResponse: raw}, nil
	}

	if !resp.Success {
		return &CreditReportResult{RawResponse: raw}, nil
	}

	// Analyze accounts for defaults, write-offs
	var hasDefaults, hasWriteoffs, hasSettlements bool
	var totalLoans int64
	var delinquencies int
	for _, acct := range resp.Data.Accounts {
		switch acct.AccountStatus {
		case "Defaulted":
			hasDefaults = true
			delinquencies++
		case "Written-off":
			hasWriteoffs = true
		case "Settled":
			hasSettlements = true
		}
		if acct.OverdueAmount > 0 {
			delinquencies++
		}
		totalLoans += int64(acct.CurrentBalance * 100) // convert to paise
	}

	score := resp.Data.CreditScore
	now := time.Now()

	return &CreditReportResult{
		BureauScore:      &score,
		HasDefaults:      hasDefaults,
		HasWriteoffs:     hasWriteoffs,
		HasSettlements:   hasSettlements,
		TotalActiveLoans: totalLoans,
		DelinquencyCount: delinquencies,
		FraudFlag:        false, // Surepass doesn't return this directly
		ReportDate:       &now,
		PDFURL:           resp.Data.ReportURL,
		ProviderRef:      resp.Data.ClientID,
		RawResponse:      raw,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

// namesMatch does a simple fuzzy match between two names.
// Production should use a more robust string similarity algorithm.
func namesMatch(a, b string) bool {
	a = strings.ToUpper(strings.TrimSpace(a))
	b = strings.ToUpper(strings.TrimSpace(b))
	if a == b {
		return true
	}
	// Accept if one is a substring of the other (handles middle name cases)
	return strings.Contains(a, b) || strings.Contains(b, a)
}
