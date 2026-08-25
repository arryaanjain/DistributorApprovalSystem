import React, { useState } from "react";
import { Briefcase, ArrowLeft, ArrowRight, Plus, X, Tag, Sparkles } from "lucide-react";
import { Step2Data } from "@/types/onboarding";

interface Step2BusinessExperienceProps {
  step2: Step2Data;
  setStep2: React.Dispatch<React.SetStateAction<Step2Data>>;
  loading: boolean;
  onBack: () => void;
  onSubmit: (e: React.FormEvent) => void;
}

const POPULAR_SUGGESTIONS = [
  "Amul",
  "Britannia",
  "Parle",
  "Nestle",
  "Tata Consumer",
  "HUL",
  "Dabur",
  "Marico",
  "Haldiram's",
  "ITC",
];

export const Step2BusinessExperience: React.FC<Step2BusinessExperienceProps> = ({
  step2,
  setStep2,
  loading,
  onBack,
  onSubmit,
}) => {
  const [brandInput, setBrandInput] = useState("");

  const brandArray: string[] = Array.isArray(step2.existing_brands)
    ? step2.existing_brands
    : typeof step2.existing_brands === "string" && step2.existing_brands.trim().length > 0
    ? step2.existing_brands.split(",").map((s) => s.trim()).filter(Boolean)
    : [];

  const handleAddBrand = (name: string) => {
    const clean = name.trim();
    if (!clean) return;
    if (!brandArray.some((b) => b.toLowerCase() === clean.toLowerCase())) {
      const updated = [...brandArray, clean];
      setStep2({ ...step2, existing_brands: updated });
    }
    setBrandInput("");
  };

  const handleRemoveBrand = (target: string) => {
    const updated = brandArray.filter((b) => b.toLowerCase() !== target.toLowerCase());
    setStep2({ ...step2, existing_brands: updated });
  };

  const brandCount = brandArray.length;
  const brandWeightagePts = brandCount >= 3 ? 5 : brandCount >= 1 ? 3 : 0;

  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
            <Briefcase className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">Step 2: Business Experience & Distribution</h2>
            <p className="text-xs text-slate-400">Specify distribution reach, portfolio & historical trade metrics</p>
          </div>
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

        {/* Dynamic Brands Card Section */}
        <div className="p-5 bg-slate-800/40 border border-slate-700/80 rounded-2xl space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Tag className="w-4 h-4 text-indigo-400" />
              <label className="text-xs font-bold uppercase text-slate-200 tracking-wider">
                Brands Currently Distributed
              </label>
            </div>
            <div className={`px-2.5 py-1 rounded-lg text-[11px] font-semibold border flex items-center gap-1.5 ${
              brandWeightagePts === 5
                ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40"
                : brandWeightagePts === 3
                ? "bg-indigo-500/20 text-indigo-300 border-indigo-500/40"
                : "bg-slate-800 text-slate-400 border-slate-700"
            }`}>
              <Sparkles className="w-3 h-3" />
              {/* {brandWeightagePts > 0 ? `+${brandWeightagePts} Score Weightage Boost` : "+5 Max Score Weightage Boost"} */}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="text"
              value={brandInput}
              onChange={(e) => setBrandInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleAddBrand(brandInput);
                }
              }}
              placeholder="Type brand name (e.g. Amul, Britannia) and hit Enter"
              className="flex-1 px-4 py-2.5 bg-slate-900/90 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500 placeholder:text-slate-500"
            />
            <button
              type="button"
              onClick={() => handleAddBrand(brandInput)}
              className="px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl text-xs font-semibold flex items-center gap-1.5 shadow-lg shadow-indigo-600/30 transition-all"
            >
              <Plus className="w-4 h-4" /> Add Brand
            </button>
          </div>

          {/* Render Active Dynamic Cards */}
          {brandArray.length > 0 ? (
            <div className="flex flex-wrap gap-2 pt-1">
              {brandArray.map((brand, idx) => (
                <div
                  key={idx}
                  className="group flex items-center gap-2 px-3.5 py-2 bg-gradient-to-r from-indigo-950/60 to-slate-900 border border-indigo-500/30 rounded-xl text-xs font-semibold text-indigo-200 shadow-md hover:border-indigo-400 transition-all"
                >
                  <Tag className="w-3.5 h-3.5 text-indigo-400" />
                  <span>{brand}</span>
                  <button
                    type="button"
                    onClick={() => handleRemoveBrand(brand)}
                    className="p-0.5 hover:bg-indigo-500/30 rounded text-slate-400 hover:text-white transition-colors"
                  >
                    <X className="w-3.5 h-3.5" />
                  </button>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-xs text-slate-500 italic">No brands added yet. Add brands above or select from quick suggestions below to increase your credit score weightage.</p>
          )}

          {/* Quick Suggestions */}
          <div className="pt-2 border-t border-slate-700/50">
            <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block mb-2">Quick Add Suggestions:</span>
            <div className="flex flex-wrap gap-1.5">
              {POPULAR_SUGGESTIONS.map((sug) => {
                const isAdded = brandArray.some((b) => b.toLowerCase() === sug.toLowerCase());
                return (
                  <button
                    key={sug}
                    type="button"
                    disabled={isAdded}
                    onClick={() => handleAddBrand(sug)}
                    className={`px-2.5 py-1 rounded-lg text-xs font-medium transition-all flex items-center gap-1 ${
                      isAdded
                        ? "bg-slate-800 text-slate-600 border border-slate-800 cursor-not-allowed opacity-50"
                        : "bg-slate-800/80 hover:bg-indigo-600/20 text-slate-300 hover:text-indigo-300 border border-slate-700 hover:border-indigo-500/40"
                    }`}
                  >
                    <Plus className="w-3 h-3" /> {sug}
                  </button>
                );
              })}
            </div>
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
            <span>Continue to Credit Preference</span>
            <ArrowRight className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};
