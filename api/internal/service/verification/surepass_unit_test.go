package verification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSurepassUnit_PAN(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pan/pan-comprehensive" {
			t.Errorf("Expected path /pan/pan-comprehensive, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test_token" {
			t.Errorf("Expected Authorization header 'Bearer test_token'")
		}

		var req panRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.IDNumber != "CZTPJ8269A" || req.GetAddress != "yes" {
			t.Errorf("Unexpected request body: %+v", req)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"client_id": "pan_comprehensive_123",
				"pan_number": "CZTPJ8269A",
				"full_name": "ARRYAAN BHAVESH JAIN",
				"status": "valid"
			},
			"status_code": 200,
			"success": true
		}`))
	}))
	defer mockServer.Close()

	client := NewSurepassClient(mockServer.URL, mockServer.URL, "test_token", "test_token")
	res, err := client.VerifyPAN(context.Background(), "CZTPJ8269A", "ARRYAAN BHAVESH JAIN")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Status != StatusVerified {
		t.Errorf("Expected status %s, got %s", StatusVerified, res.Status)
	}
	if res.NameOnPAN != "ARRYAAN BHAVESH JAIN" {
		t.Errorf("Expected name 'ARRYAAN BHAVESH JAIN', got '%s'", res.NameOnPAN)
	}
	if res.NameMatch == nil || !*res.NameMatch {
		t.Errorf("Expected NameMatch to be true")
	}
}

func TestSurepassUnit_GST(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/corporate/gstin" {
			t.Errorf("Expected path /corporate/gstin, got %s", r.URL.Path)
		}

		var req gstRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.IDNumber != "08AKWPJ1234H1ZN" {
			t.Errorf("Unexpected GST request body: %+v", req)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"client_id": "gstin_123",
				"gstin": "08AKWPJ1234H1ZN",
				"business_name": "MINDA MARWAR PRODUCER COMPANY",
				"legal_name": "MADAN LAL JAT",
				"gstin_status": "Active",
				"date_of_registration": "2021-10-20",
				"address": "MINDA NAVA, WARD NO. 15",
				"constitution_of_business": "Proprietorship"
			},
			"status_code": 200,
			"success": true
		}`))
	}))
	defer mockServer.Close()

	client := NewSurepassClient(mockServer.URL, mockServer.URL, "test_token", "test_token")
	res, err := client.VerifyGST(context.Background(), "08AKWPJ1234H1ZN", "MINDA MARWAR PRODUCER COMPANY")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.Status != StatusVerified {
		t.Errorf("Expected status %s, got %s", StatusVerified, res.Status)
	}
	if res.TradeName != "MINDA MARWAR PRODUCER COMPANY" {
		t.Errorf("Expected TradeName 'MINDA MARWAR PRODUCER COMPANY', got '%s'", res.TradeName)
	}
}

func TestSurepassUnit_CIBIL(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/credit-report-cibil/fetch-report" {
			t.Errorf("Expected path /credit-report-cibil/fetch-report, got %s", r.URL.Path)
		}

		var req cibilRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Mobile != "9912345675" || req.PAN != "EKRPR1234F" || req.Consent != "Y" {
			t.Errorf("Unexpected CIBIL request body: %+v", req)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"data": {
				"client_id": "cibil_123",
				"credit_score": "750",
				"credit_report": [{
					"scores": [{"score": "750"}],
					"accounts": [{
						"accountType": "Credit Card",
						"currentBalance": "1000",
						"amountOverdue": "0",
						"paymentHistory": "000"
					}]
				}]
			},
			"status_code": 200,
			"success": true
		}`))
	}))
	defer mockServer.Close()

	client := NewSurepassClient(mockServer.URL, mockServer.URL, "test_token", "test_token")
	res, err := client.FetchCreditReport(context.Background(), "9912345675", "EKRPR1234F", "Vishal Rathore", "male")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if res.BureauScore == nil || *res.BureauScore != 750 {
		t.Errorf("Expected BureauScore 750, got %v", res.BureauScore)
	}
	if res.HasDefaults {
		t.Errorf("Expected HasDefaults to be false")
	}
}
