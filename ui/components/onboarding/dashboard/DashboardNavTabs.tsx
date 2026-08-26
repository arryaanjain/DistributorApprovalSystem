import React from "react";
import { Truck, ShoppingBag, CreditCard, MapPin } from "lucide-react";

export type DashboardTabType = "tracking" | "catalogue" | "credit" | "address";

interface DashboardNavTabsProps {
  activeTab: DashboardTabType;
  setActiveTab: (tab: DashboardTabType) => void;
}

export const DashboardNavTabs: React.FC<DashboardNavTabsProps> = ({
  activeTab,
  setActiveTab,
}) => {
  return (
    <div className="flex items-center gap-1.5 p-1.5 bg-slate-900/90 border border-slate-800 rounded-2xl overflow-x-auto scrollbar-none touch-pan-x -mx-1 px-2 sm:mx-0">
      <button
        onClick={() => setActiveTab("tracking")}
        className={`flex items-center gap-2 px-3.5 sm:px-5 py-2.5 rounded-xl text-xs font-bold transition-all whitespace-nowrap min-h-[40px] touch-manipulation ${
          activeTab === "tracking"
            ? "bg-indigo-600 text-white shadow-lg shadow-indigo-600/30"
            : "text-slate-400 hover:text-slate-200 active:bg-slate-800/60"
        }`}
      >
        <Truck className="w-4 h-4 shrink-0" />
        <span>Live Order Tracking</span>
      </button>

      <button
        onClick={() => setActiveTab("catalogue")}
        className={`flex items-center gap-2 px-3.5 sm:px-5 py-2.5 rounded-xl text-xs font-bold transition-all whitespace-nowrap min-h-[40px] touch-manipulation ${
          activeTab === "catalogue"
            ? "bg-indigo-600 text-white shadow-lg shadow-indigo-600/30"
            : "text-slate-400 hover:text-slate-200 active:bg-slate-800/60"
        }`}
      >
        <ShoppingBag className="w-4 h-4 shrink-0" />
        <span>Catalogue</span>
      </button>

      <button
        onClick={() => setActiveTab("credit")}
        className={`flex items-center gap-2 px-3.5 sm:px-5 py-2.5 rounded-xl text-xs font-bold transition-all whitespace-nowrap min-h-[40px] touch-manipulation ${
          activeTab === "credit"
            ? "bg-indigo-600 text-white shadow-lg shadow-indigo-600/30"
            : "text-slate-400 hover:text-slate-200 active:bg-slate-800/60"
        }`}
      >
        <CreditCard className="w-4 h-4 shrink-0" />
        <span>Credit Facility</span>
      </button>

      <button
        onClick={() => setActiveTab("address")}
        className={`flex items-center gap-2 px-3.5 sm:px-5 py-2.5 rounded-xl text-xs font-bold transition-all whitespace-nowrap min-h-[40px] touch-manipulation ${
          activeTab === "address"
            ? "bg-indigo-600 text-white shadow-lg shadow-indigo-600/30"
            : "text-slate-400 hover:text-slate-200 active:bg-slate-800/60"
        }`}
      >
        <MapPin className="w-4 h-4 shrink-0" />
        <span>Addresses</span>
      </button>
    </div>
  );
};
