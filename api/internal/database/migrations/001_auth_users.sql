-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- EXTENSIONS
-- ============================================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE user_role AS ENUM (
    'super_admin',
    'credit_manager',
    'accounts',
    'sales',
    'back_office',
    'dispatch',
    'viewer'
);

CREATE TYPE app_env AS ENUM ('development', 'staging', 'production');

-- ============================================================
-- USERS (internal employees)
-- ============================================================
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          user_role NOT NULL DEFAULT 'viewer',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- ============================================================
-- OTP VERIFICATIONS (distributor mobile auth)
-- ============================================================
CREATE TABLE otp_verifications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mobile       TEXT NOT NULL,
    otp_hash     TEXT NOT NULL,       -- bcrypt hash of the OTP
    purpose      TEXT NOT NULL,       -- 'onboarding' | 'login' | 'agreement'
    expires_at   TIMESTAMPTZ NOT NULL,
    attempts     INT NOT NULL DEFAULT 0,
    verified_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_mobile_purpose ON otp_verifications(mobile, purpose);

-- ============================================================
-- CONSENTS
-- ============================================================
CREATE TABLE consents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id  UUID,             -- may be null before distributor record is created
    mobile          TEXT NOT NULL,
    consent_type    TEXT NOT NULL,    -- 'credit_assessment' | 'data_processing' | 'agreement'
    consent_text    TEXT NOT NULL,
    consent_version TEXT NOT NULL,
    ip_address      INET,
    user_agent      TEXT,
    consented_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS consents;
DROP TABLE IF EXISTS otp_verifications;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS app_env;
DROP TYPE IF EXISTS user_role;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS "pgcrypto";
-- +goose StatementEnd
