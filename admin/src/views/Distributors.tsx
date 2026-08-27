import React, { useEffect, useState } from 'react';
import { Search, RefreshCw, CheckCircle2, Clock, FileText, History, X, ShieldCheck, Award } from 'lucide-react';
import { api } from '../services/api';

export const Distributors: React.FC = () => {
  const [distributors, setDistributors] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [searchTerm, setSearchTerm] = useState<string>('');

  // Credit Trail Modal state
  const [selectedDist, setSelectedDist] = useState<any | null>(null);
  const [trail, setTrail] = useState<any[]>([]);
  const [loadingTrail, setLoadingTrail] = useState<boolean>(false);

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

  const openCreditTrail = async (d: any) => {
    setSelectedDist(d);
    setLoadingTrail(true);
    try {
      const res = await api.getDistributorCreditTrail(d.id);
      setTrail(res.trail || []);
    } catch (err) {
      console.error('Failed loading credit decision trail', err);
      setTrail([]);
    } finally {
      setLoadingTrail(false);
    }
  };

  const filtered = distributors.filter((d) => {
    const term = searchTerm.toLowerCase();
    return (
      (d.name && d.name.toLowerCase().includes(term)) ||
      (d.mobile && d.mobile.includes(term)) ||
      (d.business_name && d.business_name.toLowerCase().includes(term))
    );
  });

  const formatINR = (paise?: number) => {
    if (!paise || paise === 0) return '₹0';
    return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(paise / 100);
  };

  const formatDate = (isoString?: string) => {
    if (!isoString) return 'N/A';
    const date = new Date(isoString);
    return date.toLocaleDateString('en-IN', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Distributor Directory</h1>
          <p className="text-sm text-slate-400 mt-1">Manage onboarded distributors, sanctioned credit limits, and credit audit trails.</p>
        </div>
        <button
          onClick={loadDistributors}
          className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors border border-slate-700 flex items-center gap-2 text-xs font-semibold"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh Directory</span>
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
          <div className="py-12 text-center text-slate-500 text-sm">Loading distributor directory...</div>
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
                  <th className="py-3.5 px-4 font-semibold">Agreement Status</th>
                  <th className="py-3.5 px-4 font-semibold">Account Status</th>
                  <th className="py-3.5 px-4 font-semibold text-right">Audit Trail</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {filtered.map((d) => (
                  <tr key={d.id} className="hover:bg-slate-800/30 transition-colors">
                    <td className="py-3.5 px-4 font-bold text-white flex items-center gap-2">
                      <div className="w-7 h-7 rounded-lg bg-indigo-600/20 text-indigo-300 flex items-center justify-center font-bold text-xs">
                        {d.name ? d.name.charAt(0) : 'D'}
                      </div>
                      <span>{d.name || 'Onboarding Applicant'}</span>
                    </td>
                    <td className="py-3.5 px-4 font-mono text-xs text-slate-400">{d.mobile}</td>
                    <td className="py-3.5 px-4 text-slate-300">{d.business_name || 'Pending Onboarding'}</td>
                    <td className="py-3.5 px-4 font-bold text-emerald-400">
                      {formatINR(d.approved_limit_paise)}
                    </td>
                    <td className="py-3.5 px-4">
                      {d.agreement_status === 'SIGNED' ? (
                        <span className="inline-flex items-center gap-1 text-xs text-emerald-300 bg-emerald-500/10 px-2.5 py-1 rounded-full border border-emerald-500/20 font-medium">
                          <FileText className="w-3 h-3 text-emerald-400" /> SureSign Executed
                        </span>
                      ) : d.agreement_status === 'GENERATED' ? (
                        <span className="inline-flex items-center gap-1 text-xs text-amber-300 bg-amber-500/10 px-2.5 py-1 rounded-full border border-amber-500/20 font-medium">
                          <Clock className="w-3 h-3 text-amber-400" /> Pending Signature
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 text-xs text-slate-400 bg-slate-800/60 px-2.5 py-1 rounded-full border border-slate-700 font-medium">
                          Not Executed
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 px-4">
                      {d.is_active ? (
                        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          <CheckCircle2 className="w-3.5 h-3.5" /> Credit Active
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                          <Clock className="w-3.5 h-3.5" /> Onboarding / Pending
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 px-4 text-right">
                      <button
                        onClick={() => openCreditTrail(d)}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 text-xs font-semibold transition-colors"
                      >
                        <History className="w-3.5 h-3.5" />
                        <span>Credit Trail</span>
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Credit Limit Trail Modal */}
      {selectedDist && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-2xl overflow-hidden shadow-2xl space-y-0">
            {/* Modal Header */}
            <div className="p-6 border-b border-slate-800 flex items-center justify-between bg-slate-900/80">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
                  <Award className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="text-lg font-bold text-white">Credit Sanction Audit Trail</h3>
                  <p className="text-xs text-slate-400">
                    Historical limit decisions & risk scoring for <span className="text-slate-200 font-semibold">{selectedDist.name || selectedDist.mobile}</span>
                  </p>
                </div>
              </div>
              <button
                onClick={() => setSelectedDist(null)}
                className="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Modal Body */}
            <div className="p-6 max-h-[70vh] overflow-y-auto space-y-4">
              {loadingTrail ? (
                <div className="py-12 text-center text-slate-400 text-sm flex items-center justify-center gap-2">
                  <RefreshCw className="w-4 h-4 animate-spin text-indigo-400" />
                  <span>Fetching credit decision history...</span>
                </div>
              ) : trail.length === 0 ? (
                <div className="py-12 text-center text-slate-400 text-sm">
                  No credit decision records found for this distributor yet.
                </div>
              ) : (
                <div className="relative border-l-2 border-slate-800 ml-4 pl-6 space-y-6">
                  {trail.map((item, idx) => (
                    <div key={item.id || idx} className="relative group">
                      {/* Timeline dot */}
                      <div className="absolute -left-[31px] top-1.5 w-3.5 h-3.5 rounded-full bg-indigo-500 border-2 border-slate-900 ring-4 ring-indigo-500/20" />

                      <div className="bg-slate-800/40 border border-slate-800 rounded-xl p-4 space-y-3 hover:border-slate-700 transition-colors">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
                              Decision #{trail.length - idx}
                            </span>
                            <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                              {item.policy_version || 'v1.0'}
                            </span>
                          </div>
                          <span className="text-xs text-slate-500 font-mono">
                            {formatDate(item.decided_at)}
                          </span>
                        </div>

                        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 pt-1">
                          <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800/80">
                            <span className="text-[10px] font-medium text-slate-400 uppercase tracking-wider block">Approved Limit</span>
                            <span className="text-base font-black text-emerald-400">
                              {formatINR(item.approved_limit_paise)}
                            </span>
                          </div>

                          <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800/80">
                            <span className="text-[10px] font-medium text-slate-400 uppercase tracking-wider block">Risk Grade</span>
                            <span className="text-sm font-bold text-indigo-300 flex items-center gap-1 mt-0.5">
                              <ShieldCheck className="w-3.5 h-3.5 text-indigo-400" />
                              Grade {item.risk_grade || 'A'} (Score: {item.total_score || 'N/A'})
                            </span>
                          </div>

                          <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800/80">
                            <span className="text-[10px] font-medium text-slate-400 uppercase tracking-wider block">Payment Terms</span>
                            <span className="text-sm font-semibold text-slate-200 mt-0.5 block">
                              {item.payment_terms || '15 Days Credit'}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Modal Footer */}
            <div className="p-4 border-t border-slate-800 bg-slate-900/80 flex justify-end">
              <button
                onClick={() => setSelectedDist(null)}
                className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition-colors border border-slate-700"
              >
                Close Audit Trail
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

