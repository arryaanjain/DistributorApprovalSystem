import React from "react";
import { Briefcase, ArrowLeft, ArrowRight } from "lucide-react";
import { Step2Data } from "@/types/onboarding";

interface Step2BusinessExperienceProps {
  step2: Step2Data;
  setStep2: React.Dispatch<React.SetStateAction<Step2Data>>;
  loading: boolean;
  onBack: () => void;
  onSubmit: (e: React.FormEvent) => void;
}

export const Step2BusinessExperience: React.FC<Step2BusinessExperienceProps> = ({
  step2,
  setStep2,
  loading,
  onBack,
  onSubmit,
}) => {
  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl max-w-3xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
          <Briefcase className="w-5 h-5" />
        </div>
        <div>
          <h2 className="text-xl font-bold text-white">Step 2: Business Experience & Distribution</h2>
          <p className="text-xs text-slate-400">Specify distribution reach and historical trade metrics</p>
        </div>
      </div>

      <form onSubmit={onSubmit} className="space-y-6">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Distribution Experience (Years) *
            </label>
            <input
              type="number"
              step="0.5"
              value={step2.distribution_experience_years}
              onChange={(e) =>
                setStep2({ ...step2, distribution_experience_years: parseFloat(e.target.value || "0") })
              }
              required
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Serviced Retailers / Wholesalers Count *
            </label>
            <input
              type="number"
              value={step2.serviced_retailers_wholesalers_count}
              onChange={(e) =>
                setStep2({
                  ...step2,
                  serviced_retailers_wholesalers_count: parseInt(e.target.value || "0", 10),
                })
              }
              required
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Interested Business Role *
            </label>
            <select
              value={step2.interested_business_role}
              onChange={(e) => setStep2({ ...step2, interested_business_role: e.target.value })}
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            >
              <option value="Super Stockist">Super Stockist</option>
              <option value="Distributor">Authorized Distributor</option>
              <option value="Wholesaler">Wholesaler</option>
              <option value="Retailer">Retail Outlet</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
              Total Vintage Years in Trade
            </label>
            <input
              type="number"
              value={step2.vintage_years}
              onChange={(e) => setStep2({ ...step2, vintage_years: parseFloat(e.target.value || "0") })}
              className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>

        <div>
          <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
            Approx Monthly Business Volume (₹ INR)
          </label>
          <input
            type="number"
            value={step2.approx_monthly_business_inr}
            onChange={(e) =>
              setStep2({ ...step2, approx_monthly_business_inr: parseInt(e.target.value || "0", 10) })
            }
            className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
          />
        </div>

        <div>
          <label className="block text-xs font-semibold uppercase text-slate-400 mb-1">
            Existing Brands Currently Distributed
          </label>
          <input
            type="text"
            value={step2.existing_brands}
            onChange={(e) => setStep2({ ...step2, existing_brands: e.target.value })}
            placeholder="e.g. Amul, Britannia, Fortune"
            className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500"
          />
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
            <span>Continue to Credit Preference</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};
