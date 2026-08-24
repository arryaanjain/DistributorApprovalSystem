# Kresconet Distributor Credit Platform — Implementation Plan

## Current State

All three projects are freshly scaffolded:

| Project | Stack | Port | Status |
|---------|-------|------|--------|
| `api/` | Go 1.26, bare `main.go` | TBD | Empty |
| `ui/` | Next.js 16 + Tailwind v4 (PostCSS) | 3000 | Boilerplate |
| `admin/` | React 19 + Vite 8 + Tailwind v4 (Vite plugin) | 5173 | Boilerplate |

**Constraints from user:**
- APIs versioned under `/v1/`
- `ui/` — Tailwind class utilities only, no custom CSS files
- `admin/` — Tailwind class utilities only, no custom CSS files

---

## Technology Decisions

### API (`api/`)

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Router | `chi` | Lightweight, stdlib-compatible, middleware-friendly |
| DB | PostgreSQL 16 | Relational integrity for financial data |
| ORM/Query | `sqlc` | Type-safe generated Go from SQL, no runtime reflection |
| Migrations | `goose` | Simple, SQL-based, versioned |
| Queue | `river` (PG-backed) or Redis + `asynq` | Background jobs without extra infra initially |
| Auth | JWT (access) + secure cookie (refresh) | OTP-based for distributors, password for employees |
| Config | `viper` | Env + file based config |
| Validation | `go-playground/validator` | Struct tag validation |
| Logging | `slog` (stdlib) | Structured, zero-dep |
| Testing | stdlib `testing` + `testify` | Standard Go testing |

### Database

PostgreSQL with `goose` migrations. All credit policy parameters stored as configurable records (Phase 0 requirement).

### UI (`ui/`) — Distributor-Facing

Next.js 16 App Router. Tailwind v4 utility classes only. Server components where possible, client components for interactive forms.

### Admin (`admin/`) — Internal Dashboard

React 19 + Vite 8 + React Router. Tailwind v4 utility classes only. SPA with client-side routing.

---

## Directory Structure

### API

```
api/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── config/                  # App config, env loading
│   ├── database/                # DB connection, migrations
│   │   └── migrations/          # goose SQL migrations
│   ├── middleware/               # Auth, RBAC, rate-limit, audit
│   ├── models/                  # Domain structs
│   ├── handler/                 # HTTP handlers grouped by domain
│   │   ├── auth.go
│   │   ├── onboarding.go
│   │   ├── distributor.go
│   │   ├── verification.go
│   │   ├── credit.go
│   │   ├── risk.go
│   │   ├── offer.go
│   │   ├── agreement.go
│   │   ├── catalogue.go
│   │   ├── order.go
│   │   ├── payment.go
│   │   ├── invoice.go
│   │   ├── outstanding.go
│   │   ├── collections.go
│   │   ├── behaviour.go
│   │   ├── credit_enhancement.go
│   │   ├── credit_control.go
│   │   ├── admin.go
│   │   ├── audit.go
│   │   └── notification.go
│   ├── router/                  # chi router setup, v1 prefix
│   ├── service/                 # Business logic layer
│   │   ├── auth/
│   │   ├── onboarding/
│   │   ├── verification/        # PAN, GST, FSSAI, Bank abstractions
│   │   ├── credit/              # Scoring engine, decision engine
│   │   ├── risk/                # Hard-risk override engine
│   │   ├── order/
│   │   ├── payment/
│   │   ├── invoice/
│   │   ├── collections/
│   │   ├── behaviour/           # Behavioural credit engine
│   │   ├── enhancement/         # Auto credit enhancement
│   │   ├── agreement/
│   │   └── audit/
│   ├── repository/              # DB queries (sqlc generated)
│   ├── policy/                  # Credit policy config loader
│   ├── queue/                   # Background job definitions
│   └── pkg/                     # Shared utilities
│       ├── response/            # Standard JSON responses
│       ├── errors/              # App error types
│       ├── crypto/              # Hashing, encryption helpers
│       └── idempotency/         # Idempotency key logic
├── sql/
│   ├── queries/                 # sqlc query files
│   └── schema/                  # Reference schema
├── sqlc.yaml
├── go.mod
└── go.sum
```

### UI (Next.js)

