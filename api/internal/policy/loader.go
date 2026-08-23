// Package policy loads the active credit policy from the database and
// exposes it to the scoring and decision engines.
// Credit rules are never hard-coded — everything comes from the DB.
package policy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ScoreBand maps a score range to an eligibility and max credit limit.
type ScoreBand struct {
	MinScore        int
	MaxScore        int
	Eligibility     string // "CREDIT" | "ADVANCE_ONLY"
	MaxCreditPaise  int64
	DisplayLabel    string
}

// LadderStep represents one rung in the credit enhancement ladder.
type LadderStep struct {
	StepOrder          int
	LimitPaise         int64
	DisplayLabel       string
	MinCycles          int
	MinOntimePct       float64
	MinUtilisationPct  float64
	AutoApprove        bool
	ApprovalRole       string
}

// OverdueThreshold maps overdue day ranges to automated actions.
type OverdueThreshold struct {
	Tier          int
	FromDays      int
	ToDays        *int
	Label         string
	ActionCodes   []string
	AutoRestrict  bool
	AutoHold      bool
}

// NonGSTRule defines constraints for distributors without GST.
type NonGSTRule struct {
	MaxInitialLimitPaise  int64
	RequiresAltEvidence   bool
	AcceptableEvidence    []string
}

// ApprovalAuthority maps limit ranges to required approval roles.
type ApprovalAuthority struct {
	FromLimitPaise int64
	ToLimitPaise   *int64
	RequiredRole   string
	Label          string
}

// EnhancementRule defines conditions for credit limit upgrades.
type EnhancementRule struct {
	FromLimitPaise    int64
	ToLimitPaise      int64
	RequiredCycles    int
	RequiredOntimePct float64
	RequiredUtilPct   float64
	NoCurrentFlags    bool
	AutoApprove       bool
	ApprovalRole      string
	Label             string
}

// RiskGrade maps score ranges to letter grades.
type RiskGrade struct {
	Grade          string
	MinScore       int
	MaxScore       int
	Label          string
	MaxLimitPaise  *int64
}

// Policy is the fully-loaded active credit policy.
type Policy struct {
	ID          string
	Version     string
	Name        string

	ScoreBands          []ScoreBand
	CreditLadder        []LadderStep
	OverdueThresholds   []OverdueThreshold
	NonGSTRule          *NonGSTRule
	ApprovalAuthorities []ApprovalAuthority
	EnhancementRules    []EnhancementRule
	RiskGrades          []RiskGrade

	LoadedAt time.Time
}

// Loader caches the active policy and refreshes it on demand.
type Loader struct {
	mu     sync.RWMutex
	db     *pgxpool.Pool
	cached *Policy
}

// NewLoader creates a policy loader backed by the given pool.
func NewLoader(db *pgxpool.Pool) *Loader {
	return &Loader{db: db}
}

// Active returns the cached active policy, loading it from DB if needed.
func (l *Loader) Active(ctx context.Context) (*Policy, error) {
	l.mu.RLock()
	if l.cached != nil {
		p := l.cached
		l.mu.RUnlock()
		return p, nil
	}
	l.mu.RUnlock()

	return l.Reload(ctx)
}

// Reload forces a fresh load of the active policy from DB.
func (l *Loader) Reload(ctx context.Context) (*Policy, error) {
	p, err := l.load(ctx)
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	l.cached = p
	l.mu.Unlock()
	return p, nil
}

// Invalidate clears the cached policy (call after a policy update).
func (l *Loader) Invalidate() {
	l.mu.Lock()
	l.cached = nil
	l.mu.Unlock()
}

