-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- DISTRIBUTORS (Phase 1)
-- ============================================================

CREATE TYPE constitution_type AS ENUM (
    'proprietorship',
    'partnership',
    'llp',
    'private_limited',
    'public_limited',
    'huf',
    'trust',
    'other'
);

CREATE TYPE application_status AS ENUM (
    'draft',
    'mobile_verified',
    'basic_submitted',
    'business_submitted',
    'statutory_submitted',
    'bank_submitted',
    'preference_submitted',
    'consent_given',
    'verification_pending',
    'verification_complete',
    'scoring',
    'scored',
    'offer_generated',
    'offer_accepted',
    'offer_declined',
    'agreement_pending',
    'agreement_signed',
    'credit_active',
    'advance_only',
    'hold',
    'rejected',
    'blocked'
);

CREATE TYPE payment_preference AS ENUM (
    'advance_100',       -- 100% Advance
    'partial_delivery',  -- Partial Advance + Delivery
    'partial_receipt',   -- Partial Advance + Receipt
    'partial_15d',       -- Partial Advance + 15 Days
    'cod',               -- COD
    '15_days',
    '30_days',
    'bill_to_bill'
);

CREATE TYPE exposure_class AS ENUM (
    'low_no',     -- LOW/NO EXPOSURE
    'short',      -- SHORT EXPOSURE
    'standard',   -- STANDARD CREDIT
    'extended'    -- EXTENDED CREDIT
);

-- Master distributor record
CREATE TABLE distributors (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mobile      TEXT NOT NULL UNIQUE,
    email       TEXT,
    name        TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_distributors_mobile ON distributors(mobile);
CREATE INDEX idx_distributors_email  ON distributors(email);

-- Business profile
CREATE TABLE business_profiles (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id        UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    business_name         TEXT NOT NULL,
    constitution          constitution_type NOT NULL,
    address_line1         TEXT NOT NULL,
    address_line2         TEXT,
    city                  TEXT NOT NULL,
    state                 TEXT NOT NULL,
    pin                   TEXT NOT NULL,
    vintage_years         NUMERIC(4,1),     -- years in business
    fmcg_experience_years NUMERIC(4,1),
    approx_monthly_business_paise BIGINT,
    retailer_count        INT,
    salesperson_count     INT,
    existing_brands       TEXT[],
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_business_profiles_distributor ON business_profiles(distributor_id);

-- Statutory details
CREATE TABLE business_documents (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    pan            TEXT,
    gst_number     TEXT,
    fssai_number   TEXT,
    udyam_number   TEXT,
    shop_est_number TEXT,
    has_gst        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_business_documents_distributor ON business_documents(distributor_id);
CREATE INDEX idx_business_documents_pan ON business_documents(pan) WHERE pan IS NOT NULL;
CREATE INDEX idx_business_documents_gst ON business_documents(gst_number) WHERE gst_number IS NOT NULL;

-- Bank details
CREATE TABLE bank_details (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id  UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    account_number  TEXT NOT NULL,
    ifsc            TEXT NOT NULL,
    account_holder  TEXT NOT NULL,
    bank_name       TEXT,
    branch          TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_bank_details_distributor ON bank_details(distributor_id);

-- Applications (one per onboarding attempt)
CREATE TABLE applications (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id   UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    status           application_status NOT NULL DEFAULT 'draft',
    payment_preference payment_preference,
    exposure_class   exposure_class,
    -- Duplicate detection flags
    is_duplicate_suspect BOOLEAN NOT NULL DEFAULT FALSE,
    duplicate_reason     TEXT,
    -- Submission tracking
    submitted_at     TIMESTAMPTZ,
    completed_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_applications_distributor ON applications(distributor_id);
CREATE INDEX idx_applications_status ON applications(status);

-- Application event log (immutable state machine audit)
CREATE TABLE application_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    from_status    application_status,
    to_status      application_status NOT NULL,
    actor_type     TEXT NOT NULL,   -- 'distributor' | 'system' | 'employee'
    actor_id       UUID,
    reason         TEXT,
    metadata       JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_application_events_application ON application_events(application_id);
CREATE INDEX idx_application_events_created ON application_events(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS application_events;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS bank_details;
DROP TABLE IF EXISTS business_documents;
DROP TABLE IF EXISTS business_profiles;
DROP TABLE IF EXISTS distributors;
DROP TYPE IF EXISTS exposure_class;
DROP TYPE IF EXISTS payment_preference;
DROP TYPE IF EXISTS application_status;
DROP TYPE IF EXISTS constitution_type;
-- +goose StatementEnd
