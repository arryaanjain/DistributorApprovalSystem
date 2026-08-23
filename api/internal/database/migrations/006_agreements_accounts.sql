-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- AGREEMENTS (Phase 10)
-- ============================================================

CREATE TYPE agreement_status AS ENUM (
    'draft',
    'sent',
    'signed',
    'expired',
    'revoked'
);

-- Agreement templates (versioned)
CREATE TABLE agreement_versions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version      TEXT NOT NULL UNIQUE,
    template     TEXT NOT NULL,      -- agreement template text with placeholders
    is_active    BOOLEAN NOT NULL DEFAULT FALSE,
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_agreement_versions_active ON agreement_versions(is_active) WHERE is_active = TRUE;

-- Generated distributor agreements
CREATE TABLE distributor_agreements (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id       UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    application_id       UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    credit_decision_id   UUID NOT NULL REFERENCES credit_decisions(id) ON DELETE CASCADE,
    agreement_version_id UUID NOT NULL REFERENCES agreement_versions(id),
    -- Content
    agreement_text       TEXT NOT NULL,   -- filled template
    agreement_hash       TEXT NOT NULL,   -- SHA-256 of agreement_text
    -- Credit terms embedded in agreement
    credit_limit_paise   BIGINT NOT NULL,
    credit_period        credit_period_code NOT NULL,
    max_outstanding_days INT,
    -- Surepass Esign integration
    esign_request_id     TEXT,
    esign_doc_url        TEXT,
    -- Status
    status               agreement_status NOT NULL DEFAULT 'draft',
    sent_at              TIMESTAMPTZ,
    signed_at            TIMESTAMPTZ,
    expires_at           TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_distributor_agreements_distributor ON distributor_agreements(distributor_id);
CREATE INDEX idx_distributor_agreements_application ON distributor_agreements(application_id);

-- Signature records
CREATE TABLE agreement_signatures (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agreement_id   UUID NOT NULL REFERENCES distributor_agreements(id) ON DELETE CASCADE,
    signer_type    TEXT NOT NULL,    -- 'distributor' | 'witness' | 'company'
    signer_id      UUID,             -- distributor_id or user_id
    signer_name    TEXT NOT NULL,
    signed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address     INET,
    user_agent     TEXT,
    device_info    JSONB,
    esign_ref      TEXT             -- Surepass signature reference
);

CREATE INDEX idx_agreement_signatures_agreement ON agreement_signatures(agreement_id);

-- ============================================================
-- CREDIT ACCOUNTS (Phase 14, 15)
-- ============================================================

CREATE TYPE account_status AS ENUM (
    'active',
    'restricted',
    'hold',
    'blocked',
    'closed'
);

-- Live credit account per distributor
CREATE TABLE credit_accounts (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id       UUID NOT NULL UNIQUE REFERENCES distributors(id) ON DELETE CASCADE,
    agreement_id         UUID NOT NULL REFERENCES distributor_agreements(id) ON DELETE CASCADE,
    status               account_status NOT NULL DEFAULT 'active',
    -- Current approved terms
    approved_limit_paise BIGINT NOT NULL,
    credit_period        credit_period_code NOT NULL,
    max_outstanding_days INT,
    -- Utilisation (maintained by triggers / application)
    current_outstanding_paise BIGINT NOT NULL DEFAULT 0,
    -- Metadata
    activated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_reviewed_at     TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Constraint: outstanding never exceeds limit
    CONSTRAINT chk_outstanding_within_limit
        CHECK (current_outstanding_paise <= approved_limit_paise)
);

-- History of limit changes (append-only)
CREATE TABLE credit_limit_history (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id        UUID NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    from_limit_paise  BIGINT,
    to_limit_paise    BIGINT NOT NULL,
    reason            TEXT NOT NULL,
    policy_version    TEXT,
    approved_by       UUID REFERENCES users(id),
    effective_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_limit_history_account ON credit_limit_history(account_id);

-- History of period changes (append-only)
CREATE TABLE credit_term_history (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id        UUID NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    from_period       credit_period_code,
    to_period         credit_period_code NOT NULL,
    reason            TEXT NOT NULL,
    approved_by       UUID REFERENCES users(id),
    effective_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_term_history_account ON credit_term_history(account_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS credit_term_history;
DROP TABLE IF EXISTS credit_limit_history;
DROP TABLE IF EXISTS credit_accounts;
DROP TABLE IF EXISTS agreement_signatures;
DROP TABLE IF EXISTS distributor_agreements;
DROP TABLE IF EXISTS agreement_versions;
DROP TYPE IF EXISTS account_status;
DROP TYPE IF EXISTS agreement_status;
-- +goose StatementEnd
