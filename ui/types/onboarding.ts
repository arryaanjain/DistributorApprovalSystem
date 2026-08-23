export type OnboardingStep =
  | "auth"                // 0. Login / OTP
  | "step1_business_det"  // 1. Business Details
  | "step2_business_exp"  // 2. Business Experience & Distribution Details
  | "step3_credit_pref"   // 3. Credit Preference
  | "step4_order_req"     // 4. Order Requirement (Branch: Yes Order / No Sample)
  | "step5_kyc_gst"       // 5. KYC & GST
  | "step6_auth"          // 6. Authorization
  | "step7_bank"          // 7. Bank Account (Optional)
  | "step8_approval"      // 8. Approval / E-Signing
  | "step9_dashboard";    // 9. Dashboard

export interface ProductItem {
  id: string;
  sku: string;
  name: string;
  description?: string;
  category: string;
  price_paise: number;
  moq: number;
  is_sample: boolean;
  is_regular: boolean;
}

export interface Step1Data {
  name: string;
  email: string;
  business_name: string;
  constitution: string;
  address_line1: string;
  address_line2: string;
  city: string;
  state: string;
  pin: string;
}

export interface Step2Data {
  distribution_experience_years: number;
  serviced_retailers_wholesalers_count: number;
  interested_business_role: string;
  vintage_years: number;
  approx_monthly_business_inr: number;
  existing_brands: string;
}

export interface Step3Data {
  preference: string;
}

export interface Step5Data {
  pan: string;
  has_gst: boolean;
  gst_number: string;
  fssai_number: string;
  udyam_number: string;
  shop_est_number: string;
}

export interface Step6Data {
  authorized: boolean;
  signature_name: string;
}

export interface Step7Data {
  account_number: string;
  ifsc: string;
  account_holder: string;
  bank_name: string;
  branch: string;
}

export interface AppStatus {
  status?: string;
  distributor_id?: string;
  business_name?: string;
  constitution?: string;
  gst_number?: string;
  pan_number?: string;
  assigned_credit_limit?: number;
  approved_at?: string;
  kyc_status?: string;
  credit_status?: string;
  surepass_verified?: boolean;
}

export const STEP_LABELS: Record<string, { title: string; num: number }> = {
  step1_business_det: { title: "Business Details", num: 1 },
  step2_business_exp: { title: "Experience & Distribution", num: 2 },
  step3_credit_pref: { title: "Credit Preference", num: 3 },
  step4_order_req: { title: "Order Requirement", num: 4 },
  step5_kyc_gst: { title: "KYC & GST", num: 5 },
  step6_auth: { title: "Authorization", num: 6 },
  step7_bank: { title: "Bank Account", num: 7 },
  step8_approval: { title: "Approval / E-Sign", num: 8 },
  step9_dashboard: { title: "Dashboard", num: 9 },
};