func (l *Loader) load(ctx context.Context) (*Policy, error) {
	p := &Policy{LoadedAt: time.Now()}

	// Load base policy
	err := l.db.QueryRow(ctx,
		`SELECT id, version, name FROM credit_policies WHERE is_active = TRUE LIMIT 1`,
	).Scan(&p.ID, &p.Version, &p.Name)
	if err != nil {
		return nil, fmt.Errorf("loading active policy: %w", err)
	}

	// Score bands
	rows, err := l.db.Query(ctx,
		`SELECT min_score, max_score, eligibility, max_credit_paise, display_label
		 FROM policy_score_bands WHERE policy_id = $1 ORDER BY min_score DESC`, p.ID)
	if err != nil {
		return nil, fmt.Errorf("loading score bands: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var b ScoreBand
		if err := rows.Scan(&b.MinScore, &b.MaxScore, &b.Eligibility, &b.MaxCreditPaise, &b.DisplayLabel); err != nil {
			return nil, err
		}
		p.ScoreBands = append(p.ScoreBands, b)
	}

	// Credit ladder
	rows2, err := l.db.Query(ctx,
		`SELECT step_order, limit_paise, display_label, min_cycles,
		        min_ontime_pct, min_utilisation_pct, auto_approve, COALESCE(approval_role::TEXT, '')
		 FROM policy_credit_ladder WHERE policy_id = $1 ORDER BY step_order`, p.ID)
	if err != nil {
		return nil, fmt.Errorf("loading credit ladder: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var s LadderStep
		if err := rows2.Scan(&s.StepOrder, &s.LimitPaise, &s.DisplayLabel, &s.MinCycles,
			&s.MinOntimePct, &s.MinUtilisationPct, &s.AutoApprove, &s.ApprovalRole); err != nil {
			return nil, err
		}
		p.CreditLadder = append(p.CreditLadder, s)
	}

	// Non-GST rule
	var ng NonGSTRule
	err = l.db.QueryRow(ctx,
		`SELECT max_initial_limit_paise, requires_alt_evidence, acceptable_evidence
		 FROM policy_non_gst_rules WHERE policy_id = $1 LIMIT 1`, p.ID,
	).Scan(&ng.MaxInitialLimitPaise, &ng.RequiresAltEvidence, &ng.AcceptableEvidence)
	if err == nil {
		p.NonGSTRule = &ng
	}

	// Overdue thresholds
	rows3, err := l.db.Query(ctx,
		`SELECT tier, from_days, to_days, label, action_codes, auto_restrict, auto_hold
		 FROM policy_overdue_thresholds WHERE policy_id = $1 ORDER BY tier`, p.ID)
	if err != nil {
		return nil, fmt.Errorf("loading overdue thresholds: %w", err)
	}
	defer rows3.Close()
	for rows3.Next() {
		var t OverdueThreshold
		if err := rows3.Scan(&t.Tier, &t.FromDays, &t.ToDays, &t.Label,
			&t.ActionCodes, &t.AutoRestrict, &t.AutoHold); err != nil {
			return nil, err
		}
		p.OverdueThresholds = append(p.OverdueThresholds, t)
	}

	// Approval authorities
	rows4, err := l.db.Query(ctx,
		`SELECT from_limit_paise, to_limit_paise, required_role::TEXT, label
		 FROM policy_approval_authorities WHERE policy_id = $1 ORDER BY from_limit_paise`, p.ID)
	if err != nil {
		return nil, fmt.Errorf("loading approval authorities: %w", err)
	}
	defer rows4.Close()
	for rows4.Next() {
		var a ApprovalAuthority
		if err := rows4.Scan(&a.FromLimitPaise, &a.ToLimitPaise, &a.RequiredRole, &a.Label); err != nil {
			return nil, err
		}
		p.ApprovalAuthorities = append(p.ApprovalAuthorities, a)
	}

	// Enhancement rules
	rows5, err := l.db.Query(ctx,
		`SELECT from_limit_paise, to_limit_paise, required_cycles, required_ontime_pct,
		        required_util_pct, no_current_flags, auto_approve, COALESCE(approval_role::TEXT, ''), label
		 FROM policy_enhancement_rules WHERE policy_id = $1 ORDER BY from_limit_paise`, p.ID)
	if err != nil {
		return nil, fmt.Errorf("loading enhancement rules: %w", err)
	}
	defer rows5.Close()
	for rows5.Next() {
		var r EnhancementRule
		if err := rows5.Scan(&r.FromLimitPaise, &r.ToLimitPaise, &r.RequiredCycles, &r.RequiredOntimePct,
			&r.RequiredUtilPct, &r.NoCurrentFlags, &r.AutoApprove, &r.ApprovalRole, &r.Label); err != nil {
			return nil, err
		}
		p.EnhancementRules = append(p.EnhancementRules, r)
	}

	// Risk grades
	rows6, err := l.db.Query(ctx,
		`SELECT grade, min_score, max_score, label, max_limit_paise
		 FROM policy_risk_grades WHERE policy_id = $1 ORDER BY min_score DESC`, p.ID)
	if err != nil {
		return nil, fmt.Errorf("loading risk grades: %w", err)
	}
	defer rows6.Close()
	for rows6.Next() {
		var g RiskGrade
		if err := rows6.Scan(&g.Grade, &g.MinScore, &g.MaxScore, &g.Label, &g.MaxLimitPaise); err != nil {
			return nil, err
		}
		p.RiskGrades = append(p.RiskGrades, g)
	}

	return p, nil
}

// ScoreBandFor returns the matching ScoreBand for a given total score.
func (p *Policy) ScoreBandFor(score int) *ScoreBand {
	for i := range p.ScoreBands {
		b := &p.ScoreBands[i]
		if score >= b.MinScore && score <= b.MaxScore {
			return b
		}
	}
	return nil
}

// RiskGradeFor returns the matching RiskGrade for a given score.
func (p *Policy) RiskGradeFor(score int) *RiskGrade {
	for i := range p.RiskGrades {
		g := &p.RiskGrades[i]
		if score >= g.MinScore && score <= g.MaxScore {
			return g
		}
	}
	return nil
}

// NextLadderStep returns the next credit ladder step above currentLimitPaise.
func (p *Policy) NextLadderStep(currentLimitPaise int64) *LadderStep {
	for i := range p.CreditLadder {
		s := &p.CreditLadder[i]
		if s.LimitPaise > currentLimitPaise {
			return s
		}
	}
	return nil
}

// OverdueThresholdFor returns the matching overdue threshold for given days overdue.
func (p *Policy) OverdueThresholdFor(daysOverdue int) *OverdueThreshold {
	for i := range p.OverdueThresholds {
		t := &p.OverdueThresholds[i]
		if daysOverdue >= t.FromDays {
			if t.ToDays == nil || daysOverdue <= *t.ToDays {
				return t
			}
		}
	}
	return nil
}

// RequiredApproverFor returns the required role to approve a given credit limit.
func (p *Policy) RequiredApproverFor(limitPaise int64) string {
	for i := range p.ApprovalAuthorities {
		a := &p.ApprovalAuthorities[i]
		if limitPaise >= a.FromLimitPaise {
			if a.ToLimitPaise == nil || limitPaise <= *a.ToLimitPaise {
				return a.RequiredRole
			}
		}
	}
	return "credit_manager"
}
