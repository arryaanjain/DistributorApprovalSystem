# Implementation Plan: Fix Admin Auto-Logout & Token Rotation Issue

## Overview
The Admin portal (`admin/src/App.tsx`) currently experiences automatic logouts when the access token approaches expiry or when the page is reloaded. This is caused by a failure in the token rotation flow where:
1. `admin/src/services/api.ts` explicitly hardcodes `getRefreshToken()` to return `null` and wipes `kresconet_admin_refresh_token` from `localStorage` on `setAuthToken()`.
2. `refreshAdminSession()` sends a `POST` request to `/auth/employee/refresh` with an empty JSON body `{}`.
3. Modern web browsers strip `SameSite=Lax` HttpOnly cookies (`kresconet_admin_refresh_token`) on cross-origin `POST` fetch calls (e.g. `localhost:5173` -> `localhost:8081`).
4. Backend `EmployeeRefresh` handler receives neither a cookie nor a body parameter, returning a `401 Unauthorized` ("missing admin refresh token").
5. `AuthContext.tsx` catches the 401 during its 60-second token expiry check or API request interceptor and clears user state, forcing an instant auto-logout.

## Proposed Changes

### 1. Admin Frontend Service (`admin/src/services/api.ts`)
- Update `getRefreshToken()` to retrieve `kresconet_admin_refresh_token` from `localStorage`.
- Update `setAuthToken(token: string, refreshToken?: string)` to store `refreshToken` in `localStorage` when provided.
- Update `clearAuthToken()` to remove both `kresconet_admin_token` and `kresconet_admin_refresh_token`.
- Update `refreshAdminSession()` to include `{ refresh_token: getRefreshToken() }` in the POST request body as a resilient fallback when HttpOnly cookies are stripped in cross-origin environments.
- Update `refreshAdminSession()` to extract and save rotated `data.refresh_token` via `setAuthToken(newAccess, newRefresh)`.

### 2. Admin Auth Context & Login (`admin/src/context/AuthContext.tsx` & `admin/src/views/Login.tsx`)
- Update `login` in `AuthContext.tsx` to accept `refreshToken?: string` and pass it to `setAuthToken(token, refreshToken)`.
- Update `Login.tsx` to pass `data.refresh_token` returned by `/auth/employee/login` into the `login()` function.

### 3. Distributor Frontend API Client (`ui/lib/api.ts`)
- Update `fetchApi` in `ui/lib/api.ts` to include `body: JSON.stringify({ refresh_token: localStorage.getItem("kresconet_refresh_token") })` when calling `/auth/refresh` so distributor portal refresh is also resilient to cross-origin cookie restrictions.

### 4. Backend Auth Handler (`api/internal/handler/auth.go`)
- Review `EmployeeRefresh` and `RefreshToken` handlers to ensure `r.Body` parsing reliably extracts `refresh_token` from JSON payload whenever cookie is omitted or restricted by browser policies.

---

## Verification Plan

### Automated Tests
- Run `go test ./...` in `api/` to verify all backend auth tests pass without regressions.
- Run `npm run build` in `admin/` to verify TypeScript compilation and build output.

### Manual Verification
- Log in to the Admin Console at `http://localhost:5173`.
- Verify `kresconet_admin_token` and `kresconet_admin_refresh_token` are saved in `localStorage` and HttpOnly cookie is set.
- Trigger manual token refresh / wait for 60-second `checkTokenExpiry` interval to fire.
- Verify session remains authenticated and new access/refresh tokens are successfully rotated without auto-logout.
