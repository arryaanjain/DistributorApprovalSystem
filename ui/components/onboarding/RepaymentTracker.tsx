import React from "react";
import {
  Calendar,
  CheckCircle2,
  Clock,
  AlertTriangle,
  RefreshCw,
  ArrowUpRight,
  ShieldCheck,
} from "lucide-react";

export interface RepaymentRecord {
  id: string;
  instalment_number: number;
  due_date: string;
  amount_paise: number;
  status: "scheduled" | "pending" | "paid" | "failed" | "bounced";
  razorpay_payment_id?: string;
  paid_at?: string;
  failure_reason?: string;
}

export interface OrderCreditCycle {
  id: string;
  order_id: string;
  total_amount_paise: number;
  credit_amount_paise: number;
  instalment_amount_paise: number;
  frequency: "weekly" | "monthly";
  total_instalments: number;
  completed_instalments: number;
  status: string;
  grace_period_days: number;
  grace_period_ends_at?: string;
}

export interface UPIMandate {
  id: string;
  razorpay_subscription_id: string;
  mandate_status: string;
  razorpay_short_url?: string;
}

interface RepaymentTrackerProps {
  cycle: OrderCreditCycle;
  mandate?: UPIMandate | null;
  repayments: RepaymentRecord[];
  isGraceActive?: boolean;
  graceRemainingDays?: number;
  onRefresh?: () => void;
}

export const RepaymentTracker: React.FC<RepaymentTrackerProps> = ({
  cycle,
  mandate,
  repayments,
  isGraceActive,
  graceRemainingDays,
  onRefresh,
}) => {
  const progressPct = Math.round((cycle.completed_instalments / cycle.total_instalments) * 100);

  // Deduplicate repayments by instalment_number
  const uniqueRepayments = Array.from(
    new Map(repayments.map((r) => [r.instalment_number, r])).values()
  ).sort((a, b) => a.instalment_number - b.instalment_number);

  return (
    <div className="bg-slate-900/90 border border-slate-800 rounded-3xl p-5 sm:p-6 shadow-2xl space-y-6">
      {/* Header & Status Pill */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-5 border-b border-slate-800/80">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <span className="px-2.5 py-0.5 bg-indigo-500/10 text-indigo-400 border border-indigo-500/30 rounded-full text-[10px] font-bold uppercase tracking-wider">
              {cycle.frequency} Cycle
            </span>
            {isGraceActive && (
              <span className="px-2.5 py-0.5 bg-amber-500/20 text-amber-300 border border-amber-500/40 rounded-full text-[10px] font-bold uppercase tracking-wider flex items-center gap-1 animate-pulse">
                <AlertTriangle className="w-3 h-3" /> {graceRemainingDays}d Grace Period Active
              </span>
            )}
          </div>
          <h4 className="text-base sm:text-lg font-bold text-white flex items-center gap-2">
            Order Credit Repayment Tracker
          </h4>
          <p className="text-xs text-slate-400">
            Automated UPI Mandate: {mandate?.mandate_status || "Active"} ({mandate?.razorpay_subscription_id || "Razorpay Sub"})
          </p>
        </div>

        {onRefresh && (
          <button
            onClick={onRefresh}
            className="self-start sm:self-auto p-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl transition-colors flex items-center gap-1.5 text-xs font-semibold"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            <span>Sync Status</span>
          </button>
        )}
      </div>

      {/* Progress Bar & Summary */}
      <div className="p-4 bg-slate-950/70 border border-slate-800/80 rounded-2xl space-y-3">
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-400">Repayment Completion</span>
          <span className="font-bold text-emerald-400">
            {cycle.completed_instalments} of {cycle.total_instalments} Paid ({progressPct}%)
          </span>
        </div>
        <div className="w-full h-2.5 bg-slate-800 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-emerald-500 to-teal-400 rounded-full transition-all duration-500"
            style={{ width: `${progressPct}%` }}
          />
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-2 text-xs">
          <div>
            <span className="text-slate-400 block text-[11px]">Total Financed</span>
            <span className="font-bold text-white">₹{(cycle.credit_amount_paise / 100).toLocaleString("en-IN")}</span>
          </div>
          <div>
            <span className="text-slate-400 block text-[11px]">Instalment Size</span>
            <span className="font-bold text-emerald-400">
              ₹{(cycle.instalment_amount_paise / 100).toLocaleString("en-IN")}
            </span>
          </div>
          <div>
            <span className="text-slate-400 block text-[11px]">Mandate Status</span>
            <span className="font-bold text-indigo-400 capitalize">{mandate?.mandate_status || "Active"}</span>
          </div>
          <div>
            <span className="text-slate-400 block text-[11px]">Cycle Status</span>
            <span className="font-bold text-white capitalize">{cycle.status}</span>
          </div>
        </div>
      </div>

      {/* Repayment Schedule Table */}
      <div className="space-y-3">
        <h5 className="text-xs font-bold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
          <Calendar className="w-3.5 h-3.5 text-indigo-400" /> Repayment Schedule
        </h5>
        <div className="border border-slate-800 rounded-2xl overflow-hidden">
          <table className="w-full text-left border-collapse text-xs">
            <thead>
              <tr className="bg-slate-950/80 border-b border-slate-800 text-slate-400 font-semibold">
                <th className="py-3 px-4">#</th>
                <th className="py-3 px-4">Due Date</th>
                <th className="py-3 px-4">Amount</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4 text-right">Reference</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 bg-slate-900/40">
              {uniqueRepayments.map((rep) => {
                const dueDateFormatted = new Date(rep.due_date).toLocaleDateString("en-IN", {
                  day: "2-digit",
                  month: "short",
                  year: "numeric",
                });
                return (
                  <tr key={rep.id} className="hover:bg-slate-800/30 transition-colors">
                    <td className="py-3 px-4 font-semibold text-slate-300">{rep.instalment_number}</td>
                    <td className="py-3 px-4 text-slate-300 font-medium">{dueDateFormatted}</td>
                    <td className="py-3 px-4 font-bold text-white">
                      ₹{(rep.amount_paise / 100).toLocaleString("en-IN")}
                    </td>
                    <td className="py-3 px-4">
                      {rep.status === "paid" && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 font-semibold text-[11px]">
                          <CheckCircle2 className="w-3 h-3" /> Paid
                        </span>
                      )}
                      {rep.status === "scheduled" && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-300 border border-indigo-500/30 font-medium text-[11px]">
                          <Clock className="w-3 h-3" /> Scheduled
                        </span>
                      )}
                      {(rep.status === "failed" || rep.status === "bounced") && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-rose-500/10 text-rose-400 border border-rose-500/30 font-semibold text-[11px]">
                          <AlertTriangle className="w-3 h-3" /> Failed
                        </span>
                      )}
                    </td>
                    <td className="py-3 px-4 text-right font-mono text-[11px] text-slate-400">
                      {rep.razorpay_payment_id || "Auto-Debit Pending"}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
