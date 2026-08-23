import React from "react";
import {
  ShoppingCart,
  CheckCircle2,
  ChevronRight,
  Sparkles,
  Package,
  ArrowRight,
} from "lucide-react";
import { ProductItem } from "@/types/onboarding";

interface Step4OrderRequirementProps {
  orderChoice: "none" | "full" | "sample";
  setOrderChoice: (val: "none" | "full" | "sample") => void;
  regularProducts: ProductItem[];
  sampleProducts: ProductItem[];
  orderQuantities: Record<string, number>;
  handleQuantityChange: (id: string, delta: number) => void;
  calculateOrderTotal: () => number;
  onOpenOrderReview: () => void;
  onInitiateSampleBooking: (item: ProductItem) => void;
}

export const Step4OrderRequirement: React.FC<Step4OrderRequirementProps> = ({
  orderChoice,
  setOrderChoice,
  regularProducts,
  sampleProducts,
  orderQuantities,
  handleQuantityChange,
  calculateOrderTotal,
  onOpenOrderReview,
  onInitiateSampleBooking,
}) => {
  // Check if any selected product violates MOQ
  const hasMoqViolation = regularProducts.some(
    (p) => (orderQuantities[p.id] || 0) > 0 && orderQuantities[p.id] < p.moq
  );

  const totalItemsCount = regularProducts.reduce(
    (acc, p) => acc + (orderQuantities[p.id] || 0),
    0
  );

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
            <ShoppingCart className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">Step 4: Order Requirement Selection</h2>
            <p className="text-xs text-slate-400">Choose to place a catalog inventory order or book a sample kit</p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 my-6">
          {/* YES PATH CARD */}
          <div
            onClick={() => setOrderChoice("full")}
            className={`p-6 rounded-2xl border cursor-pointer transition-all ${
              orderChoice === "full"
                ? "bg-indigo-600/20 border-indigo-500 text-white ring-2 ring-indigo-500/30"
                : "bg-slate-800/40 border-slate-800 text-slate-400 hover:border-slate-700"
            }`}
          >
            <div className="flex items-center justify-between mb-3">
              <span className="px-3 py-1 bg-indigo-500/20 text-indigo-300 rounded-full text-xs font-bold uppercase tracking-wider">
                Option A: Full Order
              </span>
              {orderChoice === "full" && <CheckCircle2 className="w-5 h-5 text-indigo-400" />}
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Place Catalog Order Now</h3>
            <p className="text-xs text-slate-300 leading-relaxed mb-4">
              Select items from our active product catalog. Immediate full order unlocks priority credit scoring and statutory verification.
            </p>
            <div className="text-xs text-indigo-400 font-semibold flex items-center gap-1">
              Browse Catalogue ({regularProducts.length} items) <ChevronRight className="w-4 h-4" />
            </div>
          </div>

          {/* NO PATH CARD */}
          <div
            onClick={() => setOrderChoice("sample")}
            className={`p-6 rounded-2xl border cursor-pointer transition-all ${
              orderChoice === "sample"
                ? "bg-amber-500/20 border-amber-500 text-white ring-2 ring-amber-500/30"
                : "bg-slate-800/40 border-slate-800 text-slate-400 hover:border-slate-700"
            }`}
          >
            <div className="flex items-center justify-between mb-3">
              <span className="px-3 py-1 bg-amber-500/20 text-amber-300 rounded-full text-xs font-bold uppercase tracking-wider flex items-center gap-1">
                <Sparkles className="w-3 h-3" /> Option B: Book Sample Kit
              </span>
              {orderChoice === "sample" && <CheckCircle2 className="w-5 h-5 text-amber-400" />}
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Order Trial Sample Kit</h3>
            <p className="text-xs text-slate-300 leading-relaxed mb-4">
              Test product quality first. Pay nominal sample charge via Razorpay to activate Instant Trial Status and bypass remaining credit steps.
            </p>
            <div className="text-xs text-amber-400 font-semibold flex items-center gap-1">
              Book Trial Kit <ChevronRight className="w-4 h-4" />
            </div>
          </div>
        </div>
      </div>

      {/* Render Catalog if Option A selected */}
      {orderChoice === "full" && (
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl">
          <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
            <Package className="w-5 h-5 text-indigo-400" />
            Regular Products Catalog
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            {regularProducts.map((p) => {
              const qty = orderQuantities[p.id] || 0;
              const isSubMoq = qty > 0 && qty < p.moq;

              const onPlus = () => {
                if (qty === 0) {
                  // Jump straight to MOQ when starting from 0
                  handleQuantityChange(p.id, p.moq);
                } else {
                  handleQuantityChange(p.id, 1);
                }
              };

              const onMinus = () => {
                if (qty <= p.moq) {
                  // Reset to 0 when decrementing at or below MOQ
                  handleQuantityChange(p.id, -qty);
                } else {
                  handleQuantityChange(p.id, -1);
                }
              };

              return (
                <div
                  key={p.id}
                  className={`p-4 border rounded-xl flex flex-col justify-between transition-all ${
                    isSubMoq
                      ? "bg-rose-950/20 border-rose-500/50"
                      : qty > 0
                      ? "bg-slate-800/80 border-indigo-500/50"
                      : "bg-slate-800/40 border-slate-700/60"
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="font-semibold text-white text-sm">{p.name}</div>
                      <div className="text-xs text-slate-400 mt-0.5">
                        SKU: {p.sku} |{" "}
                        <span className="font-semibold text-indigo-300">MOQ: {p.moq} units</span>
                      </div>
                      <div className="text-sm font-bold text-emerald-400 mt-1">
                        ₹{(p.price_paise / 100).toLocaleString("en-IN")}
                      </div>
                    </div>

                    <div className="flex items-center gap-2 bg-slate-900 px-3 py-1.5 rounded-xl border border-slate-700">
                      <button
                        type="button"
                        onClick={onMinus}
                        className="w-6 h-6 bg-slate-800 text-slate-300 rounded hover:bg-slate-700 font-bold flex items-center justify-center transition-colors"
                      >
                        -
                      </button>
                      <span className="w-8 text-center text-sm font-semibold">{qty}</span>
                      <button
                        type="button"
                        onClick={onPlus}
                        className="w-6 h-6 bg-indigo-600 text-white rounded hover:bg-indigo-500 font-bold flex items-center justify-center transition-colors"
                      >
                        +
                      </button>
                    </div>
                  </div>

                  {isSubMoq && (
                    <div className="mt-2 text-[11px] font-semibold text-rose-400 bg-rose-500/10 px-2.5 py-1 rounded-lg border border-rose-500/20">
                      ⚠️ Minimum order quantity for {p.name} is {p.moq}
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          <div className="flex items-center justify-between pt-4 border-t border-slate-800">
            <div>
              <span className="text-xs text-slate-400">Total Order Value:</span>
              <div className="text-2xl font-extrabold text-emerald-400">
                ₹{calculateOrderTotal().toLocaleString("en-IN")}
              </div>
            </div>

            <button
              type="button"
              onClick={onOpenOrderReview}
              disabled={totalItemsCount === 0 || hasMoqViolation}
              className="px-6 py-3 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              <span>Review & Submit Order</span>
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Render Sample Kits if Option B selected */}
      {orderChoice === "sample" && (
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl">
          <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-amber-400" />
            Available Trial Sample Kits
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
            {sampleProducts.map((sp) => (
              <div key={sp.id} className="p-5 bg-gradient-to-br from-slate-800/80 to-slate-900 border border-amber-500/30 rounded-2xl flex flex-col justify-between">
                <div>
                  <span className="px-2.5 py-0.5 bg-amber-500/20 text-amber-300 rounded-md text-[11px] font-semibold uppercase">
                    Sample Trial Kit
                  </span>
                  <h4 className="font-bold text-white text-base mt-2">{sp.name}</h4>
                  <p className="text-xs text-slate-400 mt-1">{sp.description || "Official trial bundle for distributor evaluation"}</p>
                </div>

                <div className="mt-6 flex items-center justify-between pt-4 border-t border-slate-700/50">
                  <div>
                    <div className="text-xs text-slate-400">Sample Price</div>
                    <div className="text-xl font-bold text-amber-400">
                      ₹{(sp.price_paise / 100).toLocaleString("en-IN")}
                    </div>
                  </div>

                  <button
                    type="button"
                    onClick={() => onInitiateSampleBooking(sp)}
                    className="px-4 py-2 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-400 hover:to-orange-400 text-slate-950 font-bold text-xs rounded-xl shadow-lg shadow-amber-500/20"
                  >
                    Book via Razorpay
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
