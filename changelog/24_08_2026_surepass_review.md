# Surepass Integration Audit — Payload & Response Mismatches

Full audit of the Surepass client at [surepass.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/service/verification/surepass.go) and [esign.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/service/verification/esign.go) compared against the actual Surepass Playground API contracts you provided.

---

## Summary of Issues Found

| # | Endpoint | Severity | Issue |
|---|----------|----------|-------|
| 1 | PAN Comprehensive | 🔴 Critical | Wrong request field name & missing required fields; wrong response field names; wrong status value check |
| 2 | Corporate GSTIN | 🔴 Critical | Wrong request field; wrong response field names for legal_name, address, constitution |
| 3 | CIBIL Credit Report | 🔴 Critical | Wrong endpoint path; wrong/missing request fields (`name`, `gender`, `consent`); response model entirely wrong |
| 4 | CIBIL PDF Report | 🟡 Missing | Not implemented at all; PDF URL endpoint different from JSON report |
| 5 | E-Sign | 🟡 Medium | Missing `accept_virtual_sign`, `track_location`, `enforce_geo_location`, `skip_otp`, `skip_email` in config |
| 6 | Base URL | 🔴 Critical | Default is `kyc-api.surepass.io` but CIBIL uses `app.surepass.app/production` |
| 7 | Bank Verification | 🟢 OK | Payload structure is reasonable (Surepass docs not provided, so cannot fully verify) |

---

## Detailed Findings

### 1. PAN Comprehensive (`/pan-comprehensive`)

> [!CAUTION]
> Request payload and response parsing are both wrong — PAN verification will silently fail.

#### Request — Current vs Correct

```diff
 // Current request struct:
 type panRequest struct {
-    IDNUMBER string `json:"id_number"`
+    IDNumber   string `json:"id_number"`
+    GetAddress string `json:"get_address"`   // "yes" to get address
+    GetContact string `json:"get_contact"`   // "yes" to get contact
+    MaskedAadhaarVariant string `json:"masked_aadhaar_variant"` // "v1"
 }
```

> [!NOTE]
> The field name `IDNUMBER` in the Go struct is fine — its JSON tag `id_number` is correct. But the API requires additional fields `get_address`, `get_contact`, and `masked_aadhaar_variant` for the comprehensive variant. Without these, the API may return a partial or error response.

#### Response — Current vs Actual API

```diff
 type surepassPANResponse struct {
     Data struct {
         ClientID   string `json:"client_id"`
-        IDNumber   string `json:"id_number"`
-        Name       string `json:"name"`
-        Status     string `json:"status"`
+        PANNumber  string `json:"pan_number"`
+        FullName   string `json:"full_name"`
+        Status     string `json:"status"`       // returns "valid" (lowercase!)
+        Gender     string `json:"gender"`
+        DOB        string `json:"dob"`
+        Email      string `json:"email"`
+        Phone      string `json:"phone_number"`
+        Category   string `json:"category"`
+        AadhaarLinked bool `json:"aadhaar_linked"`
+        MaskedAadhaar string `json:"masked_aadhaar"`
+        Address    struct { ... } `json:"address"`
     } `json:"data"`
     StatusCode int    `json:"status_code"`
-    Message    string `json:"message"`
+    Message    *string `json:"message"`    // null in success
+    MessageCode string `json:"message_code"`
     Success    bool   `json:"success"`
 }
```

#### Status Check Bug

```diff
-    if resp.Data.Status != "VALID" {  // ❌ API returns "valid" (lowercase)
+    if resp.Data.Status != "valid" {  // ✅ Actual API value
```

---

### 2. Corporate GSTIN (`/corporate/gstin`)

> [!CAUTION]
> Endpoint path and response field names are wrong.

#### Endpoint Path

```diff
-    raw, err := c.post(ctx, "/corporate-gstin", ...)   // ❌ Wrong path
+    raw, err := c.post(ctx, "/corporate/gstin", ...)   // ✅ Actual API path
```

