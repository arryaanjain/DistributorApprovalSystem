import React, { useState } from "react";
import {
  ShoppingCart,
  Trash2,
  Plus,
  Minus,
  Calendar,
  Clock,
  ArrowRight,
  Sparkles,
  CreditCard,
  ShieldCheck,
  X,
  ChevronUp,
  ChevronDown,
} from "lucide-react";
import { ProductItem } from "@/types/onboarding";

export interface CartItem {
  product: ProductItem;
  quantity: number;
}

interface WholesaleCartProps {
  cartItems: CartItem[];
  onUpdateQuantity: (productId: string, quantity: number) => void;
  onRemoveItem: (productId: string) => void;
  onClearCart: () => void;
  onCheckout: () => void;
  loading: boolean;
  availableCreditPaise?: number;
}

export const WholesaleCart: React.FC<WholesaleCartProps> = ({
  cartItems,
  onUpdateQuantity,
  onRemoveItem,
  onClearCart,
  onCheckout,
  loading,
  availableCreditPaise = 50000000,
}) => {
  const [selectedPreviewFreq, setSelectedPreviewFreq] = useState<"weekly" | "monthly">("weekly");
  const [isExpanded, setIsExpanded] = useState<boolean>(true);

  if (cartItems.length === 0) return null;

  const totalPaise = cartItems.reduce(
    (acc, item) => acc + item.product.price_paise * item.quantity,
    0
  );

  const totalItemsCount = cartItems.reduce((acc, item) => acc + item.quantity, 0);

  // Repayment Calculations
  const weeklyInstalment5Paise = Math.ceil(totalPaise / 5);
  const weeklyInstalment3Paise = Math.ceil(totalPaise / 3);
  const monthlyInstalment2Paise = Math.ceil(totalPaise / 2);

  return (
    <div className="fixed bottom-4 right-4 sm:right-6 z-40 w-full max-w-lg shadow-2xl">
      <div className="bg-slate-900/95 border border-indigo-500/40 rounded-3xl backdrop-blur-xl overflow-hidden shadow-2xl transition-all">
        {/* Cart Header Bar */}
        <div
          onClick={() => setIsExpanded(!isExpanded)}
          className="p-4 bg-gradient-to-r from-indigo-950 via-slate-900 to-slate-950 border-b border-slate-800 flex items-center justify-between cursor-pointer select-none"
        >
          <div className="flex items-center gap-3">
            <div className="relative">
              <div className="w-10 h-10 rounded-2xl bg-indigo-600/20 border border-indigo-500/40 flex items-center justify-center text-indigo-400">
                <ShoppingCart className="w-5 h-5" />
              </div>
              <span className="absolute -top-1 -right-1 w-5 h-5 rounded-full bg-indigo-500 text-white font-bold text-[10px] flex items-center justify-center border border-slate-900">
                {totalItemsCount}
              </span>
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h4 className="font-bold text-white text-sm sm:text-base">Wholesale Order Cart</h4>
                <span className="text-[10px] font-bold text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20">
                  {cartItems.length} SKU(s)
                </span>
              </div>
              <p className="text-xs text-slate-400">Total: ₹{(totalPaise / 100).toLocaleString("en-IN")}</p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              className="p-1.5 rounded-xl bg-slate-800/80 hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
            >
              {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronUp className="w-4 h-4" />}
            </button>
          </div>
        </div>

        {/* Collapsible Cart Body */}
        {isExpanded && (
          <div className="p-4 sm:p-5 space-y-4 max-h-[65vh] overflow-y-auto">
            {/* Cart Items List */}
            <div className="space-y-2.5">
              <div className="flex items-center justify-between text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                <span>Selected Products</span>
                <button
                  type="button"
                  onClick={onClearCart}
                  className="text-rose-400 hover:text-rose-300 transition-colors flex items-center gap-1"
                >
                  <Trash2 className="w-3 h-3" /> Clear Cart
                </button>
              </div>

              <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
                {cartItems.map((item) => {
                  const itemTotalPaise = item.product.price_paise * item.quantity;
                  const moq = item.product.moq || 1;

                  return (
                    <div
                      key={item.product.id}
                      className="p-3 bg-slate-950/80 border border-slate-800/80 rounded-2xl flex items-center justify-between gap-3 text-xs"
                    >
                      <div className="space-y-0.5 flex-1 min-w-0">
                        <h5 className="font-bold text-white truncate">{item.product.name}</h5>
                        <p className="text-[11px] text-slate-400">
                          ₹{(item.product.price_paise / 100).toLocaleString("en-IN")} / {item.product.unit || "case"}
                        </p>
                      </div>

                      {/* Quantity Selector */}
                      <div className="flex items-center gap-1 bg-slate-900 border border-slate-800 rounded-xl p-1 shrink-0">
                        <button
                          type="button"
                          onClick={() =>
                            onUpdateQuantity(item.product.id, item.quantity - 1)
                          }
                          className="w-6 h-6 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 flex items-center justify-center transition-colors"
                        >
                          <Minus className="w-3 h-3" />
                        </button>
                        <span className="w-8 text-center font-bold text-white text-xs">{item.quantity}</span>
                        <button
                          type="button"
                          onClick={() =>
                            onUpdateQuantity(item.product.id, item.quantity + 1)
                          }
                          className="w-6 h-6 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 flex items-center justify-center transition-colors"
                        >
                          <Plus className="w-3 h-3" />
                        </button>
                      </div>

                      <div className="text-right shrink-0">
                        <div className="font-bold text-emerald-400">
                          ₹{(itemTotalPaise / 100).toLocaleString("en-IN")}
                        </div>
                        <button
                          type="button"
                          onClick={() => onRemoveItem(item.product.id)}
                          className="text-slate-500 hover:text-rose-400 transition-colors text-[10px]"
                        >
                          Remove
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Repayment Bifurcation Comparison & Difference Display */}
            <div className="p-4 bg-gradient-to-br from-slate-950 via-slate-900 to-indigo-950/60 border border-indigo-500/30 rounded-2xl space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-white flex items-center gap-1.5">
                  <Sparkles className="w-4 h-4 text-indigo-400" /> Order Credit Repayment Bifurcation
                </span>
                <span className="text-[10px] text-indigo-300 bg-indigo-500/20 px-2 py-0.5 rounded-full border border-indigo-500/30">
                  Razorpay UPI Subscriptions
                </span>
              </div>

              {/* Weekly vs Monthly Toggle */}
              <div className="grid grid-cols-2 gap-2 p-1 bg-slate-950/80 border border-slate-800 rounded-xl text-xs">
                <button
                  type="button"
                  onClick={() => setSelectedPreviewFreq("weekly")}
                  className={`py-1.5 rounded-lg font-bold transition-all flex items-center justify-center gap-1.5 ${
                    selectedPreviewFreq === "weekly"
                      ? "bg-indigo-600 text-white shadow-md"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <Calendar className="w-3.5 h-3.5" />
                  <span>Weekly Plan</span>
                </button>
                <button
                  type="button"
                  onClick={() => setSelectedPreviewFreq("monthly")}
                  className={`py-1.5 rounded-lg font-bold transition-all flex items-center justify-center gap-1.5 ${
                    selectedPreviewFreq === "monthly"
                      ? "bg-indigo-600 text-white shadow-md"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <Clock className="w-3.5 h-3.5" />
                  <span>Monthly Plan</span>
                </button>
              </div>

              {/* Detailed Difference Comparison Breakdown */}
              {selectedPreviewFreq === "weekly" ? (
                <div className="space-y-2 text-xs pt-1">
                  <div className="p-3 bg-slate-900/80 border border-slate-800 rounded-xl flex items-center justify-between">
                    <div>
                      <span className="font-bold text-white block">5 Weekly Instalments</span>
                      <span className="text-[11px] text-slate-400">Auto-debit every 7 days</span>
                    </div>
                    <div className="text-right">
                      <span className="text-sm font-black text-emerald-400">
                        ₹{(weeklyInstalment5Paise / 100).toLocaleString("en-IN")}
                      </span>
                      <span className="text-[10px] text-slate-400 block">/ week</span>
                    </div>
                  </div>
                  <div className="flex justify-between text-[11px] text-slate-400 px-1">
                    <span>Alternative (3 Wks Option):</span>
                    <span className="font-semibold text-slate-300">₹{(weeklyInstalment3Paise / 100).toLocaleString("en-IN")} / week</span>
                  </div>
                </div>
              ) : (
                <div className="space-y-2 text-xs pt-1">
                  <div className="p-3 bg-slate-900/80 border border-slate-800 rounded-xl flex items-center justify-between">
                    <div>
                      <span className="font-bold text-white block">2 Monthly Instalments</span>
                      <span className="text-[11px] text-slate-400">Auto-debit every 30 days</span>
                    </div>
                    <div className="text-right">
                      <span className="text-sm font-black text-emerald-400">
                        ₹{(monthlyInstalment2Paise / 100).toLocaleString("en-IN")}
                      </span>
                      <span className="text-[10px] text-slate-400 block">/ month</span>
                    </div>
                  </div>
                  <div className="flex justify-between text-[11px] text-slate-400 px-1">
                    <span>Difference vs 5-Wk Plan:</span>
                    <span className="font-semibold text-indigo-300">
                      ₹{((monthlyInstalment2Paise - weeklyInstalment5Paise) / 100).toLocaleString("en-IN")} larger per payment
                    </span>
                  </div>
                </div>
              )}
            </div>

            {/* Total Order Amount & Checkout Trigger */}
            <div className="pt-2 space-y-3">
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Total Order Amount</span>
                <span className="text-lg font-black text-white">
                  ₹{(totalPaise / 100).toLocaleString("en-IN")}
                </span>
              </div>

              <button
                type="button"
                disabled={loading}
                onClick={onCheckout}
                className="w-full py-3.5 bg-gradient-to-r from-indigo-600 via-violet-600 to-indigo-600 hover:from-indigo-500 hover:to-violet-500 disabled:opacity-50 text-white font-bold text-xs rounded-2xl shadow-xl shadow-indigo-600/30 flex items-center justify-center gap-2 transition-all min-h-[46px] touch-manipulation"
              >
                {loading ? (
                  <span>Creating Order...</span>
                ) : (
                  <>
                    <ShieldCheck className="w-4 h-4 text-emerald-400" />
                    <span>Checkout & Setup Credit Cycle</span>
                    <ArrowRight className="w-4 h-4" />
                  </>
                )}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
