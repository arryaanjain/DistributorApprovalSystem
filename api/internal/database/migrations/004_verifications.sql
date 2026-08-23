-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- VERIFICATION LAYER (Phase 3)
-- Each verification provider returns a normalized status.
-- ============================================================

CREATE TYPE verification_status AS ENUM (
    'pending',
    'verified',
    'partially_verified',
    'mismatch',
    'failed',
    'unavailable'
);

-- PAN verifications
CREATE TABLE pan_verifications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    application_id    UUID REFERENCES applications(id) ON DELETE CASCADE,
    pan               TEXT NOT NULL,
    status            verification_status NOT NULL DEFAULT 'pending',
    name_on_pan       TEXT,
    name_match        BOOLEAN,
    raw_response      JSONB,        -- full Surepass response (encrypted in prod)
    provider          TEXT NOT NULL DEFAULT 'surepass',
    provider_ref      TEXT,         -- Surepass transaction ID
    verified_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pan_verifications_distributor ON pan_verifications(distributor_id);
CREATE INDEX idx_pan_verifications_pan ON pan_verifications(pan);

-- GST verifications
CREATE TABLE gst_verifications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    application_id    UUID REFERENCES applications(id) ON DELETE CASCADE,
    gst_number        TEXT NOT NULL,
    status            verification_status NOT NULL DEFAULT 'pending',
    legal_name        TEXT,
    trade_name        TEXT,
    registration_date DATE,
    gst_status        TEXT,         -- 'Active' | 'Cancelled' | etc.
    address           TEXT,
    constitution      TEXT,
    name_match        BOOLEAN,
    raw_response      JSONB,
    provider          TEXT NOT NULL DEFAULT 'surepass',
    provider_ref      TEXT,
    verified_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gst_verifications_distributor ON gst_verifications(distributor_id);
CREATE INDEX idx_gst_verifications_gst ON gst_verifications(gst_number);

-- FSSAI verifications
CREATE TABLE fssai_verifications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    application_id    UUID REFERENCES applications(id) ON DELETE CASCADE,
    fssai_number      TEXT NOT NULL,
    status            verification_status NOT NULL DEFAULT 'pending',
    holder_name       TEXT,
    expiry_date       DATE,
    raw_response      JSONB,
    provider          TEXT NOT NULL DEFAULT 'surepass',
    provider_ref      TEXT,
    verified_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fssai_verifications_distributor ON fssai_verifications(distributor_id);

-- Bank verifications (name match / penny-drop)
CREATE TABLE bank_verifications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    application_id    UUID REFERENCES applications(id) ON DELETE CASCADE,
    account_number    TEXT NOT NULL,
    ifsc              TEXT NOT NULL,
    status            verification_status NOT NULL DEFAULT 'pending',
    account_holder    TEXT,
    name_match        BOOLEAN,
    bank_name         TEXT,
    raw_response      JSONB,
    provider          TEXT NOT NULL DEFAULT 'surepass',
    provider_ref      TEXT,
    verified_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bank_verifications_distributor ON bank_verifications(distributor_id);

-- Credit bureau reports (CIBIL via Surepass)
CREATE TABLE credit_reports (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    application_id    UUID REFERENCES applications(id) ON DELETE CASCADE,
    pan               TEXT,
    mobile            TEXT,
    bureau_score      INT,
    has_defaults      BOOLEAN,
    has_writeoffs     BOOLEAN,
    has_settlements   BOOLEAN,
    total_active_loans BIGINT,       -- paise
    delinquency_count INT,
    fraud_flag        BOOLEAN NOT NULL DEFAULT FALSE,
    report_date       DATE,
    pdf_url           TEXT,          -- signed URL to PDF report
    raw_response      JSONB,
    provider          TEXT NOT NULL DEFAULT 'surepass_cibil',
    provider_ref      TEXT,
    fetched_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_reports_distributor ON credit_reports(distributor_id);
CREATE INDEX idx_credit_reports_pan ON credit_reports(pan) WHERE pan IS NOT NULL;

-- Alternative business evidence (Phase 4 — non-GST route)
CREATE TABLE alternative_evidence (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id  UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    application_id  UUID REFERENCES applications(id) ON DELETE CASCADE,
    evidence_type   TEXT NOT NULL,   -- 'FSSAI' | 'UDYAM' | 'SHOP_EST' | 'SHOP_PHOTO' | 'SIGNBOARD' | 'ADDRESS'
    document_url    TEXT,            -- signed URL to uploaded file
    notes           TEXT,
    verified_by     UUID REFERENCES users(id),
    verified_at     TIMESTAMPTZ,
    status          verification_status NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alternative_evidence_distributor ON alternative_evidence(distributor_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS alternative_evidence;
DROP TABLE IF EXISTS credit_reports;
DROP TABLE IF EXISTS bank_verifications;
DROP TABLE IF EXISTS fssai_verifications;
DROP TABLE IF EXISTS gst_verifications;
DROP TABLE IF EXISTS pan_verifications;
DROP TYPE IF EXISTS verification_status;
-- +goose StatementEnd
