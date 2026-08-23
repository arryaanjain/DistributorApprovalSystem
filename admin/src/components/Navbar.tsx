import React from 'react';
import { Search, Bell, ShieldCheck } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export const Navbar: React.FC = () => {
  const { user } = useAuth();

  return (
    <header className="h-16 glass-panel border-b border-slate-800/80 px-8 flex items-center justify-between sticky top-0 z-20">
      {/* Search Input */}
      <div className="relative w-72">
        <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          placeholder="Search GST, PAN, Distributor..."
          className="w-full bg-slate-900/60 border border-slate-700/60 rounded-xl pl-10 pr-4 py-1.5 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition-colors"
        />
      </div>

      {/* Right Controls */}
      <div className="flex items-center gap-4">
        {/* Environment Badge */}
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 text-xs font-semibold">
          <ShieldCheck className="w-3.5 h-3.5 text-indigo-400" />
          <span>Production Pipeline</span>
        </div>

        {/* Notifications */}
        <button className="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition-colors relative">
          <Bell className="w-4 h-4" />
          <span className="w-2 h-2 rounded-full bg-indigo-500 absolute top-1.5 right-1.5" />
        </button>

        {/* User Profile */}
        <div className="flex items-center gap-3 pl-2 border-l border-slate-800">
          <div className="w-8 h-8 rounded-full bg-indigo-600/30 border border-indigo-500/40 flex items-center justify-center text-indigo-300 font-bold text-xs">
            {user?.name ? user.name.charAt(0) : 'A'}
          </div>
          <div className="hidden sm:block">
            <p className="text-xs font-bold text-slate-200">{user?.name || 'Risk Officer'}</p>
            <p className="text-[10px] text-slate-400 capitalize">{user?.role || 'Credit Manager'}</p>
          </div>
        </div>
      </div>
    </header>
  );
};
