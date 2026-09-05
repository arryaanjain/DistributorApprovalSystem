package ordercycle

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ProcessGracePeriodExpirations checks for order credit cycles whose grace period has expired and triggers a credit hold
func (s *Service) ProcessGracePeriodExpirations(ctx context.Context) error {
	overdueCycles, err := s.repo.GetOverdueCyclesWithExpiredGrace(ctx)
	if err != nil {
		return fmt.Errorf("getting overdue cycles: %w", err)
	}

	for _, cycle := range overdueCycles {
		log.Printf("[OrderCycleJob] Expired grace period detected for cycle %s (Order %s, Distributor %s). Placing Credit Account on HOLD.",
			cycle.ID, cycle.OrderID, cycle.DistributorID)

		// 1. Mark cycle as defaulted and record hold trigger
		if err := s.repo.SetCycleHoldTriggered(ctx, cycle.ID); err != nil {
			log.Printf("[OrderCycleJob] Error setting hold triggered on cycle %s: %v", cycle.ID, err)
			continue
		}

		// 2. Put distributor credit account on HOLD status
		if err := s.repo.UpdateAccountStatus(ctx, cycle.DistributorID, "hold"); err != nil {
			log.Printf("[OrderCycleJob] Error placing distributor %s account on hold: %v", cycle.DistributorID, err)
		}
	}

	return nil
}

// StartGracePeriodWatcher starts a background ticker that checks expired grace periods every minute
func (s *Service) StartGracePeriodWatcher(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := s.ProcessGracePeriodExpirations(ctx); err != nil {
					log.Printf("[OrderCycleWatcher] Error processing grace periods: %v", err)
				}
			}
		}
	}()
}
