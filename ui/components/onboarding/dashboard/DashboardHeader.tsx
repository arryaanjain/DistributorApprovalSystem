import React from "react";
import { LogOut } from "lucide-react";
import { AppStatus } from "@/types/onboarding";

interface DashboardHeaderProps {
  trialActivated: boolean;
  appStatus?: AppStatus | null;
  totalOrdersCount: number;
  onSignOut?: () => void;
}

export const DashboardHeader: React.FC<DashboardHeaderProps> = ({
  trialActivated,
  appStatus,
  totalOrdersCount,
  onSignOut,
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
    <div className="bg-gradient-to-r from-slate-900 via-indigo-950/50 to-slate-900 border border-indigo-500/20 rounded-2xl sm:rounded-3xl p-5 sm:p-8 flex flex-col lg:flex-row items-start lg:items-center justify-between gap-5 sm:gap-6 shadow-2xl">
      <div className="space-y-1.5 sm:space-y-2 w-full lg:w-auto">
        <div className="flex flex-wrap items-center gap-2 sm:gap-3">
          <span className="px-3 py-1 bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 rounded-full text-[11px] sm:text-xs font-bold uppercase tracking-wider">
            {trialActivated
              ? "Active Trial Partner"
              : isApprovedCredit
              ? "Approved Credit Partner"
              : "Registered Partner"}
          </span>
          <span className="text-[11px] sm:text-xs text-slate-400 font-mono">
            ID: {appStatus?.distributor_id ? appStatus.distributor_id.slice(0, 8) : "DIST-ACTIVE"}
          </span>
        </div>
        <h2 className="text-xl sm:text-2xl lg:text-3xl font-black text-white tracking-tight">
          Kresconet Distributor Portal
        </h2>
        <p className="text-xs text-slate-400 max-w-2xl leading-relaxed">
          Manage your revolving commercial credit line, catalogue dispatches, and real-time trial kit logistics.
        </p>
      </div>

      <div className="grid grid-cols-2 sm:flex sm:flex-wrap items-center gap-3 w-full lg:w-auto pt-2 lg:pt-0 border-t lg:border-t-0 border-slate-800/80">
        <div className="bg-slate-950/80 p-3.5 sm:px-5 sm:py-3.5 rounded-xl sm:rounded-2xl border border-slate-800 text-center min-w-[120px]">
          <div className="text-[10px] sm:text-[11px] font-medium text-slate-400 uppercase tracking-wider">
            Available Credit
          </div>
          <div className="text-base sm:text-lg font-black text-emerald-400 mt-0.5">
            {renderCreditDisplay()}
          </div>
        </div>

        <div className="bg-slate-950/80 p-3.5 sm:px-5 sm:py-3.5 rounded-xl sm:rounded-2xl border border-slate-800 text-center min-w-[100px]">
          <div className="text-[10px] sm:text-[11px] font-medium text-slate-400 uppercase tracking-wider">
            Total Orders
          </div>
          <div className="text-base sm:text-lg font-black text-white mt-0.5">
            {totalOrdersCount}
          </div>
        </div>

        {onSignOut && (
          <button
            onClick={onSignOut}
            type="button"
            className="col-span-2 sm:col-span-1 w-full sm:w-auto px-4 py-3 bg-rose-500/10 hover:bg-rose-500/20 active:bg-rose-500/30 text-rose-400 border border-rose-500/30 rounded-xl sm:rounded-2xl text-xs font-bold transition-all flex items-center justify-center gap-2 shadow-lg shadow-rose-500/5 touch-manipulation min-h-[44px]"
          >
            <LogOut className="w-4 h-4" />
            <span>Sign Out</span>
          </button>
        )}
      </div>
    </div>
  );
};
