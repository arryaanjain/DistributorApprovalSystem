-- +goose Up
-- +goose StatementBegin

-- Add ON DELETE CASCADE to applications & application_events
ALTER TABLE applications DROP CONSTRAINT IF EXISTS applications_distributor_id_fkey;
ALTER TABLE applications ADD CONSTRAINT applications_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;

ALTER TABLE application_events DROP CONSTRAINT IF EXISTS application_events_application_id_fkey;
ALTER TABLE application_events ADD CONSTRAINT application_events_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to verification tables
ALTER TABLE pan_verifications DROP CONSTRAINT IF EXISTS pan_verifications_distributor_id_fkey;
ALTER TABLE pan_verifications ADD CONSTRAINT pan_verifications_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE pan_verifications DROP CONSTRAINT IF EXISTS pan_verifications_application_id_fkey;
ALTER TABLE pan_verifications ADD CONSTRAINT pan_verifications_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE gst_verifications DROP CONSTRAINT IF EXISTS gst_verifications_distributor_id_fkey;
ALTER TABLE gst_verifications ADD CONSTRAINT gst_verifications_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE gst_verifications DROP CONSTRAINT IF EXISTS gst_verifications_application_id_fkey;
ALTER TABLE gst_verifications ADD CONSTRAINT gst_verifications_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE bank_verifications DROP CONSTRAINT IF EXISTS bank_verifications_distributor_id_fkey;
ALTER TABLE bank_verifications ADD CONSTRAINT bank_verifications_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE bank_verifications DROP CONSTRAINT IF EXISTS bank_verifications_application_id_fkey;
ALTER TABLE bank_verifications ADD CONSTRAINT bank_verifications_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE credit_reports DROP CONSTRAINT IF EXISTS credit_reports_distributor_id_fkey;
ALTER TABLE credit_reports ADD CONSTRAINT credit_reports_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE credit_reports DROP CONSTRAINT IF EXISTS credit_reports_application_id_fkey;
ALTER TABLE credit_reports ADD CONSTRAINT credit_reports_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to credit engine tables
ALTER TABLE credit_scores DROP CONSTRAINT IF EXISTS credit_scores_distributor_id_fkey;
ALTER TABLE credit_scores ADD CONSTRAINT credit_scores_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE credit_scores DROP CONSTRAINT IF EXISTS credit_scores_application_id_fkey;
ALTER TABLE credit_scores ADD CONSTRAINT credit_scores_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE risk_flags DROP CONSTRAINT IF EXISTS risk_flags_distributor_id_fkey;
ALTER TABLE risk_flags ADD CONSTRAINT risk_flags_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE risk_flags DROP CONSTRAINT IF EXISTS risk_flags_application_id_fkey;
ALTER TABLE risk_flags ADD CONSTRAINT risk_flags_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE credit_decisions DROP CONSTRAINT IF EXISTS credit_decisions_distributor_id_fkey;
ALTER TABLE credit_decisions ADD CONSTRAINT credit_decisions_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE credit_decisions DROP CONSTRAINT IF EXISTS credit_decisions_application_id_fkey;
ALTER TABLE credit_decisions ADD CONSTRAINT credit_decisions_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE credit_offers DROP CONSTRAINT IF EXISTS credit_offers_distributor_id_fkey;
ALTER TABLE credit_offers ADD CONSTRAINT credit_offers_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE credit_offers DROP CONSTRAINT IF EXISTS credit_offers_application_id_fkey;
ALTER TABLE credit_offers ADD CONSTRAINT credit_offers_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to agreement & account tables
ALTER TABLE distributor_agreements DROP CONSTRAINT IF EXISTS distributor_agreements_distributor_id_fkey;
ALTER TABLE distributor_agreements ADD CONSTRAINT distributor_agreements_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;
ALTER TABLE distributor_agreements DROP CONSTRAINT IF EXISTS distributor_agreements_application_id_fkey;
ALTER TABLE distributor_agreements ADD CONSTRAINT distributor_agreements_application_id_fkey FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE;

ALTER TABLE credit_accounts DROP CONSTRAINT IF EXISTS credit_accounts_distributor_id_fkey;
ALTER TABLE credit_accounts ADD CONSTRAINT credit_accounts_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to order tables
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_distributor_id_fkey;
ALTER TABLE orders ADD CONSTRAINT orders_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;

ALTER TABLE payment_proofs DROP CONSTRAINT IF EXISTS payment_proofs_distributor_id_fkey;
ALTER TABLE payment_proofs ADD CONSTRAINT payment_proofs_distributor_id_fkey FOREIGN KEY (distributor_id) REFERENCES distributors(id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No down migration needed for cascade constraints
-- +goose StatementEnd
