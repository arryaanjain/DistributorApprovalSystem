-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- FINANCIAL LEDGER (Phases 15, 16, 20)
-- ============================================================

CREATE TYPE invoice_status AS ENUM (
    'open',
    'partially_paid',
    'paid',
    'overdue',
    'written_off'
);

CREATE TYPE overdue_tier AS ENUM (
    'current',
    'overdue_1_3',
    'overdue_4_7',
    'overdue_8_15',
    'serious'
);

-- Invoices (immutable financial records)
CREATE TABLE invoices (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number       TEXT NOT NULL UNIQUE,
    distributor_id       UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    account_id           UUID NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    order_id             UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    invoice_date         DATE NOT NULL,
    due_date             DATE NOT NULL,
    -- Amounts
    total_paise          BIGINT NOT NULL,
    credit_paise         BIGINT NOT NULL,    -- portion on credit
    advance_paise        BIGINT NOT NULL,    -- portion paid upfront
    outstanding_paise    BIGINT NOT NULL,    -- remaining to pay
    -- Status
    status               invoice_status NOT NULL DEFAULT 'open',
    days_outstanding     INT NOT NULL DEFAULT 0,
    overdue_tier         overdue_tier NOT NULL DEFAULT 'current',
    -- Timestamps
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_distributor ON invoices(distributor_id);
CREATE INDEX idx_invoices_account ON invoices(account_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_invoices_overdue ON invoices(distributor_id, status)
    WHERE status IN ('open', 'partially_paid', 'overdue');

-- Payments received against invoices
CREATE TABLE invoice_payments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id       UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    distributor_id   UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    amount_paise     BIGINT NOT NULL,
    payment_mode     TEXT NOT NULL,
    utr_reference    TEXT,
    payment_date     DATE NOT NULL,
    recorded_by      UUID REFERENCES users(id),
    notes            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoice_payments_invoice ON invoice_payments(invoice_id);

-- Payment allocation (which payment covers which invoice)
CREATE TABLE payment_allocations (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id       UUID NOT NULL REFERENCES invoice_payments(id) ON DELETE CASCADE,
    invoice_id       UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount_paise     BIGINT NOT NULL,
    allocated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Credit transactions ledger (immutable append-only)
CREATE TABLE credit_transactions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id    UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    account_id        UUID NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    transaction_type  TEXT NOT NULL,  -- 'CREDIT_UTILIZED' | 'PAYMENT_RECEIVED' | 'CREDIT_REVERSED'
    amount_paise      BIGINT NOT NULL,
    reference_id      UUID,           -- order_id or invoice_id
    reference_type    TEXT,           -- 'order' | 'invoice'
    balance_after_paise BIGINT NOT NULL,
    notes             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_transactions_account ON credit_transactions(account_id);
CREATE INDEX idx_credit_transactions_distributor ON credit_transactions(distributor_id);

-- Outstanding ledger snapshot (point-in-time reference)
CREATE TABLE outstanding_ledger (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id        UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    account_id            UUID NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    snapshot_date         DATE NOT NULL,
    total_outstanding_paise BIGINT NOT NULL,
    total_overdue_paise   BIGINT NOT NULL,
    oldest_overdue_days   INT NOT NULL DEFAULT 0,
    invoice_count         INT NOT NULL DEFAULT 0,
    overdue_invoice_count INT NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_outstanding_ledger_daily ON outstanding_ledger(account_id, snapshot_date);

-- Collections actions log (Phase 16)
CREATE TABLE collections_actions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id   UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    invoice_id       UUID REFERENCES invoices(id) ON DELETE CASCADE,
    action_type      TEXT NOT NULL,   -- 'REMINDER_SENT' | 'ACCOUNT_RESTRICTED' | 'CREDIT_HOLD' | 'ESCALATED'
    channel          TEXT,            -- 'email' | 'whatsapp' | 'system'
    overdue_tier     overdue_tier,
    message          TEXT,
    triggered_by     TEXT NOT NULL,   -- 'system' | 'employee'
    triggered_by_id  UUID REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collections_actions_distributor ON collections_actions(distributor_id);

-- Overdue events (Phase 16)
CREATE TABLE overdue_events (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id       UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    distributor_id   UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    account_id       UUID NOT NULL REFERENCES credit_accounts(id) ON DELETE CASCADE,
    tier             overdue_tier NOT NULL,
    days_overdue     INT NOT NULL,
    auto_restricted  BOOLEAN NOT NULL DEFAULT FALSE,
    auto_held        BOOLEAN NOT NULL DEFAULT FALSE,
    detected_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_overdue_events_distributor ON overdue_events(distributor_id);
CREATE INDEX idx_overdue_events_invoice ON overdue_events(invoice_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS overdue_events;
DROP TABLE IF EXISTS collections_actions;
DROP TABLE IF EXISTS outstanding_ledger;
DROP TABLE IF EXISTS credit_transactions;
DROP TABLE IF EXISTS payment_allocations;
DROP TABLE IF EXISTS invoice_payments;
DROP TABLE IF EXISTS invoices;
DROP TYPE IF EXISTS overdue_tier;
DROP TYPE IF EXISTS invoice_status;
-- +goose StatementEnd