#### Request Payload

```diff
 type gstRequest struct {
-    GSTIN string `json:"gstin"`        // ❌ Wrong field name
+    IDNumber string `json:"id_number"` // ✅ Actual API expects "id_number"
 }
```

#### Response Field Mapping

```diff
 type surepassGSTResponse struct {
     Data struct {
         ClientID         string `json:"client_id"`
         GSTINStatus      string `json:"gstin_status"`
-        LegalName        string `json:"legal_name_of_business"` // ❌
+        LegalName        string `json:"legal_name"`             // ✅
-        TradeName        string `json:"trade_name"`
+        BusinessName     string `json:"business_name"`          // ✅ (this is the trade name)
         DateOfReg        string `json:"date_of_registration"`
-        PrincipalAddress string `json:"principal_place_address"` // ❌
+        Address          string `json:"address"`                 // ✅
         ConstitutionType string `json:"constitution_of_business"`
+        GSTIN            string `json:"gstin"`
+        PANNumber        string `json:"pan_number"`
+        TaxpayerType     string `json:"taxpayer_type"`
+        AadhaarValidation string `json:"aadhaar_validation"`
     } `json:"data"`
 }
```

#### Date Parsing Bug

```diff
-    if t, err := time.Parse("02/01/2006", resp.Data.DateOfReg); err == nil {
+    if t, err := time.Parse("2006-01-02", resp.Data.DateOfReg); err == nil {
     // Actual API returns "2021-10-20" (ISO format), not "DD/MM/YYYY"
```

---

### 3. CIBIL Credit Report

> [!CAUTION]
> The entire endpoint, request, and response structures are wrong.

#### Endpoint Path & Base URL

```diff
-    // Current: uses baseURL + "/credit-report-cibil"
-    // baseURL = "https://kyc-api.surepass.io/api/v1"
+    // Actual CIBIL endpoint is on a DIFFERENT host:
+    // "https://app.surepass.app/production/api/v1/credit-report-cibil/fetch-report"
```

> [!IMPORTANT]
> CIBIL lives on `app.surepass.app/production`, NOT `kyc-api.surepass.io`. This requires either a second base URL or building the full URL for CIBIL calls.

#### Request Payload

```diff
 type cibilRequest struct {
-    MobileNumber string `json:"mobile_number"` // ❌
-    PAN          string `json:"pan"`
+    Mobile  string `json:"mobile"`   // ✅
+    PAN     string `json:"pan"`
+    Name    string `json:"name"`     // ✅ REQUIRED
+    Gender  string `json:"gender"`   // ✅ REQUIRED ("male"/"female")
+    Consent string `json:"consent"`  // ✅ REQUIRED ("Y")
 }
```

#### Response Model

The current code expects a flat `credit_score` int and `accounts` array — **neither exist in the actual response**.

**Actual response structure:**
- `data.credit_score` is a **string** `"750"`, not int
- `data.credit_report` is a deeply nested array of report objects, each containing:
  - `scores[].score` (string)
  - `accounts[]` with fields like `amountOverdue`, `currentBalance`, `paymentHistory`, etc.
  - `response.consumerSummaryresp.accountSummary`
- There is no `data.report_url` — PDF comes from a **separate endpoint** (`/fetch-report-pdf`)

---

### 4. CIBIL PDF Report — Not Implemented

> [!WARNING]
> The PDF report endpoint (`/credit-report-cibil/fetch-report-pdf`) is entirely missing. The current code tries to read `report_url` from the JSON report endpoint, but that field doesn't exist there. PDF reports are a separate API call with the same request payload, returning `credit_report_link`.

---

### 5. E-Sign Config Gaps

The [esign.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/service/verification/esign.go) payload struct has the right fields but the values are not being set:

