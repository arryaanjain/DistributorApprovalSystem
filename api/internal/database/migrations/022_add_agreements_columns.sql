-- +goose Up
-- +goose StatementBegin

-- Modify distributor_agreements to add missing agreement columns for order-level credit agreements
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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS esign_provider_ref;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS document_url;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS approved_period_days;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS approved_limit_paise;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS version;
ALTER TABLE distributor_agreements DROP COLUMN IF EXISTS agreement_number;
-- +goose StatementEnd
