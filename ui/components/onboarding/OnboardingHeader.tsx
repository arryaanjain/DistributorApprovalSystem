import React from "react";
import { Building2, ShieldCheck, LogOut } from "lucide-react";

interface OnboardingHeaderProps {
  token: string | null;
  onSignOut: () => void;
}

export const OnboardingHeader: React.FC<OnboardingHeaderProps> = ({
  token,
  onSignOut,
}) => {
  return (
    <header className="border-b border-slate-800 bg-slate-900/80 backdrop-blur-lg sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-gradient-to-tr from-indigo-600 to-violet-500 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/30">
            <Building2 className="w-6 h-6 text-white" />
          </div>
          <div>
            <span className="font-bold text-lg text-white tracking-tight">KRESCONET</span>
            <span className="ml-2 text-xs font-semibold px-2 py-0.5 bg-indigo-500/10 text-indigo-400 border border-indigo-500/30 rounded-full">
              Distributor Onboarding
            </span>
          </div>
        </div>

        {token && (
          <div className="flex items-center gap-3">
            <button
              onClick={onSignOut}
              type="button"
              className="px-3.5 py-2 bg-rose-500/10 hover:bg-rose-500/20 active:bg-rose-500/30 text-rose-400 border border-rose-500/30 rounded-xl text-xs font-semibold transition-all flex items-center gap-1.5 touch-manipulation min-h-[38px]"
            >
              <LogOut className="w-3.5 h-3.5" />
              <span>Sign Out</span>
            </button>
          </div>
        )}
      </div>
    </header>
  );
};
