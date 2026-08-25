import React from "react";
import { Building2, ArrowRight } from "lucide-react";
import { Step1Data } from "@/types/onboarding";

interface Step1BusinessDetailsProps {
  step1: Step1Data;
  setStep1: React.Dispatch<React.SetStateAction<Step1Data>>;
  loading: boolean;
  onSubmit: (e: React.FormEvent) => void;
}

export const Step1BusinessDetails: React.FC<Step1BusinessDetailsProps> = ({
  step1,
  setStep1,
  loading,
  onSubmit,
}) => {
  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl max-w-3xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
          <Building2 className="w-5 h-5" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-white">Step 1: Business Details</h2>
          <p className="text-xs text-slate-400">Provide legal entity & primary contact details</p>
        </div>
      </div>

      <form onSubmit={onSubmit} className="space-y-6">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Business Owner Name *
            </label>
            <input
              type="text"
              value={step1.name}
              onChange={(e) => setStep1({ ...step1, name: e.target.value })}
              required
              placeholder="Rajesh Kumar"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Email Address *
            </label>
            <input
              type="email"
              value={step1.email}
              onChange={(e) => setStep1({ ...step1, email: e.target.value })}
              required
              placeholder="rajesh@krescotraders.com"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Business / Firm Name *
            </label>
            <input
              type="text"
              value={step1.business_name}
              onChange={(e) => setStep1({ ...step1, business_name: e.target.value })}
              required
              placeholder="Kresco Traders Pvt Ltd"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Constitution Type *
            </label>
            <select
              value={step1.constitution}
              onChange={(e) => setStep1({ ...step1, constitution: e.target.value })}
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            >
              <option value="proprietorship">Sole Proprietorship</option>
              <option value="partnership">Partnership Firm</option>
              <option value="llp">Limited Liability Partnership (LLP)</option>
              <option value="private_limited">Private Limited (Pvt Ltd)</option>
              <option value="public_limited">Public Limited</option>
            </select>
          </div>
        </div>

        <div>
          <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
            Registered Address Line 1 *
          </label>
          <input
            type="text"
            value={step1.address_line1}
            onChange={(e) => setStep1({ ...step1, address_line1: e.target.value })}
            required
            placeholder="Plot No. 42, GIDC Industrial Estate"
            className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
          />
        </div>

        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">City *</label>
            <input
              type="text"
              value={step1.city}
              onChange={(e) => setStep1({ ...step1, city: e.target.value })}
              required
              placeholder="Ahmedabad"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">State *</label>
            <input
              type="text"
              value={step1.state}
              onChange={(e) => setStep1({ ...step1, state: e.target.value })}
              required
              placeholder="Gujarat"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">PIN Code *</label>
            <input
              type="text"
              value={step1.pin}
              onChange={(e) => setStep1({ ...step1, pin: e.target.value })}
              required
              maxLength={6}
              placeholder="380015"
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>

        <div className="flex justify-end pt-4">
          <button
            type="submit"
            disabled={loading}
            className="px-6 py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-xl shadow-lg shadow-indigo-600/30 transition-all flex items-center gap-2"
          >
            <span>Save & Proceed to Step 2</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};
