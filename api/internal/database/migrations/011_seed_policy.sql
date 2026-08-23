-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- SEED: Initial Credit Policy v1.0
-- All monetary values stored in paise (1 INR = 100 paise)
-- ============================================================

-- Insert the first credit policy
INSERT INTO credit_policies (id, version, name, description, is_active, activated_at)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    '1.0',
    'Kresconet Credit Policy v1.0',
    'Initial credit policy. Score bands follow the automated prequalification model. Ladder runs to ₹3,00,000.',
    TRUE,
    NOW()
);

-- Score bands → initial credit offer
-- Score 85-100: up to ₹50,000
INSERT INTO policy_score_bands (policy_id, min_score, max_score, eligibility, max_credit_paise, display_label)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 85, 100, 'CREDIT',       5000000, '₹50,000'),
    ('a0000000-0000-0000-0000-000000000001', 75,  84, 'CREDIT',       3500000, '₹35,000'),
    ('a0000000-0000-0000-0000-000000000001', 65,  74, 'CREDIT',       2500000, '₹25,000'),
    ('a0000000-0000-0000-0000-000000000001', 55,  64, 'CREDIT',       1500000, '₹15,000'),
    ('a0000000-0000-0000-0000-000000000001',  0,  54, 'ADVANCE_ONLY',       0, 'Advance Only');

-- Credit ladder (the long-term enhancement path)
INSERT INTO policy_credit_ladder
    (policy_id, step_order, limit_paise, display_label, min_cycles, min_ontime_pct, min_utilisation_pct, auto_approve, approval_role)
VALUES
    ('a0000000-0000-0000-0000-000000000001', 1,   1500000, '₹15,000',   0, 0,    0,    FALSE, 'credit_manager'),
    ('a0000000-0000-0000-0000-000000000001', 2,   2500000, '₹25,000',   0, 0,    0,    FALSE, 'credit_manager'),
    ('a0000000-0000-0000-0000-000000000001', 3,   3500000, '₹35,000',   0, 0,    0,    FALSE, 'credit_manager'),
    ('a0000000-0000-0000-0000-000000000001', 4,   5000000, '₹50,000',   3, 90.0, 60.0, FALSE, 'credit_manager'),
    ('a0000000-0000-0000-0000-000000000001', 5,  10000000, '₹1,00,000', 6, 92.0, 65.0, FALSE, 'credit_manager'),
    ('a0000000-0000-0000-0000-000000000001', 6,  15000000, '₹1,50,000', 9, 94.0, 70.0, FALSE, 'accounts'),
    ('a0000000-0000-0000-0000-000000000001', 7,  20000000, '₹2,00,000',12, 95.0, 70.0, FALSE, 'accounts'),
    ('a0000000-0000-0000-0000-000000000001', 8,  30000000, '₹3,00,000',18, 97.0, 75.0, FALSE, 'credit_manager');

-- Credit periods
INSERT INTO policy_credit_periods (policy_id, code, label, days) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'COD',        'Cash on Delivery',  NULL),
    ('a0000000-0000-0000-0000-000000000001', 'RECEIPT',    'On Receipt',        NULL),
    ('a0000000-0000-0000-0000-000000000001', '15D',        '15 Days',           15),
    ('a0000000-0000-0000-0000-000000000001', '30D',        '30 Days',           30),
    ('a0000000-0000-0000-0000-000000000001', 'BTB',        'Bill-to-Bill',      NULL);

-- Risk grades
INSERT INTO policy_risk_grades (policy_id, grade, min_score, max_score, label, description, max_limit_paise) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'A',  90, 100, 'Excellent', 'Very low risk',    5000000),
    ('a0000000-0000-0000-0000-000000000001', 'B',  75,  89, 'Good',      'Low risk',         3500000),
    ('a0000000-0000-0000-0000-000000000001', 'C',  60,  74, 'Fair',      'Moderate risk',    2500000),
    ('a0000000-0000-0000-0000-000000000001', 'D',  45,  59, 'Poor',      'High risk',        1500000),
    ('a0000000-0000-0000-0000-000000000001', 'E',   0,  44, 'Very Poor', 'Very high risk',          0);

