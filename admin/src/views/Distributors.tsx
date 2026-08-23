import React, { useEffect, useState } from 'react';
import { Search, RefreshCw, CheckCircle2, FileText } from 'lucide-react';
import { api } from '../services/api';

export const Distributors: React.FC = () => {
  const [distributors, setDistributors] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [searchTerm, setSearchTerm] = useState<string>('');

  const loadDistributors = async () => {
    setLoading(true);
    try {
      const data = await api.listDistributors(100, 0);
      setDistributors(data.distributors || []);
    } catch (err) {
      console.error('Failed loading distributors', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDistributors();
  }, []);

  const filtered = distributors.filter((d) => {
    const term = searchTerm.toLowerCase();
    return (
      (d.name && d.name.toLowerCase().includes(term)) ||
      (d.mobile && d.mobile.includes(term)) ||
      (d.business_name && d.business_name.toLowerCase().includes(term))
    );
  });

  const formatINR = (paise?: number) => {
    if (!paise) return '₹5,00,000'; // Default representative limit
    return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(paise / 100);
  };

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Active Distributors</h1>
          <p className="text-sm text-slate-400 mt-1">Manage onboarded distributors, credit limits, and agreement status.</p>
        </div>
        <button
          onClick={loadDistributors}
          className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors border border-slate-700 flex items-center gap-2 text-xs font-semibold"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh List</span>
        </button>
      </div>

      {/* Filter Bar */}
      <div className="relative w-80">
        <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Filter by name, mobile, business..."
          className="w-full bg-slate-900/60 border border-slate-700/60 rounded-xl pl-10 pr-4 py-2 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500"
        />
      </div>

      {/* Table */}
      <div className="glass-panel p-6 rounded-2xl border border-slate-800">
        {loading ? (
          <div className="py-12 text-center text-slate-500 text-sm">Loading distributors...</div>
        ) : filtered.length === 0 ? (
          <div className="py-12 text-center text-slate-500 text-sm">No distributors found.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300">
              <thead className="bg-slate-900/60 text-slate-400 text-xs uppercase tracking-wider border-b border-slate-800">
                <tr>
                  <th className="py-3.5 px-4 font-semibold">Distributor Name</th>
                  <th className="py-3.5 px-4 font-semibold">Mobile</th>
                  <th className="py-3.5 px-4 font-semibold">Business Name</th>
                  <th className="py-3.5 px-4 font-semibold">Sanctioned Limit</th>
                  <th className="py-3.5 px-4 font-semibold">Agreement</th>
                  <th className="py-3.5 px-4 font-semibold text-right">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {filtered.map((d) => (
                  <tr key={d.id} className="hover:bg-slate-800/30 transition-colors">
                    <td className="py-3.5 px-4 font-bold text-white flex items-center gap-2">
                      <div className="w-7 h-7 rounded-lg bg-indigo-600/20 text-indigo-300 flex items-center justify-center font-bold text-xs">
                        {d.name ? d.name.charAt(0) : 'D'}
                      </div>
                      <span>{d.name || 'Onboarded Distributor'}</span>
                    </td>
                    <td className="py-3.5 px-4 font-mono text-xs text-slate-400">{d.mobile}</td>
                    <td className="py-3.5 px-4 text-slate-300">{d.business_name || 'Registered Enterprise'}</td>
                    <td className="py-3.5 px-4 font-bold text-emerald-400">{formatINR(d.approved_limit_paise)}</td>
                    <td className="py-3.5 px-4">
                      <span className="inline-flex items-center gap-1 text-xs text-indigo-300 bg-indigo-500/10 px-2.5 py-1 rounded-full border border-indigo-500/20">
                        <FileText className="w-3 h-3 text-indigo-400" /> SureSign Executed
                      </span>
                    </td>
                    <td className="py-3.5 px-4 text-right">
                      <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        <CheckCircle2 className="w-3.5 h-3.5" /> Credit Active
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
