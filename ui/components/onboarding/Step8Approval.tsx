import React from "react";
import { Sparkles, CheckCircle2 } from "lucide-react";

interface Step8ApprovalProps {
  trialActivated: boolean;
  onGoToDashboard: () => void;
}

export const Step8Approval: React.FC<Step8ApprovalProps> = ({
  trialActivated,
  onGoToDashboard,
}) => {
  return (
    <div className="max-w-3xl mx-auto text-center py-8">
      <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl">
        {trialActivated ? (
          <div className="space-y-6">
            <div className="w-20 h-20 bg-gradient-to-tr from-amber-500 to-orange-500 rounded-3xl flex items-center justify-center mx-auto shadow-xl shadow-amber-500/30">
              <Sparkles className="w-10 h-10 text-slate-950" />
            </div>
            <div>
              <span className="px-3 py-1 bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded-full text-xs font-bold uppercase tracking-wider">
                Trial Status Active
              </span>
              <h2 className="text-2xl font-bold text-white mt-3">Sample Kit Ordered & Trial Activated</h2>
              <p className="text-sm text-slate-400 max-w-md mx-auto mt-2">
                Your sample product trial has been booked successfully! You now have trial catalog privileges.
              </p>
            </div>

            <button
              onClick={onGoToDashboard}
              className="px-8 py-3.5 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-400 hover:to-orange-400 text-slate-950 font-bold rounded-xl shadow-lg shadow-amber-500/20 transition-all"
            >
              Go to Trial Dashboard
            </button>
          </div>
        ) : (
          <div className="space-y-6">
            <div className="w-20 h-20 bg-gradient-to-tr from-emerald-500 to-teal-500 rounded-3xl flex items-center justify-center mx-auto shadow-xl shadow-emerald-500/30">
              <CheckCircle2 className="w-10 h-10 text-slate-950" />
            </div>
            <div>
              <span className="px-3 py-1 bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 rounded-full text-xs font-bold uppercase tracking-wider">
                Underwriting Review In-Progress
              </span>
              <h2 className="text-2xl font-bold text-white mt-3">Onboarding Submission Completed</h2>
              <p className="text-sm text-slate-400 max-w-md mx-auto mt-2">
                Your business profile and credit application have been logged. Final credit limit approval and e-signing agreement will be rendered shortly.
              </p>
            </div>

            <button
              onClick={onGoToDashboard}
              className="px-8 py-3.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 transition-all"
            >
              Enter Distributor Portal
            </button>
          </div>
        )}
      </div>
    </div>
  );
};
