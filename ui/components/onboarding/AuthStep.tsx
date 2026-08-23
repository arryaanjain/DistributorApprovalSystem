import React from "react";
import { Phone, ArrowRight } from "lucide-react";

interface AuthStepProps {
  mobile: string;
  setMobile: (val: string) => void;
  otp: string;
  setOtp: (val: string) => void;
  devOtp: string | null;
  otpSent: boolean;
  loading: boolean;
  onSendOtp: (e: React.FormEvent) => void;
  onVerifyOtp: (e: React.FormEvent) => void;
}

export const AuthStep: React.FC<AuthStepProps> = ({
  mobile,
  setMobile,
  otp,
  setOtp,
  devOtp,
  otpSent,
  loading,
  onSendOtp,
  onVerifyOtp,
}) => {
  return (
    <div className="max-w-md mx-auto py-12">
      <div className="bg-slate-900/80 border border-indigo-500/20 rounded-2xl p-8 shadow-2xl backdrop-blur-xl">
        <div className="text-center mb-8">
          <div className="w-16 h-16 bg-gradient-to-tr from-indigo-600 to-violet-500 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-lg shadow-indigo-500/30">
            <Phone className="w-8 h-8 text-white" />
          </div>
          <h2 className="text-2xl font-bold text-white">Distributor Login</h2>
          <p className="text-sm text-slate-400 mt-1">
            Enter your mobile number to begin or resume your onboarding application
          </p>
        </div>

        {!otpSent ? (
          <form onSubmit={onSendOtp} className="space-y-5">
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
                Mobile Number (+91)
              </label>
              <div className="relative">
                <span className="absolute left-3.5 top-3.5 text-slate-400 font-semibold text-sm">
                  +91
                </span>
                <input
                  type="tel"
                  value={mobile}
                  onChange={(e) => setMobile(e.target.value)}
                  required
                  maxLength={10}
                  placeholder="9876543210"
                  className="w-full pl-14 pr-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl focus:outline-none focus:border-indigo-500 text-slate-100 placeholder-slate-500"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3.5 px-4 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 transition-all flex items-center justify-center gap-2"
            >
              {loading ? (
                <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <>
                  <span>Send Verification OTP</span>
                  <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </form>
        ) : (
          <form onSubmit={onVerifyOtp} className="space-y-5">
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
                Enter 6-Digit Verification Code
              </label>
              <input
                type="text"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                required
                maxLength={6}
                placeholder="123456"
                className="w-full text-center tracking-[0.5em] text-xl font-mono py-3 bg-slate-800/80 border border-slate-700 rounded-xl focus:outline-none focus:border-indigo-500 text-slate-100"
              />
              {devOtp && (
                <div className="mt-2 text-xs text-amber-400 text-center font-mono">
                  Dev OTP Auto-filled: {devOtp}
                </div>
              )}
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3.5 px-4 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 transition-all flex items-center justify-center gap-2"
            >
              {loading ? (
                <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <>
                  <span>Verify OTP & Continue</span>
                  <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </form>
        )}
      </div>
    </div>
  );
};
