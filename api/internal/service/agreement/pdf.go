package agreement

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
)

type AgreementPDFData struct {
	AgreementID      string
	DistributorName  string
	BusinessName     string
	Constitution     string
	PAN              string
	GST              string
	Address          string
	CityStatePIN     string
	Mobile           string
	Email            string
	CreditLimitPaise int64
	PaymentTermsDays int
	InterestRatePct  float64
	EffectiveDate    time.Time
}

type OrderCreditPDFData struct {
	AgreementID           string
	OrderID               string
	OrderNumber           string
	DistributorName       string
	BusinessName          string
	Constitution          string
	PAN                   string
	GST                   string
	Address               string
	CityStatePIN          string
	Mobile                string
	Email                 string
	TotalOrderPaise       int64
	AdvancePaidPaise       int64
	CreditPaise           int64
	Frequency             string // 'weekly' | 'monthly'
	TotalInstalments      int
	InstalmentAmountPaise int64
	GracePeriodDays       int
	EffectiveDate         time.Time
}

type GeneratedPDF struct {
	PDFBytes []byte
	SignX    int
	SignY    int
	PageNum  int
}

// GenerateAgreementPDF generates a styled PDF for onboarding credit agreements.
func GenerateAgreementPDF(data *AgreementPDFData) (*GeneratedPDF, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	// Header
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(30, 41, 59)
	pdf.CellFormat(0, 10, "KRESCONET DISTRIBUTOR CREDIT AGREEMENT", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(0, 6, fmt.Sprintf("Agreement Ref: %s  |  Date: %s", data.AgreementID, data.EffectiveDate.Format("02 Jan 2006")), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Divider line
	pdf.SetDrawColor(226, 232, 240)
	pdf.SetLineWidth(0.5)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	// Section 1: Parties
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 7, "1. PARTIES TO THE AGREEMENT", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(51, 65, 85)
	partyText := fmt.Sprintf(
		"This Credit Agreement (\"Agreement\") is entered into between KRESCONET ENTERPRISE LIMITED (\"Lender/Platform\") and the Distributor detailed below:\n\n"+
			"• Business Name: %s (%s)\n"+
			"• Authorized Signatory: %s\n"+
			"• PAN: %s  |  GSTIN: %s\n"+
			"• Address: %s, %s\n"+
			"• Contact Mobile: %s  |  Email: %s",
		data.BusinessName, data.Constitution,
		data.DistributorName,
		data.PAN, data.GST,
		data.Address, data.CityStatePIN,
		data.Mobile, data.Email,
	)
	pdf.MultiCell(0, 4.5, partyText, "", "L", false)
	pdf.Ln(5)

	// Section 2: Credit Limit & Commercial Terms
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 7, "2. CREDIT LIMIT AND COMMERCIAL TERMS", "", 1, "L", false, 0, "")

	limitINR := float64(data.CreditLimitPaise) / 100.0
	termsText := fmt.Sprintf(
		"• Sanctioned Credit Limit: INR %.2f\n"+
			"• Interest-Free Repayment Cycle: %d Days from Order Invoice Date\n"+
			"• Overdue Interest Penalty: %.1f%% per month on outstanding balances beyond repayment cycle\n"+
			"• Settlement Mode: Direct Bank Transfer / Virtual Account Payment / NACH Mandate",
		limitINR, data.PaymentTermsDays, data.InterestRatePct,
	)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(51, 65, 85)
	pdf.MultiCell(0, 4.5, termsText, "", "L", false)
	pdf.Ln(5)

	// Section 3: Terms & Conditions
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 7, "3. TERMS AND CONDITIONS", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(71, 85, 105)
	tnc := "1. The Distributor agrees to utilize the sanctioned credit limit exclusively for purchasing goods via the Kresconet platform.\n" +
		"2. The Lender reserves the right to review, suspend, or modify the credit limit based on payment history and risk scoring.\n" +
		"3. Any default in payment beyond 60 days will attract credit block, bureau reporting (CIBIL), and legal recovery proceedings.\n" +
		"4. This agreement is governed by the laws of India and subject to exclusive jurisdiction of courts in Mumbai."
	pdf.MultiCell(0, 4, tnc, "", "L", false)
	pdf.Ln(8)

	// Signature Box Coordinates
	signY := int(pdf.GetY() * 2.83)
	signX := 109

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(85, 6, "For KRESCONET ENTERPRISE LTD", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, "For DISTRIBUTOR (AUTHORIZED SIGNATORY)", "", 1, "L", false, 0, "")

	pdf.Ln(12)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(148, 163, 184)
	pdf.CellFormat(85, 5, "[ Digitally Authorized ]", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, "[ Surepass SureSign Digital Signature ]", "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate agreement PDF: %w", err)
	}

	return &GeneratedPDF{
		PDFBytes: buf.Bytes(),
		SignX:    signX,
		SignY:    signY,
		PageNum:  1,
	}, nil
}

