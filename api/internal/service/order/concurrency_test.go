package order_test

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestCreditGuardConcurrency simulates simultaneous order dispatch requests against an available credit balance.
func TestCreditGuardConcurrency(t *testing.T) {
	approvedLimit := int64(5000000) // ₹50,000 in paise
	var currentCredit int64 = 0
	var mu sync.Mutex

	orderAmount := int64(2000000) // ₹20,000 in paise
	numWorkers := 5

	var successfulDispatches int64 = 0
	var rejectedDispatches int64 = 0

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()

			mu.Lock()
			availableCredit := approvedLimit - currentCredit
			if orderAmount <= availableCredit {
				// Simulates atomic update inside DB transaction FOR UPDATE
				currentCredit += orderAmount
				mu.Unlock()
				atomic.AddInt64(&successfulDispatches, 1)
			} else {
				mu.Unlock()
				atomic.AddInt64(&rejectedDispatches, 1)
			}
		}()
	}

	wg.Wait()

	// With ₹50,000 limit and ₹20,000 order size, exactly 2 dispatches should succeed (₹40k used)
	// and 3 dispatches must be blocked by the Credit Guard Invariant!
	if successfulDispatches != 2 {
		t.Errorf("Expected exactly 2 successful dispatches, got %d", successfulDispatches)
	}
	if rejectedDispatches != 3 {
		t.Errorf("Expected exactly 3 blocked dispatches, got %d", rejectedDispatches)
	}
	if currentCredit > approvedLimit {
		t.Errorf("INVARIANT VIOLATION: Current credit %d exceeded approved limit %d", currentCredit, approvedLimit)
	}
}
