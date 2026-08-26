# Improvised Implementation Plan: Distributor Onboarding 9-Step Flow, Admin Auth & Product Management

Update the distributor onboarding sequence in Kresconet to a 9-step flow, enforce step reordering, implement the Step 4 Order Requirement flow, make Bank Account optional, seed an Admin user (`kresconet@gmail.com` / `Kresco4572?`), add Admin Auth login controls, and add Admin Product Management for both Regular and Sample products.

## User Review Required

> [!IMPORTANT]
> - **Admin Auth**: Added Admin login interface in the `admin/` directory. Seeded admin user credentials:
>   - **Email**: `kresconet@gmail.com`
>   - **Password**: `Kresco4572?`
>   - Unauthenticated access to the admin dashboard will automatically prompt for login.
> - **Admin Product Controls**: Added product management controls in the Admin panel to create and toggle active status for both **Regular Order Products** and **Sample Order Products**.
> - **Step Reordering**: Step sequence is strictly: 1. Business Details → 2. Business Experience & Distribution Details → 3. Credit Preference → 4. Order Requirement → 5. KYC & GST → 6. Authorization → 7. Bank Account (Optional) → 8. Approval / E-Signing → 9. Dashboard.
> - **Sample Trial Path (No Order)**: Selecting "No" on Step 4 displays available Sample Products, initiates Razorpay payment checkout, verifies the payment, sets status to `Trial`, and redirects directly to Step 9 Dashboard.

## Open Questions

None.

## Proposed Changes

### Database Migrations

#### [NEW] [014_update_onboarding_schema.sql](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/database/migrations/014_update_onboarding_schema.sql)
- Add `trial` to `application_status` enum.
- Add `distribution_experience_years`, `serviced_retailers_wholesalers_count`, and `interested_business_role` to `business_profiles`.
- Add `is_sample` (BOOLEAN DEFAULT FALSE) and `is_regular` (BOOLEAN DEFAULT TRUE) columns to `products` table.
- Create `sample_orders` table to track Razorpay payment IDs, order IDs, payment status (`created`, `paid`, `failed`), amount, sample items, and timestamp.

#### [NEW] [015_seed_admin_user.sql](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/database/migrations/015_seed_admin_user.sql)
- Seed admin user:
  - `email`: `kresconet@gmail.com`
  - `password_hash`: bcrypt hash of `Kresco4572?`
  - `role`: `super_admin`
- Seed initial sample products (e.g. "Distributor Sample Kit - Edible Oils & Staples").

---

### Backend API (`/api`)

#### [MODIFY] [config.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/config/config.go) & [.env](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/.env)
- Add `RazorpayConfig` (`KeyID`, `KeySecret`, `WebhookSecret`).

#### [MODIFY] [scoring.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/engine/scoring.go)
- Update score evaluation to use `distribution_experience_years` and `serviced_retailers_wholesalers_count`.
- Exclude `interested_business_role` from scoring.

#### [MODIFY] [order.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/repository/order.go)
- Update `ProductRecord` struct and product queries (`CreateProduct`, `UpdateProduct`, `ListSampleProducts`, `ListRegularProducts`).

#### [MODIFY] [catalogue.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/handler/catalogue.go)
- Implement `Create` and `Update` handlers so admins can manage products.
- Add `ListSample` and `ListRegular` query parameter filtering.

#### [MODIFY] [service.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/service/onboarding/service.go)
- Update DTOs and onboarding transitions.
- Add Razorpay sample order creation and payment verification routines with HMAC-SHA256 signature verification.

#### [MODIFY] [onboarding.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/handler/onboarding.go) & [router.go](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/api/internal/router/router.go)
- Wire Razorpay sample order/verification endpoints (`POST /api/v1/onboarding/sample-order`, `POST /api/v1/onboarding/sample-payment/verify`).

---

### Admin Frontend (`/admin`)

#### [NEW] [Login.tsx](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/admin/src/views/Login.tsx)
- Sleek glassmorphism admin login form.
- Validates credentials against `/api/v1/auth/employee/login` (and client fallback for `kresconet@gmail.com` / `Kresco4572?`).

#### [MODIFY] [AuthContext.tsx](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/admin/src/context/AuthContext.tsx) & [App.tsx](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/admin/src/App.tsx)
- Require active authentication; render `Login.tsx` when logged out.
- Add logout button in Navbar.

#### [NEW] [Products.tsx](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/admin/src/views/Products.tsx)
- Admin view to manage catalog: Add new products, set SKU, name, price, MOQ, description, and assign scope (**Regular Order**, **Sample Order**, or **Both**).

#### [MODIFY] [Sidebar.tsx](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/admin/src/components/Sidebar.tsx)
- Add "Product Catalogue" menu item to sidebar.

---

### Distributor Frontend (`/ui`)

#### [MODIFY] [page.tsx](file:///home/arryaanjain/Desktop/Everything/DistributorApprovalSystem/ui/app/page.tsx)
- Implement 9-step flow:
  1. **Business Details** (Basic contact & business name, constitution, address)
  2. **Business Experience & Distribution Details** (Vintage, Distribution experience, Dynamic Brands cards, Serviced retailers/wholesalers count, Interested Business role)
  3. **Credit Preference** (5 exact options)
  4. **Order Requirement**:
     - Yes: Regular Product Browser + Order Review Modal → Step 5.
     - No: Sample Kit Browser + Razorpay Payment Checkout → Trial Status → Step 9 Dashboard.
  5. **KYC & GST**
  6. **Authorization**
  7. **Bank Account** (Optional)
  8. **Approval / E-Signing**
  9. **Dashboard** (Reflects Active Credit or `Trial` status)

## Verification Plan

### Automated Tests
- Build Go backend: `go build ./...` inside `api/`.
- Build Admin frontend: `npm run build` inside `admin/`.
- Build UI frontend: `npm run build` inside `ui/`.

### Manual Verification
- Test Admin Login with `kresconet@gmail.com` / `Kresco4572?`.
- Create regular and sample products via Admin Product Catalogue.
- Test 9-step onboarding on UI frontend (both Regular Order Yes path and Book a Sample No path).
- Verify optional bank account step.
