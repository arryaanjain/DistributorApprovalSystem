import React, { useState } from "react";
import {
  FileText,
  CreditCard,
  CheckCircle2,
  Calendar,
  Sparkles,
  ArrowRight,
  ShieldCheck,
  ExternalLink,
  AlertCircle,
  Clock,
  ChevronRight,
  X,
} from "lucide-react";

interface CreditOrderFlowProps {
  orderId: string;
  orderNumber: string;
  totalAmountPaise: number;
  advancePaidPaise?: number;
  availableCreditPaise: number;
  token: string;
  onSuccess: () => void;
  onClose: () => void;
}

export const CreditOrderFlow: React.FC<CreditOrderFlowProps> = ({
  orderId,
  orderNumber,
  totalAmountPaise,
  advancePaidPaise = 0,
  availableCreditPaise,
  token,
  onSuccess,
  onClose,
}) => {
  const [step, setStep] = useState<"configure" | "esign" | "mandate" | "complete">("configure");
  const [frequency, setFrequency] = useState<"weekly" | "monthly">("weekly");
  const [instalments, setInstalments] = useState<number>(frequency === "weekly" ? 5 : 2);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const [cycleData, setCycleData] = useState<any>(null);
  const [signingUrl, setSigningUrl] = useState<string | null>(null);
  const [mandateUrl, setMandateUrl] = useState<string | null>(null);

  // Environment flag — set NEXT_PUBLIC_ENV=production to hide dev-only simulate buttons
  const isDev = process.env.NEXT_PUBLIC_ENV !== "production";

  const effectiveTotalPaise =
    cycleData?.total_amount_paise && cycleData.total_amount_paise > 0
      ? cycleData.total_amount_paise
      : totalAmountPaise;

  const effectiveAdvancePaise =
    cycleData?.advance_paid_paise !== undefined && cycleData.advance_paid_paise !== null
      ? cycleData.advance_paid_paise
      : advancePaidPaise;

  const netCreditPaise = effectiveTotalPaise - effectiveAdvancePaise;
  const instalmentAmountPaise = Math.ceil(netCreditPaise / instalments);

  const handleFrequencyChange = (newFreq: "weekly" | "monthly") => {
    setFrequency(newFreq);
    setInstalments(newFreq === "weekly" ? 5 : 2);
  };

  const handleInitCycle = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("http://localhost:8081/api/v1/credit-cycles/init", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          order_id: orderId,
          frequency,
          total_instalments: instalments,
          redirect_url: window.location.origin + `/orders?order_id=${orderId}&esign=completed`,
        }),
      });

      const resBody = await res.json();
      if (!res.ok || resBody.success === false) {
        throw new Error(resBody.error?.message || resBody.message || "Failed to initialize order credit cycle");
      }

      const data = resBody.data || resBody;
      setCycleData(data.cycle);
      setSigningUrl(data.signing_url);
      setStep("esign");
    } catch (err: any) {
      setError(err.message || "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  const handleSimulateESignCompletion = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`http://localhost:8081/api/v1/credit-cycles/orders/${orderId}/complete-esign`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const resBody = await res.json();
      if (!res.ok || resBody.success === false) {
        throw new Error(resBody.error?.message || resBody.message || "Failed to complete eSign authorization");
      }

      const data = resBody.data || resBody;
      setMandateUrl(data.short_url);
      setStep("mandate");
    } catch (err: any) {
      setError(err.message || "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  const handleSimulateMandateCompletion = () => {
    setStep("complete");
    setTimeout(() => {
      onSuccess();
    }, 1500);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md">
      <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        {/* Modal Header */}
        <div className="p-6 bg-gradient-to-r from-slate-900 via-indigo-950/50 to-slate-900 border-b border-slate-800 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
              <Sparkles className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white">Order Credit Cycle Setup</h3>
              <p className="text-xs text-slate-400">Order #{orderNumber} • ₹{(netCreditPaise / 100).toLocaleString("en-IN")}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-full bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white flex items-center justify-center transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Progress Tracker Bar */}
        <div className="px-6 py-3 bg-slate-950/60 border-b border-slate-800/60 flex items-center justify-between text-xs">
          <div className={`flex items-center gap-1.5 font-medium ${step === "configure" ? "text-indigo-400" : "text-emerald-400"}`}>
            <span className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] ${step === "configure" ? "bg-indigo-500/20 text-indigo-400 border border-indigo-500/40" : "bg-emerald-500/20 text-emerald-400"}`}>1</span>
            <span>Bifurcation</span>
          </div>
          <ChevronRight className="w-3.5 h-3.5 text-slate-600" />
          <div className={`flex items-center gap-1.5 font-medium ${step === "esign" ? "text-indigo-400" : step === "mandate" || step === "complete" ? "text-emerald-400" : "text-slate-500"}`}>
            <span className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] ${step === "esign" ? "bg-indigo-500/20 text-indigo-400 border border-indigo-500/40" : step === "mandate" || step === "complete" ? "bg-emerald-500/20 text-emerald-400" : "bg-slate-800 text-slate-500"}`}>2</span>
            <span>Surepass eSign</span>
          </div>
          <ChevronRight className="w-3.5 h-3.5 text-slate-600" />
          <div className={`flex items-center gap-1.5 font-medium ${step === "mandate" ? "text-indigo-400" : step === "complete" ? "text-emerald-400" : "text-slate-500"}`}>
            <span className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] ${step === "mandate" ? "bg-indigo-500/20 text-indigo-400 border border-indigo-500/40" : step === "complete" ? "bg-emerald-500/20 text-emerald-400" : "bg-slate-800 text-slate-500"}`}>3</span>
            <span>Razorpay Mandate</span>
          </div>
        </div>

        {/* Modal Content */}
        <div className="p-6 overflow-y-auto space-y-6 flex-1">
          {error && (
            <div className="p-4 bg-rose-500/10 border border-rose-500/30 rounded-2xl flex items-center gap-3 text-xs text-rose-300">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* STEP 1: CONFIGURE */}
          {step === "configure" && (
            <div className="space-y-6">
              <div>
                <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider block mb-2">
                  Select Cycle Frequency
                </label>
                <div className="grid grid-cols-2 gap-3">
                  <button
                    type="button"
                    onClick={() => handleFrequencyChange("weekly")}
                    className={`p-4 rounded-2xl border text-left transition-all ${
                      frequency === "weekly"
                        ? "bg-indigo-950/60 border-indigo-500/80 shadow-lg shadow-indigo-500/10 text-white"
                        : "bg-slate-950/40 border-slate-800 text-slate-400 hover:border-slate-700"
                    }`}
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-bold text-sm">Weekly Cycles</span>
                      <Calendar className="w-4 h-4 text-indigo-400" />
                    </div>
                    <p className="text-[11px] text-slate-400">Series of at most 5 weekly payments</p>
                  </button>

                  <button
                    type="button"
                    onClick={() => handleFrequencyChange("monthly")}
                    className={`p-4 rounded-2xl border text-left transition-all ${
                      frequency === "monthly"
                        ? "bg-indigo-950/60 border-indigo-500/80 shadow-lg shadow-indigo-500/10 text-white"
                        : "bg-slate-950/40 border-slate-800 text-slate-400 hover:border-slate-700"
                    }`}
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-bold text-sm">Monthly Cycles</span>
                      <Clock className="w-4 h-4 text-indigo-400" />
                    </div>
                    <p className="text-[11px] text-slate-400">Series of at most 2 monthly payments</p>
                  </button>
                </div>
              </div>

              <div>
                <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider block mb-2">
                  Number of Instalments
                </label>
                <div className="flex items-center gap-2">
                  {Array.from({ length: frequency === "weekly" ? 5 : 2 }).map((_, i) => {
                    const num = i + 1;
                    return (
                      <button
                        key={num}
                        type="button"
                        onClick={() => setInstalments(num)}
                        className={`flex-1 py-3 rounded-xl border text-center font-bold text-sm transition-all ${
                          instalments === num
                            ? "bg-gradient-to-r from-indigo-600 to-violet-600 border-indigo-500 text-white shadow-lg"
                            : "bg-slate-950/60 border-slate-800 text-slate-400 hover:border-slate-700"
                        }`}
                      >
                        {num} {frequency === "weekly" ? (num === 1 ? "Wk" : "Wks") : num === 1 ? "Mo" : "Mos"}
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* Repayment Breakdown Card */}
              <div className="p-5 bg-gradient-to-br from-slate-950 via-slate-900 to-indigo-950/40 border border-slate-800 rounded-2xl space-y-3">
                <h4 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
                  <CreditCard className="w-4 h-4 text-indigo-400" /> Instalment Bifurcation Summary
                </h4>
                <div className="space-y-2 pt-1 border-t border-slate-800/80 text-xs">
                  <div className="flex justify-between text-slate-400">
                    <span>Financed Credit Portion:</span>
                    <span className="font-semibold text-white">₹{(netCreditPaise / 100).toLocaleString("en-IN")}</span>
                  </div>
                  <div className="flex justify-between text-slate-400">
                    <span>Instalment Amount:</span>
                    <span className="font-black text-emerald-400 text-sm">
                      ₹{(instalmentAmountPaise / 100).toLocaleString("en-IN")} / {frequency === "weekly" ? "week" : "month"}
                    </span>
                  </div>
                  <div className="flex justify-between text-slate-400">
                    <span>Auto-Debit Mandate Provider:</span>
                    <span className="font-medium text-slate-300">Razorpay Subscriptions (UPI)</span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* STEP 2: SUREPASS ESIGN */}
          {step === "esign" && (
            <div className="space-y-6 text-center py-4">
              <div className="w-16 h-16 rounded-3xl bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400 mx-auto">
                <FileText className="w-8 h-8" />
              </div>
              <div className="space-y-2">
                <h4 className="text-lg font-bold text-white">Digital Surepass eSign Consent Required</h4>
                <p className="text-xs text-slate-400 max-w-md mx-auto leading-relaxed">
                  As per Kresconet Credit Policy, digital authorization via Surepass eSign is required for order credit cycle #{orderNumber}.
                </p>
              </div>

              <div className="p-4 bg-slate-950/80 border border-slate-800 rounded-2xl text-left text-xs space-y-2">
                <div className="flex items-center justify-between text-slate-300 font-medium">
                  <span>Agreement Reference:</span>
                  <span className="font-mono text-indigo-400">{cycleData?.agreement_id || "KRESCO-ORD-AGR"}</span>
                </div>
                <div className="flex items-center justify-between text-slate-300 font-medium">
                  <span>Surepass eSign Provider:</span>
                  <span className="text-emerald-400 flex items-center gap-1"><ShieldCheck className="w-3.5 h-3.5" /> Verified Legal eSign</span>
                </div>
              </div>

              <div className="space-y-3 pt-2">
                {signingUrl && (
                  <a
                    href={signingUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="w-full py-3.5 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs rounded-xl shadow-lg shadow-indigo-600/30 flex items-center justify-center gap-2 transition-all"
                  >
                    <span>Open Surepass eSign Window</span>
                    <ExternalLink className="w-4 h-4" />
                  </a>
                )}
                {isDev && (
                  <button
                    type="button"
                    onClick={handleSimulateESignCompletion}
                    disabled={loading}
                    className="w-full py-3.5 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-bold text-xs rounded-xl shadow-lg shadow-emerald-600/20 flex items-center justify-center gap-2 transition-all disabled:opacity-50"
                  >
                    <CheckCircle2 className="w-4 h-4" />
                    <span>[DEV] Confirm eSign Completed &amp; Proceed to Mandate</span>
                  </button>
                )}
              </div>
            </div>
          )}

          {/* STEP 3: RAZORPAY MANDATE */}
          {step === "mandate" && (
            <div className="space-y-6 text-center py-4">
              <div className="w-16 h-16 rounded-3xl bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 mx-auto">
                <CreditCard className="w-8 h-8" />
              </div>
              <div className="space-y-2">
                <h4 className="text-lg font-bold text-white">Setup Razorpay UPI e-Mandate</h4>
                <p className="text-xs text-slate-400 max-w-md mx-auto leading-relaxed">
                  eSign consent verified! Complete the UPI e-Mandate authorization for automated {frequency} auto-debit of ₹{(instalmentAmountPaise / 100).toLocaleString("en-IN")}.
                </p>
              </div>

              {mandateUrl && (
                <a
                  href={mandateUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="w-full py-3.5 bg-slate-950 border border-indigo-500/50 hover:border-indigo-500 text-indigo-300 font-bold text-xs rounded-xl shadow-lg flex items-center justify-center gap-2 transition-all"
                >
                  <span>Authorize Mandate on Razorpay</span>
                  <ExternalLink className="w-4 h-4" />
                </a>
              )}

              {isDev && (
                <button
                  type="button"
                  onClick={handleSimulateMandateCompletion}
                  className="w-full py-3.5 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-bold text-xs rounded-xl shadow-lg shadow-emerald-600/20 flex items-center justify-center gap-2 transition-all"
                >
                  <CheckCircle2 className="w-4 h-4" />
                  <span>[DEV] Verify Mandate &amp; Activate Order Cycle</span>
                </button>
              )}
            </div>
          )}

          {/* STEP 4: COMPLETE */}
          {step === "complete" && (
            <div className="space-y-4 text-center py-8">
              <div className="w-20 h-20 rounded-full bg-emerald-500/20 border border-emerald-500/40 flex items-center justify-center text-emerald-400 mx-auto animate-bounce">
                <CheckCircle2 className="w-10 h-10" />
              </div>
              <h4 className="text-xl font-black text-white">Order Credit Cycle Activated!</h4>
              <p className="text-xs text-slate-300 max-w-sm mx-auto">
                Surepass eSign consent received and Razorpay UPI Mandate is active. Repayments will automatically execute {frequency}.
              </p>
            </div>
          )}
        </div>

        {/* Modal Footer */}
        {step === "configure" && (
          <div className="p-6 bg-slate-950 border-t border-slate-800 flex items-center justify-between">
            <button
              type="button"
              onClick={onClose}
              className="px-5 py-3 bg-slate-800 hover:bg-slate-700 text-slate-300 font-bold text-xs rounded-xl transition-all"
            >
              Cancel
            </button>
            <button
              type="button"
              disabled={loading}
              onClick={handleInitCycle}
              className="px-6 py-3 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-bold text-xs rounded-xl shadow-lg shadow-indigo-600/30 flex items-center gap-2 transition-all disabled:opacity-50"
            >
              {loading ? <span>Initializing...</span> : <span>Proceed to Surepass eSign</span>}
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        )}
      </div>
    </div>
  );
};