```
ui/
├── app/
│   ├── layout.tsx
│   ├── page.tsx                 # Landing / lead capture
│   ├── onboarding/
│   │   ├── layout.tsx           # Multi-step form wrapper
│   │   ├── page.tsx             # Redirect to step 1
│   │   ├── mobile/page.tsx      # Mobile + OTP
│   │   ├── basic/page.tsx       # Name, email, etc.
│   │   ├── business/page.tsx    # Business details
│   │   ├── statutory/page.tsx   # PAN, GST, FSSAI
│   │   ├── bank/page.tsx        # Bank details
│   │   ├── preference/page.tsx  # Payment preference
│   │   └── consent/page.tsx     # Consent + submit
│   ├── offer/
│   │   └── [id]/page.tsx        # View credit offer
│   ├── agreement/
│   │   └── [id]/page.tsx        # View & sign agreement
│   ├── catalogue/
│   │   └── page.tsx             # Product catalogue
│   ├── order/
│   │   ├── page.tsx             # Create order
│   │   └── [id]/page.tsx        # Order detail + payment
│   └── dashboard/
│       └── page.tsx             # Distributor self-service
├── components/
│   ├── ui/                      # Reusable form elements
│   └── onboarding/              # Step-specific components
├── lib/
│   ├── api.ts                   # API client
│   └── validation.ts            # Client-side validation
└── ...
```

### Admin (React + Vite)

```
admin/
├── src/
│   ├── main.tsx
│   ├── App.tsx                  # Router setup
│   ├── pages/
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Applications.tsx     # Onboarding pipeline
│   │   ├── ApplicationDetail.tsx
│   │   ├── Distributors.tsx
│   │   ├── DistributorDetail.tsx
│   │   ├── CreditControl.tsx
│   │   ├── Orders.tsx
│   │   ├── OrderReview.tsx      # First human intervention
│   │   ├── Invoices.tsx
│   │   ├── Outstanding.tsx
│   │   ├── Collections.tsx
│   │   ├── CreditEnhancement.tsx
│   │   ├── PolicyConfig.tsx
│   │   ├── AuditLog.tsx
│   │   └── Reports.tsx
│   ├── components/
│   │   ├── layout/              # Sidebar, header, etc.
│   │   └── shared/              # Tables, cards, badges
│   ├── lib/
│   │   ├── api.ts               # Admin API client
│   │   └── auth.ts              # Admin auth helpers
│   └── index.css                # @import "tailwindcss" only
└── ...
```

---

## Development Milestones

### Milestone 1 — Foundation (Phases 0, 24, 25, 26, 27)

> Database, auth, config, API skeleton, background job infra

#### [NEW] `api/cmd/server/main.go`
Replace bare `main.go`. Initialize config, DB, router, queue worker, and HTTP server.

#### [NEW] `api/internal/config/config.go`
Load from env/file: DB DSN, JWT secret, OTP settings, API keys, server port.

#### [NEW] `api/internal/database/database.go`
PostgreSQL connection pool via `pgx`. Health check endpoint.

#### [NEW] `api/internal/database/migrations/` (goose SQL)

Core schema tables (abridged — full list in Phase 24):

**Auth & RBAC:** `users`, `roles`, `permissions`, `role_permissions`, `user_roles`

**Policy (Phase 0):** `credit_policies`, `credit_policy_versions` — stores all configurable parameters:
- Score bands & initial offers (85-100→₹50k, 75-84→₹35k, etc.)
- Credit ladder steps (₹15k→₹25k→…→₹3L)
- Credit periods, risk grades, overdue thresholds
- Non-GST caps, approval authorities, max exposure
- Enhancement rules, behavioural thresholds

**Distributor:** `distributors`, `business_profiles`, `business_documents`

**Applications:** `applications`, `application_events`

**Verification:** `pan_verifications`, `gst_verifications`, `fssai_verifications`, `bank_verifications`, `credit_reports`

**Credit:** `credit_scores`, `credit_score_components`, `risk_flags`, `credit_decisions`, `credit_offers`

**Agreements:** `distributor_agreements`, `agreement_versions`, `agreement_signatures`

**Financial:** `credit_accounts`, `credit_limit_history`, `credit_term_history`, `orders`, `order_items`, `invoices`, `invoice_payments`, `payment_allocations`, `credit_transactions`, `outstanding_ledger`

**Behavioural:** `behaviour_scores`, `behaviour_events`, `credit_reviews`, `credit_enhancements`, `credit_reductions`

**Operations:** `overdue_events`, `collections_actions`, `approval_requests`, `approval_actions`

