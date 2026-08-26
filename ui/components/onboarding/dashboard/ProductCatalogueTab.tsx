import React from "react";
import { ShoppingBag } from "lucide-react";
import { ProductItem } from "@/types/onboarding";

interface ProductCatalogueTabProps {
  regularProducts: ProductItem[];
}

export const ProductCatalogueTab: React.FC<ProductCatalogueTabProps> = ({ regularProducts }) => {
  return (
    <div className="bg-slate-900/70 border border-slate-800 rounded-2xl sm:rounded-3xl p-4 sm:p-6 shadow-xl space-y-5">
      <div>
        <h3 className="text-base sm:text-lg font-bold text-white flex items-center gap-2">
          <ShoppingBag className="w-5 h-5 text-indigo-400" /> Product Catalogue & Wholesale Requirements
        </h3>
        <p className="text-xs text-slate-400 mt-1">
          Browse approved distributor products with minimum order quantities (MOQ).
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3.5 sm:gap-4">
        {regularProducts.map((p) => (
          <div
            key={p.id}
            className="p-4 sm:p-5 bg-slate-950/80 border border-slate-800 rounded-xl sm:rounded-2xl space-y-3 hover:border-indigo-500/40 transition-all flex flex-col justify-between"
          >
            <div className="space-y-2">
              <div className="flex justify-between items-start gap-2">
                <div>
                  <h4 className="font-bold text-white text-sm leading-snug">{p.name}</h4>
                  <span className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block mt-0.5">
                    {p.category}
                  </span>
                </div>
                <span className="text-[11px] font-bold text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20 shrink-0">
                  MOQ: {p.moq || 10}
                </span>
              </div>
            </div>

            <div className="pt-2 border-t border-slate-800/80 flex items-center justify-between gap-2">
              <div>
                <div className="text-[10px] text-slate-500">Unit Price</div>
                <div className="text-sm sm:text-base font-black text-white">
                  ₹{(p.price_paise / 100).toLocaleString("en-IN")}
                </div>
              </div>

              <button
                type="button"
                className="px-3.5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white text-xs font-bold shadow-md shadow-indigo-600/20 transition-all min-h-[38px] touch-manipulation"
              >
                Order Stock
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
