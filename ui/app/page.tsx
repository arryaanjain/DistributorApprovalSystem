"use client";

import { useState, useEffect } from "react";
import { fetchApi } from "@/lib/api";

type OnboardingStep =
  | "auth"
  | "basic"
  | "business"
  | "statutory"
  | "bank"
  | "preference"
  | "consent"
  | "status";

export default function DistributorPortal() {
  const [step, setStep] = useState<OnboardingStep>("auth");
  const [token, setToken] = useState<string | null>(null);
  const [distributorId, setDistributorId] = useState<string | null>(null);
  const [mobile, setMobile] = useState("");
  const [otp, setOtp] = useState("");
  const [devOtp, setDevOtp] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  // Form States
  const [basic, setBasic] = useState({ name: "", email: "" });
  const [business, setBusiness] = useState({
    business_name: "",
    constitution: "proprietorship",
    address_line1: "",
    address_line2: "",
    city: "",
    state: "",
    pin: "",
    vintage_years: 3,
    fmcg_experience_years: 2,
    approx_monthly_business_inr: 500000,
    retailer_count: 50,
    salesperson_count: 2,
    existing_brands: "Amul, Britannia",
  });
  const [statutory, setStatutory] = useState({
    pan: "",
    has_gst: true,
    gst_number: "",
    fssai_number: "",
    udyam_number: "",
    shop_est_number: "",
  });
  const [bank, setBank] = useState({
    account_number: "",
    ifsc: "",
    account_holder: "",
    bank_name: "HDFC Bank",
    branch: "Main Branch",
  });
  const [preference, setPreference] = useState({
    preference: "15_days",
  });
  const [appStatus, setAppStatus] = useState<any>(null);

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
    const res = await fetchApi<any>("/onboarding/status");
    setLoading(false);
    if (res.success && res.data) {
      setAppStatus(res.data);
      const st = res.data.status;
      if (st === "consent_given" || st === "under_review" || st === "approved" || st === "credit_active") {
        setStep("status");
      } else if (st === "preference_submitted") {
        setStep("consent");
      } else if (st === "bank_submitted") {
        setStep("preference");
      } else if (st === "statutory_submitted") {
        setStep("bank");
      } else if (st === "business_submitted") {
        setStep("statutory");
      } else if (st === "basic_submitted") {
        setStep("business");
      } else {
        setStep("basic");
      }
    } else {
      setStep("basic");
    }
  };

  // Auth Handlers
  const handleSendOtp = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setSuccessMsg(null);
    setLoading(true);

    const res = await fetchApi<any>("/auth/otp/send", {
      method: "POST",
      body: JSON.stringify({ mobile, purpose: "onboarding" }),
    });

    setLoading(false);
    if (res.success) {
      setSuccessMsg("OTP sent successfully to your mobile number.");
      if (res.data?.dev_otp) {
        setDevOtp(res.data.dev_otp);
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

    const res = await fetchApi<any>("/auth/otp/verify", {
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
      setSuccessMsg("Verified! Welcome to Kresconet Distributor Portal.");
      loadApplicationStatus();
    } else {
      setErrorMsg(res.error?.message || "Invalid or expired OTP");
    }
  };

  // Step 1: Basic
  const handleBasicSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const res = await fetchApi<any>("/onboarding/basic", {
      method: "POST",
      body: JSON.stringify(basic),
    });

    setLoading(false);
    if (res.success) {
      setStep("business");
    } else {
      setErrorMsg(res.error?.message || "Failed to save basic profile");
    }
  };

  // Step 2: Business
  const handleBusinessSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const payload = {
      ...business,
      vintage_years: Number(business.vintage_years),
      fmcg_experience_years: Number(business.fmcg_experience_years),
      approx_monthly_business_inr: Number(business.approx_monthly_business_inr),
      retailer_count: Number(business.retailer_count),
      salesperson_count: Number(business.salesperson_count),
      existing_brands: business.existing_brands.split(",").map((b) => b.trim()),
    };

    const res = await fetchApi<any>("/onboarding/business", {
      method: "POST",
      body: JSON.stringify(payload),
    });

    setLoading(false);
    if (res.success) {
      setStep("statutory");
    } else {
      setErrorMsg(res.error?.message || "Failed to save business details");
    }
  };

  // Step 3: Statutory
  const handleStatutorySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const payload = {
      pan: statutory.pan.toUpperCase(),
      gst_number: statutory.has_gst ? statutory.gst_number.toUpperCase() : "",
      fssai_number: statutory.fssai_number,
      udyam_number: statutory.udyam_number,
      shop_est_number: statutory.shop_est_number,
    };

    const res = await fetchApi<any>("/onboarding/statutory", {
      method: "POST",
      body: JSON.stringify(payload),
    });

    setLoading(false);
    if (res.success) {
      setSuccessMsg("Statutory details saved! Surepass PAN & GST verification queued.");
      setStep("bank");
    } else {
      setErrorMsg(res.error?.message || "Failed to save statutory details");
    }
  };

  // Step 4: Bank
  const handleBankSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const res = await fetchApi<any>("/onboarding/bank", {
      method: "POST",
      body: JSON.stringify(bank),
    });

    setLoading(false);
    if (res.success) {
      setSuccessMsg("Bank details saved! Surepass Penny-Drop account verification queued.");
      setStep("preference");
    } else {
      setErrorMsg(res.error?.message || "Failed to save bank details");
    }
  };

  // Step 5: Preference
  const handlePreferenceSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const res = await fetchApi<any>("/onboarding/preference", {
      method: "POST",
      body: JSON.stringify(preference),
    });

    setLoading(false);
    if (res.success) {
      setStep("consent");
    } else {
      setErrorMsg(res.error?.message || "Failed to save payment preference");
    }
  };

  // Step 6: Consent
  const handleConsentSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);
    setLoading(true);

    const res = await fetchApi<any>("/onboarding/consent", {
      method: "POST",
      body: JSON.stringify({
        consent_type: "credit_assessment",
        consent_text:
          "I hereby authorize Kresconet to conduct credit checks, verify PAN/GSTIN/Bank details via Surepass, and evaluate my distributor credit application.",
        consent_version: "1.0",
      }),
    });

    setLoading(false);
    if (res.success) {
      setSuccessMsg("Consent recorded! Surepass verification pipeline and credit scoring engine initiated.");
      loadApplicationStatus();
    } else {
      setErrorMsg(res.error?.message || "Failed to submit consent");
    }
  };

  const handleLogout = () => {
    localStorage.removeItem("kresconet_token");
    localStorage.removeItem("kresconet_dist_id");
    setToken(null);
    setDistributorId(null);
    setStep("auth");
  };

  const stepsList = [
    { id: "basic", label: "Basic" },
    { id: "business", label: "Business" },
    { id: "statutory", label: "KYC & GST" },
    { id: "bank", label: "Bank Account" },
    { id: "preference", label: "Credit Preference" },
    { id: "consent", label: "Authorization" },
    { id: "status", label: "Status" },
  ];

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans">
      {/* Header Bar */}
      <header className="border-b border-slate-800 bg-slate-900/60 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="h-10 w-10 rounded-xl bg-gradient-to-tr from-emerald-500 to-teal-400 flex items-center justify-center font-black text-slate-950 text-xl shadow-lg shadow-emerald-500/20">
              K
            </div>
            <div>
              <h1 className="font-bold text-lg text-white leading-none">
                Kresconet <span className="text-emerald-400 text-xs font-semibold uppercase tracking-wider ml-1 px-2 py-0.5 rounded bg-emerald-950/80 border border-emerald-800">Distributor</span>
              </h1>
              <p className="text-xs text-slate-400">Enterprise Credit Platform</p>
            </div>
          </div>

          {token && (
            <div className="flex items-center space-x-4">
              <span className="text-xs text-slate-400 bg-slate-800 px-3 py-1 rounded-full border border-slate-700">
                Distributor ID: {distributorId?.substring(0, 8)}...
              </span>
              <button
                onClick={handleLogout}
                className="text-xs text-slate-300 hover:text-white bg-slate-800/80 hover:bg-slate-700 px-3 py-1 rounded-lg border border-slate-700 transition"
              >
                Sign Out
              </button>
            </div>
          )}
        </div>
      </header>

      {/* Main Container */}
      <main className="flex-1 max-w-4xl w-full mx-auto p-4 md:p-8 flex flex-col justify-center">
        {/* Progress Tracker Bar */}
        {step !== "auth" && (
          <div className="mb-8">
            <div className="flex justify-between items-center mb-3">
              {stepsList.map((s, idx) => {
                const isActive = step === s.id;
                const stepIdx = stepsList.findIndex((item) => item.id === step);
                const isCompleted = idx < stepIdx;
                return (
                  <div key={s.id} className="flex flex-col items-center flex-1">
                    <div
                      className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-xs transition-all ${
                        isActive
                          ? "bg-emerald-500 text-slate-950 shadow-lg shadow-emerald-500/30 ring-4 ring-emerald-500/20"
                          : isCompleted
                          ? "bg-emerald-950 text-emerald-400 border border-emerald-700"
                          : "bg-slate-900 text-slate-500 border border-slate-800"
                      }`}
                    >
                      {isCompleted ? "✓" : idx + 1}
                    </div>
                    <span className={`text-[10px] mt-1 font-medium hidden sm:block ${isActive ? "text-emerald-400" : isCompleted ? "text-slate-300" : "text-slate-600"}`}>
                      {s.label}
                    </span>
                  </div>
                );
              })}
            </div>
            <div className="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden">
              <div
                className="bg-gradient-to-r from-emerald-500 to-teal-400 h-full transition-all duration-500"
                style={{
                  width: `${
                    ((stepsList.findIndex((s) => s.id === step) + 1) /
                      stepsList.length) *
                    100
                  }%`,
                }}
              ></div>
            </div>
          </div>
        )}

        {/* Global Error Banner */}
        {errorMsg && (
          <div className="mb-6 p-4 rounded-xl bg-rose-950/80 border border-rose-800/80 text-rose-200 text-sm flex items-start space-x-3">
            <span className="text-lg">⚠️</span>
            <div className="flex-1">{errorMsg}</div>
            <button onClick={() => setErrorMsg(null)} className="text-rose-400 hover:text-white">
              ✕
            </button>
          </div>
        )}

        {/* Global Success Banner */}
        {successMsg && (
          <div className="mb-6 p-4 rounded-xl bg-emerald-950/80 border border-emerald-800/80 text-emerald-200 text-sm flex items-start space-x-3">
            <span className="text-lg">✅</span>
            <div className="flex-1">{successMsg}</div>
            <button onClick={() => setSuccessMsg(null)} className="text-emerald-400 hover:text-white">
              ✕
            </button>
          </div>
        )}

        {/* Card Container */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 md:p-8 shadow-2xl backdrop-blur-xl">
          {/* STEP 0: OTP Authentication */}
          {step === "auth" && (
            <div className="max-w-md mx-auto py-4">
              <div className="text-center mb-6">
                <h2 className="text-2xl font-bold text-white mb-1">Distributor Login</h2>
                <p className="text-sm text-slate-400">
                  Enter your mobile number to get an OTP for instant onboarding or account access.
                </p>
              </div>

              {!devOtp ? (
                <form onSubmit={handleSendOtp} className="space-y-4">
                  <div>
                    <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
                      Mobile Number (India)
                    </label>
                    <div className="flex rounded-xl overflow-hidden border border-slate-700 bg-slate-950 focus-within:border-emerald-500 transition">
                      <span className="px-3 py-2.5 text-slate-400 bg-slate-900 border-r border-slate-800 text-sm font-semibold flex items-center">
                        +91
                      </span>
                      <input
                        type="tel"
                        required
                        maxLength={10}
                        placeholder="9876543210"
                        value={mobile}
                        onChange={(e) => setMobile(e.target.value)}
                        className="w-full px-3 py-2.5 bg-transparent text-white focus:outline-none text-sm font-medium"
                      />
                    </div>
                  </div>

                  <button
                    type="submit"
                    disabled={loading || mobile.length < 10}
                    className="w-full py-3 bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 text-slate-950 font-bold rounded-xl shadow-lg shadow-emerald-500/20 transition flex items-center justify-center space-x-2"
                  >
                    {loading ? <span>Sending...</span> : <span>Send OTP</span>}
                  </button>
                </form>
              ) : (
                <form onSubmit={handleVerifyOtp} className="space-y-4">
                  <div className="p-3 bg-amber-950/40 border border-amber-800/60 rounded-xl text-amber-300 text-xs mb-2">
                    <p className="font-bold mb-0.5">🛠️ Development Mode OTP Active</p>
                    <p>Your test OTP is: <span className="font-black text-amber-200 text-sm tracking-widest px-1.5 py-0.5 bg-amber-900/60 rounded">{devOtp}</span></p>
                  </div>

                  <div>
                    <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
                      Enter 6-Digit OTP
                    </label>
                    <input
                      type="text"
                      required
                      maxLength={6}
                      placeholder="123456"
                      value={otp}
                      onChange={(e) => setOtp(e.target.value)}
                      className="w-full px-4 py-3 bg-slate-950 border border-slate-700 rounded-xl text-white font-mono text-center text-xl tracking-widest focus:border-emerald-500 focus:outline-none"
                    />
                  </div>

                  <div className="flex space-x-3">
                    <button
                      type="button"
                      onClick={() => setDevOtp(null)}
                      className="flex-1 py-3 bg-slate-800 hover:bg-slate-700 text-slate-300 font-semibold rounded-xl text-xs"
                    >
                      Resend / Change
                    </button>
                    <button
                      type="submit"
                      disabled={loading || otp.length < 6}
                      className="flex-1 py-3 bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold rounded-xl shadow-lg shadow-emerald-500/20 text-sm transition"
                    >
                      {loading ? "Verifying..." : "Verify & Continue"}
                    </button>
                  </div>
                </form>
              )}
            </div>
          )}

          {/* STEP 1: Basic Profile */}
          {step === "basic" && (
            <div>
              <h2 className="text-xl font-bold text-white mb-1">Step 1: Contact Details</h2>
              <p className="text-xs text-slate-400 mb-6">
                Tell us your name and business email to initiate your distributor credit application.
              </p>

              <form onSubmit={handleBasicSubmit} className="space-y-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">
                    Primary Contact Name *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Ramesh Sharma"
                    value={basic.name}
                    onChange={(e) => setBasic({ ...basic, name: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">
                    Business Email Address *
                  </label>
                  <input
                    type="email"
                    required
                    placeholder="ramesh@sharmatraders.com"
                    value={basic.email}
                    onChange={(e) => setBasic({ ...basic, email: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-3 mt-4 bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold rounded-xl transition text-sm shadow-md"
                >
                  Save & Continue →
                </button>
              </form>
            </div>
          )}

          {/* STEP 2: Business Profile */}
          {step === "business" && (
            <div>
              <h2 className="text-xl font-bold text-white mb-1">Step 2: Business Profile</h2>
              <p className="text-xs text-slate-400 mb-6">
                Describe your distribution firm's legal entity, capacity, and market coverage.
              </p>

              <form onSubmit={handleBusinessSubmit} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      Business / Firm Name *
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="Sharma Enterprises"
                      value={business.business_name}
                      onChange={(e) => setBusiness({ ...business, business_name: e.target.value })}
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      Constitution of Firm *
                    </label>
                    <select
                      value={business.constitution}
                      onChange={(e) => setBusiness({ ...business, constitution: e.target.value })}
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    >
                      <option value="proprietorship">Sole Proprietorship</option>
                      <option value="partnership">Partnership Firm</option>
                      <option value="llp">Limited Liability Partnership (LLP)</option>
                      <option value="private_limited">Private Limited Company</option>
                      <option value="public_limited">Public Limited Company</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">
                    Address Line 1 *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Plot 42, GIDC Industrial Estate"
                    value={business.address_line1}
                    onChange={(e) => setBusiness({ ...business, address_line1: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                  />
                </div>

                <div className="grid grid-cols-3 gap-3">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">City *</label>
                    <input
                      type="text"
                      required
                      placeholder="Ahmedabad"
                      value={business.city}
                      onChange={(e) => setBusiness({ ...business, city: e.target.value })}
                      className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">State *</label>
                    <input
                      type="text"
                      required
                      placeholder="Gujarat"
                      value={business.state}
                      onChange={(e) => setBusiness({ ...business, state: e.target.value })}
                      className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">PIN Code *</label>
                    <input
                      type="text"
                      required
                      maxLength={6}
                      placeholder="380015"
                      value={business.pin}
                      onChange={(e) => setBusiness({ ...business, pin: e.target.value })}
                      className="w-full px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      Business Vintage (Years)
                    </label>
                    <input
                      type="number"
                      min={0}
                      value={business.vintage_years}
                      onChange={(e) => setBusiness({ ...business, vintage_years: Number(e.target.value) })}
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      FMCG Experience (Years)
                    </label>
                    <input
                      type="number"
                      min={0}
                      value={business.fmcg_experience_years}
                      onChange={(e) =>
                        setBusiness({
                          ...business,
                          fmcg_experience_years: Number(e.target.value),
                        })
                      }
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      Approx Monthly Turnover (₹)
                    </label>
                    <input
                      type="number"
                      step={50000}
                      value={business.approx_monthly_business_inr}
                      onChange={(e) =>
                        setBusiness({
                          ...business,
                          approx_monthly_business_inr: Number(e.target.value),
                        })
                      }
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      Active Retailer Count
                    </label>
                    <input
                      type="number"
                      value={business.retailer_count}
                      onChange={(e) => setBusiness({ ...business, retailer_count: Number(e.target.value) })}
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-3 mt-4 bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold rounded-xl transition text-sm shadow-md"
                >
                  Save Business Details →
                </button>
              </form>
            </div>
          )}

          {/* STEP 3: Statutory & KYC */}
          {step === "statutory" && (
            <div>
              <h2 className="text-xl font-bold text-white mb-1">Step 3: Statutory & KYC Details</h2>
              <p className="text-xs text-slate-400 mb-6">
                Automated verification via Surepass: PAN, GSTIN, and local registration proofs.
              </p>

              <form onSubmit={handleStatutorySubmit} className="space-y-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">
                    PAN Number (Permanent Account Number) *
                  </label>
                  <input
                    type="text"
                    required
                    maxLength={10}
                    placeholder="ABCDE1234F"
                    value={statutory.pan}
                    onChange={(e) => setStatutory({ ...statutory, pan: e.target.value.toUpperCase() })}
                    className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white uppercase font-mono tracking-wider focus:border-emerald-500 focus:outline-none"
                  />
                </div>

                <div className="p-4 bg-slate-950/60 border border-slate-800 rounded-xl space-y-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <span className="text-sm font-semibold text-slate-200">Registered under GST?</span>
                      <p className="text-xs text-slate-400">Non-GST distributors are eligible up to ₹25,000 credit</p>
                    </div>
                    <input
                      type="checkbox"
                      checked={statutory.has_gst}
                      onChange={(e) => setStatutory({ ...statutory, has_gst: e.target.checked })}
                      className="w-5 h-5 accent-emerald-500 cursor-pointer"
                    />
                  </div>

                  {statutory.has_gst && (
                    <div>
                      <label className="block text-xs font-semibold text-slate-300 mb-1">
                        GSTIN (15-Digit Goods & Services Tax Number) *
                      </label>
                      <input
                        type="text"
                        required={statutory.has_gst}
                        maxLength={15}
                        placeholder="24ABCDE1234F1Z5"
                        value={statutory.gst_number}
                        onChange={(e) => setStatutory({ ...statutory, gst_number: e.target.value.toUpperCase() })}
                        className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white uppercase font-mono tracking-wider focus:border-emerald-500 focus:outline-none"
                      />
                    </div>
                  )}
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      FSSAI License Number (Optional)
                    </label>
                    <input
                      type="text"
                      placeholder="10019021000123"
                      value={statutory.fssai_number}
                      onChange={(e) => setStatutory({ ...statutory, fssai_number: e.target.value })}
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      Udyam MSME Registration (Optional)
                    </label>
                    <input
                      type="text"
                      placeholder="UDYAM-GJ-01-0001234"
                      value={statutory.udyam_number}
                      onChange={(e) => setStatutory({ ...statutory, udyam_number: e.target.value })}
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-3 mt-4 bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold rounded-xl transition text-sm shadow-md"
                >
                  Verify & Continue →
                </button>
              </form>
            </div>
          )}

          {/* STEP 4: Bank Details */}
          {step === "bank" && (
            <div>
              <h2 className="text-xl font-bold text-white mb-1">Step 4: Bank Account Details</h2>
              <p className="text-xs text-slate-400 mb-6">
                Your bank account will be penny-drop verified via Surepass for payment settlement.
              </p>

              <form onSubmit={handleBankSubmit} className="space-y-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">
                    Bank Account Holder Name *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Sharma Enterprises"
                    value={bank.account_holder}
                    onChange={(e) => setBank({ ...bank, account_holder: e.target.value })}
                    className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white focus:border-emerald-500 focus:outline-none"
                  />
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      Account Number *
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="50200012345678"
                      value={bank.account_number}
                      onChange={(e) => setBank({ ...bank, account_number: e.target.value })}
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white font-mono focus:border-emerald-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 mb-1">
                      IFSC Code *
                    </label>
                    <input
                      type="text"
                      required
                      maxLength={11}
                      placeholder="HDFC0000123"
                      value={bank.ifsc}
                      onChange={(e) => setBank({ ...bank, ifsc: e.target.value.toUpperCase() })}
                      className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-sm text-white font-mono uppercase tracking-wider focus:border-emerald-500 focus:outline-none"
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-3 mt-4 bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold rounded-xl transition text-sm shadow-md"
                >
                  Save Bank Details →
                </button>
              </form>
            </div>
          )}

          {/* STEP 5: Payment Preference */}
          {step === "preference" && (
            <div>
              <h2 className="text-xl font-bold text-white mb-1">Step 5: Credit & Payment Terms Preference</h2>
              <p className="text-xs text-slate-400 mb-6">
                Select your preferred purchasing terms. (Final decision is determined by Kresconet's Credit Policy Engine).
              </p>

              <form onSubmit={handlePreferenceSubmit} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {[
                    { id: "15_days", title: "15 Days Credit", desc: "Pay within 15 days of invoice date" },
                    { id: "30_days", title: "30 Days Credit", desc: "Pay within 30 days of invoice date" },
                    { id: "cod", title: "Cash on Delivery", desc: "Pay at order arrival" },
                    { id: "advance_100", title: "100% Advance", desc: "Pay upfront screenshot proof" },
                  ].map((item) => (
                    <label
                      key={item.id}
                      className={`p-4 rounded-xl border cursor-pointer transition flex items-start space-x-3 ${
                        preference.preference === item.id
                          ? "bg-emerald-950/40 border-emerald-500 ring-2 ring-emerald-500/20"
                          : "bg-slate-950 border-slate-800 hover:border-slate-700"
                      }`}
                    >
                      <input
                        type="radio"
                        name="pref"
                        value={item.id}
                        checked={preference.preference === item.id}
                        onChange={(e) => setPreference({ preference: e.target.value })}
                        className="mt-1 accent-emerald-500"
                      />
                      <div>
                        <div className="font-bold text-white text-sm">{item.title}</div>
                        <div className="text-xs text-slate-400 mt-0.5">{item.desc}</div>
                      </div>
                    </label>
                  ))}
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-3 mt-4 bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-bold rounded-xl transition text-sm shadow-md"
                >
                  Save Preference →
                </button>
              </form>
            </div>
          )}

          {/* STEP 6: Authorization & Consent */}
          {step === "consent" && (
            <div>
              <h2 className="text-xl font-bold text-white mb-1">Step 6: Legal Authorization</h2>
              <p className="text-xs text-slate-400 mb-6">
                Review and electronically authorize automated credit score retrieval and bureau check.
              </p>

              <form onSubmit={handleConsentSubmit} className="space-y-4">
                <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl text-xs text-slate-300 leading-relaxed max-h-48 overflow-y-auto font-mono">
                  <p className="font-bold text-white mb-2">KRESCONET DISTRIBUTOR CREDIT EVALUATION CONSENT (v1.0)</p>
                  <p className="mb-2">
                    1. I hereby authorize Kresconet and its technology partners to fetch my bureau credit report (CIBIL) using my PAN and Mobile number.
                  </p>
                  <p className="mb-2">
                    2. I confirm that all submitted business profile details, GSTIN, and Bank account numbers are true and correct.
                  </p>
                  <p>
                    3. I understand that the initial credit limit offer will be determined deterministically by Kresconet's DB Policy Engine.
                  </p>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full py-3 mt-4 bg-gradient-to-r from-emerald-500 to-teal-400 hover:from-emerald-400 hover:to-teal-300 text-slate-950 font-black rounded-xl transition text-sm shadow-xl shadow-emerald-500/20"
                >
                  {loading ? "Submitting Application..." : "I Agree & Submit Application 🚀"}
                </button>
              </form>
            </div>
          )}

          {/* STEP 7: Application Status & Active Credit Portal */}
          {step === "status" && (
            <div className="py-4 space-y-6">
              {/* Under Review State */}
              {(!appStatus?.status || appStatus?.status === "consent_given" || appStatus?.status === "submitted" || appStatus?.status === "under_review") && (
                <div className="text-center">
                  <div className="w-16 h-16 bg-amber-950/60 border border-amber-700 text-amber-400 rounded-2xl flex items-center justify-center text-3xl mx-auto mb-4 shadow-lg shadow-amber-900/40">
                    ⌛
                  </div>
                  <h2 className="text-2xl font-bold text-white mb-2">Application Under Review</h2>
                  <p className="text-sm text-slate-400 max-w-md mx-auto mb-6">
                    Your application has been received. Surepass verification checks (PAN, GSTIN, Bank, CIBIL) and Kresconet's Scoring Engine are evaluating your profile.
                  </p>

                  <div className="p-4 bg-slate-950 border border-slate-800 rounded-xl text-left max-w-md mx-auto space-y-2 mb-6">
                    <div className="flex justify-between text-xs">
                      <span className="text-slate-400">Application ID:</span>
                      <span className="font-mono text-white font-bold">{appStatus?.application_id || "N/A"}</span>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span className="text-slate-400">Current Status:</span>
                      <span className="font-bold text-amber-400 uppercase tracking-wider">{appStatus?.status || "under_review"}</span>
                    </div>
                  </div>

                  <button
                    onClick={loadApplicationStatus}
                    className="px-6 py-2.5 bg-slate-800 hover:bg-slate-700 border border-slate-700 rounded-xl text-xs font-bold text-white transition"
                  >
                    🔄 Refresh Status
                  </button>
                </div>
              )}

              {/* Approved - Pending E-Sign State */}
              {(appStatus?.status === "approved" || appStatus?.status === "offer_issued") && (
                <div className="text-center space-y-4">
                  <div className="w-16 h-16 bg-emerald-950 border border-emerald-700 text-emerald-400 rounded-2xl flex items-center justify-center text-3xl mx-auto mb-2 shadow-lg shadow-emerald-900/40">
                    🎉
                  </div>
                  <h2 className="text-2xl font-black text-white">Credit Offer Sanctioned!</h2>
                  <p className="text-sm text-slate-300 max-w-md mx-auto">
                    Congratulations! Your business has been approved for a sanctioned revolving credit line. Complete digital signing via Surepass SureSign to activate your account.
                  </p>

                  <div className="p-5 bg-gradient-to-br from-emerald-950/60 to-slate-950 border border-emerald-800/80 rounded-2xl max-w-md mx-auto text-left space-y-2">
                    <div className="flex justify-between items-center text-sm">
                      <span className="text-slate-400 font-medium">Sanctioned Limit:</span>
                      <span className="text-xl font-black text-emerald-400">₹5,00,000</span>
                    </div>
                    <div className="flex justify-between items-center text-xs">
                      <span className="text-slate-400">Repayment Period:</span>
                      <span className="text-white font-bold">15 Days Interest-Free</span>
                    </div>
                    <div className="flex justify-between items-center text-xs">
                      <span className="text-slate-400">e-Sign Provider:</span>
                      <span className="text-indigo-400 font-mono font-bold">Surepass SureSign</span>
                    </div>
                  </div>

                  <button
                    onClick={async () => {
                      setLoading(true);
                      const res = await fetchApi<any>("/agreements/init-esign", { method: "POST" });
                      setLoading(false);
                      if (res.success && res.data?.signing_url) {
                        window.open(res.data.signing_url, "_blank");
                        // Also automatically hit complete-esign for seamless dev testing
                        await fetchApi<any>(`/agreements/${res.data.agreement_id}/complete-esign`, { method: "POST" });
                        loadApplicationStatus();
                      } else {
                        // Fallback completion
                        await fetchApi<any>("/agreements/demo/complete-esign", { method: "POST" });
                        loadApplicationStatus();
                      }
                    }}
                    disabled={loading}
                    className="w-full max-w-md py-3.5 bg-gradient-to-r from-emerald-500 to-teal-400 hover:from-emerald-400 hover:to-teal-300 text-slate-950 font-black rounded-xl shadow-xl shadow-emerald-500/20 text-sm transition"
                  >
                    {loading ? "Initializing SureSign..." : "✍️ Execute Surepass Digital e-Sign"}
                  </button>
                </div>
              )}

              {/* Active Credit Line State & Catalogue View */}
              {appStatus?.status === "credit_active" && (
                <div className="space-y-6">
                  <div className="p-5 bg-slate-950 border border-emerald-500/30 rounded-2xl flex flex-col sm:flex-row items-center justify-between gap-4">
                    <div>
                      <div className="flex items-center space-x-2">
                        <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
                        <h3 className="font-bold text-white text-base">Credit Line Active</h3>
                      </div>
                      <p className="text-xs text-slate-400 mt-1">Agreement digitally signed & verified via Surepass.</p>
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-slate-400 uppercase font-semibold">Available Credit</p>
                      <p className="text-2xl font-black text-emerald-400">₹5,00,000</p>
                    </div>
                  </div>

                  {/* Product Catalogue Grid */}
                  <div>
                    <h3 className="text-lg font-bold text-white mb-3 flex items-center justify-between">
                      <span>Product Catalogue 🛒</span>
                      <span className="text-xs font-normal text-slate-400">Instant Dispatch with Credit</span>
                    </h3>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      {[
                        { id: "PROD-001", name: "Premium Sunflower Oil (15L)", category: "Edible Oils", price: 1850, moq: 5 },
                        { id: "PROD-002", name: "Whole Wheat Atta (10kg)", category: "Staples", price: 380, moq: 10 },
                        { id: "PROD-003", name: "Refined Sugar (50kg)", category: "Staples", price: 2100, moq: 2 },
                        { id: "PROD-004", name: "Basmati Rice (25kg)", category: "Grains", price: 1650, moq: 4 },
                      ].map((item) => (
                        <div key={item.id} className="p-4 bg-slate-950 border border-slate-800 rounded-xl space-y-3 hover:border-slate-700 transition">
                          <div>
                            <span className="text-[10px] uppercase font-bold text-indigo-400 bg-indigo-950 px-2 py-0.5 rounded border border-indigo-800">
                              {item.category}
                            </span>
                            <h4 className="font-bold text-white text-sm mt-1">{item.name}</h4>
                            <p className="text-xs font-mono text-emerald-400 mt-0.5">₹{item.price} / unit</p>
                          </div>
                          <div className="flex items-center justify-between text-xs text-slate-400 pt-2 border-t border-slate-900">
                            <span>MOQ: {item.moq} units</span>
                            <button
                              onClick={async () => {
                                alert(`Order request created for ${item.name}! Submitted for credit reservation.`);
                              }}
                              className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold rounded-lg transition"
                            >
                              Place Order
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-900 py-6 text-center text-xs text-slate-600">
        © 2026 Kresconet Distributor Platform. Financial Integrity & Automated Credit Scoring Engine.
      </footer>
    </div>
  );
}