**Audit:** `audit_logs`, `notifications`, `otp_verifications`, `consents`

> [!IMPORTANT]
> All history tables use append-only pattern (e.g., `credit_limit_history` stores transitions like `50000 → 100000` with reason, policy version, approver, timestamp).

#### [NEW] `api/internal/router/router.go`
chi router with versioned prefix:
```
/api/v1/auth/...
/api/v1/onboarding/...
/api/v1/distributors/...
/api/v1/verification/...
/api/v1/credit/...
/api/v1/risk/...
/api/v1/offers/...
/api/v1/agreements/...
/api/v1/catalogue/...
/api/v1/orders/...
/api/v1/payments/...
/api/v1/invoices/...
/api/v1/outstanding/...
/api/v1/collections/...
/api/v1/behaviour/...
/api/v1/credit-enhancement/...
/api/v1/credit-control/...
/api/v1/admin/...
/api/v1/audit/...
/api/v1/notifications/...
```

#### [NEW] `api/internal/middleware/`
- `auth.go` — JWT validation, session management
- `rbac.go` — Role-based access control
- `audit.go` — Audit trail middleware
- `ratelimit.go` — Rate limiting
- `idempotency.go` — Idempotency key enforcement for financial endpoints

#### [NEW] `api/internal/queue/`
Background job infrastructure. Jobs: `verify-pan`, `verify-gst`, `verify-bank`, `fetch-credit-report`, `calculate-credit-score`, `generate-agreement`, `send-notification`, `process-payment-webhook`, `calculate-overdue`, `calculate-behaviour-score`, `evaluate-credit-enhancement`, `evaluate-credit-reduction`.

#### [NEW] `api/internal/policy/loader.go`
Load credit policy from DB. Expose current active policy version. Never hard-code credit numbers.

---

### Milestone 2 — Onboarding & Verification (Phases 1, 2, 3, 4)

#### API Handlers

**`handler/auth.go`** — OTP send/verify for distributors, employee login.

**`handler/onboarding.go`** — Multi-step application:
1. Mobile + OTP
2. Basic details (name, email)
3. Business details (name, address, constitution, vintage, FMCG exp, monthly business, retailers, salespersons, brands)
4. Statutory details (PAN, GST, FSSAI, Udyam/Shop & Est.)
5. Bank details (account, IFSC, holder name)
6. Payment preference (8 options → classified as LOW/NO, SHORT, STANDARD, EXTENDED exposure)
7. Consent

**`handler/verification.go`** — Unified verification abstraction:
```
VerificationService
├── PAN     → validity, name match
├── GST     → status, legal name, trade name, reg date, address, constitution
├── FSSAI   → where applicable
├── Bank    → account, IFSC, holder name, ownership match
├── Business Registration
└── Credit Bureau
```
Each returns: `VERIFIED | PARTIALLY_VERIFIED | MISMATCH | FAILED | UNAVAILABLE | PENDING`

**Duplicate Detection** — Before credit assessment, check: mobile, PAN, GST, bank account. Flag suspicious/duplicate applications rather than blind rejection.

**Non-GST Route (Phase 4)** — No GST ≠ rejection. Accept alternative evidence: FSSAI, Udyam, Shop & Est., photos, address evidence. Non-GST cap: ₹25,000 max initial credit regardless of score.

#### UI Pages

**`ui/app/onboarding/`** — Multi-step form flow with progress indicator. Each step validates before proceeding. All Tailwind utility classes.

---

### Milestone 3 — Credit Engine (Phases 5, 6, 7, 8, 9, 10)

#### Credit/Risk Data Layer (Phase 5)
**`service/credit/external.go`** — Fetch external risk data (credit score, repayment history, defaults, write-offs, fraud indicators). Credit score alone never determines the decision.

#### Scoring Engine (Phase 6)
**`service/credit/scoring.go`** — Deterministic rules engine (NOT an LLM):

| Component | Weight |
|-----------|--------|
| Credit / repayment risk | 30 |
| Identity / KYC | 15 |
| Business verification | 15 |
| Business vintage | 10 |
| Distribution experience | 10 |
| Business capacity | 10 |
| Data consistency / fraud | 10 |
| **Total** | **100** |

Granular sub-factors: PAN/identity, bank verification, GST/alternative, vintage, credit risk, FMCG experience, retailer coverage, business scale, verification quality.

