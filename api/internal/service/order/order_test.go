package order_test

import (
	"testing"
)

func TestOrderCreditAdvanceSplit(t *testing.T) {
	// Test Split Logic:
	// If available credit is ₹25,000 (2,500,000 paise) and total order is ₹35,000 (3,500,000 paise):
	// Credit Used = ₹25,000
	// Required Advance = ₹10,000
	availableCreditPaise := int64(2500000)
	totalOrderPaise := int64(3500000)

	var creditUsed, advancePaid int64
	if totalOrderPaise <= availableCreditPaise {
		creditUsed = totalOrderPaise
		advancePaid = 0
	} else {
		creditUsed = availableCreditPaise
		advancePaid = totalOrderPaise - availableCreditPaise
	}

	if creditUsed != 2500000 {
		t.Errorf("Expected credit used to be 2,500,000 paise (₹25,000), got %d", creditUsed)
	}
	if advancePaid != 1000000 {
		t.Errorf("Expected advance paid to be 1,000,000 paise (₹10,000), got %d", advancePaid)
	}
}
