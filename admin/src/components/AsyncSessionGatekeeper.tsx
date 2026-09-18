import React, { useState, useEffect } from 'react';
import { ShieldCheck, RefreshCw, Key, Database, AlertCircle } from 'lucide-react';

interface AsyncSessionGatekeeperProps {
  error?: string | null;
  onRetry?: () => void;
}

export const AsyncSessionGatekeeper: React.FC<AsyncSessionGatekeeperProps> = ({ error, onRetry }) => {
  const [step, setStep] = useState<number>(1);

  useEffect(() => {
    const t1 = setTimeout(() => setStep(2), 250);
    const t2 = setTimeout(() => setStep(3), 600);
    return () => {
      clearTimeout(t1);
      clearTimeout(t2);
    };
  }, []);

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4 font-sans relative overflow-hidden">
      {/* Background glow effects */}
      <div className="absolute -top-40 -left-40 w-96 h-96 bg-indigo-600/20 rounded-full blur-3xl pointer-events-none" />
      <div className="absolute -bottom-40 -right-40 w-96 h-96 bg-violet-600/20 rounded-full blur-3xl pointer-events-none" />

      <div className="w-full max-w-md bg-slate-900/90 border border-indigo-500/30 rounded-3xl p-8 shadow-2xl backdrop-blur-xl relative z-10">
        <div className="text-center mb-6">
          <div className="w-16 h-16 bg-gradient-to-tr from-indigo-600 to-violet-500 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-lg shadow-indigo-500/30 relative">
            <ShieldCheck className="w-8 h-8 text-white animate-pulse" />
            <div className="absolute -top-1 -right-1 w-4 h-4 bg-indigo-400 rounded-full animate-ping" />
          </div>
          <h2 className="text-xl font-bold text-white tracking-tight">Kresconet Admin Console</h2>
          <p className="text-xs text-slate-400 mt-1">Verifying Credentials & Synchronizing Auth Session</p>
        </div>

        {error ? (
          <div className="p-4 bg-rose-500/10 border border-rose-500/30 rounded-xl text-rose-300 text-xs text-center space-y-3">
            <div className="flex items-center justify-center gap-2 text-rose-400 font-semibold">
              <AlertCircle className="w-4 h-4" />
              <span>Session Verification Failed</span>
            </div>
            <p className="text-slate-400 text-[11px]">{error}</p>
            {onRetry && (
              <button
                onClick={onRetry}
                className="mt-2 w-full py-2 bg-rose-600/30 hover:bg-rose-600/50 text-white rounded-lg transition-colors font-medium text-xs flex items-center justify-center gap-2"
              >
                <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                <span>Retry Token Rotation</span>
              </button>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            {/* Step 1: Read local state */}
            <div
              className={`p-3.5 rounded-xl border transition-all duration-300 flex items-center gap-3 ${
                step >= 1
                  ? 'bg-indigo-950/40 border-indigo-500/30 text-indigo-200'
                  : 'bg-slate-900/40 border-slate-800 text-slate-600'
              }`}
            >
              <Key className={`w-4 h-4 ${step >= 1 ? 'text-indigo-400 animate-bounce' : ''}`} />
              <div className="flex-1 min-w-0">
                <p className="text-xs font-semibold">1. Checking Auth Tokens</p>
                <p className="text-[10px] text-slate-400 truncate">Reading local access and refresh storage</p>
              </div>
              {step >= 1 && <div className="w-2 h-2 rounded-full bg-indigo-400 animate-ping" />}
            </div>

            {/* Step 2: Postgres Token Rotation */}
            <div
              className={`p-3.5 rounded-xl border transition-all duration-300 flex items-center gap-3 ${
                step >= 2
                  ? 'bg-indigo-950/40 border-indigo-500/30 text-indigo-200'
                  : 'bg-slate-900/40 border-slate-800 text-slate-600'
              }`}
            >
              <Database className={`w-4 h-4 ${step >= 2 ? 'text-indigo-400 animate-pulse' : ''}`} />
              <div className="flex-1 min-w-0">
                <p className="text-xs font-semibold">2. Requesting Postgres Token Rotation</p>
                <p className="text-[10px] text-slate-400 truncate">Exchanging HttpOnly Cookie / Body token</p>
              </div>
              {step >= 2 && <RefreshCw className="w-3.5 h-3.5 text-indigo-400 animate-spin" />}
            </div>

            {/* Step 3: Verifying Role Permissions */}
            <div
              className={`p-3.5 rounded-xl border transition-all duration-300 flex items-center gap-3 ${
                step >= 3
                  ? 'bg-indigo-950/40 border-indigo-500/30 text-indigo-200'
                  : 'bg-slate-900/40 border-slate-800 text-slate-600'
              }`}
            >
              <ShieldCheck className={`w-4 h-4 ${step >= 3 ? 'text-emerald-400' : ''}`} />
              <div className="flex-1 min-w-0">
                <p className="text-xs font-semibold">3. Restoring User Session</p>
                <p className="text-[10px] text-slate-400 truncate">Validating internal employee role permissions</p>
              </div>
            </div>

            {/* Progress bar */}
            <div className="pt-2">
              <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-indigo-500 to-violet-500 transition-all duration-500"
                  style={{ width: `${(step / 3) * 100}%` }}
                />
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
