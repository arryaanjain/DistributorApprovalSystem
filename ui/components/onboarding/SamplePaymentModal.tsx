import React from "react";
import { Sparkles, Lock } from "lucide-react";
import { ProductItem } from "@/types/onboarding";

interface SamplePaymentModalProps {
  selectedSampleItem: ProductItem;
  loading: boolean;
  onClose: () => void;
  onCompletePayment: () => void;
}

export const SamplePaymentModal: React.FC<SamplePaymentModalProps> = ({
  selectedSampleItem,
  loading,
  onClose,
  onCompletePayment,
}) => {
  return (
    <div className="fixed inset-0 z-50 bg-slate-950/85 backdrop-blur-md flex items-center justify-center p-4">
      <div className="w-full max-w-sm bg-slate-900 border border-amber-500/40 rounded-2xl p-6 shadow-2xl text-center">
        <div className="w-12 h-12 bg-amber-500/20 text-amber-400 rounded-2xl flex items-center justify-center mx-auto mb-4 border border-amber-500/30">
          <Sparkles className="w-6 h-6" />
        </div>

        <h3 className="text-lg font-bold text-white mb-1">Razorpay Sample Trial Payment</h3>
        <p className="text-xs text-slate-400 mb-4">{selectedSampleItem.name}</p>

        <div className="bg-slate-800/80 p-4 rounded-xl border border-slate-700 mb-6 text-center">
          <div className="text-xs text-slate-400">Amount Payable</div>
          <div className="text-2xl font-extrabold text-amber-400 mt-1">
            ₹{(selectedSampleItem.price_paise / 100).toLocaleString("en-IN")}
          </div>
        </div>

        <div className="space-y-3">
          <button
            type="button"
            onClick={onCompletePayment}
            disabled={loading}
            className="w-full py-3 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-400 hover:to-orange-400 text-slate-950 font-bold text-sm rounded-xl shadow-lg shadow-amber-500/30 flex items-center justify-center gap-2"
          >
            {loading ? (
              <div className="w-5 h-5 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
            ) : (
              <>
                <Lock className="w-4 h-4" />
                <span>Simulate Razorpay Payment</span>
              </>
            )}
          </button>

          <button
            type="button"
            onClick={onClose}
            className="w-full py-2 text-xs text-slate-400 hover:text-slate-200"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
};
