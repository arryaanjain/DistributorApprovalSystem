-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- CATALOGUE, ORDERS, PAYMENTS (Phases 11, 12)
-- ============================================================

CREATE TYPE order_status AS ENUM (
    'draft',
    'pending_advance',       -- waiting for advance payment
    'advance_received',      -- advance screenshot uploaded
    'advance_verified',      -- admin verified advance payment
    'pending_review',        -- awaiting first human review
    'approved',
    'dispatched',
    'delivered',
    'completed',
    'cancelled',
    'hold'
);

-- Product catalogue
CREATE TABLE catalogue_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku           TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    description   TEXT,
    category      TEXT,
    unit          TEXT NOT NULL DEFAULT 'case',
    price_paise   BIGINT NOT NULL,
    mrp_paise     BIGINT,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Orders (Phase 11 — order ≠ credit limit)
CREATE TABLE orders (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id       UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    account_id           UUID REFERENCES credit_accounts(id) ON DELETE CASCADE,
    order_number         TEXT NOT NULL UNIQUE,
    status               order_status NOT NULL DEFAULT 'draft',
    -- Financials
    subtotal_paise       BIGINT NOT NULL,
    tax_paise            BIGINT NOT NULL DEFAULT 0,
    total_paise          BIGINT NOT NULL,
    -- Credit split (Phase 11 core logic)
    advance_required_paise BIGINT NOT NULL DEFAULT 0,
    credit_utilized_paise  BIGINT NOT NULL DEFAULT 0,
    -- Manual payment tracking (Phase 12)
    advance_paid_paise   BIGINT NOT NULL DEFAULT 0,
    advance_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    advance_verified_by  UUID REFERENCES users(id),
    advance_verified_at  TIMESTAMPTZ,
    -- First human review (Phase 13)
    reviewed_by          UUID REFERENCES users(id),
    reviewed_at          TIMESTAMPTZ,
    review_action        TEXT,       -- 'approved' | 'hold' | 'revised'
    review_notes         TEXT,
    -- Idempotency
    idempotency_key      TEXT UNIQUE,
    -- Delivery
    dispatched_at        TIMESTAMPTZ,
    delivered_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_distributor ON orders(distributor_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_account ON orders(account_id);

CREATE TABLE order_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    catalogue_id    UUID NOT NULL REFERENCES catalogue_items(id),
    sku             TEXT NOT NULL,
    name            TEXT NOT NULL,
    quantity        INT NOT NULL,
    unit_price_paise BIGINT NOT NULL,
    subtotal_paise  BIGINT NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items(order_id);

-- Manual payment proofs (Phase 12 — distributor uploads screenshot)
CREATE TABLE payment_proofs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id         UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    distributor_id   UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    amount_paise     BIGINT NOT NULL,
    payment_mode     TEXT NOT NULL,  -- 'neft' | 'rtgs' | 'imps' | 'upi' | 'cash' | 'cheque'
    utr_reference    TEXT,
    screenshot_url   TEXT NOT NULL,  -- signed URL to uploaded screenshot
    notes            TEXT,
    -- Admin verification
    verified         BOOLEAN NOT NULL DEFAULT FALSE,
    verified_by      UUID REFERENCES users(id),
    verified_at      TIMESTAMPTZ,
    rejection_reason TEXT,
    submitted_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_proofs_order ON payment_proofs(order_id);
CREATE INDEX idx_payment_proofs_distributor ON payment_proofs(distributor_id);
CREATE INDEX idx_payment_proofs_unverified ON payment_proofs(verified) WHERE verified = FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment_proofs;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS catalogue_items;
DROP TYPE IF EXISTS order_status;
-- +goose StatementEnd
