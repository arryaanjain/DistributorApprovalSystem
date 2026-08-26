-- +goose Up
-- +goose StatementBegin

ALTER TABLE sample_orders
    ADD COLUMN IF NOT EXISTS shiprocket_order_id TEXT,
    ADD COLUMN IF NOT EXISTS shipment_id TEXT,
    ADD COLUMN IF NOT EXISTS awb_code TEXT,
    ADD COLUMN IF NOT EXISTS courier_name TEXT,
    ADD COLUMN IF NOT EXISTS label_url TEXT,
    ADD COLUMN IF NOT EXISTS manifest_url TEXT,
    ADD COLUMN IF NOT EXISTS pickup_status TEXT,
    ADD COLUMN IF NOT EXISTS package_weight NUMERIC(8,2),
    ADD COLUMN IF NOT EXISTS package_length NUMERIC(8,2),
    ADD COLUMN IF NOT EXISTS package_breadth NUMERIC(8,2),
    ADD COLUMN IF NOT EXISTS package_height NUMERIC(8,2);

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS shiprocket_order_id TEXT,
    ADD COLUMN IF NOT EXISTS shipment_id TEXT,
    ADD COLUMN IF NOT EXISTS awb_code TEXT,
    ADD COLUMN IF NOT EXISTS courier_name TEXT,
    ADD COLUMN IF NOT EXISTS label_url TEXT,
    ADD COLUMN IF NOT EXISTS manifest_url TEXT,
    ADD COLUMN IF NOT EXISTS pickup_status TEXT,
    ADD COLUMN IF NOT EXISTS package_weight NUMERIC(8,2),
    ADD COLUMN IF NOT EXISTS package_length NUMERIC(8,2),
    ADD COLUMN IF NOT EXISTS package_breadth NUMERIC(8,2),
    ADD COLUMN IF NOT EXISTS package_height NUMERIC(8,2);

CREATE INDEX IF NOT EXISTS idx_sample_orders_awb ON sample_orders(awb_code);
CREATE INDEX IF NOT EXISTS idx_orders_awb ON orders(awb_code);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_orders_awb;
DROP INDEX IF EXISTS idx_sample_orders_awb;

ALTER TABLE orders
    DROP COLUMN IF EXISTS package_height,
    DROP COLUMN IF EXISTS package_breadth,
    DROP COLUMN IF EXISTS package_length,
    DROP COLUMN IF EXISTS package_weight,
    DROP COLUMN IF EXISTS pickup_status,
    DROP COLUMN IF EXISTS manifest_url,
    DROP COLUMN IF EXISTS label_url,
    DROP COLUMN IF EXISTS courier_name,
    DROP COLUMN IF EXISTS awb_code,
    DROP COLUMN IF EXISTS shipment_id,
    DROP COLUMN IF EXISTS shiprocket_order_id;

ALTER TABLE sample_orders
    DROP COLUMN IF EXISTS package_height,
    DROP COLUMN IF EXISTS package_breadth,
    DROP COLUMN IF EXISTS package_length,
    DROP COLUMN IF EXISTS package_weight,
    DROP COLUMN IF EXISTS pickup_status,
    DROP COLUMN IF EXISTS manifest_url,
    DROP COLUMN IF EXISTS label_url,
    DROP COLUMN IF EXISTS courier_name,
    DROP COLUMN IF EXISTS awb_code,
    DROP COLUMN IF EXISTS shipment_id,
    DROP COLUMN IF EXISTS shiprocket_order_id;

-- +goose StatementEnd
