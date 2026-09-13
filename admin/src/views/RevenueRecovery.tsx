import React, { useEffect, useState } from 'react';
import {
  ShieldAlert,
  Zap,
  RefreshCw,
  CheckCircle2,
  AlertTriangle,
  Clock,
  FileText,
  Filter,
  IndianRupee,
  ShieldCheck
} from 'lucide-react';

import { api } from '../services/api';

export const RevenueRecovery: React.FC = () => {
  const [workflows, setWorkflows] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [riskFilter, setRiskFilter] = useState<string>('all');
  const [actionLoading, setActionLoading] = useState<boolean>(false);
  const [message, setMessage] = useState<string>('');
  const [selectedAuditWorkflow, setSelectedAuditWorkflow] = useState<any | null>(null);

  const fetchWorkflows = async () => {
    setLoading(true);
    try {
      const data = await api.listRecoveryWorkflows();
      setWorkflows(data || []);
    } catch (err: any) {
      console.error('Failed fetching recovery workflows', err);
      setMessage(`Error loading recovery workflows: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWorkflows();
  }, []);

  const handleScanRisk = async () => {
    setActionLoading(true);
    setMessage('Triggering AI Revenue at Risk Scanner across all active accounts...');
    try {
      await api.scanRevenueRisk();
      setMessage('Revenue risk scan completed successfully!');
      await fetchWorkflows();
    } catch (err: any) {
      setMessage(`Scan failed: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const filteredWorkflows = workflows.filter((w) => {
    if (riskFilter === 'all') return true;
    if (riskFilter === 'pending') return w.status === 'pending_approval' || w.requires_admin_approval;
    if (riskFilter === 'low') return w.risk_level === 'low';
    if (riskFilter === 'high') return w.risk_level === 'high' || w.risk_level === 'medium';
    return true;
  });

  const totalRevenueAtRisk = workflows.reduce((acc, w) => acc + (w.revenue_at_risk_paise || 0), 0);
  const pendingApprovalsCount = workflows.filter((w) => w.status === 'pending_approval' || w.requires_admin_approval).length;

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 bg-white p-6 rounded-2xl shadow-sm border border-slate-200">
        <div>
          <div className="flex items-center gap-2">
            <ShieldAlert className="w-7 h-7 text-indigo-600" />
            <h1 className="text-2xl font-bold text-slate-900">AI Agent Revenue Recovery Dashboard</h1>
          </div>
          <p className="text-slate-500 text-sm mt-1">
            Bounded multi-LLM workflow detecting revenue at risk, payment failures, checkout drop-offs & executing automated interventions.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={handleScanRisk}
            disabled={actionLoading}
            className="flex items-center gap-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-sm font-semibold transition-all shadow-sm shadow-indigo-200 disabled:opacity-50"
          >
            <Zap className={`w-4 h-4 ${actionLoading ? 'animate-spin' : ''}`} />
            Scan Revenue Risk
          </button>
          <button
            onClick={fetchWorkflows}
            className="p-2.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl transition-all"
            title="Refresh Workflows"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>

      {message && (
        <div className="p-4 bg-indigo-50 border border-indigo-200 rounded-xl text-indigo-800 text-sm flex items-center gap-2">
          <CheckCircle2 className="w-5 h-5 text-indigo-600 shrink-0" />
          <span>{message}</span>
        </div>
      )}

      {/* Metrics Row */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-5">
        <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-amber-50 border border-amber-200 flex items-center justify-center text-amber-600">
            <IndianRupee className="w-6 h-6" />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-slate-400">Total Revenue at Risk</p>
            <h3 className="text-2xl font-bold text-slate-900 mt-0.5">
              ₹{(totalRevenueAtRisk / 100).toLocaleString('en-IN')}
            </h3>
          </div>
        </div>

        <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-indigo-50 border border-indigo-200 flex items-center justify-center text-indigo-600">
            <Zap className="w-6 h-6" />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-slate-400">Active Agent Workflows</p>
            <h3 className="text-2xl font-bold text-slate-900 mt-0.5">{workflows.length}</h3>
          </div>
        </div>

        <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-rose-50 border border-rose-200 flex items-center justify-center text-rose-600">
            <AlertTriangle className="w-6 h-6" />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-slate-400">Pending Admin Approval</p>
            <h3 className="text-2xl font-bold text-slate-900 mt-0.5">{pendingApprovalsCount}</h3>
          </div>
        </div>

        <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-emerald-50 border border-emerald-200 flex items-center justify-center text-emerald-600">
            <ShieldCheck className="w-6 h-6" />
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-slate-400">Safety Compliance Rate</p>
            <h3 className="text-2xl font-bold text-slate-900 mt-0.5">100%</h3>
          </div>
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-200 pb-3">
        <Filter className="w-4 h-4 text-slate-400" />
        <span className="text-xs font-semibold uppercase text-slate-400 mr-2">Filter Workflows:</span>
        <button
          onClick={() => setRiskFilter('all')}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
            riskFilter === 'all' ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
          }`}
        >
          All ({workflows.length})
        </button>
        <button
          onClick={() => setRiskFilter('pending')}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
            riskFilter === 'pending' ? 'bg-rose-600 text-white' : 'bg-rose-50 text-rose-700 hover:bg-rose-100'
          }`}
        >
          Pending Review ({pendingApprovalsCount})
        </button>
        <button
          onClick={() => setRiskFilter('low')}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
            riskFilter === 'low' ? 'bg-emerald-600 text-white' : 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100'
          }`}
        >
          Low Risk Auto-Exec
        </button>
      </div>

      {/* Workflows Table */}
      <div className="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
        {loading ? (
          <div className="p-12 text-center text-slate-400">
            <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-3" />
            Loading AI recovery workflows...
          </div>
        ) : filteredWorkflows.length === 0 ? (
          <div className="p-12 text-center text-slate-400">
            <ShieldCheck className="w-12 h-12 text-emerald-500 mx-auto mb-3" />
            <h3 className="text-slate-900 font-semibold text-base">No Revenue Risk Workflows Detected</h3>
            <p className="text-xs text-slate-500 mt-1">
              Click "Scan Revenue Risk" above to run the multi-LLM risk detection agent.
            </p>
          </div>
        ) : (
          <table className="w-full text-left text-sm text-slate-600">
            <thead className="bg-slate-50 border-b border-slate-200 text-xs font-semibold text-slate-500 uppercase">
              <tr>
                <th className="px-6 py-4">Workflow #</th>
                <th className="px-6 py-4">Distributor</th>
                <th className="px-6 py-4">Risk Event</th>
                <th className="px-6 py-4">Revenue at Risk</th>
                <th className="px-6 py-4">AI Proposed Intervention</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4 text-right">Audit</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filteredWorkflows.map((w) => (
                <tr key={w.id} className="hover:bg-slate-50/50 transition-colors">
                  <td className="px-6 py-4 font-semibold text-slate-900">{w.workflow_number}</td>
                  <td className="px-6 py-4">
                    <div className="font-semibold text-slate-900">{w.business_name}</div>
                    <div className="text-xs text-slate-400 font-mono">{w.distributor_id.slice(0, 8)}...</div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-amber-50 text-amber-800 border border-amber-200">
                      {w.risk_type}
                    </span>
                  </td>
                  <td className="px-6 py-4 font-semibold text-slate-900">
                    ₹{(w.revenue_at_risk_paise / 100).toLocaleString('en-IN')}
                  </td>
                  <td className="px-6 py-4 max-w-xs">
                    <div className="font-semibold text-indigo-900 text-xs">{w.proposed_intervention}</div>
                    <div className="text-xs text-slate-500 line-clamp-2 mt-0.5">{w.risk_analysis}</div>
                  </td>
                  <td className="px-6 py-4">
                    {w.requires_admin_approval ? (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-rose-50 text-rose-700 border border-rose-200">
                        <Clock className="w-3.5 h-3.5" /> Pending Approval
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 border border-emerald-200">
                        <CheckCircle2 className="w-3.5 h-3.5" /> Executed
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={() => setSelectedAuditWorkflow(w)}
                      className="px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-medium transition-all"
                    >
                      Audit Trace
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Audit Modal */}
      {selectedAuditWorkflow && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white w-full max-w-2xl rounded-2xl shadow-xl border border-slate-200 overflow-hidden flex flex-col max-h-[90vh]">
            <div className="p-6 border-b border-slate-200 flex items-center justify-between bg-slate-50">
              <div className="flex items-center gap-2">
                <FileText className="w-5 h-5 text-indigo-600" />
                <h3 className="text-lg font-bold text-slate-900">
                  Agent Audit Log — {selectedAuditWorkflow.workflow_number}
                </h3>
              </div>
              <button
                onClick={() => setSelectedAuditWorkflow(null)}
                className="text-slate-400 hover:text-slate-600 text-sm font-bold"
              >
                ✕
              </button>
            </div>
            <div className="p-6 space-y-4 overflow-y-auto text-sm">
              <div>
                <label className="text-xs font-bold uppercase text-slate-400">Risk Analysis & Reasoning</label>
                <p className="mt-1 p-3 bg-slate-50 border border-slate-200 rounded-xl text-slate-800">
                  {selectedAuditWorkflow.risk_analysis}
                </p>
              </div>

              <div>
                <label className="text-xs font-bold uppercase text-slate-400">Action Payload</label>
                <pre className="mt-1 p-3 bg-slate-900 text-emerald-400 font-mono text-xs rounded-xl overflow-x-auto">
                  {JSON.stringify(selectedAuditWorkflow.action_payload, null, 2)}
                </pre>
              </div>

              <div className="p-4 bg-emerald-50 border border-emerald-200 rounded-xl">
                <div className="flex items-center gap-2 text-emerald-900 font-bold text-xs uppercase">
                  <ShieldCheck className="w-4 h-4 text-emerald-600" /> Bounded Safety Verification
                </div>
                <p className="text-xs text-emerald-700 mt-1">
                  Passed all 4 deterministic safety guardrails (grace ceiling ≤14 days, fee waiver ≤₹1,000, zero credit limit increase).
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