// GenerateOrderCreditAgreementPDF generates a PDF agreement specifically for an Order Credit Cycle.
func GenerateOrderCreditAgreementPDF(data *OrderCreditPDFData) (*GeneratedPDF, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	// Header
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(30, 41, 59)
	pdf.CellFormat(0, 10, "KRESCONET ORDER CREDIT & E-MANDATE AGREEMENT", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(0, 6, fmt.Sprintf("Agreement Ref: %s  |  Order: %s  |  Date: %s", data.AgreementID, data.OrderNumber, data.EffectiveDate.Format("02 Jan 2006")), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Divider line
	pdf.SetDrawColor(226, 232, 240)
	pdf.SetLineWidth(0.5)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	// Section 1: Parties
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 7, "1. PARTIES TO THE ORDER CREDIT CYCLE", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(51, 65, 85)
	partyText := fmt.Sprintf(
		"This Order Credit Cycle Agreement (\"Order Agreement\") is executed between KRESCONET ENTERPRISE LIMITED (\"Lender\") and:\n\n"+
			"• Business Name: %s (%s)\n"+
			"• Authorized Signatory: %s\n"+
			"• PAN: %s  |  GSTIN: %s\n"+
			"• Address: %s, %s\n"+
			"• Mobile: %s  |  Email: %s",
		data.BusinessName, data.Constitution,
		data.DistributorName,
		data.PAN, data.GST,
		data.Address, data.CityStatePIN,
		data.Mobile, data.Email,
	)
	pdf.MultiCell(0, 4.5, partyText, "", "L", false)
	pdf.Ln(5)

	// Section 2: Order & Repayment BIFURCATION
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 7, "2. ORDER BIFURCATION & REPAYMENT SCHEDULE", "", 1, "L", false, 0, "")

	totalINR := float64(data.TotalOrderPaise) / 100.0
	advanceINR := float64(data.AdvancePaidPaise) / 100.0
	creditINR := float64(data.CreditPaise) / 100.0
	instINR := float64(data.InstalmentAmountPaise) / 100.0

	termsText := fmt.Sprintf(
		"• Order Number: %s\n"+
			"• Total Order Value: INR %.2f\n"+
			"• Upfront Advance Paid: INR %.2f\n"+
			"• Net Credit Financed: INR %.2f\n"+
			"• Repayment Plan: %d %s instalment(s) of INR %.2f each\n"+
			"• Mandate Provider: Razorpay UPI Subscriptions / E-Mandate",
		data.OrderNumber, totalINR, advanceINR, creditINR,
		data.TotalInstalments, data.Frequency, instINR,
	)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(51, 65, 85)
	pdf.MultiCell(0, 4.5, termsText, "", "L", false)
	pdf.Ln(5)

	// Section 3: UPI MANDATE & DEFAULT CONSEQUENCES
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 7, "3. MANDATE OBLIGATION & CREDIT HOLD TERMS", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(71, 85, 105)
	tnc := fmt.Sprintf(
		"1. The Distributor agrees to immediately setup an auto-debit UPI E-Mandate via Razorpay for the %d instalments.\n"+
			"2. Automated transfers shall occur on scheduled dates. The Distributor must maintain sufficient balance.\n"+
			"3. In the event of a mandate charge failure, a Grace Period of %d days is granted to rectify the payment.\n"+
			"4. If repayment is not completed within %d days, this Order Agreement FAILS, and the Distributor's overall credit limit and account status will be placed on IMMEDIATE HOLD.",
		data.TotalInstalments, data.GracePeriodDays, data.GracePeriodDays,
	)
	pdf.MultiCell(0, 4, tnc, "", "L", false)
	pdf.Ln(8)

	// Signature Box Coordinates
	signY := int(pdf.GetY() * 2.83)
	signX := 109

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(85, 6, "For KRESCONET ENTERPRISE LTD", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, "For DISTRIBUTOR (AUTHORIZED SIGNATORY)", "", 1, "L", false, 0, "")

	pdf.Ln(12)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(148, 163, 184)
	pdf.CellFormat(85, 5, "[ Digitally Authorized ]", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, "[ Surepass SureSign Digital Signature ]", "", 1, "L", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate order credit agreement PDF: %w", err)
	}

	return &GeneratedPDF{
		PDFBytes: buf.Bytes(),
		SignX:    signX,
		SignY:    signY,
		PageNum:  1,
	}, nil
}
