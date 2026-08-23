-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- CREDIT SCORING ENGINE (Phases 6, 7, 8)
-- ============================================================

CREATE TYPE eligibility_decision AS ENUM (
    'credit',
    'advance_only',
    'hold',
    'blocked'
);

CREATE TYPE credit_period_code AS ENUM (
    'cod',
    'receipt',
    '15_days',
    '30_days',
    'bill_to_bill'
);

CREATE TYPE risk_grade AS ENUM ('A', 'B', 'C', 'D', 'E');

-- Scoring result for an application
CREATE TABLE credit_scores (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id    UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    policy_version    TEXT NOT NULL,   -- which policy was used
    total_score       NUMERIC(5,2) NOT NULL,
    risk_grade        risk_grade NOT NULL,
    -- Component scores (weights from policy)
    score_credit_risk     NUMERIC(5,2),
    score_identity_kyc    NUMERIC(5,2),
    score_biz_verification NUMERIC(5,2),
    score_biz_vintage     NUMERIC(5,2),
    score_dist_experience NUMERIC(5,2),
    score_biz_capacity    NUMERIC(5,2),
    score_data_consistency NUMERIC(5,2),
    -- Inputs snapshot (for explainability)
    inputs            JSONB NOT NULL,  -- all values used in scoring
    calculated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    calculated_by     TEXT NOT NULL DEFAULT 'system'  -- 'system' | user id
);

CREATE INDEX idx_credit_scores_application ON credit_scores(application_id);
CREATE INDEX idx_credit_scores_distributor ON credit_scores(distributor_id);

-- Individual score component detail (for deep explainability)
CREATE TABLE credit_score_components (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_score_id  UUID NOT NULL REFERENCES credit_scores(id) ON DELETE CASCADE,
    component_name   TEXT NOT NULL,
    raw_value        TEXT,
    normalized_score NUMERIC(5,2),
    weight           NUMERIC(5,2),
    weighted_score   NUMERIC(5,2),
    notes            TEXT
);

CREATE INDEX idx_score_components_score ON credit_score_components(credit_score_id);

-- Hard risk flags (Phase 7) — evaluated BEFORE score decision
CREATE TABLE risk_flags (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id   UUID REFERENCES applications(id) ON DELETE CASCADE,
    distributor_id   UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    flag_code        TEXT NOT NULL,   -- 'INVALID_PAN' | 'BANK_MISMATCH' | 'FRAUD_INDICATOR' | ...
    flag_description TEXT NOT NULL,
    severity         TEXT NOT NULL,   -- 'hard' | 'soft'
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    triggered_by     TEXT NOT NULL,   -- 'system' | 'employee'
    triggered_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_by      UUID REFERENCES users(id),
    resolved_at      TIMESTAMPTZ,
    resolution_notes TEXT
);

CREATE INDEX idx_risk_flags_distributor ON risk_flags(distributor_id);
CREATE INDEX idx_risk_flags_application ON risk_flags(application_id);
CREATE INDEX idx_risk_flags_active ON risk_flags(distributor_id, is_active) WHERE is_active = TRUE;

-- Final credit decisions (3-part: eligibility + limit + period)
CREATE TABLE credit_decisions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id      UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    distributor_id      UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    credit_score_id     UUID REFERENCES credit_scores(id) ON DELETE CASCADE,
    policy_version      TEXT NOT NULL,
    -- Decision 1: Eligibility
    eligibility         eligibility_decision NOT NULL,
    -- Decision 2: Credit Limit
    approved_limit_paise BIGINT,           -- null if advance_only/blocked
    -- Decision 3: Credit Period
    approved_period     credit_period_code,
    -- Additional controls
    max_outstanding_days INT,
    hard_flags_present  BOOLEAN NOT NULL DEFAULT FALSE,
    -- Who/what made the decision
    decision_source     TEXT NOT NULL,     -- 'auto' | 'employee' | 'override'
    decided_by          UUID REFERENCES users(id),  -- null if auto
    decision_notes      TEXT,
    decided_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_decisions_application ON credit_decisions(application_id);
CREATE INDEX idx_credit_decisions_distributor ON credit_decisions(distributor_id);

-- Credit offers presented to the distributor
CREATE TABLE credit_offers (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id       UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    distributor_id       UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    credit_decision_id   UUID NOT NULL REFERENCES credit_decisions(id) ON DELETE CASCADE,
    risk_grade           risk_grade NOT NULL,
    offered_limit_paise  BIGINT NOT NULL,
    offered_period       credit_period_code NOT NULL,
    max_outstanding_days INT,
    offer_text           TEXT,            -- human-readable summary
    expires_at           TIMESTAMPTZ NOT NULL,
    accepted_at          TIMESTAMPTZ,
    declined_at          TIMESTAMPTZ,
    decision             TEXT,            -- 'accepted' | 'declined' | 'advance_only' | null
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_offers_application ON credit_offers(application_id);
CREATE INDEX idx_credit_offers_distributor ON credit_offers(distributor_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS credit_offers;
DROP TABLE IF EXISTS credit_decisions;
DROP TABLE IF EXISTS risk_flags;
DROP TABLE IF EXISTS credit_score_components;
DROP TABLE IF EXISTS credit_scores;
DROP TYPE IF EXISTS risk_grade;
DROP TYPE IF EXISTS credit_period_code;
DROP TYPE IF EXISTS eligibility_decision;
-- +goose StatementEnd
