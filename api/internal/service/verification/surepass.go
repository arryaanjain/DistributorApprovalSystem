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
	"regexp"
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
// PAN, GST, Bank use baseURL (kyc-api.surepass.io).
// CIBIL uses cibilBaseURL (app.surepass.app/production).
type SurepassClient struct {
	baseURL      string
	cibilBaseURL string
	token        string
	cibilToken   string
	httpClient   *http.Client
}

// NewSurepassClient creates a Surepass API client.
// cibilBaseURL defaults to "https://app.surepass.app/production/api/v1".
// cibilToken defaults to token if empty.
func NewSurepassClient(baseURL, cibilBaseURL, token, cibilToken string) *SurepassClient {
	if cibilBaseURL == "" {
		cibilBaseURL = baseURL
	}
	if cibilToken == "" {
		cibilToken = token
	}
	return &SurepassClient{
		baseURL:      baseURL,
		cibilBaseURL: cibilBaseURL,
		token:        token,
		cibilToken:   cibilToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *SurepassClient) post(ctx context.Context, path string, body interface{}) (json.RawMessage, error) {
	return c.doPost(ctx, c.baseURL+path, body, c.token)
}

func (c *SurepassClient) postCIBIL(ctx context.Context, path string, body interface{}) (json.RawMessage, error) {
	return c.doPost(ctx, c.cibilBaseURL+path, body, c.cibilToken)
}

func (c *SurepassClient) doPost(ctx context.Context, url string, body interface{}, bearerToken string) (json.RawMessage, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
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
	IDNumber             string `json:"id_number"`
	GetAddress           string `json:"get_address"`
	GetContact           string `json:"get_contact"`
	MaskedAadhaarVariant string `json:"masked_aadhaar_variant"`
}

// PANAddress holds address data returned by PAN Comprehensive.
type PANAddress struct {
	Full       string `json:"full"`
	Line1      string `json:"line_1"`
	Line2      string `json:"line_2"`
	StreetName string `json:"street_name"`
	Zip        string `json:"zip"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
}

type surepassPANResponse struct {
	Data struct {
		ClientID      string      `json:"client_id"`
		PANNumber     string      `json:"pan_number"`
		FullName      string      `json:"full_name"`
		FullNameSplit []string    `json:"full_name_split"`
		Status        string      `json:"status"`   // "valid" (lowercase)
		Category      string      `json:"category"` // "person" / "company"
		DOB           string      `json:"dob"`
		Gender        string      `json:"gender"`
		Email         interface{} `json:"email"`
		PhoneNumber   interface{} `json:"phone_number"`
		AadhaarLinked bool        `json:"aadhaar_linked"`
		MaskedAadhaar string      `json:"masked_aadhaar"`
		Address       interface{} `json:"address"`
	} `json:"data"`
	StatusCode  int         `json:"status_code"`
	Message     interface{} `json:"message"`
	MessageCode string      `json:"message_code"`
	Success     bool        `json:"success"`
}

// VerifyPAN calls Surepass pan/pan-comprehensive and returns a normalized PANResult.
func (c *SurepassClient) VerifyPAN(ctx context.Context, pan, expectedName string) (*PANResult, error) {
	raw, err := c.post(ctx, "/pan/pan-comprehensive", panRequest{
		IDNumber:             pan,
		GetAddress:           "yes",
		GetContact:           "yes",
		MaskedAadhaarVariant: "v1",
	})
	if err != nil {
		return &PANResult{Status: StatusUnavailable, RawResponse: raw}, nil
	}

	var resp surepassPANResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return &PANResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	// Accept if status_code == 200 OR success is true OR full_name is present
	if !resp.Success && resp.StatusCode != 200 && resp.Data.FullName == "" {
		return &PANResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	status := StatusVerified
	nameMatch := false
	if resp.Data.Status != "" && resp.Data.Status != "valid" {
		status = StatusFailed
	} else if expectedName != "" {
		nameMatch = NamesMatch(resp.Data.FullName, expectedName)
		if !nameMatch {
			status = StatusMismatch
		}
	}

	return &PANResult{
		Status:      status,
		NameOnPAN:   resp.Data.FullName,
		NameMatch:   &nameMatch,
		ProviderRef: resp.Data.ClientID,
		RawResponse: raw,
	}, nil
}

// ────────────────────────────────────────────────────────────────────────────
// GST (corporate/gstin)
// ────────────────────────────────────────────────────────────────────────────

type gstRequest struct {
	IDNumber string `json:"id_number"`
}

type surepassGSTResponse struct {
	Data struct {
		ClientID          string      `json:"client_id"`
		GSTIN             string      `json:"gstin"`
		PANNumber         string      `json:"pan_number"`
		BusinessName      string      `json:"business_name"`
		LegalName         string      `json:"legal_name"`
		GSTINStatus       string      `json:"gstin_status"`
		DateOfReg         string      `json:"date_of_registration"`
		Address           interface{} `json:"address"`
		PrincipalAddress  string      `json:"principal_place_address"`
		ConstitutionType  string      `json:"constitution_of_business"`
		TaxpayerType      string      `json:"taxpayer_type"`
		AadhaarValidation string      `json:"aadhaar_validation"`
	} `json:"data"`
	StatusCode  int         `json:"status_code"`
	Message     interface{} `json:"message"`
	MessageCode string      `json:"message_code"`
	Success     bool        `json:"success"`
}

// VerifyGST calls Surepass corporate/gstin and returns a normalized GSTResult.
func (c *SurepassClient) VerifyGST(ctx context.Context, gstin, expectedUserName, expectedBizName string) (*GSTResult, error) {
	raw, err := c.post(ctx, "/corporate/gstin", gstRequest{IDNumber: gstin})
	if err != nil {
		return &GSTResult{Status: StatusUnavailable, RawResponse: raw}, nil
	}

	var resp surepassGSTResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return &GSTResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	// Accept if status_code == 200 OR success is true OR legal_name/business_name is present
	if !resp.Success && resp.StatusCode != 200 && resp.Data.LegalName == "" && resp.Data.BusinessName == "" {
		return &GSTResult{Status: StatusFailed, RawResponse: raw}, nil
	}

	status := StatusVerified
	if resp.Data.GSTINStatus != "" && resp.Data.GSTINStatus != "Active" {
		status = StatusPartiallyVerified
	}

	nameMatch := false
	if expectedUserName != "" || expectedBizName != "" {
		nameMatch = NamesMatch(resp.Data.LegalName, expectedUserName) ||
			NamesMatch(resp.Data.LegalName, expectedBizName) ||
			NamesMatch(resp.Data.BusinessName, expectedUserName) ||
			NamesMatch(resp.Data.BusinessName, expectedBizName)
		if !nameMatch {
			status = StatusMismatch
		}
	}

	var regDate *time.Time
	if resp.Data.DateOfReg != "" {
		// Surepass returns ISO 8601 format: "2021-10-20"
		if t, err := time.Parse("2006-01-02", resp.Data.DateOfReg); err == nil {
			regDate = &t
		}
	}

	addressStr := ""
	if str, ok := resp.Data.Address.(string); ok {
		addressStr = str
	} else if resp.Data.PrincipalAddress != "" {
		addressStr = resp.Data.PrincipalAddress
	}

	return &GSTResult{
		Status:           status,
		LegalName:        resp.Data.LegalName,
		TradeName:        resp.Data.BusinessName,
		RegistrationDate: regDate,
		GSTStatus:        resp.Data.GSTINStatus,
		Address:          addressStr,
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
// Hosted on app.surepass.app/production (separate from KYC base URL)
// ────────────────────────────────────────────────────────────────────────────

type cibilRequest struct {
	Mobile  string `json:"mobile"`
	PAN     string `json:"pan"`
	Name    string `json:"name"`
	Gender  string `json:"gender"`  // "male" / "female"
	Consent string `json:"consent"` // "Y"
}

// Actual Surepass CIBIL response structure (deeply nested).
type surepassCIBILResponse struct {
	Data struct {
		ClientID     string `json:"client_id"`
		CreditScore  string `json:"credit_score"` // string, e.g. "750"
		CreditReport []struct {
			Scores []struct {
				Score         string `json:"score"`
				ScoreCardName string `json:"scoreCardName"`
			} `json:"scores"`
			Accounts []struct {
				AccountType    string `json:"accountType"`
				AccountNumber  string `json:"accountNumber"`
				CurrentBalance string `json:"currentBalance"`
				AmountOverdue  string `json:"amountOverdue"`
				DateOpened     string `json:"dateOpened"`
				DateClosed     string `json:"dateClosed"`
				PaymentHistory string `json:"paymentHistory"`
				OwnershipInd   string `json:"ownershipIndicator"`
			} `json:"accounts"`
			Response struct {
				ConsumerSummary struct {
					AccountSummary struct {
						TotalAccounts    int    `json:"totalAccounts"`
						OverdueAccounts  int    `json:"overdueAccounts"`
						ZeroBalAccounts  int    `json:"zeroBalanceAccounts"`
						HighCreditAmount int64  `json:"highCreditAmount"`
						CurrentBalance   int64  `json:"currentBalance"`
						OverdueBalance   int64  `json:"overdueBalance"`
						RecentDateOpened string `json:"recentDateOpened"`
					} `json:"accountSummary"`
				} `json:"consumerSummaryresp"`
			} `json:"response"`
		} `json:"credit_report"`
	} `json:"data"`
	StatusCode  int     `json:"status_code"`
	Message     *string `json:"message"`
	MessageCode string  `json:"message_code"`
	Success     bool    `json:"success"`
}

// CIBIL PDF response structure.
type surepassCIBILPDFResponse struct {
	Data struct {
		ClientID         string  `json:"client_id"`
		CreditScore      string  `json:"credit_score"`
		CreditReportLink string  `json:"credit_report_link"`
	} `json:"data"`
	StatusCode  int     `json:"status_code"`
	Message     *string `json:"message"`
	MessageCode string  `json:"message_code"`
	Success     bool    `json:"success"`
}

// FetchCreditReport calls Surepass credit-report-cibil/fetch-report (JSON).
// name and gender are required by the API.
func (c *SurepassClient) FetchCreditReport(ctx context.Context, mobile, pan, name, gender string) (*CreditReportResult, error) {
	if gender == "" {
		gender = "male"
	}
	raw, err := c.postCIBIL(ctx, "/credit-report-cibil/fetch-report", cibilRequest{
		Mobile:  mobile,
		PAN:     pan,
		Name:    name,
		Gender:  gender,
		Consent: "Y",
	})
	if err != nil {
		return &CreditReportResult{RawResponse: raw}, fmt.Errorf("cibil request error: %w", err)
	}

	var resp surepassCIBILResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return &CreditReportResult{RawResponse: raw}, fmt.Errorf("cibil response unmarshal error: %w", err)
	}

	if !resp.Success {
		errMsg := "unknown error"
		if resp.Message != nil && *resp.Message != "" {
			errMsg = *resp.Message
		} else if resp.MessageCode != "" {
			errMsg = resp.MessageCode
		}
		return &CreditReportResult{RawResponse: raw}, fmt.Errorf("surepass cibil error: %s (status code %d)", errMsg, resp.StatusCode)
	}

	// Parse credit score from string
	score := 0
	if resp.Data.CreditScore != "" {
		fmt.Sscanf(resp.Data.CreditScore, "%d", &score)
	} else if len(resp.Data.CreditReport) > 0 && len(resp.Data.CreditReport[0].Scores) > 0 {
		fmt.Sscanf(resp.Data.CreditReport[0].Scores[0].Score, "%d", &score)
	}

	// Analyze accounts for defaults, write-offs from the first report block
	var hasDefaults, hasWriteoffs, hasSettlements bool
	var totalLoans int64
	var delinquencies int

	if len(resp.Data.CreditReport) > 0 {
		report := resp.Data.CreditReport[0]
		for _, acct := range report.Accounts {
			// Parse overdue amount
			var overdueAmt float64
			fmt.Sscanf(acct.AmountOverdue, "%f", &overdueAmt)
			var curBal float64
			fmt.Sscanf(acct.CurrentBalance, "%f", &curBal)

			if overdueAmt >= 5000 {
				delinquencies++
				hasDefaults = true
			}
			// Check payment history for genuine write-offs (WOF)
			history := strings.ToUpper(acct.PaymentHistory)
			if strings.Contains(history, "WOF") || strings.Contains(history, "WRITE") {
				hasWriteoffs = true
			}
			if strings.Contains(history, "SET") || strings.Contains(history, "SETTLE") {
				hasSettlements = true
			}
			totalLoans += int64(curBal * 100) // convert to paise
		}

		// Also use the summary if available
		summary := report.Response.ConsumerSummary.AccountSummary
		if summary.OverdueAccounts > 0 {
			hasDefaults = true
		}
	}

	now := time.Now()

	return &CreditReportResult{
		BureauScore:      &score,
		HasDefaults:      hasDefaults,
		HasWriteoffs:     hasWriteoffs,
		HasSettlements:   hasSettlements,
		TotalActiveLoans: totalLoans,
		DelinquencyCount: delinquencies,
		FraudFlag:        false,
		ReportDate:       &now,
		ProviderRef:      resp.Data.ClientID,
		RawResponse:      raw,
	}, nil
}

// FetchCreditReportPDF calls Surepass credit-report-cibil/fetch-report-pdf.
// Returns the S3 presigned URL for the CIBIL PDF report.
func (c *SurepassClient) FetchCreditReportPDF(ctx context.Context, mobile, pan, name, gender string) (pdfURL string, clientID string, err error) {
	if gender == "" {
		gender = "male"
	}
	raw, err := c.postCIBIL(ctx, "/credit-report-cibil/fetch-report-pdf", cibilRequest{
		Mobile:  mobile,
		PAN:     pan,
		Name:    name,
		Gender:  gender,
		Consent: "Y",
	})
	if err != nil {
		return "", "", fmt.Errorf("cibil pdf request: %w", err)
	}

	var resp surepassCIBILPDFResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", "", fmt.Errorf("cibil pdf parse: %w", err)
	}

	if !resp.Success {
		return "", "", fmt.Errorf("cibil pdf failed: status_code=%d", resp.StatusCode)
	}

	return resp.Data.CreditReportLink, resp.Data.ClientID, nil
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

// levenshteinDistance computes edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n1, n2 := len(r1), len(r2)
	if n1 == 0 {
		return n2
	}
	if n2 == 0 {
		return n1
	}
	matrix := make([][]int, n1+1)
	for i := range matrix {
		matrix[i] = make([]int, n2+1)
		matrix[i][0] = i
	}
	for j := 0; j <= n2; j++ {
		matrix[0][j] = j
	}
	for i := 1; i <= n1; i++ {
		for j := 1; j <= n2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			min := matrix[i-1][j] + 1
			if matrix[i][j-1]+1 < min {
				min = matrix[i][j-1] + 1
			}
			if matrix[i-1][j-1]+cost < min {
				min = matrix[i-1][j-1] + cost
			}
			matrix[i][j] = min
		}
	}
	return matrix[n1][n2]
}

// NamesMatch performs a robust, case-insensitive token, abbreviation, and edit distance match.
func NamesMatch(a, b string) bool {
	a = strings.ToUpper(strings.TrimSpace(a))
	b = strings.ToUpper(strings.TrimSpace(b))
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}

	reg := regexp.MustCompile(`[^A-Z0-9\s]`)
	cleanA := reg.ReplaceAllString(a, "")
	cleanB := reg.ReplaceAllString(b, "")
	if cleanA == cleanB || strings.Contains(cleanA, cleanB) || strings.Contains(cleanB, cleanA) {
		return true
	}

	normalizeWord := func(w string) string {
		switch w {
		case "PVT", "PVTLTD":
			return "PRIVATE"
		case "LTD":
			return "LIMITED"
		case "AGENCY", "AGENCIES":
			return "AGENCY"
		case "ENTERPRISE", "ENTERPRISES":
			return "ENTERPRISE"
		case "TRADER", "TRADERS":
			return "TRADER"
		case "STORE", "STORES":
			return "STORE"
		case "DISTRIBUTOR", "DISTRIBUTORS":
			return "DISTRIBUTOR"
		case "SUPPLIER", "SUPPLIERS":
			return "SUPPLIER"
		default:
			return w
		}
	}

	wordsA := strings.Fields(cleanA)
	wordsB := strings.Fields(cleanB)
	if len(wordsA) == 0 || len(wordsB) == 0 {
		return false
	}

	var fullA, fullB []string
	for _, w := range wordsA {
		if len(w) > 1 {
			fullA = append(fullA, normalizeWord(w))
		}
	}
	for _, w := range wordsB {
		if len(w) > 1 {
			fullB = append(fullB, normalizeWord(w))
		}
	}

	if len(fullA) > 0 && len(fullB) > 0 {
		smaller, larger := fullA, fullB
		if len(fullA) > len(fullB) {
			smaller, larger = fullB, fullA
		}

		matches := 0
		for _, sw := range smaller {
			for _, lw := range larger {
				if sw == lw || strings.HasPrefix(lw, sw) || strings.HasPrefix(sw, lw) || (len(sw) >= 4 && len(lw) >= 4 && levenshteinDistance(sw, lw) <= 2) {
					matches++
					break
				}
			}
		}

		requiredMatches := len(smaller)
		if len(smaller) >= 2 {
			requiredMatches = 2
		}
		if matches >= requiredMatches {
			return true
		}
	}

	return false
}
