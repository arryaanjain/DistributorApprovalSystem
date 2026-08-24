"use client";

import { useState, useEffect } from "react";
import { fetchApi } from "@/lib/api";
import { AlertCircle, CheckCircle2 } from "lucide-react";
import {
  OnboardingStep,
  ProductItem,
  Step1Data,
  Step2Data,
  Step3Data,
  Step5Data,
  Step6Data,
  Step7Data,
  AppStatus,
} from "@/types/onboarding";

import { OnboardingHeader } from "@/components/onboarding/OnboardingHeader";
import { OnboardingStepNav } from "@/components/onboarding/OnboardingStepNav";
import { AuthStep } from "@/components/onboarding/AuthStep";
import { Step1BusinessDetails } from "@/components/onboarding/Step1BusinessDetails";
import { Step2BusinessExperience } from "@/components/onboarding/Step2BusinessExperience";
import { Step3CreditPreference } from "@/components/onboarding/Step3CreditPreference";
import { Step4OrderRequirement } from "@/components/onboarding/Step4OrderRequirement";
import { Step5KycGst } from "@/components/onboarding/Step5KycGst";
import { Step6Authorization } from "@/components/onboarding/Step6Authorization";
import { Step7BankAccount } from "@/components/onboarding/Step7BankAccount";
import { Step8Approval } from "@/components/onboarding/Step8Approval";
import { Step9Dashboard } from "@/components/onboarding/Step9Dashboard";
import { OrderReviewModal } from "@/components/onboarding/OrderReviewModal";
import { SamplePaymentModal } from "@/components/onboarding/SamplePaymentModal";

