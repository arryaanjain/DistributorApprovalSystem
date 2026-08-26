import React, { useState } from "react";
import { Sparkles, Lock, Truck, ShieldCheck, CheckCircle, AlertCircle } from "lucide-react";
import { ProductItem } from "@/types/onboarding";

export interface AddressFormData {
  address_line1: string;
  address_line2: string;
  city: string;
  state: string;
  pin: string;
  phone: string;
}

interface SamplePaymentModalProps {
  selectedSampleItem: ProductItem;
  loading: boolean;
  onClose: () => void;
  onCompletePayment: (address: AddressFormData) => void;
}

export const SamplePaymentModal: React.FC<SamplePaymentModalProps> = ({
  selectedSampleItem,
  loading,
  onClose,
  onCompletePayment,
}) => {
  const [address, setAddress] = useState<AddressFormData>({
    address_line1: "",
    address_line2: "",
    city: "",
    state: "",
    pin: "",
    phone: "",
  });

  const [formError, setFormError] = useState<string | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!address.address_line1.trim() || !address.city.trim() || !address.state.trim() || !address.pin.trim()) {
      setFormError("Please fill out complete shipping address (Address Line 1, City, State, PIN).");
      return;
    }
    setFormError(null);
    onCompletePayment(address);
  };

  const amountINR = (selectedSampleItem.price_paise / 100).toLocaleString("en-IN");

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/85 backdrop-blur-md flex items-center justify-center p-4">
      <div className="w-full max-w-lg bg-slate-900 border border-amber-500/40 rounded-3xl p-6 shadow-2xl text-left space-y-5 max-h-[90vh] overflow-y-auto relative">
        
        {/* Header */}
        <div className="flex items-center justify-between border-b border-slate-800 pb-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-amber-500/20 text-amber-400 rounded-2xl flex items-center justify-center border border-amber-500/30">
              <Sparkles className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white">Sample Kit Shipping & Payment</h3>
              <p className="text-xs text-slate-400">{selectedSampleItem.name}</p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            disabled={loading}
            className="w-8 h-8 rounded-full bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white flex items-center justify-center transition-colors text-sm font-bold"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Summary Box */}
          <div className="bg-gradient-to-r from-amber-950/30 via-slate-800/80 to-slate-900 p-4 rounded-2xl border border-amber-500/30 flex items-center justify-between">
            <div>
              <span className="text-[11px] uppercase tracking-wider text-amber-400/90 font-semibold flex items-center gap-1">
                <ShieldCheck className="w-3.5 h-3.5" /> Amount Payable
              </span>
              <div className="text-2xl font-black text-amber-400 mt-0.5">
                ₹{amountINR}
              </div>
            </div>
            <span className="text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-3 py-1.5 rounded-full font-bold flex items-center gap-1">
              <CheckCircle className="w-3 h-3 text-emerald-400" />
              Activates Trial Status
            </span>
          </div>

          {/* Shipping Address Section */}
          <div className="space-y-3">
            <div className="flex items-center gap-2 text-xs font-bold text-slate-300 uppercase tracking-wider">
              <Truck className="w-4 h-4 text-amber-400" />
              <span>Sample Delivery Address</span>
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">
                Street Address / Line 1 <span className="text-rose-400">*</span>
              </label>
              <input
                type="text"
                value={address.address_line1}
                onChange={(e) => setAddress({ ...address, address_line1: e.target.value })}
                placeholder="Plot/Building No., Street Name"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2.5 text-xs text-white placeholder-slate-600 focus:outline-none focus:border-amber-500 transition-colors"
                required
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">
                Address Line 2 (Optional)
              </label>
              <input
                type="text"
                value={address.address_line2}
                onChange={(e) => setAddress({ ...address, address_line2: e.target.value })}
                placeholder="Landmark, Area"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2.5 text-xs text-white placeholder-slate-600 focus:outline-none focus:border-amber-500 transition-colors"
              />
            </div>

            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">
                  City <span className="text-rose-400">*</span>
                </label>
                <input
                  type="text"
                  value={address.city}
                  onChange={(e) => setAddress({ ...address, city: e.target.value })}
                  placeholder="City"
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-xs text-white placeholder-slate-600 focus:outline-none focus:border-amber-500 transition-colors"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">
                  State <span className="text-rose-400">*</span>
                </label>
                <input
                  type="text"
                  value={address.state}
                  onChange={(e) => setAddress({ ...address, state: e.target.value })}
                  placeholder="State"
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-xs text-white placeholder-slate-600 focus:outline-none focus:border-amber-500 transition-colors"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-400 mb-1">
                  Pincode <span className="text-rose-400">*</span>
                </label>
                <input
                  type="text"
                  value={address.pin}
                  onChange={(e) => setAddress({ ...address, pin: e.target.value })}
                  placeholder="PIN"
                  maxLength={6}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-2 text-xs text-white placeholder-slate-600 focus:outline-none focus:border-amber-500 transition-colors"
                  required
                />
              </div>
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-400 mb-1">
                Contact Phone Number
              </label>
              <input
                type="tel"
                value={address.phone}
                onChange={(e) => setAddress({ ...address, phone: e.target.value })}
                placeholder="10-digit mobile number for delivery"
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2.5 text-xs text-white placeholder-slate-600 focus:outline-none focus:border-amber-500 transition-colors"
              />
            </div>
          </div>

          {formError && (
            <div className="text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 p-3 rounded-xl flex items-center gap-2">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{formError}</span>
            </div>
          )}

          {/* Action Buttons */}
          <div className="pt-2 space-y-2">
            <button
              type="submit"
              disabled={loading}
              className="w-full py-3.5 bg-gradient-to-r from-amber-500 via-orange-500 to-amber-600 hover:from-amber-400 hover:to-orange-400 text-slate-950 font-extrabold text-sm rounded-xl shadow-lg shadow-amber-500/20 flex items-center justify-center gap-2 transition-all disabled:opacity-50"
            >
              {loading ? (
                <div className="flex items-center gap-2">
                  <div className="w-4 h-4 border-2 border-slate-950 border-t-transparent rounded-full animate-spin" />
                  <span>Processing Sample Order...</span>
                </div>
              ) : (
                <>
                  <Lock className="w-4 h-4" />
                  <span>Proceed to Razorpay Checkout</span>
                </>
              )}
            </button>

            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="w-full py-2 text-xs text-slate-400 hover:text-slate-200 transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
