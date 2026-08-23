import React from "react";
import { ShieldCheck, ArrowLeft, ArrowRight } from "lucide-react";
import { Step6Data } from "@/types/onboarding";

interface Step6AuthorizationProps {
  step6: Step6Data;
  setStep6: React.Dispatch<React.SetStateAction<Step6Data>>;
  loading: boolean;
  onBack: () => void;
  onSubmit: (e: React.FormEvent) => void;
}

export const Step6Authorization: React.FC<Step6AuthorizationProps> = ({
  step6,
  setStep6,
  loading,
  onBack,
  onSubmit,
}) => {
  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl max-w-3xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
          <ShieldCheck className="w-5 h-5" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-white">Step 6: Legal Authorization Declaration</h2>
          <p className="text-xs text-slate-400">Review & digitally consent to credit checks</p>
        </div>
      </div>

      <form onSubmit={onSubmit} className="space-y-6">
        <div className="p-5 bg-slate-800/50 border border-slate-700/60 rounded-xl text-xs text-slate-300 space-y-3 leading-relaxed">
          <p className="font-semibold text-slate-100">Distributor Authorization Terms:</p>
          <p>
            1. I certify that all information provided in this onboarding application is true, correct, and complete to the best of my knowledge.
          </p>
          <p>
            2. I authorize Kresconet and its financial partners to obtain credit reports from authorized credit bureaus (Experian / CIBIL) and perform automated verification of GST and bank details.
          </p>
        </div>

        <div className="flex items-start gap-3 p-4 bg-indigo-950/30 border border-indigo-500/30 rounded-xl">
          <input
            type="checkbox"
            id="authorize"
            checked={step6.authorized}
            onChange={(e) => setStep6({ ...step6, authorized: e.target.checked })}
            className="mt-0.5 w-5 h-5 text-indigo-600 rounded bg-slate-900 border-slate-700"
          />
          <label htmlFor="authorize" className="text-xs text-slate-200 cursor-pointer">
            I accept and confirm that I am an authorized signatory of the applicant entity.
          </label>
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
            disabled={loading || !step6.authorized}
            className="px-6 py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 transition-all disabled:opacity-50 flex items-center gap-2"
          >
            <span>Grant Authorization & Continue</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};