#### Hard-Risk Override (Phase 7)
**`service/risk/hard_flags.go`** — Executes BEFORE final score decision. Hard flag = credit blocked regardless of score. Flags: invalid PAN, identity mismatch, bank mismatch, serious default, fraud indicator, suspicious duplicate, previous Kresconet default, manipulated documents, adverse repayment.

#### Decision Engine (Phase 8)
**`service/credit/decision.go`** — Three SEPARATE decisions:

1. **Eligibility:** CREDIT / ADVANCE_ONLY / HOLD / BLOCKED
2. **Credit Limit:** ₹15k / ₹25k / ₹35k / ₹50k / ₹1L / ₹1.5L / ₹2L / ₹3L
3. **Credit Period:** COD / Receipt / 15 days / 30 days / Bill-to-Bill

Initial automated prequalification bands:
| Score | Initial Offer |
|-------|---------------|
| 85–100 | Up to ₹50,000 |
| 75–84 | Up to ₹35,000 |
| 65–74 | Up to ₹25,000 |
| 55–64 | Up to ₹15,000 |
| <55 | Advance only |

#### Credit Offer (Phase 9)
Present: risk grade, pre-approved limit, approved payment terms, max outstanding age. Distributor can accept credit OR purchase on advance.

#### Digital Agreement (Phase 10)
Generate from verified data. Include all legal terms (identity, PAN/GST, credit limit, payment terms, due dates, default provisions, etc.). Record: version, hash, distributor, timestamp, IP/device, acceptance, status. Credit activated only after SIGNED status.

> [!WARNING]
> The user's specification requires legal vetting before implementing the agreement template. We will build the generation infrastructure but the actual legal text must be reviewed by counsel before production use.

---

### Milestone 4 — Orders & Payments (Phases 11, 12, 13, 14)

#### Product Catalogue + Order Engine (Phase 11)
**Key rule:** Order value ≠ credit limit. If order (₹1,80,000) > credit (₹50,000), system calculates: Pay Now ₹1,30,000 + Credit ₹50,000. Order is NOT rejected for exceeding credit.

#### Payment Collection (Phase 12)
Before credit exposure: Required Advance → Payment Gateway → Webhook → Server verification → Payment ledger. Never trust frontend payment-success alone.

#### First Human Review (Phase 13)
**Architectural feature, not a bug.** Employee receives pre-computed dashboard: distributor details, all verifications, score, risk grade, preference, pre-approved limit/period, agreement, order value, advance received, proposed exposure, risk flags. Actions: `APPROVE_AND_RELEASE | HOLD | REVISE_TERMS`. Every action = authenticated employee + audit event. This is exception review, not reassessment.

#### Dispatch Credit Guard (Phase 14)
Backend invariant: `Available Credit = Approved Limit − Current Outstanding`. `Existing Outstanding + New Credit ≤ Approved Limit`. System ENFORCES this. Even admin cannot bypass without formal authorized override workflow.

---

### Milestone 5 — Post-Dispatch Operations (Phases 15, 16, 17, 18, 19, 20)

#### Invoice & Outstanding Ledger (Phase 15)
Proper financial ledger: invoice, date, credit/advance amounts, due date, payments, allocation, outstanding, days outstanding, utilisation, status. Immutable transaction records.

#### Automated Collections (Phase 16)
Escalation tiers: Before Due → Due Date → 1-3 days → 4-7 → 8-15 → Serious Overdue → Payment Failure. Automate: reminders, restrictions, credit holds, escalation, notifications.

#### Behavioural Credit Engine (Phase 17)
After first transaction, calculate from Kresconet's own data: on-time %, days late, successful cycles, utilisation, average order, frequency, max outstanding, payment failures, claims, relationship duration. Formula: `Initial Risk Score + Kresconet Behaviour Score = Current Creditworthiness`. Over time, internal score outweighs external.

#### Auto Credit Enhancement (Phase 18)
Ladder: ₹30k→₹50k→₹1L→₹1.5L→₹2L→₹3L. Based on payment discipline + utilisation + purchase behaviour + risk status. System RECOMMENDS (with reasoning), then routes to approval authority or auto-approves per policy.

#### Credit Reduction/Suspension (Phase 19)
Triggers: late payments, default, bounce, adverse info, fraud, business deterioration, excessive claims, agreement breach. Statuses: ACTIVE / RESTRICTED / HOLD / BLOCKED.

