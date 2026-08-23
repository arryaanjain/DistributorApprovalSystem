-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- CREDIT POLICY CONFIGURATION (Phase 0)
-- All business rules stored here — never hard-coded.
-- ============================================================

CREATE TABLE credit_policies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version     TEXT NOT NULL UNIQUE,   -- e.g. '1.0', '1.1'
    name        TEXT NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT FALSE,
    activated_at TIMESTAMPTZ,
    activated_by UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID REFERENCES users(id)
);

-- Only one policy active at a time
CREATE UNIQUE INDEX idx_credit_policies_active ON credit_policies(is_active) WHERE is_active = TRUE;

-- Score bands → initial credit offer mapping
CREATE TABLE policy_score_bands (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id      UUID NOT NULL REFERENCES credit_policies(id) ON DELETE CASCADE,
    min_score      INT NOT NULL,
    max_score      INT NOT NULL,
    eligibility    TEXT NOT NULL,        -- 'CREDIT' | 'ADVANCE_ONLY'
    max_credit_paise BIGINT NOT NULL,    -- stored in paise (1 INR = 100 paise)
    display_label  TEXT NOT NULL,
    CONSTRAINT chk_score_band CHECK (min_score <= max_score)
);

-- Credit ladder steps
CREATE TABLE policy_credit_ladder (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id       UUID NOT NULL REFERENCES credit_policies(id) ON DELETE CASCADE,
    step_order      INT NOT NULL,
    limit_paise     BIGINT NOT NULL,
    display_label   TEXT NOT NULL,       -- e.g. '₹50,000'
    min_cycles      INT NOT NULL DEFAULT 0,
    min_ontime_pct  NUMERIC(5,2) NOT NULL DEFAULT 0,
    min_utilisation_pct NUMERIC(5,2) NOT NULL DEFAULT 0,
    auto_approve    BOOLEAN NOT NULL DEFAULT FALSE,
    approval_role   user_role,
    UNIQUE (policy_id, step_order)
);

-- Credit periods available
CREATE TABLE policy_credit_periods (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id    UUID NOT NULL REFERENCES credit_policies(id) ON DELETE CASCADE,
    code         TEXT NOT NULL,   -- 'COD' | 'RECEIPT' | '15D' | '30D' | 'BTB'
    label        TEXT NOT NULL,
    days         INT,             -- null for COD/RECEIPT/BTB
    is_active    BOOLEAN NOT NULL DEFAULT TRUE
);

-- Risk grades
CREATE TABLE policy_risk_grades (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id     UUID NOT NULL REFERENCES credit_policies(id) ON DELETE CASCADE,
    grade         TEXT NOT NULL,   -- 'A' | 'B' | 'C' | 'D' | 'E'
    min_score     INT NOT NULL,
    max_score     INT NOT NULL,
    label         TEXT NOT NULL,
    description   TEXT,
    max_limit_paise BIGINT
);

-- Overdue thresholds that trigger automated actions
CREATE TABLE policy_overdue_thresholds (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id       UUID NOT NULL REFERENCES credit_policies(id) ON DELETE CASCADE,
    tier            INT NOT NULL,         -- 1,2,3...
    from_days       INT NOT NULL,
    to_days         INT,                  -- null = open-ended
    label           TEXT NOT NULL,
    action_codes    TEXT[] NOT NULL,      -- ['SEND_REMINDER','RESTRICT_CREDIT','NOTIFY_MANAGER']
    auto_restrict   BOOLEAN NOT NULL DEFAULT FALSE,
    auto_hold       BOOLEAN NOT NULL DEFAULT FALSE
);

-- Non-GST cap
CREATE TABLE policy_non_gst_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id       UUID NOT NULL REFERENCES credit_policies(id) ON DELETE CASCADE,
    max_initial_limit_paise BIGINT NOT NULL,  -- ₹25,000 default
    requires_alt_evidence   BOOLEAN NOT NULL DEFAULT TRUE,
    acceptable_evidence     TEXT[] NOT NULL   -- ['FSSAI','UDYAM','SHOP_EST','PHOTO']
);

-- Approval authority thresholds
CREATE TABLE policy_approval_authorities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id       UUID NOT NULL REFERENCES credit_policies(id) ON DELETE CASCADE,
    from_limit_paise BIGINT NOT NULL,
    to_limit_paise   BIGINT,              -- null = no upper bound
    required_role    user_role NOT NULL,
    label            TEXT NOT NULL
);

-- Behavioural enhancement thresholds
CREATE TABLE policy_enhancement_rules (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id           UUID NOT NULL REFERENCES credit_policies(id) ON DELETE CASCADE,
    from_limit_paise    BIGINT NOT NULL,
    to_limit_paise      BIGINT NOT NULL,
    required_cycles     INT NOT NULL,
    required_ontime_pct NUMERIC(5,2) NOT NULL,
    required_util_pct   NUMERIC(5,2) NOT NULL,
    no_current_flags    BOOLEAN NOT NULL DEFAULT TRUE,
    auto_approve        BOOLEAN NOT NULL DEFAULT FALSE,
    approval_role       user_role,
    label               TEXT NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS policy_enhancement_rules;
DROP TABLE IF EXISTS policy_approval_authorities;
DROP TABLE IF EXISTS policy_non_gst_rules;
DROP TABLE IF EXISTS policy_overdue_thresholds;
DROP TABLE IF EXISTS policy_risk_grades;
DROP TABLE IF EXISTS policy_credit_periods;
DROP TABLE IF EXISTS policy_credit_ladder;
DROP TABLE IF EXISTS policy_score_bands;
DROP TABLE IF EXISTS credit_policies;
-- +goose StatementEnd
