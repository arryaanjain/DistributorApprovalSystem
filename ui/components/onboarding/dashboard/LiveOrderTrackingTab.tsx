import React, { useState } from "react";
import { Truck, RefreshCw, Search, Package, Sparkles, ShoppingBag, CheckCircle2, Clock, Eye } from "lucide-react";

interface LiveOrderTrackingTabProps {
  sampleOrders: any[];
  catalogOrders: any[];
  loadingOrders: boolean;
  onRefresh: () => void;
  onSelectOrder: (order: any, type: "sample" | "commercial") => void;
}

export const LiveOrderTrackingTab: React.FC<LiveOrderTrackingTabProps> = ({
  sampleOrders,
  catalogOrders,
  loadingOrders,
  onRefresh,
  onSelectOrder,
}) => {
  const [orderFilter, setOrderFilter] = useState<"all" | "sample" | "commercial">("all");
  const [searchQuery, setSearchQuery] = useState<string>("");

  const totalOrdersCount = catalogOrders.length + sampleOrders.length;

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "DISPATCHED":
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-[11px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shadow-sm">
            <Truck className="w-3.5 h-3.5" /> Dispatched
          </span>
        );
      case "PAID":
      case "APPROVED":
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-[11px] font-bold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 shadow-sm">
            <CheckCircle2 className="w-3.5 h-3.5" /> {status === "PAID" ? "Confirmed" : "Approved"}
          </span>
        );
      case "PENDING_REVIEW":
      case "CREATED":
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-[11px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20 shadow-sm">
            <Clock className="w-3.5 h-3.5" /> Processing
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[11px] font-semibold bg-slate-800 text-slate-300">
            {status}
          </span>
        );
    }
  };

  const combinedOrders = [
    ...sampleOrders.map((s) => ({ ...s, _type: "sample" })),
    ...catalogOrders.map((c) => ({ ...c, _type: "commercial" })),
  ].filter((item) => {
    if (orderFilter === "sample" && item._type !== "sample") return false;
    if (orderFilter === "commercial" && item._type !== "commercial") return false;
    if (searchQuery.trim() !== "") {
      const q = searchQuery.toLowerCase();
      const ref = (item.order_number || item.id || item.razorpay_order_id || "").toLowerCase();
      return ref.includes(q);
    }
    return true;
  });

  return (
    <div className="bg-slate-900/70 border border-slate-800 rounded-2xl sm:rounded-3xl p-4 sm:p-6 shadow-xl space-y-5">
      {/* Header section */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 border-b border-slate-800/80 pb-4">
        <div>
          <h3 className="text-base sm:text-lg font-bold text-white flex items-center gap-2">
            <Truck className="w-5 h-5 text-indigo-400" /> Live Order Tracking Pipeline
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Track current status, payment verification, and Shiprocket warehouse dispatch for all placed orders.
          </p>
        </div>

        <button
          onClick={onRefresh}
          type="button"
          className="w-full sm:w-auto p-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 active:bg-slate-600 text-slate-300 text-xs font-semibold border border-slate-700 flex items-center justify-center gap-1.5 transition-all min-h-[40px] touch-manipulation"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loadingOrders ? "animate-spin" : ""}`} />
          <span>Sync Orders</span>
        </button>
      </div>

      {/* Sub-Filters and Search Bar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-xl border border-slate-800 overflow-x-auto scrollbar-none">
          <button
            onClick={() => setOrderFilter("all")}
            type="button"
            className={`flex-1 sm:flex-none px-3 py-1.5 rounded-lg text-xs font-bold transition-all min-h-[36px] touch-manipulation ${
              orderFilter === "all" ? "bg-slate-800 text-white" : "text-slate-400 hover:text-slate-200"
            }`}
          >
            All ({totalOrdersCount})
          </button>
          <button
            onClick={() => setOrderFilter("sample")}
            type="button"
            className={`flex-1 sm:flex-none px-3 py-1.5 rounded-lg text-xs font-bold transition-all min-h-[36px] touch-manipulation ${
              orderFilter === "sample" ? "bg-slate-800 text-amber-400" : "text-slate-400 hover:text-slate-200"
            }`}
          >
            Sample Kits ({sampleOrders.length})
          </button>
          <button
            onClick={() => setOrderFilter("commercial")}
            type="button"
            className={`flex-1 sm:flex-none px-3 py-1.5 rounded-lg text-xs font-bold transition-all min-h-[36px] touch-manipulation ${
              orderFilter === "commercial" ? "bg-slate-800 text-indigo-400" : "text-slate-400 hover:text-slate-200"
            }`}
          >
            Catalogue ({catalogOrders.length})
          </button>
        </div>

        <div className="relative w-full sm:w-64">
          <Search className="w-3.5 h-3.5 absolute left-3 top-3 text-slate-500" />
          <input
            type="text"
            placeholder="Search order ref..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-xl pl-9 pr-4 py-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-colors"
          />
        </div>
      </div>

      {/* Order Cards List */}
      {loadingOrders ? (
        <div className="py-12 text-center text-xs text-slate-500">Loading order tracking data...</div>
      ) : combinedOrders.length === 0 ? (
        <div className="py-10 text-center text-xs text-slate-500 space-y-2 border border-dashed border-slate-800 rounded-2xl p-4">
          <Package className="w-8 h-8 mx-auto text-slate-600" />
          <p className="font-semibold text-slate-400">No matching orders found.</p>
          <p className="text-slate-600 text-[11px]">Placed orders will show detailed status updates here.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {combinedOrders.map((item) => {
            const isSample = item._type === "sample";
            const refNum = isSample ? `#SMP-${item.id.slice(0, 8)}` : `#${item.order_number || item.id.slice(0, 8)}`;
            const totalAmt = isSample ? item.amount_paise : item.total_amount_paise;

            return (
              <div
                key={item.id}
                className="p-3.5 sm:p-4 rounded-xl sm:rounded-2xl bg-slate-950/80 border border-slate-800/90 hover:border-slate-700 transition-all flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3.5 group"
              >
                <div className="flex items-start sm:items-center gap-3 w-full sm:w-auto">
                  <div
                    className={`w-10 h-10 sm:w-11 sm:h-11 rounded-xl sm:rounded-2xl flex items-center justify-center shrink-0 border ${
                      isSample
                        ? "bg-amber-500/10 border-amber-500/20 text-amber-400"
                        : "bg-indigo-500/10 border-indigo-500/20 text-indigo-400"
                    }`}
                  >
                    {isSample ? <Sparkles className="w-4 h-4 sm:w-5 sm:h-5" /> : <ShoppingBag className="w-4 h-4 sm:w-5 sm:h-5" />}
                  </div>

                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-xs font-bold text-white truncate">
                        {isSample ? "Sample Kit Trial Order" : "Commercial Order"}
                      </span>
                      <span className="text-[11px] font-mono font-bold text-indigo-300">{refNum}</span>
                    </div>
                    <div className="text-[11px] text-slate-400 mt-0.5 flex flex-wrap items-center gap-2 sm:gap-3">
                      <span>Date: {new Date(item.created_at).toLocaleDateString("en-IN")}</span>
                      {!isSample && item.credit_used_paise > 0 && (
                        <span className="text-emerald-400 font-semibold">
                          Credit: ₹{(item.credit_used_paise / 100).toLocaleString("en-IN")}
                        </span>
                      )}
                      {isSample && item.razorpay_order_id && (
                        <span className="font-mono text-slate-400 hidden sm:inline">RP: {item.razorpay_order_id}</span>
                      )}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-3 w-full sm:w-auto justify-between sm:justify-end border-t sm:border-t-0 border-slate-800/80 pt-2.5 sm:pt-0">
                  <div className="text-left sm:text-right">
                    <div className="text-xs sm:text-sm font-black text-white">
                      ₹{(totalAmt / 100).toLocaleString("en-IN")}
                    </div>
                    <div className="text-[10px] text-slate-500 font-medium">Incl. Taxes</div>
                  </div>

                  {getStatusBadge(item.status)}

                  <button
                    onClick={() => onSelectOrder(item, isSample ? "sample" : "commercial")}
                    type="button"
                    className="px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 active:bg-slate-600 text-slate-300 hover:text-white transition-all border border-slate-700 text-xs font-bold flex items-center gap-1 shrink-0 touch-manipulation min-h-[38px]"
                  >
                    <Eye className="w-3.5 h-3.5 text-indigo-400" />
                    <span>Track</span>
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
