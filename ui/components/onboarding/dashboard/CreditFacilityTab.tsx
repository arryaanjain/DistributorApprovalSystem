import React, { useState, useEffect } from "react";
import { CreditCard, ShieldCheck, ArrowRight, Sparkles, RefreshCw, Calendar } from "lucide-react";
import { AppStatus } from "@/types/onboarding";
import { RepaymentTracker, OrderCreditCycle, UPIMandate, RepaymentRecord } from "../RepaymentTracker";
import { CreditHoldBanner } from "../CreditHoldBanner";

interface CreditFacilityTabProps {
  trialActivated: boolean;
  appStatus?: AppStatus | null;
  onContinueFullOnboarding?: () => void;
  token?: string;
}

export const CreditFacilityTab: React.FC<CreditFacilityTabProps> = ({
  trialActivated,
  appStatus,
  onContinueFullOnboarding,
  token,
}) => {
  const [cycles, setCycles] = useState<OrderCreditCycle[]>([]);
  const [selectedCycleDetails, setSelectedCycleDetails] = useState<any>(null);
  const [loading, setLoading] = useState<boolean>(false);

  const creditLimitPaise = appStatus?.assigned_credit_limit;
  const isApprovedCredit =
    appStatus?.status === "credit_active" ||
    appStatus?.status === "approved" ||
    appStatus?.status === "offer_generated" ||
    appStatus?.status === "offer_accepted" ||
    appStatus?.status === "agreement_signed";

  const fetchCycles = async () => {
    const storedToken = token || localStorage.getItem("distributor_token") || localStorage.getItem("kresconet_token");
    if (!storedToken) return;
    setLoading(true);
    try {
      const res = await fetch("http://localhost:8081/api/v1/credit-cycles/me", {
        headers: { Authorization: `Bearer ${storedToken}` },
      });
      if (res.ok) {
        const resBody = await res.json();
        const data = resBody.data || resBody;
        setCycles(data.cycles || []);
        if (data.cycles && data.cycles.length > 0) {
          fetchCycleDetails(data.cycles[0].order_id);
        }
      }
    } catch (err) {
      console.error("Error fetching credit cycles:", err);
    } finally {
      setLoading(false);
    }
  };

  const fetchCycleDetails = async (orderId: string) => {
    const storedToken = token || localStorage.getItem("distributor_token") || localStorage.getItem("kresconet_token");
    if (!storedToken) return;
    try {
      const res = await fetch(`http://localhost:8081/api/v1/credit-cycles/orders/${orderId}`, {
        headers: { Authorization: `Bearer ${storedToken}` },
      });
      if (res.ok) {
        const resBody = await res.json();
        const data = resBody.data || resBody;
        setSelectedCycleDetails(data);
      }
    } catch (err) {
      console.error("Error fetching cycle details:", err);
    }
  };

  useEffect(() => {
    fetchCycles();
  }, [token]);

  const renderCreditDisplay = () => {
    if (trialActivated && !isApprovedCredit) return "Trial Status";
    if (isApprovedCredit && creditLimitPaise && creditLimitPaise > 0) {
      return `₹ ${(creditLimitPaise / 100).toLocaleString("en-IN")}`;
    }
    if (appStatus?.status === "advance_only") {
      return "Advance Only";
    }
    return "Advance / COD";
  };

  const isHold = appStatus?.status === "hold" || selectedCycleDetails?.cycle?.status === "defaulted";
  const isGraceActive = selectedCycleDetails?.is_grace_active;

  return (
    <div className="space-y-6">
      {/* Credit Hold Banner */}
      {isHold && <CreditHoldBanner status="hold" />}
      {isGraceActive && (
        <CreditHoldBanner
          status="grace_period"
          graceRemainingDays={selectedCycleDetails?.grace_remaining_days}
        />
      )}

      {/* Trial to Credit Transition Banner */}
      {trialActivated && !isApprovedCredit && onContinueFullOnboarding && (
        <div className="p-5 sm:p-6 bg-gradient-to-br from-indigo-950/80 via-slate-900 to-indigo-900/70 border border-indigo-500/40 rounded-2xl sm:rounded-3xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 shadow-2xl backdrop-blur-xl">
          <div className="space-y-1.5 max-w-2xl">
            <div className="flex items-center gap-2">
              <span className="px-2.5 py-0.5 bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded-full text-[11px] font-bold uppercase tracking-wider flex items-center gap-1">
                <Sparkles className="w-3 h-3" /> Upgrade Account
              </span>
              <h4 className="text-sm sm:text-base font-bold text-white">Interested in commercial orders on credit?</h4>
            </div>
            <p className="text-xs text-slate-300 leading-relaxed">
              Satisfied with your trial sample kit? Continue your onboarding process to complete statutory (KYC & GST) verification and unlock up to ₹5,00,000 in revolving commercial credit.
            </p>
          </div>
          <button
            type="button"
            onClick={onContinueFullOnboarding}
            className="w-full sm:w-auto px-5 py-3 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-bold text-xs rounded-xl shadow-lg shadow-indigo-600/30 flex items-center justify-center gap-2 shrink-0 transition-all touch-manipulation min-h-[44px]"
          >
            <span>Continue Onboarding for Credit</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Credit Summary Cards */}
      <div className="bg-slate-900/70 border border-slate-800 rounded-2xl sm:rounded-3xl p-4 sm:p-6 shadow-xl space-y-5">
        <div>
          <h3 className="text-base sm:text-lg font-bold text-white flex items-center gap-2">
            <CreditCard className="w-5 h-5 text-emerald-400" /> Credit Account & Sanctioned Limit
          </h3>
          <p className="text-xs text-slate-400 mt-1">Review active revolving credit terms, order cycles, and signed agreements.</p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3.5 sm:gap-4">
          <div className="p-4 sm:p-5 bg-slate-950/80 border border-slate-800 rounded-xl sm:rounded-2xl space-y-1">
            <div className="text-xs text-slate-400">Sanctioned Limit</div>
            <div className="text-lg sm:text-xl font-black text-emerald-400">{renderCreditDisplay()}</div>
          </div>

          <div className="p-4 sm:p-5 bg-slate-950/80 border border-slate-800 rounded-xl sm:rounded-2xl space-y-1">
            <div className="text-xs text-slate-400">Repayment Mode</div>
            <div className="text-lg sm:text-xl font-black text-white">Weekly / Monthly UPI</div>
          </div>

          <div className="p-4 sm:p-5 bg-slate-950/80 border border-slate-800 rounded-xl sm:rounded-2xl space-y-1">
            <div className="text-xs text-slate-400">Agreement Status</div>
            <div className="text-lg sm:text-xl font-black text-indigo-400 flex items-center gap-1.5">
              <ShieldCheck className="w-5 h-5 text-indigo-400 shrink-0" /> {isApprovedCredit ? "Active" : "Trial Active"}
            </div>
          </div>
        </div>
      </div>

      {/* Active Repayment Tracker if cycle details exist */}
      {selectedCycleDetails?.cycle && (
        <RepaymentTracker
          cycle={selectedCycleDetails.cycle}
          mandate={selectedCycleDetails.mandate}
          repayments={selectedCycleDetails.repayments || []}
          isGraceActive={selectedCycleDetails.is_grace_active}
          graceRemainingDays={selectedCycleDetails.grace_remaining_days}
          onRefresh={fetchCycles}
        />
      )}
    </div>
  );
};
