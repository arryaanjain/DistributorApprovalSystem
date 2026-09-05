import React from "react";
import { AlertTriangle, Lock, ShieldAlert, ArrowRight, RefreshCw } from "lucide-react";

interface CreditHoldBannerProps {
  status: "hold" | "defaulted" | "grace_period";
  graceRemainingDays?: number;
  onRetryMandate?: () => void;
}

export const CreditHoldBanner: React.FC<CreditHoldBannerProps> = ({
  status,
  graceRemainingDays = 7,
  onRetryMandate,
}) => {
  if (status === "grace_period") {
    return (
      <div className="p-4 sm:p-5 bg-gradient-to-r from-amber-950/90 via-amber-900/60 to-slate-900 border border-amber-500/40 rounded-2xl shadow-2xl backdrop-blur-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-start gap-3">
          <div className="w-10 h-10 rounded-2xl bg-amber-500/20 border border-amber-500/40 flex items-center justify-center text-amber-300 shrink-0">
            <AlertTriangle className="w-5 h-5 animate-bounce" />
          </div>
          <div className="space-y-1">
            <h4 className="text-sm font-bold text-amber-200 flex items-center gap-2">
              Action Required: UPI Mandate Transfer Failed
            </h4>
            <p className="text-xs text-slate-300 leading-relaxed max-w-2xl">
              An automated weekly/monthly mandate charge failed. Grace period active: <strong className="text-amber-300">{graceRemainingDays} day(s) remaining</strong>.
              Failure to resolve funds transfer will cause agreement failure and put your credit limit on <strong>CREDIT HOLD</strong>.
            </p>
          </div>
        </div>

        {onRetryMandate && (
          <button
            onClick={onRetryMandate}
            className="w-full sm:w-auto px-4 py-2.5 bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold text-xs rounded-xl shadow-lg flex items-center justify-center gap-2 shrink-0 transition-all"
          >
            <RefreshCw className="w-4 h-4" />
            <span>Re-authenticate UPI Mandate</span>
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="p-4 sm:p-5 bg-gradient-to-r from-rose-950/90 via-slate-900 to-rose-950/70 border border-rose-500/50 rounded-2xl shadow-2xl backdrop-blur-xl flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
      <div className="flex items-start gap-3">
        <div className="w-10 h-10 rounded-2xl bg-rose-500/20 border border-rose-500/40 flex items-center justify-center text-rose-300 shrink-0">
          <Lock className="w-5 h-5" />
        </div>
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <span className="px-2.5 py-0.5 bg-rose-500/20 text-rose-300 border border-rose-500/30 rounded-full text-[10px] font-bold uppercase tracking-wider">
              Account On Hold
            </span>
            <h4 className="text-sm font-bold text-white">Credit Limit & Ongoing Credit Stat Placed on HOLD</h4>
          </div>
          <p className="text-xs text-slate-300 leading-relaxed max-w-2xl">
            Agreement condition failed due to unfulfilled UPI Mandate repayment. New order creation on credit is temporarily restricted. Please clear pending order cycles to restore active credit limit.
          </p>
        </div>
      </div>

      {onRetryMandate && (
        <button
          onClick={onRetryMandate}
          className="w-full sm:w-auto px-4 py-2.5 bg-gradient-to-r from-rose-600 to-red-600 hover:from-rose-500 hover:to-red-500 text-white font-bold text-xs rounded-xl shadow-lg flex items-center justify-center gap-2 shrink-0 transition-all"
        >
          <span>Clear Hold & Pay Mandate</span>
          <ArrowRight className="w-4 h-4" />
        </button>
      )}
    </div>
  );
};
