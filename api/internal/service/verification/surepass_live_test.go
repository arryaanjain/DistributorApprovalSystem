//go:build surepass

package verification

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestSurepassLiveAPIs executes explicit live API integration tests against Surepass endpoints.
// This test file is excluded from default 'go test ./...' runs via the //go:build surepass tag
// to prevent accidental consumption of paid API credits.
//
// Usage:
//   SUREPASS_TOKEN="your_token" go test -tags=surepass -v ./internal/service/verification/...
func TestSurepassLiveAPIs(t *testing.T) {
	token := os.Getenv("SUREPASS_TOKEN")
	if token == "" {
		t.Skip("Skipping live Surepass integration test: SUREPASS_TOKEN environment variable not set")
	}

	baseURL := os.Getenv("SUREPASS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://kyc-api.surepass.io/api/v1"
	}
	cibilBaseURL := os.Getenv("SUREPASS_CIBIL_BASE_URL")
	if cibilBaseURL == "" {
		cibilBaseURL = "https://app.surepass.app/production/api/v1"
	}

	cibilToken := os.Getenv("SUREPASS_CIBIL_TOKEN")
	client := NewSurepassClient(baseURL, cibilBaseURL, token, cibilToken)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("PAN Verification (Corporate / Individual)", func(t *testing.T) {
		res, err := client.VerifyPAN(ctx, "CZTPJ8269A", "ARRYAAN BHAVESH JAIN")
		if err != nil {
			t.Fatalf("VerifyPAN error: %v", err)
		}
		t.Logf("PAN Result: Status=%s, Name=%s, ClientID=%s", res.Status, res.NameOnPAN, res.ProviderRef)
	})

	t.Run("Corporate GSTIN Verification", func(t *testing.T) {
		res, err := client.VerifyGST(ctx, "08AKWPJ1234H1ZN", "MINDA MARWAR PRODUCER COMPANY")
		if err != nil {
			t.Fatalf("VerifyGST error: %v", err)
		}
		t.Logf("GST Result: Status=%s, LegalName=%s, Address=%s", res.Status, res.LegalName, res.Address)
	})

	t.Run("CIBIL Credit Report JSON Fetch", func(t *testing.T) {
		res, err := client.FetchCreditReport(ctx, "9912345675", "EKRPR1234F", "Vishal Rathore", "male")
		if err != nil {
			t.Fatalf("FetchCreditReport error: %v", err)
		}
		score := 0
		if res.BureauScore != nil {
			score = *res.BureauScore
		}
		t.Logf("CIBIL Result: BureauScore=%d, TotalActiveLoansPaise=%d, Ref=%s", score, res.TotalActiveLoans, res.ProviderRef)
	})

	t.Run("CIBIL Credit Report PDF Fetch", func(t *testing.T) {
		pdfURL, clientID, err := client.FetchCreditReportPDF(ctx, "9988776655", "EKRPR1234F", "Vishal Rathore", "male")
		if err != nil {
			t.Fatalf("FetchCreditReportPDF error: %v", err)
		}
		t.Logf("CIBIL PDF Link: %s (ClientID: %s)", pdfURL, clientID)
	})

	t.Run("E-Sign Session Initialization", func(t *testing.T) {
		req := &ESignInitRequest{
			FullName:     "Arryaan Jain",
			UserEmail:    "jainarryaan@gmail.com",
			MobileNumber: "7276886316",
			PageNum:      1,
			SignX:        90,
			SignY:        468,
			RedirectURL:  "http://localhost:8080/api/v1/agreements/esign-callback",
		}
		resp, err := client.InitializeESignSession(ctx, req)
		if err != nil {
			t.Fatalf("InitializeESignSession error: %v", err)
		}
		t.Logf("ESign Session Result: URL=%s, Token=%s, ClientID=%s", resp.URL, resp.Token, resp.ClientID)
	})
}
