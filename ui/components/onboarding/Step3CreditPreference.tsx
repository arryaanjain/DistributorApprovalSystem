import React from "react";
import { CreditCard, CheckCircle2, ArrowLeft, ArrowRight } from "lucide-react";
import { Step3Data } from "@/types/onboarding";

interface Step3CreditPreferenceProps {
  step3: Step3Data;
  setStep3: React.Dispatch<React.SetStateAction<Step3Data>>;
  loading: boolean;
  onBack: () => void;
  onSubmit: (e: React.FormEvent) => void;
}

export const Step3CreditPreference: React.FC<Step3CreditPreferenceProps> = ({
  step3,
  setStep3,
  loading,
  onBack,
  onSubmit,
}) => {
  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl max-w-3xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
          <CreditCard className="w-5 h-5" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-white">Step 3: Credit Preference</h2>
          <p className="text-xs text-slate-400">Select desired payment terms & credit exposure</p>
        </div>
      </div>

      <form onSubmit={onSubmit} className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[
            { id: "advance_100", label: "100% Advance Payment", desc: "No credit exposure, priority dispatch" },
            { id: "cod", label: "Cash on Delivery (COD)", desc: "Pay upon physical receipt" },
            { id: "15_days", label: "15 Days Credit", desc: "Standard trade credit term" },
            { id: "30_days", label: "30 Days Extended Credit", desc: "For established distributors" },
          ].map((item) => (
            <div
              key={item.id}
              onClick={() => setStep3({ preference: item.id })}
              className={`p-4 rounded-xl border cursor-pointer transition-all ${
                step3.preference === item.id
                  ? "bg-indigo-600/15 border-indigo-500 text-white shadow-md shadow-indigo-500/10"
                  : "bg-slate-800/40 border-slate-800 text-slate-400 hover:border-slate-700"
              }`}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="font-semibold text-sm text-slate-100">{item.label}</span>
                {step3.preference === item.id && <CheckCircle2 className="w-4 h-4 text-indigo-400" />}
              </div>
              <p className="text-xs text-slate-400">{item.desc}</p>
            </div>
          ))}
        </div>

        <div className="p-4 bg-indigo-950/30 border border-indigo-500/20 rounded-xl text-xs text-indigo-300">
          * Note: Selected payment terms serve as your request preference. Final credit limits and terms will be determined during automated underwriting.
        </div>

        <div className="flex justify-between pt-4">
          <button
            type="button"
            onClick={onBack}
            className="px-4 py-2.5 bg-slate-800 text-slate-300 text-sm rounded-xl hover:bg-slate-700 flex items-center gap-2"
          >
            <ArrowLeft className="w-4 h-4" /> Back
          </button>
          <button
            type="submit"
            disabled={loading}
            className="px-6 py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 transition-all flex items-center gap-2"
          >
            <span>Continue to Order Requirement</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};
