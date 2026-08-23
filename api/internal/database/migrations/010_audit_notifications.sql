-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- AUDIT LOG (Phase 23) — immutable event store
-- ============================================================

CREATE TABLE audit_logs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type     TEXT NOT NULL,        -- 'APPLICATION_CREATED' | 'CREDIT_ACTIVATED' | ...
    actor_type     TEXT NOT NULL,        -- 'system' | 'distributor' | 'employee'
    actor_id       UUID,
    actor_name     TEXT,
    -- Subject of the event
    entity_type    TEXT,                 -- 'application' | 'distributor' | 'order' | ...
    entity_id      UUID,
    -- Payload
    before_state   JSONB,
    after_state    JSONB,
    metadata       JSONB,
    -- For credit decisions: full explainability
    score          NUMERIC(5,2),
    risk_grade     TEXT,
    rules_evaluated JSONB,
    policy_version TEXT,
    -- Request context
    ip_address     INET,
    user_agent     TEXT,
    request_id     TEXT,
    -- Timing
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit log is append-only — no updates or deletes
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_type, actor_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

-- ============================================================
-- NOTIFICATIONS (Phase 16, 26)
-- ============================================================

CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed', 'skipped');
CREATE TYPE notification_channel AS ENUM ('email', 'whatsapp', 'system');

CREATE TABLE notifications (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id   UUID REFERENCES distributors(id),
    user_id          UUID REFERENCES users(id),       -- for employee notifications
    channel          notification_channel NOT NULL,
    template_code    TEXT NOT NULL,
    recipient        TEXT NOT NULL,  -- email or phone
    subject          TEXT,
    body             TEXT NOT NULL,
    metadata         JSONB,
    status           notification_status NOT NULL DEFAULT 'pending',
    sent_at          TIMESTAMPTZ,
    error_message    TEXT,
    retry_count      INT NOT NULL DEFAULT 0,
    next_retry_at    TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_pending ON notifications(status, next_retry_at)
    WHERE status IN ('pending', 'failed');
CREATE INDEX idx_notifications_distributor ON notifications(distributor_id);

-- ============================================================
-- IDEMPOTENCY KEYS (Phase 28)
-- ============================================================

CREATE TABLE idempotency_keys (
    key             TEXT PRIMARY KEY,
    endpoint        TEXT NOT NULL,
    status_code     INT NOT NULL,
    response_body   JSONB NOT NULL,
    lock_expires_at TIMESTAMPTZ,
    completed       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours'
);

CREATE INDEX idx_idempotency_expiry ON idempotency_keys(expires_at);

-- ============================================================
-- DUPLICATE DETECTION (Phase 1)
-- ============================================================

-- Known duplicate suspect pairs for review
CREATE TABLE duplicate_suspects (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id_a   UUID NOT NULL REFERENCES applications(id),
    application_id_b   UUID NOT NULL REFERENCES applications(id),
    match_fields       TEXT[] NOT NULL,   -- which identifiers matched
    confidence         TEXT NOT NULL,     -- 'high' | 'medium' | 'low'
    resolved           BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_by        UUID REFERENCES users(id),
    resolution         TEXT,              -- 'legitimate' | 'duplicate' | 'fraud'
    detected_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_duplicate_suspects_unresolved ON duplicate_suspects(resolved) WHERE resolved = FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS duplicate_suspects;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS audit_logs;
DROP TYPE IF EXISTS notification_channel;
DROP TYPE IF EXISTS notification_status;
-- +goose StatementEnd