```diff
 payload.Config.Reason = "Kresconet Distributor Credit Agreement Execution"
 payload.Config.AllowDownload = true
+payload.Config.AcceptVirtualSign = true
+payload.Config.TrackLocation = true
+payload.Config.EnforceGeoLocation = true
+payload.Config.SkipOTP = true
+payload.Config.SkipEmail = true
```

---

### 6. Base URL Issue

```diff
-    v.SetDefault("SUREPASS_BASE_URL", "https://kyc-api.surepass.io/api/v1")
+    // kyc-api.surepass.io → PAN, GST, Bank
+    // app.surepass.app/production → CIBIL
+    // Two separate base URLs are needed
```

---

## Proposed Changes

### [surepass.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/service/verification/surepass.go)

#### [MODIFY] PAN Comprehensive
- Add `get_address`, `get_contact`, `masked_aadhaar_variant` to request
- Fix response fields: `full_name`, `pan_number` 
- Fix status check: `"valid"` not `"VALID"`
- Add `message_code` to response struct

#### [MODIFY] Corporate GSTIN  
- Fix endpoint path from `/corporate-gstin` to `/corporate/gstin`
- Fix request field from `gstin` to `id_number`
- Fix response fields: `legal_name`, `business_name`, `address`
- Fix date format from `DD/MM/YYYY` to `YYYY-MM-DD`

#### [MODIFY] CIBIL Credit Report
- Fix request payload: add `name`, `gender`, `consent` fields; rename `mobile_number` to `mobile`
- Use separate base URL (`https://app.surepass.app/production/api/v1`)
- Rewrite response parsing to handle actual nested structure (`credit_score` as string, `credit_report[0].scores`, `credit_report[0].accounts`, etc.)
- Extract defaults/writeoffs from `paymentHistory` field

#### [NEW] CIBIL PDF Report
- Add `FetchCreditReportPDF` method calling `/credit-report-cibil/fetch-report-pdf`
- Parse `credit_report_link` from response

---

### [esign.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/service/verification/esign.go)

#### [MODIFY] E-Sign Config
- Set `AcceptVirtualSign`, `TrackLocation`, `EnforceGeoLocation`, `SkipOTP`, `SkipEmail` to `true`

---

### [config.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/config/config.go)

#### [MODIFY] SurepassConfig  
- Add `CIBILBaseURL` field (`SUREPASS_CIBIL_BASE_URL`)
- Default to `https://app.surepass.app/production/api/v1`

---

### [.env.example](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/.env.example)

#### [MODIFY] Add CIBIL URL
- Add `SUREPASS_CIBIL_BASE_URL=https://app.surepass.app/production/api/v1`

---

## Open Questions

> [!IMPORTANT]
> 1. **Bank Verification**: You didn't provide the bank verification playground example. The current implementation (`/bank-verification`) looks reasonable but I cannot fully validate the request/response contract. Can you share the bank verification example?

> [!IMPORTANT]
> 2. **E-Sign PDF Upload**: The current code sets `pdf_pre_uploaded: true` but doesn't actually upload the PDF to Surepass before calling `/esign/initialize`. The Surepass e-Sign API requires a prior PDF upload step. Is there a separate upload endpoint you use, or should we add a `PUT /esign/upload-pdf` call before initialization?

> [!IMPORTANT]
> 3. **CIBIL Consent Flow**: The API requires `"consent": "Y"`. Is this consent collected from the distributor during onboarding (Step 3 perhaps)? We need to ensure this consent flag is persisted and passed through.

---

## Verification Plan

### Automated Tests
- Update existing Go unit tests in `internal/service/order/` and `internal/service/financial/`
- Add new test file `internal/service/verification/surepass_test.go` with mock HTTP server to validate request payloads and response parsing
- `go build ./...` and `go test ./...`

### Manual Verification
- Use actual Surepass playground tokens to test one PAN and one GST call with a known-good ID
- Verify CIBIL call goes to `app.surepass.app/production` host
