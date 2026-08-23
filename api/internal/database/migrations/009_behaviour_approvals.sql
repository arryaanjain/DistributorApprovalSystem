-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- BEHAVIOURAL CREDIT ENGINE (Phases 17, 18, 19)
-- ============================================================

-- Behavioural score (recalculated periodically from Kresconet's own data)
CREATE TABLE behaviour_scores (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id        UUID NOT NULL REFERENCES distributors(id),
    account_id            UUID NOT NULL REFERENCES credit_accounts(id),
    -- Payment behaviour
    ontime_payment_pct    NUMERIC(5,2) NOT NULL DEFAULT 0,
    avg_days_late         NUMERIC(5,2) NOT NULL DEFAULT 0,
    successful_cycles     INT NOT NULL DEFAULT 0,
    payment_failures      INT NOT NULL DEFAULT 0,
    bounces               INT NOT NULL DEFAULT 0,
    -- Utilisation behaviour
    avg_utilisation_pct   NUMERIC(5,2) NOT NULL DEFAULT 0,
    max_outstanding_paise BIGINT NOT NULL DEFAULT 0,
    -- Purchase behaviour
    avg_order_paise       BIGINT NOT NULL DEFAULT 0,
    order_frequency_days  NUMERIC(5,2),   -- avg days between orders
    claims_count          INT NOT NULL DEFAULT 0,
    -- Composite score
    behaviour_score       NUMERIC(5,2) NOT NULL DEFAULT 0,
    relationship_days     INT NOT NULL DEFAULT 0,
    -- Metadata
    calculated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    calculated_for_period TEXT         -- e.g. '2026-Q1'
);

CREATE INDEX idx_behaviour_scores_distributor ON behaviour_scores(distributor_id);
CREATE INDEX idx_behaviour_scores_latest ON behaviour_scores(distributor_id, calculated_at DESC);

-- Individual behaviour events that feed the score
CREATE TABLE behaviour_events (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id   UUID NOT NULL REFERENCES distributors(id),
    account_id       UUID NOT NULL REFERENCES credit_accounts(id),
    event_type       TEXT NOT NULL,   -- 'PAYMENT_ON_TIME' | 'PAYMENT_LATE_N_DAYS' | 'PAYMENT_BOUNCE' | 'ORDER_PLACED'
    reference_id     UUID,
    reference_type   TEXT,
    days_late        INT,
    amount_paise     BIGINT,
    occurred_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_behaviour_events_distributor ON behaviour_events(distributor_id, occurred_at DESC);

-- Credit reviews (periodic assessment)
CREATE TABLE credit_reviews (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id    UUID NOT NULL REFERENCES distributors(id),
    account_id        UUID NOT NULL REFERENCES credit_accounts(id),
    review_type       TEXT NOT NULL,   -- 'periodic' | 'triggered' | 'requested'
    trigger_reason    TEXT,
    behaviour_score_id UUID REFERENCES behaviour_scores(id),
    recommendation    TEXT NOT NULL,   -- 'ENHANCE' | 'MAINTAIN' | 'REDUCE' | 'SUSPEND'
    recommendation_notes TEXT,
    reviewed_by       UUID REFERENCES users(id),
    reviewed_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_reviews_distributor ON credit_reviews(distributor_id);

-- Credit enhancement records (Phase 18)
CREATE TABLE credit_enhancements (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id      UUID NOT NULL REFERENCES distributors(id),
    account_id          UUID NOT NULL REFERENCES credit_accounts(id),
    review_id           UUID REFERENCES credit_reviews(id),
    from_limit_paise    BIGINT NOT NULL,
    to_limit_paise      BIGINT NOT NULL,
    reason              TEXT NOT NULL,
    policy_version      TEXT,
    successful_cycles   INT,
    ontime_pct          NUMERIC(5,2),
    avg_utilisation_pct NUMERIC(5,2),
    -- Approval
    recommended_by      TEXT NOT NULL,  -- 'system' | user_id
    approved_by         UUID REFERENCES users(id),
    approved_at         TIMESTAMPTZ,
    effective_at        TIMESTAMPTZ,
    status              TEXT NOT NULL DEFAULT 'pending',  -- 'pending' | 'approved' | 'rejected'
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_enhancements_distributor ON credit_enhancements(distributor_id);
CREATE INDEX idx_credit_enhancements_pending ON credit_enhancements(status) WHERE status = 'pending';

-- Credit reductions (Phase 19)
CREATE TABLE credit_reductions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id    UUID NOT NULL REFERENCES distributors(id),
    account_id        UUID NOT NULL REFERENCES credit_accounts(id),
    review_id         UUID REFERENCES credit_reviews(id),
    from_limit_paise  BIGINT,
    to_limit_paise    BIGINT,
    from_status       account_status,
    to_status         account_status NOT NULL,
    trigger_type      TEXT NOT NULL,  -- 'late_payments' | 'default' | 'bounce' | 'fraud' | 'deterioration'
    trigger_notes     TEXT,
    initiated_by      TEXT NOT NULL,  -- 'system' | user_id
    approved_by       UUID REFERENCES users(id),
    effective_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_reductions_distributor ON credit_reductions(distributor_id);

-- ============================================================
-- APPROVAL WORKFLOW (Phase 22)
-- ============================================================

CREATE TYPE approval_status AS ENUM ('pending', 'approved', 'rejected', 'escalated', 'expired');

CREATE TABLE approval_requests (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_type     TEXT NOT NULL,   -- 'CREDIT_DECISION' | 'ENHANCEMENT' | 'REDUCTION' | 'ORDER_REVIEW'
    reference_id     UUID NOT NULL,   -- ID of the thing being approved
    reference_type   TEXT NOT NULL,
    required_role    user_role NOT NULL,
    status           approval_status NOT NULL DEFAULT 'pending',
    requested_by     UUID REFERENCES users(id),  -- null if system-initiated
    summary          TEXT,
    metadata         JSONB,
    expires_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_approval_requests_status ON approval_requests(status);
CREATE INDEX idx_approval_requests_role ON approval_requests(required_role, status);

CREATE TABLE approval_actions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id        UUID NOT NULL REFERENCES approval_requests(id),
    action            TEXT NOT NULL,   -- 'approved' | 'rejected' | 'escalated'
    actor_id          UUID NOT NULL REFERENCES users(id),
    notes             TEXT,
    metadata          JSONB,
    acted_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_approval_actions_request ON approval_actions(request_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS approval_actions;
DROP TABLE IF EXISTS approval_requests;
DROP TABLE IF EXISTS credit_reductions;
DROP TABLE IF EXISTS credit_enhancements;
DROP TABLE IF EXISTS credit_reviews;
DROP TABLE IF EXISTS behaviour_events;
DROP TABLE IF EXISTS behaviour_scores;
DROP TYPE IF EXISTS approval_status;
-- +goose StatementEnd
