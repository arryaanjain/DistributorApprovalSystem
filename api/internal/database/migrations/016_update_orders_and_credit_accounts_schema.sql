-- +goose Up
-- +goose StatementBegin

-- 1. Ensure credit_accounts table columns match repository & financial services
ALTER TABLE credit_accounts
    ADD COLUMN IF NOT EXISTS current_credit_paise BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS available_credit_paise BIGINT NOT NULL DEFAULT 0;

ALTER TABLE credit_accounts
    ALTER COLUMN agreement_id DROP NOT NULL,
    ALTER COLUMN credit_period DROP NOT NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'credit_accounts' AND column_name = 'status' AND data_type = 'USER-DEFINED'
    ) THEN
        ALTER TABLE credit_accounts ALTER COLUMN status TYPE TEXT USING status::TEXT;
    END IF;
END $$;

-- 2. Ensure orders table columns match order repository & service expectations
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS total_amount_paise BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS credit_used_paise BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS payment_proof_url TEXT,
    ADD COLUMN IF NOT EXISTS utr_reference TEXT;

ALTER TABLE orders
    ALTER COLUMN subtotal_paise DROP NOT NULL,
    ALTER COLUMN total_paise DROP NOT NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'orders' AND column_name = 'status' AND data_type = 'USER-DEFINED'
    ) THEN
        ALTER TABLE orders ALTER COLUMN status TYPE TEXT USING status::TEXT;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No-op for schema hardening compatibility
-- +goose StatementEnd
