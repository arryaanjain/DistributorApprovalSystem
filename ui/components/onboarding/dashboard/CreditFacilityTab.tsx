import React from "react";
import { CreditCard, ShieldCheck } from "lucide-react";
import { AppStatus } from "@/types/onboarding";

interface CreditFacilityTabProps {
  trialActivated: boolean;
  appStatus?: AppStatus | null;
}

export const CreditFacilityTab: React.FC<CreditFacilityTabProps> = ({
  trialActivated,
  appStatus,
}) => {
  const creditLimitPaise = appStatus?.assigned_credit_limit;
  const isApprovedCredit =
    appStatus?.status === "credit_active" ||
    appStatus?.status === "approved" ||
    appStatus?.status === "offer_generated" ||
    appStatus?.status === "offer_accepted" ||
    appStatus?.status === "agreement_signed";

  const renderCreditDisplay = () => {
    if (trialActivated) return "Trial Status";
    if (isApprovedCredit && creditLimitPaise && creditLimitPaise > 0) {
      return `₹ ${(creditLimitPaise / 100).toLocaleString("en-IN")}`;
    }
    if (appStatus?.status === "advance_only") {
      return "Advance Only";
    }
    return "Advance / COD";
  };

  return (
    <div className="bg-slate-900/70 border border-slate-800 rounded-2xl sm:rounded-3xl p-4 sm:p-6 shadow-xl space-y-5">
      <div>
        <h3 className="text-base sm:text-lg font-bold text-white flex items-center gap-2">
          <CreditCard className="w-5 h-5 text-emerald-400" /> Credit Account & Sanctioned Limit
        </h3>
        <p className="text-xs text-slate-400 mt-1">Review active revolving credit terms and signed agreements.</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3.5 sm:gap-4">
        <div className="p-4 sm:p-5 bg-slate-950/80 border border-slate-800 rounded-xl sm:rounded-2xl space-y-1">
          <div className="text-xs text-slate-400">Sanctioned Limit</div>
          <div className="text-lg sm:text-xl font-black text-emerald-400">{renderCreditDisplay()}</div>
        </div>

        <div className="p-4 sm:p-5 bg-slate-950/80 border border-slate-800 rounded-xl sm:rounded-2xl space-y-1">
          <div className="text-xs text-slate-400">Interest-Free Period</div>
          <div className="text-lg sm:text-xl font-black text-white">30 Days</div>
        </div>

        <div className="p-4 sm:p-5 bg-slate-950/80 border border-slate-800 rounded-xl sm:rounded-2xl space-y-1">
          <div className="text-xs text-slate-400">Agreement Status</div>
          <div className="text-lg sm:text-xl font-black text-indigo-400 flex items-center gap-1.5">
            <ShieldCheck className="w-5 h-5 text-indigo-400 shrink-0" /> Active
          </div>
        </div>
      </div>
    </div>
  );
};
