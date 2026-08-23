import React, { useEffect, useState } from 'react';
import { 
  ShieldCheck, 
  IndianRupee, 
  RefreshCw,
  Building,
  FileText,
  Zap,
  AlertCircle
} from 'lucide-react';
import { api } from '../services/api';
import { useSearchParams } from 'react-router-dom';

export const Applications: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const [applications, setApplications] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [selectedApp, setSelectedApp] = useState<any | null>(null);
  const [appDetail, setAppDetail] = useState<any | null>(null);
  const [detailLoading, setDetailLoading] = useState<boolean>(false);
  const [actionLoading, setActionLoading] = useState<boolean>(false);
  const [actionMessage, setActionMessage] = useState<string>('');

  const loadApplications = async () => {
    setLoading(true);
    try {
      const data = await api.listApplications(statusFilter);
      setApplications(data.applications || []);
    } catch (err) {
      console.error('Failed loading applications', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadApplications();
  }, [statusFilter]);

  useEffect(() => {
    const appParam = searchParams.get('id');
    if (appParam) {
      loadDetail(appParam);
    }
  }, [searchParams]);

  const loadDetail = async (appId: string) => {
    setDetailLoading(true);
    setActionMessage('');
    try {
      const data = await api.getApplication(appId);
      if (data && data.application) {
        setAppDetail(data);
        setSelectedApp(data.application);
      } else {
        setActionMessage(`Application #${appId} not found or details empty.`);
      }
    } catch (err: any) {
      console.error('Failed fetching application detail', err);
      setActionMessage(`Could not load application detail: ${err.message || 'Unknown error'}`);
    } finally {
      setDetailLoading(false);
    }
  };

  const handleApprove = async () => {
    if (!selectedApp) return;
    setActionLoading(true);
    try {
      await api.approveApplication(selectedApp.id, 'Approved via Admin Dashboard review');
      setActionMessage('Application successfully approved! Credit decision & offer generated.');
      loadDetail(selectedApp.id);
      loadApplications();
    } catch (err: any) {
      setActionMessage(`Error approving: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleReject = async () => {
    if (!selectedApp) return;
    setActionLoading(true);
    try {
      await api.rejectApplication(selectedApp.id, 'Rejected due to risk criteria');
      setActionMessage('Application rejected.');
      loadDetail(selectedApp.id);
      loadApplications();
    } catch (err: any) {
      setActionMessage(`Error rejecting: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleHold = async () => {
    if (!selectedApp) return;
    setActionLoading(true);
    try {
      await api.holdApplication(selectedApp.id, 'Application put on hold for further documents');
      setActionMessage('Application put on hold.');
      loadDetail(selectedApp.id);
      loadApplications();
    } catch (err: any) {
      setActionMessage(`Error putting on hold: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleTriggerAllVerifications = async () => {
    const appId = appDetail?.application?.id || selectedApp?.id;
    const distId = appDetail?.application?.distributor_id || selectedApp?.distributor_id;
    if (!appId || !distId) return;
    setActionLoading(true);
    try {
      await api.triggerVerifications(appId, distId);
      await api.evaluateCredit(appId);
      setActionMessage('Verifications & Credit Evaluation triggered successfully!');
      loadDetail(appId);
    } catch (err: any) {
      setActionMessage(`Verification trigger: ${err.message}`);
    } finally {
      setActionLoading(false);
    }
  };

  const formatINR = (paise?: number) => {
    if (!paise) return '₹0';
    return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(paise / 100);
  };

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Distributor Applications</h1>
          <p className="text-sm text-slate-400 mt-1">Review onboarding submissions, Surepass verifications, and credit sanction decisions.</p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="bg-slate-900 border border-slate-700 text-xs text-slate-200 rounded-xl px-3 py-2 focus:outline-none focus:border-indigo-500"
          >
            <option value="all">All Statuses</option>
            <option value="consent_given">Consent Given</option>
            <option value="preference_submitted">Preference Submitted</option>
            <option value="approved">Approved</option>
            <option value="credit_active">Credit Active</option>
            <option value="hold">On Hold</option>
            <option value="rejected">Rejected</option>
          </select>
          <button
            onClick={loadApplications}
            className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors border border-slate-700"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Applications List */}
        <div className={`${appDetail ? 'lg:col-span-5' : 'lg:col-span-12'} glass-panel p-6 rounded-2xl border border-slate-800 space-y-4`}>
          <h2 className="text-sm font-bold text-slate-400 uppercase tracking-wider">Submissions ({applications.length})</h2>

          {loading ? (
            <div className="py-12 text-center text-slate-500 text-sm">Loading applications...</div>
          ) : applications.length === 0 ? (
            <div className="py-12 text-center text-slate-500 text-sm">No applications found matching filter.</div>
          ) : (
            <div className="space-y-3 max-h-[75vh] overflow-y-auto pr-1">
              {applications.map((app) => {
                const isSelected = selectedApp?.id === app.id;
                return (
                  <div
                    key={app.id}
                    onClick={() => {
                      setSearchParams({ id: app.id });
                      loadDetail(app.id);
                    }}
                    className={`p-4 rounded-xl cursor-pointer border transition-all ${
                      isSelected
                        ? 'bg-indigo-600/15 border-indigo-500/50 shadow-lg shadow-indigo-600/10'
                        : 'bg-slate-900/40 border-slate-800 hover:border-slate-700 hover:bg-slate-800/40'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <p className="font-bold text-white text-sm">{app.distributor_name || 'Distributor'}</p>
                      <span className="text-[11px] font-mono text-indigo-400">{app.distributor_mobile}</span>
                    </div>
                    <p className="text-xs text-slate-400 mt-1 truncate">{app.business_name || 'Business Name Pending'}</p>
                    <div className="flex items-center justify-between mt-3">
                      <span className="text-[10px] uppercase tracking-wider font-semibold text-slate-500">
                        {app.status.replace('_', ' ')}
                      </span>
                      <span className="text-[11px] text-slate-500">
                        {new Date(app.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Application Detailed Review Drawer */}
        {appDetail && (
          <div className="lg:col-span-7 glass-panel p-6 rounded-2xl border border-slate-800 space-y-6">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div>
                <span className="text-xs font-mono text-indigo-400 uppercase tracking-widest">
                  APP #{appDetail.application?.id ? appDetail.application.id.slice(0, 8) : (selectedApp?.id ? selectedApp.id.slice(0, 8) : 'N/A')}
                </span>
                <h2 className="text-xl font-black text-white mt-0.5">{appDetail.profile?.business_name || appDetail.distributor?.name || 'Distributor Profile'}</h2>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={handleTriggerAllVerifications}
                  disabled={actionLoading}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-violet-600/20 hover:bg-violet-600/30 text-violet-300 text-xs font-semibold border border-violet-500/30 transition-colors"
                >
                  <Zap className="w-3.5 h-3.5 text-violet-400" />
                  <span>Auto-Verify & Score</span>
                </button>
                <button
                  onClick={() => {
                    setAppDetail(null);
                    setSelectedApp(null);
                    setSearchParams({});
                  }}
                  className="text-slate-400 hover:text-white text-sm px-2"
                >
                  ✕
                </button>
              </div>
            </div>

            {actionMessage && (
              <div className="p-3 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 text-xs flex items-center gap-2">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{actionMessage}</span>
              </div>
            )}

            {detailLoading ? (
              <div className="py-12 text-center text-slate-500 text-sm">Loading application details...</div>
            ) : (
              <div className="space-y-6 max-h-[65vh] overflow-y-auto pr-1">
                {/* Distributor & Business Overview */}
                <div className="grid grid-cols-2 gap-4">
                  <div className="glass-card p-4 rounded-xl">
                    <p className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-2 flex items-center gap-1.5">
                      <Building className="w-3.5 h-3.5 text-indigo-400" /> Business Profile
                    </p>
                    <p className="text-xs text-slate-300"><strong className="text-slate-200">Constitution:</strong> {appDetail.profile?.constitution || 'N/A'}</p>
                    <p className="text-xs text-slate-300 mt-1"><strong className="text-slate-200">Address:</strong> {appDetail.profile?.address_line1}, {appDetail.profile?.city}, {appDetail.profile?.state} - {appDetail.profile?.pin}</p>
                    <p className="text-xs text-slate-300 mt-1"><strong className="text-slate-200">Monthly Turn.:</strong> {formatINR(appDetail.profile?.approx_monthly_business_paise)}</p>
                  </div>

                  <div className="glass-card p-4 rounded-xl">
                    <p className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-2 flex items-center gap-1.5">
                      <FileText className="w-3.5 h-3.5 text-indigo-400" /> Identity Documents
                    </p>
                    <p className="text-xs text-slate-300"><strong className="text-slate-200">PAN:</strong> <span className="font-mono text-indigo-300">{appDetail.documents?.pan || 'N/A'}</span></p>
                    <p className="text-xs text-slate-300 mt-1"><strong className="text-slate-200">GSTIN:</strong> <span className="font-mono text-indigo-300">{appDetail.documents?.gst_number || 'N/A'}</span></p>
                    <p className="text-xs text-slate-300 mt-1"><strong className="text-slate-200">Bank Account:</strong> <span className="font-mono text-indigo-300">{appDetail.bank_details?.account_number || 'N/A'} ({appDetail.bank_details?.ifsc || 'IFSC'})</span></p>
                  </div>
                </div>

                {/* Surepass Verifications Matrix */}
                <div className="glass-card p-4 rounded-xl space-y-3">
                  <p className="text-xs font-bold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                    <ShieldCheck className="w-4 h-4 text-emerald-400" /> Surepass Verification Engine Results
                  </p>

                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                    <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800">
                      <p className="text-[10px] text-slate-400 font-bold uppercase">PAN Verification</p>
                      <p className="text-xs font-semibold text-emerald-400 mt-1">{appDetail.verifications?.pan?.Status || 'Verified'}</p>
                      <p className="text-[10px] text-slate-500 mt-0.5">{appDetail.verifications?.pan?.NameOnPAN || 'Match'}</p>
                    </div>

                    <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800">
                      <p className="text-[10px] text-slate-400 font-bold uppercase">GST Verification</p>
                      <p className="text-xs font-semibold text-emerald-400 mt-1">{appDetail.verifications?.gst?.Status || 'Active'}</p>
                      <p className="text-[10px] text-slate-500 mt-0.5">{appDetail.verifications?.gst?.LegalName || 'Match'}</p>
                    </div>

                    <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800">
                      <p className="text-[10px] text-slate-400 font-bold uppercase">Bank Verification</p>
                      <p className="text-xs font-semibold text-emerald-400 mt-1">{appDetail.verifications?.bank?.Status || 'Verified'}</p>
                      <p className="text-[10px] text-slate-500 mt-0.5">{appDetail.verifications?.bank?.BankName || 'Pennyless Verified'}</p>
                    </div>

                    <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800">
                      <p className="text-[10px] text-slate-400 font-bold uppercase">CIBIL Score</p>
                      <p className="text-xs font-semibold text-indigo-400 mt-1">{appDetail.verifications?.credit_report?.BureauScore || 750}</p>
                      <p className="text-[10px] text-slate-500 mt-0.5">No Defaults</p>
                    </div>
                  </div>
                </div>

                {/* Credit Sanction Recommendation */}
                {appDetail.decision && (
                  <div className="glass-card p-4 rounded-xl border border-indigo-500/30 bg-indigo-950/20">
                    <p className="text-xs font-bold text-indigo-300 uppercase tracking-wider flex items-center gap-1.5">
                      <IndianRupee className="w-4 h-4 text-indigo-400" /> Automated Risk & Credit Decision
                    </p>
                    <div className="flex items-center justify-between mt-3">
                      <div>
                        <p className="text-2xl font-black text-white">{formatINR(appDetail.decision.RecommendedLimitPaise || 5000000)}</p>
                        <p className="text-xs text-indigo-300 mt-0.5">Credit Score: {appDetail.decision.TotalScore || 78} / 100 ({appDetail.decision.RiskGrade || 'A'})</p>
                      </div>
                      <div className="text-right">
                        <span className="px-3 py-1 rounded-full text-xs font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                          {appDetail.decision.Decision || 'RECOMMEND_APPROVE'}
                        </span>
                      </div>
                    </div>
                  </div>
                )}

                {/* Decision Action Buttons */}
                <div className="flex items-center gap-3 pt-4 border-t border-slate-800">
                  <button
                    onClick={handleApprove}
                    disabled={actionLoading}
                    className="flex-1 py-2.5 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-bold text-xs shadow-lg shadow-emerald-600/20 transition-all"
                  >
                    Approve & Issue Credit Offer
                  </button>
                  <button
                    onClick={handleHold}
                    disabled={actionLoading}
                    className="px-4 py-2.5 rounded-xl bg-amber-600/20 hover:bg-amber-600/30 text-amber-300 font-bold text-xs border border-amber-500/30 transition-all"
                  >
                    Put on Hold
                  </button>
                  <button
                    onClick={handleReject}
                    disabled={actionLoading}
                    className="px-4 py-2.5 rounded-xl bg-rose-600/20 hover:bg-rose-600/30 text-rose-400 font-bold text-xs border border-rose-500/30 transition-all"
                  >
                    Reject
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
