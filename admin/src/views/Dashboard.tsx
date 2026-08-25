import React, { useEffect, useState } from 'react';
import { 
  Users, 
  FileCheck, 
  IndianRupee, 
  ArrowUpRight, 
  CheckCircle2, 
  Clock, 
  XCircle,
  RefreshCw
} from 'lucide-react';
import { api } from '../services/api';
import { useNavigate } from 'react-router-dom';

export const Dashboard: React.FC = () => {
  const [stats, setStats] = useState({
    totalDistributors: 0,
    totalApplications: 0,
    pendingVerifications: 0,
    approvedLimitTotal: 0,
  });
  const [recentApps, setRecentApps] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const navigate = useNavigate();

  const loadDashboardData = async () => {
    setLoading(true);
    try {
      const [appData, distData] = await Promise.all([
        api.listApplications('all', 10, 0).catch(() => ({ applications: [], total: 0 })),
        api.listDistributors(100, 0).catch(() => ({ distributors: [], total: 0 })),
      ]);

      const apps = appData.applications || [];
      const dists = distData.distributors || [];

      const pending = apps.filter((a: any) => 
        ['submitted', 'consent_given', 'preference_submitted', 'hold'].includes(a.status)
      ).length;

      setStats({
        totalDistributors: distData.total || dists.length,
        totalApplications: appData.total || apps.length,
        pendingVerifications: pending,
        approvedLimitTotal: 15000000, // ₹1.5 Cr default limit pool representation
      });

      setRecentApps(apps.slice(0, 5));
    } catch (err) {
      console.error('Failed loading dashboard stats', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDashboardData();
  }, []);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'credit_active':
      case 'approved':
      case 'offer_generated':
      case 'offer_accepted':
      case 'agreement_pending':
      case 'agreement_signed':
        return <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"><CheckCircle2 className="w-3.5 h-3.5" /> Completed</span>;
      case 'advance_only':
        return <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-sky-500/10 text-sky-400 border border-sky-500/20"><CheckCircle2 className="w-3.5 h-3.5" /> Advance Only</span>;
      case 'rejected':
      case 'blocked':
        return <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20"><XCircle className="w-3.5 h-3.5" /> Rejected</span>;
      case 'hold':
        return <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20"><Clock className="w-3.5 h-3.5" /> On Hold</span>;
      default:
        return <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20"><RefreshCw className="w-3.5 h-3.5 animate-spin" /> In Review</span>;
    }
  };

  return (
    <div className="p-8 space-y-8 max-w-7xl mx-auto">
      {/* Header Banner */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Executive Dashboard</h1>
          <p className="text-sm text-slate-400 mt-1">Real-time distributor credit applications & risk scoring overview.</p>
        </div>
        <button
          onClick={loadDashboardData}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold transition-colors border border-slate-700"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh Pipeline</span>
        </button>
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="glass-panel p-6 rounded-2xl border border-slate-800 relative overflow-hidden">
          <div className="flex items-center justify-between mb-4">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Applications</span>
            <div className="w-9 h-9 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
              <FileCheck className="w-5 h-5" />
            </div>
          </div>
          <p className="text-3xl font-black text-white">{stats.totalApplications}</p>
          <div className="mt-3 flex items-center gap-1 text-xs text-emerald-400 font-medium">
            <ArrowUpRight className="w-4 h-4" />
            <span>+12% this week</span>
          </div>
        </div>

        <div className="glass-panel p-6 rounded-2xl border border-slate-800 relative overflow-hidden">
          <div className="flex items-center justify-between mb-4">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">Pending Review</span>
            <div className="w-9 h-9 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400">
              <Clock className="w-5 h-5" />
            </div>
          </div>
          <p className="text-3xl font-black text-amber-400">{stats.pendingVerifications}</p>
          <p className="text-xs text-slate-400 mt-3 font-medium">Auto-verifications running</p>
        </div>

        <div className="glass-panel p-6 rounded-2xl border border-slate-800 relative overflow-hidden">
          <div className="flex items-center justify-between mb-4">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">Active Distributors</span>
            <div className="w-9 h-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
              <Users className="w-5 h-5" />
            </div>
          </div>
          <p className="text-3xl font-black text-white">{stats.totalDistributors}</p>
          <p className="text-xs text-emerald-400 mt-3 font-medium">Active ordering accounts</p>
        </div>

        <div className="glass-panel p-6 rounded-2xl border border-slate-800 relative overflow-hidden">
          <div className="flex items-center justify-between mb-4">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">Sanctioned Credit Pool</span>
            <div className="w-9 h-9 rounded-xl bg-violet-500/10 border border-violet-500/20 flex items-center justify-center text-violet-400">
              <IndianRupee className="w-5 h-5" />
            </div>
          </div>
          <p className="text-3xl font-black text-violet-300">₹1.5 Cr</p>
          <p className="text-xs text-slate-400 mt-3 font-medium">Total sanctioned exposure</p>
        </div>
      </div>

      {/* Recent Pipeline Activity */}
      <div className="glass-panel p-6 rounded-2xl border border-slate-800 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-bold text-white">Recent Application Pipeline</h2>
            <p className="text-xs text-slate-400">Distributors undergoing automated verification & decision scoring.</p>
          </div>
          <button
            onClick={() => navigate('/applications')}
            className="text-xs font-bold text-indigo-400 hover:text-indigo-300 transition-colors"
          >
            View All Applications →
          </button>
        </div>

        {loading ? (
          <div className="py-12 text-center text-slate-500 text-sm">Loading distributor application pipeline...</div>
        ) : recentApps.length === 0 ? (
          <div className="py-12 text-center text-slate-500 text-sm">No applications submitted yet.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm text-slate-300">
              <thead className="bg-slate-900/60 text-slate-400 text-xs uppercase tracking-wider border-b border-slate-800">
                <tr>
                  <th className="py-3.5 px-4 font-semibold">Distributor</th>
                  <th className="py-3.5 px-4 font-semibold">Mobile</th>
                  <th className="py-3.5 px-4 font-semibold">Business Name</th>
                  <th className="py-3.5 px-4 font-semibold">Status</th>
                  <th className="py-3.5 px-4 font-semibold text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {recentApps.map((app) => (
                  <tr key={app.id} className="hover:bg-slate-800/30 transition-colors">
                    <td className="py-3.5 px-4 font-bold text-white">{app.distributor_name || 'New Distributor'}</td>
                    <td className="py-3.5 px-4 text-slate-400 font-mono text-xs">{app.distributor_mobile}</td>
                    <td className="py-3.5 px-4 text-slate-300">{app.business_name || 'Pending Details'}</td>
                    <td className="py-3.5 px-4">{getStatusBadge(app.status)}</td>
                    <td className="py-3.5 px-4 text-right">
                      <button
                        onClick={() => navigate(`/applications?id=${app.id}`)}
                        className="px-3 py-1.5 rounded-lg bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 text-xs font-semibold transition-colors border border-indigo-500/30"
                      >
                        Review
                      </button>
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