export default function DistributorPortal() {
  const [step, setStep] = useState<OnboardingStep>("auth");
  const [token, setToken] = useState<string | null>(null);
  const [, setDistributorId] = useState<string | null>(null);

  // Auth State
  const [mobile, setMobile] = useState("");
  const [otp, setOtp] = useState("");
  const [devOtp, setDevOtp] = useState<string | null>(null);
  const [otpSent, setOtpSent] = useState(false);

  // General State
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Form States
  const [step1, setStep1] = useState<Step1Data>({
    name: "",
    email: "",
    business_name: "",
    constitution: "proprietorship",
    address_line1: "",
    address_line2: "",
    city: "",
    state: "",
    pin: "",
  });

  const [step2, setStep2] = useState<Step2Data>({
    distribution_experience_years: 3,
    serviced_retailers_wholesalers_count: 75,
    interested_business_role: "Distributor",
    vintage_years: 5,
    approx_monthly_business_inr: 500000,
    existing_brands: "Amul, Fortune, Parle",
  });

  const [step3, setStep3] = useState<Step3Data>({
    preference: "15_days",
  });

  const [orderChoice, setOrderChoice] = useState<"none" | "full" | "sample">("none");
  const [regularProducts, setRegularProducts] = useState<ProductItem[]>([]);
  const [sampleProducts, setSampleProducts] = useState<ProductItem[]>([]);
  const [orderQuantities, setOrderQuantities] = useState<Record<string, number>>({});
  const [showOrderReviewModal, setShowOrderReviewModal] = useState(false);

  const [samplePaymentModal, setSamplePaymentModal] = useState(false);
  const [selectedSampleItem, setSelectedSampleItem] = useState<ProductItem | null>(null);
  const [trialActivated, setTrialActivated] = useState(false);

  const [step5, setStep5] = useState<Step5Data>({
    pan: "",
    has_gst: true,
    gst_number: "",
    fssai_number: "",
    udyam_number: "",
    shop_est_number: "",
  });
  const [step5Warnings, setStep5Warnings] = useState<string[]>([]);
  const [verificationResults, setVerificationResults] = useState<{
    panVerified?: boolean;
    gstVerified?: boolean;
    panHolderName?: string;
    gstLegalName?: string;
  }>({});

  const [step6, setStep6] = useState<Step6Data>({
    authorized: false,
    signature_name: "",
  });

  const [step7, setStep7] = useState<Step7Data>({
    account_number: "",
    ifsc: "",
    account_holder: "",
    bank_name: "HDFC Bank",
    branch: "Main Branch",
  });

  const [, setAppStatus] = useState<AppStatus | null>(null);

  useEffect(() => {
    const savedToken = localStorage.getItem("kresconet_token");
    const savedDistId = localStorage.getItem("kresconet_dist_id");
    if (savedToken && savedDistId) {
      setToken(savedToken);
      setDistributorId(savedDistId);
      loadApplicationStatus();
    }
  }, []);

  const loadApplicationStatus = async () => {
    setLoading(true);
    const res = await fetchApi<AppStatus>("/onboarding/status");
    setLoading(false);
    if (res.success && res.data) {
      setAppStatus(res.data);
      const st = res.data.status;
      if (st === "trial") {
        setTrialActivated(true);
        setStep("step8_approval");
      } else if (st === "consent_given" || st === "under_review" || st === "approved" || st === "credit_active") {
        setStep("step8_approval");
      } else if (st === "bank_submitted") {
        setStep("step8_approval");
      } else if (st === "statutory_submitted") {
        setStep("step6_auth");
      } else if (st === "business_submitted") {
        setStep("step3_credit_pref");
      } else if (st === "basic_submitted") {
        setStep("step2_business_exp");
      } else {
        setStep("step1_business_det");
      }
    } else {
      setStep("step1_business_det");
    }

    fetchCatalogues();
  };

  const fetchCatalogues = async () => {
    const resReg = await fetchApi<ProductItem[]>("/catalogue");
    if (resReg.success && Array.isArray(resReg.data)) {
      setRegularProducts(resReg.data);
    }
    const resSample = await fetchApi<ProductItem[]>("/catalogue/samples");
    if (resSample.success && Array.isArray(resSample.data)) {
      setSampleProducts(resSample.data);
    }
  };

  // Auth Handlers
  const handleSendOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setSuccessMsg(null);
    setLoading(true);

    const res = await fetchApi<{ dev_otp?: string }>("/auth/otp/send", {
      method: "POST",
      body: JSON.stringify({ mobile, purpose: "onboarding" }),
    });

    setLoading(false);
    if (res.success) {
      setOtpSent(true);
      setSuccessMsg("OTP sent successfully to your mobile number.");
      if (res.data?.dev_otp) {
        setDevOtp(res.data.dev_otp);
        setOtp(res.data.dev_otp);
      }
    } else {
      setErrorMsg(res.error?.message || "Failed to send OTP");
    }
  };

  const handleVerifyOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setSuccessMsg(null);
    setLoading(true);

    const res = await fetchApi<{ token: string; distributor_id: string }>("/auth/otp/verify", {
      method: "POST",
      body: JSON.stringify({ mobile, otp, purpose: "onboarding" }),
    });

    setLoading(false);
    if (res.success && res.data) {
      const { token, distributor_id } = res.data;
      localStorage.setItem("kresconet_token", token);
      localStorage.setItem("kresconet_dist_id", distributor_id);
      setToken(token);
      setDistributorId(distributor_id);
      setSuccessMsg("Mobile Verified Successfully!");
      loadApplicationStatus();
    } else {
      setErrorMsg(res.error?.message || "Invalid or expired OTP");
    }
  };

  const handleSignOut = () => {
    localStorage.removeItem("kresconet_token");
    localStorage.removeItem("kresconet_dist_id");
    setToken(null);
    setDistributorId(null);
    setStep("auth");
  };

  // Step 1 Submission: Business Details
  const handleStep1Submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const resBasic = await fetchApi<unknown>("/onboarding/basic", {
      method: "POST",
      body: JSON.stringify({ name: step1.name, email: step1.email }),
    });

    if (!resBasic.success) {
      setLoading(false);
      setErrorMsg(resBasic.error?.message || "Failed to save basic details");
      return;
    }

    setLoading(false);
    setStep("step2_business_exp");
  };

  // Step 2 Submission: Experience & Distribution
  const handleStep2Submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const resBiz = await fetchApi<unknown>("/onboarding/business", {
      method: "POST",
      body: JSON.stringify({
        business_name: step1.business_name,
        constitution: step1.constitution,
        address_line1: step1.address_line1,
        address_line2: step1.address_line2,
        city: step1.city,
        state: step1.state,
        pin: step1.pin,
        vintage_years: Number(step2.vintage_years),
        fmcg_experience_years: Number(step2.distribution_experience_years),
        distribution_experience_years: Number(step2.distribution_experience_years),
        approx_monthly_business_inr: Number(step2.approx_monthly_business_inr),
        retailer_count: Number(step2.serviced_retailers_wholesalers_count),
        serviced_retailers_wholesalers_count: Number(step2.serviced_retailers_wholesalers_count),
        salesperson_count: 2,
        interested_business_role: step2.interested_business_role,
        existing_brands: step2.existing_brands.split(",").map((s) => s.trim()),
      }),
    });

    setLoading(false);
    if (resBiz.success) {
      setStep("step3_credit_pref");
    } else {
      setErrorMsg(resBiz.error?.message || "Failed to save experience details");
    }
  };

  // Step 3 Submission: Credit Preference
  const handleStep3Submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const res = await fetchApi<unknown>("/onboarding/preference", {
      method: "POST",
      body: JSON.stringify({ preference: step3.preference }),
    });

    setLoading(false);
    if (res.success) {
      setStep("step4_order_req");
    } else {
      setErrorMsg(res.error?.message || "Failed to save credit preference");
    }
  };

  // Step 4 Handlers
  const handleQuantityChange = (id: string, delta: number) => {
    setOrderQuantities((prev) => {
      const current = prev[id] || 0;
      const next = Math.max(0, current + delta);
      return { ...prev, [id]: next };
    });
  };

  const calculateOrderTotal = () => {
    return regularProducts.reduce((acc, p) => {
      const q = orderQuantities[p.id] || 0;
      return acc + (p.price_paise / 100) * q;
    }, 0);
  };

  const handleConfirmFullOrder = async () => {
    setErrorMsg(null);
    setLoading(true);

    const items = regularProducts
      .filter((p) => (orderQuantities[p.id] || 0) > 0)
      .map((p) => ({
        product_id: p.id,
        quantity: orderQuantities[p.id],
      }));

    if (items.length === 0) {
      setLoading(false);
      setErrorMsg("Please select at least 1 product quantity.");
      return;
    }

    const res = await fetchApi<unknown>("/orders", {
      method: "POST",
      body: JSON.stringify({ items }),
    });

    setLoading(false);
    if (res.success) {
      setShowOrderReviewModal(false);
      setSuccessMsg("Order placed successfully! Proceeding to KYC & GST verification.");
      setStep("step5_kyc_gst");
    } else {
      const msg =
        typeof res.error?.details === "string"
          ? res.error.details
          : res.error?.message || "Failed to submit order.";
      setErrorMsg(msg);
    }
  };

  const handleInitiateSampleBooking = async (item: ProductItem) => {
    setSelectedSampleItem(item);
    setErrorMsg(null);
    setLoading(true);

    const res = await fetchApi<unknown>("/onboarding/sample-order", {
      method: "POST",
      body: JSON.stringify({
        amount_paise: item.price_paise || 50000,
        items: [{ product_id: item.id, quantity: 1, product_name: item.name }],
      }),
    });

    setLoading(false);
    if (res.success && res.data) {
      setSamplePaymentModal(true);
    } else {
      setErrorMsg(res.error?.message || "Failed to initiate sample trial order");
    }
  };

  const handleCompleteRazorpayPayment = async () => {
    setLoading(true);
    setErrorMsg(null);

    const res = await fetchApi<unknown>("/onboarding/sample-payment/verify", {
      method: "POST",
      body: JSON.stringify({
        razorpay_order_id: `rzp_order_${Date.now()}`,
        razorpay_payment_id: `pay_${Date.now()}`,
        razorpay_signature: `sig_sample_${Date.now()}`,
      }),
    });

    setLoading(false);
    if (res.success) {
      setSamplePaymentModal(false);
      setTrialActivated(true);
      setStep("step8_approval");
    } else {
      setErrorMsg(res.error?.message || "Payment verification failed");
    }
  };

  // Step 5 Submission: KYC & GST
  const handleStep5Submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const res = await fetchApi<{
      pan_verified?: boolean;
      gst_verified?: boolean;
      pan_holder_name?: string;
      gst_legal_name?: string;
      warnings?: string[];
    }>("/onboarding/statutory", {
      method: "POST",
      body: JSON.stringify(step5),
    });

    setLoading(false);
    if (res.success) {
      setVerificationResults({
        panVerified: res.data?.pan_verified,
        gstVerified: res.data?.gst_verified,
        panHolderName: res.data?.pan_holder_name,
        gstLegalName: res.data?.gst_legal_name,
      });

      const warnings = res.data?.warnings || [];
      setStep5Warnings(warnings);

      const isPanOK = res.data?.pan_verified === true;
      const isGstOK = !step5.has_gst || res.data?.gst_verified === true;
      const hasWarnings = warnings.length > 0;

      // Gate Guard: Stop user on Step 5 if verification fails or warnings/mismatches exist
      if (!isPanOK || !isGstOK || hasWarnings) {
        setErrorMsg("KYC / GST verification failed or discrepancies detected. Please resolve errors before proceeding.");
        return;
      }

      setStep("step6_auth");
    } else {
      setErrorMsg(res.error?.message || "Failed to submit KYC & GST details");
    }
  };

  // Step 6 Submission: Authorization
  const handleStep6Submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!step6.authorized) {
      setErrorMsg("Please accept the legal authorization declaration.");
      return;
    }
    setErrorMsg(null);
    setLoading(true);

    const res = await fetchApi<unknown>("/onboarding/consent", {
      method: "POST",
      body: JSON.stringify({
        consent_type: "credit_assessment",
        consent_text: "I hereby authorize Kresconet to conduct background, GST & credit checks.",
        consent_version: "v2.0",
      }),
    });

    setLoading(false);
    if (res.success) {
      setStep("step7_bank");
    } else {
      setErrorMsg(res.error?.message || "Failed to save authorization consent");
    }
  };

  // Step 7 Submission: Bank Account
  const handleStep7Submit = async (e: React.FormEvent, skip = false) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const bodyData = skip
      ? { account_number: "", ifsc: "", account_holder: "" }
      : step7;

    const res = await fetchApi<unknown>("/onboarding/bank", {
      method: "POST",
      body: JSON.stringify(bodyData),
    });

    setLoading(false);
    if (res.success) {
      setStep("step8_approval");
    } else {
      setErrorMsg(res.error?.message || "Failed to submit bank details");
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans">
      <OnboardingHeader token={token} onSignOut={handleSignOut} />

      <main className="flex-1 max-w-6xl w-full mx-auto px-6 py-8">
        {step !== "auth" && step !== "step9_dashboard" && (
          <OnboardingStepNav step={step} />
        )}

        {/* Global Error Banner */}
        {errorMsg && (
          <div className="mb-6 p-4 bg-rose-500/10 border border-rose-500/30 rounded-xl text-rose-400 text-sm flex items-center justify-between">
            <div className="flex items-center gap-3">
              <AlertCircle className="w-5 h-5 flex-shrink-0" />
              <span>{errorMsg}</span>
            </div>
            <button onClick={() => setErrorMsg(null)} className="text-rose-400 hover:text-white">
              ✕
            </button>
          </div>
        )}

        {/* Global Success Banner */}
        {successMsg && (
          <div className="mb-6 p-4 bg-emerald-500/10 border border-emerald-500/30 rounded-xl text-emerald-400 text-sm flex items-center justify-between">
            <div className="flex items-center gap-3">
              <CheckCircle2 className="w-5 h-5 flex-shrink-0" />
              <span>{successMsg}</span>
            </div>
            <button onClick={() => setSuccessMsg(null)} className="text-emerald-400 hover:text-white">
              ✕
            </button>
          </div>
        )}

        {/* STEP 0: AUTH */}
        {step === "auth" && (
          <AuthStep
            mobile={mobile}
            setMobile={setMobile}
            otp={otp}
            setOtp={setOtp}
            devOtp={devOtp}
            otpSent={otpSent}
            loading={loading}
            onSendOtp={handleSendOtp}
            onVerifyOtp={handleVerifyOtp}
          />
        )}

        {/* STEP 1: BUSINESS DETAILS */}
        {step === "step1_business_det" && (
          <Step1BusinessDetails
            step1={step1}
            setStep1={setStep1}
            loading={loading}
            onSubmit={handleStep1Submit}
          />
        )}

        {/* STEP 2: BUSINESS EXPERIENCE */}
        {step === "step2_business_exp" && (
          <Step2BusinessExperience
            step2={step2}
            setStep2={setStep2}
            loading={loading}
            onBack={() => setStep("step1_business_det")}
            onSubmit={handleStep2Submit}
          />
        )}

        {/* STEP 3: CREDIT PREFERENCE */}
        {step === "step3_credit_pref" && (
          <Step3CreditPreference
            step3={step3}
            setStep3={setStep3}
            loading={loading}
            onBack={() => setStep("step2_business_exp")}
            onSubmit={handleStep3Submit}
          />
        )}

        {/* STEP 4: ORDER REQUIREMENT */}
        {step === "step4_order_req" && (
          <Step4OrderRequirement
            orderChoice={orderChoice}
            setOrderChoice={setOrderChoice}
            regularProducts={regularProducts}
            sampleProducts={sampleProducts}
            orderQuantities={orderQuantities}
            handleQuantityChange={handleQuantityChange}
            calculateOrderTotal={calculateOrderTotal}
            onOpenOrderReview={() => setShowOrderReviewModal(true)}
            onInitiateSampleBooking={handleInitiateSampleBooking}
          />
        )}

        {/* STEP 5: KYC & GST */}
        {step === "step5_kyc_gst" && (
          <Step5KycGst
            step5={step5}
            setStep5={setStep5}
            loading={loading}
            onBack={() => setStep("step4_order_req")}
            onSubmit={handleStep5Submit}
            verificationWarnings={step5Warnings}
            verificationResults={verificationResults}
            step1Name={step1.name}
            step1BusinessName={step1.business_name}
          />
        )}

        {/* STEP 6: AUTHORIZATION */}
        {step === "step6_auth" && (
          <Step6Authorization
            step6={step6}
            setStep6={setStep6}
            loading={loading}
            onBack={() => setStep("step5_kyc_gst")}
            onSubmit={handleStep6Submit}
          />
        )}

        {/* STEP 7: BANK ACCOUNT */}
        {step === "step7_bank" && (
          <Step7BankAccount
            step7={step7}
            setStep7={setStep7}
            loading={loading}
            onSubmit={handleStep7Submit}
          />
        )}

        {/* STEP 8: APPROVAL */}
        {step === "step8_approval" && (
          <Step8Approval
            trialActivated={trialActivated}
            onGoToDashboard={() => setStep("step9_dashboard")}
          />
        )}

        {/* STEP 9: DASHBOARD */}
        {step === "step9_dashboard" && (
          <Step9Dashboard
            trialActivated={trialActivated}
            regularProducts={regularProducts}
          />
        )}
      </main>

      {/* MODALS */}
      {showOrderReviewModal && (
        <OrderReviewModal
          regularProducts={regularProducts}
          orderQuantities={orderQuantities}
          calculateOrderTotal={calculateOrderTotal}
          loading={loading}
          errorMsg={errorMsg}
          onClose={() => {
            setErrorMsg(null);
            setShowOrderReviewModal(false);
          }}
          onConfirm={handleConfirmFullOrder}
        />
      )}

      {samplePaymentModal && selectedSampleItem && (
        <SamplePaymentModal
          selectedSampleItem={selectedSampleItem}
          loading={loading}
          onClose={() => setSamplePaymentModal(false)}
          onCompletePayment={handleCompleteRazorpayPayment}
        />
      )}
    </div>
  );
}
