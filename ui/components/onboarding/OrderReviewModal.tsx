import React from "react";
import { AlertCircle, AlertTriangle } from "lucide-react";
import { ProductItem } from "@/types/onboarding";

interface OrderReviewModalProps {
  regularProducts: ProductItem[];
  orderQuantities: Record<string, number>;
  calculateOrderTotal: () => number;
  loading: boolean;
  errorMsg: string | null;
  onClose: () => void;
  onConfirm: () => void;
}

export const OrderReviewModal: React.FC<OrderReviewModalProps> = ({
  regularProducts,
  orderQuantities,
  calculateOrderTotal,
  loading,
  errorMsg,
  onClose,
  onConfirm,
}) => {
  const subMoqItems = regularProducts.filter(
    (p) => (orderQuantities[p.id] || 0) > 0 && orderQuantities[p.id] < p.moq
  );

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-slate-900 border border-slate-700 rounded-2xl p-6 shadow-2xl">
        <h3 className="text-lg font-bold text-white mb-2">Review Order Details</h3>
        <p className="text-xs text-slate-400 mb-4">
          Confirm your selected items before proceeding to KYC & statutory verification.
        </p>

        {/* Modal Backend Error Banner */}
        {errorMsg && (
          <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/30 rounded-xl text-rose-400 text-xs flex items-center gap-2">
            <AlertCircle className="w-4 h-4 flex-shrink-0" />
            <span>{errorMsg}</span>
          </div>
        )}

        {/* Modal MOQ Violation Warning */}
        {subMoqItems.length > 0 && (
          <div className="mb-4 p-3 bg-amber-500/10 border border-amber-500/30 rounded-xl text-amber-300 text-xs space-y-1">
            <div className="flex items-center gap-2 font-semibold">
              <AlertTriangle className="w-4 h-4 flex-shrink-0 text-amber-400" />
              <span>Minimum Order Quantity Violation:</span>
            </div>
            {subMoqItems.map((p) => (
              <div key={p.id} className="pl-6 text-[11px] text-amber-200">
                • {p.name}: selected {orderQuantities[p.id]}, minimum required is {p.moq}
              </div>
            ))}
          </div>
        )}

        <div className="space-y-2 mb-4 max-h-60 overflow-y-auto divide-y divide-slate-800">
          {regularProducts
            .filter((p) => (orderQuantities[p.id] || 0) > 0)
            .map((p) => {
              const qty = orderQuantities[p.id];
              const isViolating = qty < p.moq;
              return (
                <div key={p.id} className="pt-2 flex justify-between text-xs items-center">
                  <div>
                    <div className="font-semibold text-white flex items-center gap-1.5">
                      <span>{p.name}</span>
                      {isViolating && (
                        <span className="px-1.5 py-0.5 bg-rose-500/20 text-rose-300 rounded text-[10px]">
                          Below MOQ ({p.moq})
                        </span>
                      )}
                    </div>
                    <div className="text-slate-400">
                      {qty} x ₹{(p.price_paise / 100).toLocaleString("en-IN")}
                    </div>
                  </div>
                  <div className="font-bold text-emerald-400">
                    ₹{((qty * p.price_paise) / 100).toLocaleString("en-IN")}
                  </div>
                </div>
              );
            })}
        </div>

        <div className="pt-4 border-t border-slate-800 flex justify-between items-center mb-6">
          <span className="text-sm text-slate-300 font-semibold">Total Amount</span>
          <span className="text-xl font-extrabold text-emerald-400">
            ₹{calculateOrderTotal().toLocaleString("en-IN")}
          </span>
        </div>

        <div className="flex gap-3">
          <button
            type="button"
            onClick={onClose}
            className="w-1/2 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl transition-colors"
          >
            Modify Order
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={loading || subMoqItems.length > 0}
            className="w-1/2 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-indigo-600/30 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 transition-all"
          >
            {loading ? (
              <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            ) : (
              <span>Confirm & Continue</span>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};
