-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS addresses (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    address_type   TEXT NOT NULL DEFAULT 'shipping',
    address_line1  TEXT NOT NULL,
    address_line2  TEXT,
    city           TEXT NOT NULL,
    state          TEXT NOT NULL,
    pin            TEXT NOT NULL,
    phone          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_addresses_distributor ON addresses(distributor_id);

ALTER TABLE sample_orders
    ADD COLUMN IF NOT EXISTS address_id UUID REFERENCES addresses(id) ON DELETE SET NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sample_orders DROP COLUMN IF EXISTS address_id;
DROP TABLE IF EXISTS addresses;
-- +goose StatementEnd
