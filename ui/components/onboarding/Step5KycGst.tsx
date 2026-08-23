import React from "react";
import { FileCheck, ArrowLeft, ArrowRight } from "lucide-react";
import { Step5Data } from "@/types/onboarding";

interface Step5KycGstProps {
  step5: Step5Data;
  setStep5: React.Dispatch<React.SetStateAction<Step5Data>>;
  loading: boolean;
  onBack: () => void;
  onSubmit: (e: React.FormEvent) => void;
}

export const Step5KycGst: React.FC<Step5KycGstProps> = ({
  step5,
  setStep5,
  loading,
  onBack,
  onSubmit,
}) => {
  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl max-w-3xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
          <FileCheck className="w-5 h-5" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-white">Step 5: KYC & GST Verification</h2>
          <p className="text-xs text-slate-400">Submit PAN & GST credentials for real-time verification</p>
        </div>
      </div>

      <form onSubmit={onSubmit} className="space-y-6">
        <div>
          <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
            PAN Number (10 Digits) *
          </label>
          <input
            type="text"
            value={step5.pan}
            onChange={(e) => setStep5({ ...step5, pan: e.target.value.toUpperCase() })}
            required
            maxLength={10}
            placeholder="ABCDE1234F"
            className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm font-mono text-slate-100 focus:outline-none focus:border-indigo-500 uppercase tracking-widest"
          />
        </div>

        <div className="p-4 bg-slate-800/50 border border-slate-700/60 rounded-xl space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm font-semibold text-white">Does your business have a GST Registration?</span>
            <input
              type="checkbox"
              checked={step5.has_gst}
              onChange={(e) => setStep5({ ...step5, has_gst: e.target.checked })}
              className="w-5 h-5 text-indigo-600 rounded bg-slate-900 border-slate-700"
            />
          </div>

          {step5.has_gst && (
            <div>
              <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
                GSTIN (15 Digits)
              </label>
              <input
                type="text"
                value={step5.gst_number}
                onChange={(e) => setStep5({ ...step5, gst_number: e.target.value.toUpperCase() })}
                maxLength={15}
                placeholder="24ABCDE1234F1Z5"
                className="w-full px-4 py-3 bg-slate-800 border border-slate-700 rounded-xl text-sm font-mono text-slate-100 focus:outline-none focus:border-indigo-500 uppercase tracking-widest"
              />
            </div>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              FSSAI License Number (Optional)
            </label>
            <input
              type="text"
              value={step5.fssai_number}
              onChange={(e) => setStep5({ ...step5, fssai_number: e.target.value })}
              placeholder="10020021000123"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Udyam / MSME Registration (Optional)
            </label>
            <input
              type="text"
              value={step5.udyam_number}
              onChange={(e) => setStep5({ ...step5, udyam_number: e.target.value })}
              placeholder="UDYAM-GJ-01-0001234"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>
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
            <span>Submit KYC & Continue</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};
