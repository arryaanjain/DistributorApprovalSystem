-- +goose Up
-- +goose StatementBegin

-- 1. Extend application_status enum with 'trial'
ALTER TYPE application_status ADD VALUE IF NOT EXISTS 'trial';

-- 2. Extend business_profiles with new step 2 fields
ALTER TABLE business_profiles
    ADD COLUMN IF NOT EXISTS distribution_experience_years NUMERIC(4,1),
    ADD COLUMN IF NOT EXISTS serviced_retailers_wholesalers_count INT,
    ADD COLUMN IF NOT EXISTS interested_business_role TEXT;

-- 3. Ensure products table exists (rename catalogue_items if present, or create)
DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'catalogue_items')
       AND NOT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'products') THEN
        ALTER TABLE catalogue_items RENAME TO products;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS products (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku           TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    description   TEXT,
    category      TEXT NOT NULL DEFAULT 'Edible Oils',
    unit          TEXT NOT NULL DEFAULT 'case',
    price_paise   BIGINT NOT NULL DEFAULT 0,
    mrp_paise     BIGINT,
    moq           INT NOT NULL DEFAULT 1,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    is_sample     BOOLEAN NOT NULL DEFAULT FALSE,
    is_regular    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS moq INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS is_sample BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_regular BOOLEAN NOT NULL DEFAULT TRUE;

-- 4. Ensure order_items table has missing columns used by repository
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS product_id UUID,
    ADD COLUMN IF NOT EXISTS product_name TEXT,
    ADD COLUMN IF NOT EXISTS total_price_paise BIGINT;

ALTER TABLE order_items
    ALTER COLUMN sku DROP NOT NULL,
    ALTER COLUMN name DROP NOT NULL,
    ALTER COLUMN subtotal_paise DROP NOT NULL,
    ALTER COLUMN catalogue_id DROP NOT NULL;

-- 5. Create sample_orders table for Razorpay Book a Sample flow
CREATE TABLE IF NOT EXISTS sample_orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distributor_id       UUID NOT NULL REFERENCES distributors(id) ON DELETE CASCADE,
    razorpay_order_id   TEXT NOT NULL UNIQUE,
    razorpay_payment_id TEXT,
    razorpay_signature  TEXT,
    amount_paise        BIGINT NOT NULL DEFAULT 50000, -- ₹500 default sample price
    status              TEXT NOT NULL DEFAULT 'CREATED',
    items               JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sample_orders_distributor ON sample_orders(distributor_id);
CREATE INDEX IF NOT EXISTS idx_sample_orders_razorpay ON sample_orders(razorpay_order_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sample_orders;
ALTER TABLE products DROP COLUMN IF EXISTS is_regular;
ALTER TABLE products DROP COLUMN IF EXISTS is_sample;
ALTER TABLE products DROP COLUMN IF EXISTS moq;
ALTER TABLE business_profiles DROP COLUMN IF EXISTS interested_business_role;
ALTER TABLE business_profiles DROP COLUMN IF EXISTS serviced_retailers_wholesalers_count;
ALTER TABLE business_profiles DROP COLUMN IF EXISTS distribution_experience_years;
-- +goose StatementEnd