#### Bill-to-Bill Engine (Phase 20)
Special mode with BOTH max credit limit AND max outstanding age. Breach either = stop further credit dispatch.

---

### Milestone 6 — Admin Dashboard (Phases 21, 22, 23)

#### Manual Credit Control (Phase 21)
Admin dashboard pages: Applications pipeline (New→Pending→Scoring→Approved→Rejected→Hold), Distributor management (Active→Restricted→Hold→Blocked), Credit overview (limit, utilisation, outstanding, overdue, available, risk grade, behaviour score), Orders requiring intervention.

#### Approval Authority Engine (Phase 22)
RBAC workflow: Sales → Credit/Accounts → Manager/Director → Back Office → Dispatch. Sales cannot independently grant credit. No verbal approvals. Implemented as RBAC + approval workflows.

#### Audit Layer (Phase 23)
Every decision produces immutable audit record. Events: APPLICATION_CREATED through CREDIT_SUSPENDED. For every credit decision: score, risk grade, rules evaluated, inputs used, external results, hard flags, approved limit/period, policy version, timestamp, source, approver.

---

### Milestone 7 — Production Readiness (Phases 28, 29, 30, 31, 32)

#### Idempotency & Financial Integrity (Phase 28)
Idempotency keys for payment/order operations. Credit exposure: calculate → lock → recheck → create → commit (prevent double-consumption of available credit).

#### Testing (Phase 29)
- Scoring: test all bands (100, 85, 75, 65, 55, 54)
- Hard flags: high score + fraud = block
- Non-GST: high score + no GST = ₹25k cap
- Orders: credit math (limit - outstanding = available, pay-now calculation)
- Concurrency: two simultaneous orders against same credit
- Overdue: escalation tiers
- Enhancement: successful cycles → upgrade, late → no upgrade, bounce → hold
- Reduction: adverse event → restriction/hold/block

#### Shadow Mode / Pilot (Phase 30)
System calculates decisions but humans make real decisions. Compare system vs human. Tune weights, thresholds, flags.

#### Gradual Automation (Phase 31)
7 stages from "automated onboarding only" to "human intervention only for exceptions". Target: 90%+ clean applications need no manual work until actual order.

#### Production Monitoring (Phase 32)
KPIs: funnel (leads→orders), credit (exposure, utilisation), risk (default, overdue, bounce, fraud), automation (% automated, review time), behaviour (payment days, cycles, enhancements).

---

## Open Questions

> [!IMPORTANT]
> **Database hosting:** Are you using a managed PostgreSQL service (e.g., Supabase, Neon, RDS) or self-hosted? This affects connection pooling and migration strategy.

> [!IMPORTANT]
> **External verification providers:** Which specific APIs will be used for PAN, GST, FSSAI, Bank, and Credit Bureau verification? (e.g., Surepass, Signzy, Gridlines, CIBIL). We'll build the abstraction layer first with mock implementations, but need to know eventual providers.

> [!IMPORTANT]
> **Payment gateway:** Which gateway for advance payments? (Razorpay, Cashfree, PayU). Affects webhook integration in Phase 12.

> [!IMPORTANT]
> **Notification channels:** The spec mentions "configured communication channels" for collections. Which channels? (SMS, WhatsApp, Email, Push). Which providers?

> [!IMPORTANT]
> **Go module path:** Current `go.mod` has `github.com/arryaanjain/DisributorApprovalSystem.git` — should this be corrected to `DistributorApprovalSystem` (typo in "Disributor")?

> [!IMPORTANT]
> **Deployment target:** Where will the API be deployed? (VPS, Cloud Run, ECS, K8s). Affects Dockerfile and CI/CD setup.

> [!IMPORTANT]
> **Implementation priority:** Given the scale (32 phases), should we start with Milestone 1 (Foundation + DB + Auth) and iterate, or do you want to see more milestones planned in detail first?

---

## Verification Plan

### Automated Tests
- `go test ./...` for all service, handler, and repository layers
- Scoring engine unit tests covering all bands and edge cases
- Hard-flag override tests
- Credit math invariant tests (outstanding + new ≤ limit)
- Concurrency tests for credit consumption
- Integration tests with test database for full flows

### Manual Verification
- Run all three services locally (`api`, `ui`, `admin`)
- Walk through complete onboarding flow in UI
- Verify credit decision pipeline end-to-end
- Test admin dashboard review workflow
- Validate audit trail completeness
