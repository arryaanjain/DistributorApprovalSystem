import React from "react";
import { Building, ArrowRight } from "lucide-react";
import { Step7Data } from "@/types/onboarding";

interface Step7BankAccountProps {
  step7: Step7Data;
  setStep7: React.Dispatch<React.SetStateAction<Step7Data>>;
  loading: boolean;
  onSubmit: (e: React.FormEvent, skip?: boolean) => void;
}

export const Step7BankAccount: React.FC<Step7BankAccountProps> = ({
  step7,
  setStep7,
  loading,
  onSubmit,
}) => {
  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
            <Building className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">Step 7: Bank Account Details</h2>
            <p className="text-xs text-slate-400">Optional step for direct debit & automated settlements</p>
          </div>
        </div>

        <span className="px-3 py-1 bg-amber-500/10 text-amber-300 border border-amber-500/30 rounded-full text-xs font-semibold">
          Optional
        </span>
      </div>

      <form onSubmit={(e) => onSubmit(e, false)} className="space-y-6">
        <div>
          <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
            Bank Account Number
          </label>
          <input
            type="text"
            value={step7.account_number}
            onChange={(e) => setStep7({ ...step7, account_number: e.target.value })}
            placeholder="50100234567890"
            className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm font-mono text-slate-100 focus:outline-none focus:border-indigo-500"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              IFSC Code
            </label>
            <input
              type="text"
              value={step7.ifsc}
              onChange={(e) => setStep7({ ...step7, ifsc: e.target.value.toUpperCase() })}
              placeholder="HDFC0001234"
              maxLength={11}
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm font-mono text-slate-100 focus:outline-none focus:border-indigo-500 uppercase"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Account Holder Name
            </label>
            <input
              type="text"
              value={step7.account_holder}
              onChange={(e) => setStep7({ ...step7, account_holder: e.target.value })}
              placeholder="Kresco Traders Pvt Ltd"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>

        <div className="flex items-center justify-between pt-4">
          <button
            type="button"
            onClick={(e) => onSubmit(e, true)}
            className="px-4 py-2.5 bg-slate-800 text-slate-300 text-sm font-medium rounded-xl hover:bg-slate-700 border border-slate-700"
          >
            Skip Bank Account for Now
          </button>

          <button
            type="submit"
            disabled={loading}
            className="px-6 py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 transition-all flex items-center gap-2"
          >
            <span>Submit Bank & Finish</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};
