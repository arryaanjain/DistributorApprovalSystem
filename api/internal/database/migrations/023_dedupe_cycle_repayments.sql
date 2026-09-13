-- +goose Up
-- +goose StatementBegin

-- 1. Remove any existing duplicate repayment records for the same cycle and instalment number
DELETE FROM cycle_repayments a USING cycle_repayments b
WHERE a.cycle_id = b.cycle_id
  AND a.instalment_number = b.instalment_number
  AND a.id > b.id;

-- 2. Add UNIQUE constraint on cycle_id and instalment_number
ALTER TABLE cycle_repayments ADD CONSTRAINT unique_cycle_instalment UNIQUE (cycle_id, instalment_number);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cycle_repayments DROP CONSTRAINT IF EXISTS unique_cycle_instalment;
-- +goose StatementEnd
