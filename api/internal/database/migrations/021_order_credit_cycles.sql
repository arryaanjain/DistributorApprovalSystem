-- +goose Up
-- +goose StatementBegin

-- 1. Modify distributor_agreements to support order-level agreements & repository fields
ALTER TABLE distributor_agreements ALTER COLUMN application_id DROP NOT NULL;
ALTER TABLE distributor_agreements ALTER COLUMN credit_decision_id DROP NOT NULL;
ALTER TABLE distributor_agreements ALTER COLUMN agreement_version_id DROP NOT NULL;
ALTER TABLE distributor_agreements ALTER COLUMN agreement_text DROP NOT NULL;
ALTER TABLE distributor_agreements ALTER COLUMN agreement_hash DROP NOT NULL;
ALTER TABLE distributor_agreements ALTER COLUMN credit_limit_paise DROP NOT NULL;
ALTER TABLE distributor_agreements ALTER COLUMN credit_period DROP NOT NULL;

ALTER TABLE distributor_agreements ADD COLUMN IF NOT EXISTS order_id UUID REFERENCES orders(id) ON DELETE CASCADE;
ALTER TABLE distributor_agreements ADD COLUMN IF NOT EXISTS agreement_number TEXT;
ALTER TABLE distributor_agreements ADD COLUMN IF NOT EXISTS version TEXT;
ALTER TABLE distributor_agreements ADD COLUMN IF NOT EXISTS approved_limit_paise BIGINT DEFAULT 0;
ALTER TABLE distributor_agreements ADD COLUMN IF NOT EXISTS approved_period_days INT DEFAULT 30;
ALTER TABLE distributor_agreements ADD COLUMN IF NOT EXISTS document_url TEXT;
ALTER TABLE distributor_agreements ADD COLUMN IF NOT EXISTS esign_provider_ref TEXT;

-- 2. Modify orders to track cycle properties
ALTER TABLE orders ADD COLUMN IF NOT EXISTS is_cycle_order BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS cycle_frequency TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS total_instalments INT NOT NULL DEFAULT 1;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS grace_period_days INT NOT NULL DEFAULT 7;

-- 3. Order Credit Cycles table
CREATE TABLE IF NOT EXISTS order_credit_cycles (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id                UUID NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    distributor_id          UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    account_id              UUID REFERENCES credit_accounts(id) ON DELETE SET NULL,
    agreement_id            UUID REFERENCES distributor_agreements(id) ON DELETE SET NULL,
    total_amount_paise      BIGINT NOT NULL,
    advance_paid_paise      BIGINT NOT NULL DEFAULT 0,
    credit_amount_paise     BIGINT NOT NULL,
    instalment_amount_paise BIGINT NOT NULL,
    frequency               TEXT NOT NULL, -- 'weekly' | 'monthly'
    total_instalments       INT NOT NULL,  -- <=5 for weekly, <=2 for monthly
    completed_instalments   INT NOT NULL DEFAULT 0,
    status                  TEXT NOT NULL DEFAULT 'draft', -- 'draft', 'esign_pending', 'esign_completed', 'mandate_pending', 'mandate_active', 'repaying', 'completed', 'defaulted'
    grace_period_days       INT NOT NULL DEFAULT 7,
    grace_period_ends_at    TIMESTAMPTZ,
    hold_triggered_at       TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_credit_cycles_distributor ON order_credit_cycles(distributor_id);
CREATE INDEX IF NOT EXISTS idx_order_credit_cycles_status ON order_credit_cycles(status);
CREATE INDEX IF NOT EXISTS idx_order_credit_cycles_order ON order_credit_cycles(order_id);

-- 4. UPI Mandates (Razorpay Subscriptions) table
CREATE TABLE IF NOT EXISTS upi_mandates (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id                    UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    cycle_id                    UUID NOT NULL REFERENCES order_credit_cycles(id) ON DELETE CASCADE,
    distributor_id              UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    razorpay_plan_id            TEXT NOT NULL,
    razorpay_subscription_id    TEXT NOT NULL UNIQUE,
    razorpay_short_url          TEXT,
    mandate_status              TEXT NOT NULL DEFAULT 'created', -- 'created', 'authenticated', 'active', 'paused', 'failed', 'cancelled', 'completed'
    frequency                   TEXT NOT NULL, -- 'weekly' | 'monthly'
    total_count                 INT NOT NULL,
    paid_count                  INT NOT NULL DEFAULT 0,
    amount_per_instalment_paise BIGINT NOT NULL,
    start_at                    TIMESTAMPTZ,
    next_charge_at              TIMESTAMPTZ,
    authenticated_at            TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_upi_mandates_cycle ON upi_mandates(cycle_id);
CREATE INDEX IF NOT EXISTS idx_upi_mandates_distributor ON upi_mandates(distributor_id);
CREATE INDEX IF NOT EXISTS idx_upi_mandates_rzp_sub ON upi_mandates(razorpay_subscription_id);

-- 5. Cycle Repayments table
CREATE TABLE IF NOT EXISTS cycle_repayments (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_id             UUID NOT NULL REFERENCES order_credit_cycles(id) ON DELETE CASCADE,
    mandate_id           UUID REFERENCES upi_mandates(id) ON DELETE SET NULL,
    instalment_number    INT NOT NULL,
    due_date             DATE NOT NULL,
    amount_paise         BIGINT NOT NULL,
    status               TEXT NOT NULL DEFAULT 'scheduled', -- 'scheduled', 'pending', 'paid', 'failed', 'bounced'
    razorpay_payment_id  TEXT,
    razorpay_invoice_id  TEXT,
    paid_at              TIMESTAMPTZ,
    failure_reason       TEXT,
    retry_count          INT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cycle_repayments_cycle ON cycle_repayments(cycle_id);
CREATE INDEX IF NOT EXISTS idx_cycle_repayments_due_date ON cycle_repayments(due_date);
CREATE INDEX IF NOT EXISTS idx_cycle_repayments_status ON cycle_repayments(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cycle_repayments;
DROP TABLE IF EXISTS upi_mandates;
DROP TABLE IF EXISTS order_credit_cycles;

ALTER TABLE orders DROP COLUMN IF EXISTS grace_period_days;
ALTER TABLE orders DROP COLUMN IF EXISTS total_instalments;
ALTER TABLE orders DROP COLUMN IF EXISTS cycle_frequency;
ALTER TABLE orders DROP COLUMN IF EXISTS is_cycle_order;

ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS esign_provider_ref;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS document_url;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS approved_period_days;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS approved_limit_paise;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS version;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS agreement_number;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS order_id;
-- +goose StatementEnd