-- Overdue thresholds
INSERT INTO policy_overdue_thresholds
    (policy_id, tier, from_days, to_days, label, action_codes, auto_restrict, auto_hold) VALUES
    ('a0000000-0000-0000-0000-000000000001', 1,  1,  3, '1–3 Days',       ARRAY['SEND_REMINDER'],                                    FALSE, FALSE),
    ('a0000000-0000-0000-0000-000000000001', 2,  4,  7, '4–7 Days',       ARRAY['SEND_REMINDER','NOTIFY_MANAGER'],                   FALSE, FALSE),
    ('a0000000-0000-0000-0000-000000000001', 3,  8, 15, '8–15 Days',      ARRAY['SEND_REMINDER','RESTRICT_CREDIT','NOTIFY_MANAGER'], TRUE,  FALSE),
    ('a0000000-0000-0000-0000-000000000001', 4, 16, NULL,'Serious Overdue',ARRAY['CREDIT_HOLD','ESCALATE','NOTIFY_DIRECTOR'],        TRUE,  TRUE);

-- Non-GST rules (Phase 4): cap at ₹25,000 initial
INSERT INTO policy_non_gst_rules (policy_id, max_initial_limit_paise, requires_alt_evidence, acceptable_evidence) VALUES
    ('a0000000-0000-0000-0000-000000000001', 2500000, TRUE,
     ARRAY['FSSAI','UDYAM','SHOP_EST','SHOP_PHOTO','SIGNBOARD','ADDRESS_PROOF']);

-- Approval authorities
INSERT INTO policy_approval_authorities (policy_id, from_limit_paise, to_limit_paise, required_role, label) VALUES
    ('a0000000-0000-0000-0000-000000000001',       0,   5000000, 'credit_manager', 'Up to ₹50,000 — Credit Manager'),
    ('a0000000-0000-0000-0000-000000000001', 5000001,  15000000, 'accounts',       '₹50,001–₹1,50,000 — Accounts'),
    ('a0000000-0000-0000-0000-000000000001',15000001,  NULL,     'credit_manager', 'Above ₹1,50,000 — Director/Authorised Manager');

-- Credit enhancement rules (Phase 18)
INSERT INTO policy_enhancement_rules
    (policy_id, from_limit_paise, to_limit_paise, required_cycles, required_ontime_pct, required_util_pct, no_current_flags, auto_approve, approval_role, label)
VALUES
    ('a0000000-0000-0000-0000-000000000001',  3500000,  5000000, 3,  90.0, 60.0, TRUE, FALSE, 'credit_manager', '₹35k → ₹50k'),
    ('a0000000-0000-0000-0000-000000000001',  5000000, 10000000, 6,  92.0, 65.0, TRUE, FALSE, 'credit_manager', '₹50k → ₹1L'),
    ('a0000000-0000-0000-0000-000000000001', 10000000, 15000000, 9,  94.0, 70.0, TRUE, FALSE, 'accounts',       '₹1L → ₹1.5L'),
    ('a0000000-0000-0000-0000-000000000001', 15000000, 20000000,12,  95.0, 70.0, TRUE, FALSE, 'accounts',       '₹1.5L → ₹2L'),
    ('a0000000-0000-0000-0000-000000000001', 20000000, 30000000,18,  97.0, 75.0, TRUE, FALSE, 'credit_manager', '₹2L → ₹3L');

-- Seed: initial agreement template
INSERT INTO agreement_versions (id, version, template, is_active)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    '1.0',
    E'KRESCONET DISTRIBUTOR CREDIT AGREEMENT\n\nVersion: 1.0\n\nThis agreement is entered into between:\n1. {{distributor_name}} (PAN: {{pan}}, GST: {{gst_number}}) — Distributor\n2. Kresconet — Company\n\nCREDIT TERMS\nApproved Credit Limit: {{credit_limit}}\nPayment Terms: {{credit_period}}\nMaximum Outstanding Age: {{max_outstanding_days}} days\n\n[LEGAL TEXT — TO BE REVIEWED BY COUNSEL BEFORE PRODUCTION USE]\n\nBy signing, the Distributor agrees to all terms and conditions.\n\nDistributor Signature: _______________  Date: {{date}}',
    TRUE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM agreement_versions WHERE id = 'b0000000-0000-0000-0000-000000000001';
DELETE FROM policy_enhancement_rules WHERE policy_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM policy_approval_authorities WHERE policy_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM policy_non_gst_rules WHERE policy_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM policy_overdue_thresholds WHERE policy_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM policy_risk_grades WHERE policy_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM policy_credit_periods WHERE policy_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM policy_credit_ladder WHERE policy_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM policy_score_bands WHERE policy_id = 'a0000000-0000-0000-0000-000000000001';
DELETE FROM credit_policies WHERE id = 'a0000000-0000-0000-0000-000000000001';
-- +goose StatementEnd
