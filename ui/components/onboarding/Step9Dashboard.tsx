import React from "react";
import { ProductItem, AppStatus } from "@/types/onboarding";

interface Step9DashboardProps {
  trialActivated: boolean;
  regularProducts: ProductItem[];
  appStatus?: AppStatus | null;
}

export const Step9Dashboard: React.FC<Step9DashboardProps> = ({
  trialActivated,
  regularProducts,
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
    <div className="space-y-8">
      <div className="bg-gradient-to-r from-slate-900 via-indigo-950/40 to-slate-900 border border-indigo-500/20 rounded-3xl p-8 flex flex-col md:flex-row items-center justify-between gap-6 shadow-2xl">
        <div>
          <span className="px-3 py-1 bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 rounded-full text-xs font-semibold uppercase">
            {trialActivated ? "Active Trial Partner" : isApprovedCredit ? "Approved Credit Partner" : "Registered Partner"}
          </span>
          <h2 className="text-2xl font-bold text-white mt-2">Welcome to Kresconet Portal</h2>
          <p className="text-xs text-slate-400 mt-1">
            Manage inventory orders, track credit lines, and review sample shipments.
          </p>
        </div>

        <div className="flex gap-4">
          <div className="bg-slate-900/80 px-6 py-4 rounded-2xl border border-slate-800 text-center">
            <div className="text-xs text-slate-400">Available Credit</div>
            <div className="text-xl font-extrabold text-emerald-400 mt-1">
              {renderCreditDisplay()}
            </div>
          </div>

          <div className="bg-slate-900/80 px-6 py-4 rounded-2xl border border-slate-800 text-center">
            <div className="text-xs text-slate-400">Active Orders</div>
            <div className="text-xl font-extrabold text-white mt-1">1</div>
          </div>
        </div>
      </div>

      {/* Dashboard Catalogue Section */}
      <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 shadow-xl">
        <h3 className="text-lg font-bold text-white mb-4">Quick Order Catalogue</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {regularProducts.slice(0, 6).map((p) => (
            <div key={p.id} className="p-4 bg-slate-800/40 border border-slate-700/60 rounded-xl">
              <div className="font-semibold text-white text-sm">{p.name}</div>
              <div className="text-xs text-slate-400 mt-1">Category: {p.category}</div>
              <div className="text-sm font-bold text-emerald-400 mt-3">
                ₹{(p.price_paise / 100).toLocaleString("en-IN")}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
