import React, { useEffect, useState } from 'react';
import { RefreshCw, CheckCircle2, Code } from 'lucide-react';
import { api } from '../services/api';

export const Policy: React.FC = () => {
  const [policyData, setPolicyData] = useState<any | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [reloading, setReloading] = useState<boolean>(false);
  const [message, setMessage] = useState<string>('');

  const loadPolicy = async () => {
    setLoading(true);
    try {
      const data = await api.getPolicy();
      setPolicyData(data);
    } catch (err: any) {
      console.error('Failed loading policy', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPolicy();
  }, []);

  const handleReload = async () => {
    setReloading(true);
    setMessage('');
    try {
      const res = await api.reloadPolicy();
      setPolicyData(res.policy || res);
      setMessage('Active credit policy hot-reloaded successfully from database!');
    } catch (err: any) {
      setMessage(`Policy reload error: ${err.message}`);
    } finally {
      setReloading(false);
    }
  };

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Credit Risk & Underwriting Policy</h1>
          <p className="text-sm text-slate-400 mt-1">Live policy score matrix, credit limits ladder, and hot-reload controls.</p>
        </div>
        <button
          onClick={handleReload}
          disabled={reloading}
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-lg shadow-indigo-600/20 transition-all"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${reloading ? 'animate-spin' : ''}`} />
          <span>Hot Reload Policy</span>
        </button>
      </div>

      {message && (
        <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
          <span>{message}</span>
        </div>
      )}

      {/* Policy Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="glass-panel p-6 rounded-2xl border border-slate-800">
          <p className="text-xs font-bold text-slate-400 uppercase tracking-wider">Active Version</p>
          <p className="text-2xl font-black text-white mt-2">{policyData?.Version || policyData?.version || 'v2.1-PROD'}</p>
          <p className="text-xs text-emerald-400 mt-2 flex items-center gap-1 font-medium">
            <CheckCircle2 className="w-3.5 h-3.5" /> Enforced by Rule Engine
          </p>
        </div>

        <div className="glass-panel p-6 rounded-2xl border border-slate-800">
          <p className="text-xs font-bold text-slate-400 uppercase tracking-wider">Default Max Sanction</p>
          <p className="text-2xl font-black text-indigo-300 mt-2">₹50,00,000</p>
          <p className="text-xs text-slate-400 mt-2">Per distributor evaluation cap</p>
        </div>

        <div className="glass-panel p-6 rounded-2xl border border-slate-800">
          <p className="text-xs font-bold text-slate-400 uppercase tracking-wider">Hard Knockout Rules</p>
          <p className="text-2xl font-black text-rose-400 mt-2">4 Guards</p>
          <p className="text-xs text-slate-400 mt-2">CIBIL Default, Write-off, Invalid PAN, GST Inactive</p>
        </div>
      </div>

      {/* JSON Viewer */}
      <div className="glass-panel p-6 rounded-2xl border border-slate-800 space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
            <Code className="w-4 h-4 text-indigo-400" /> Loaded Policy Document (JSON)
          </h2>
        </div>
        {loading ? (
          <div className="py-12 text-center text-slate-500 text-sm">Loading policy...</div>
        ) : (
          <pre className="bg-slate-950 p-4 rounded-xl border border-slate-800 text-xs font-mono text-indigo-300 overflow-x-auto max-h-[50vh]">
            {JSON.stringify(policyData, null, 2)}
          </pre>
        )}
      </div>
    </div>
  );
};
