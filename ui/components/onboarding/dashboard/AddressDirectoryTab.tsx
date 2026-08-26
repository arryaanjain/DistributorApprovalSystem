import React from "react";
import { MapPin } from "lucide-react";

export const AddressDirectoryTab: React.FC = () => {
  return (
    <div className="bg-slate-900/70 border border-slate-800 rounded-2xl sm:rounded-3xl p-4 sm:p-6 shadow-xl space-y-5">
      <div>
        <h3 className="text-base sm:text-lg font-bold text-white flex items-center gap-2">
          <MapPin className="w-5 h-5 text-amber-400" /> Delivery Address Directory
        </h3>
        <p className="text-xs text-slate-400 mt-1">Registered warehouse and commercial shipping destinations.</p>
      </div>

      <div className="p-4 sm:p-5 bg-slate-950/80 border border-amber-500/20 rounded-xl sm:rounded-2xl space-y-2">
        <div className="flex flex-wrap justify-between items-center gap-2">
          <span className="text-xs font-bold text-amber-400 uppercase tracking-wider">
            Primary Shipping Address
          </span>
          <span className="text-[10px] bg-amber-500/10 text-amber-300 border border-amber-500/20 px-2.5 py-0.5 rounded-full font-semibold">
            Default Destination
          </span>
        </div>
        <p className="text-xs sm:text-sm font-medium text-white">Registered Distributor Warehouse Facility</p>
        <p className="text-xs text-slate-400 leading-relaxed">
          Saved during sample booking and verified via statutory onboarding.
        </p>
      </div>
    </div>
  );
};
