import React, { useEffect, useState } from 'react';
import { 
  ShieldCheck, 
  IndianRupee, 
  RefreshCw,
  Building,
  FileText,
  Zap,
  AlertCircle,
  Download
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
  const [overrideLimit, setOverrideLimit] = useState<string>('');
  const [overrideDays, setOverrideDays] = useState<number>(15);

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
        if (data.decision) {
          const recPaise = data.decision.approved_limit_paise ?? data.decision.ApprovedLimitPaise ?? data.decision.RecommendedLimitPaise ?? 0;
          setOverrideLimit((recPaise / 100).toString());
          setOverrideDays(data.decision.approved_period_days ?? data.decision.ApprovedPeriodDays ?? 15);
        }
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
      const parsedLimitPaise = overrideLimit !== '' ? Math.max(0, Math.round(Number(overrideLimit) * 100)) : undefined;
      await api.approveApplication(selectedApp.id, 'Approved via Admin Dashboard review', parsedLimitPaise, overrideDays);
      setActionMessage('Application successfully approved & completed with granted credit limit!');
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

  const handleFetchCibilReport = async () => {
    const appId =
      appDetail?.application?.id ||
      appDetail?.application?.ID ||
      selectedApp?.id ||
      selectedApp?.ID ||
      searchParams.get('id');

    const distId =
      appDetail?.application?.distributor_id ||
      appDetail?.application?.DistributorID ||
      selectedApp?.distributor_id ||
      selectedApp?.DistributorID ||
      appDetail?.distributor?.id ||
      appDetail?.distributor?.ID;

    if (!appId || !distId) {
      setActionMessage('Cannot fetch CIBIL report: Missing Application ID or Distributor ID.');
      return;
    }
    setActionLoading(true);
    setActionMessage('Fetching CIBIL Credit Bureau Report via Surepass...');
    try {
      await api.triggerVerifications(appId, distId);
      setActionMessage('CIBIL Credit Bureau Report fetched successfully!');
      await loadDetail(appId);
    } catch (err: any) {
      console.error('CIBIL fetch failed:', err);
      setActionMessage(`CIBIL fetch error: ${err.message || 'Unknown error'}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleCalculateCreditScore = async () => {
    const appId =
      appDetail?.application?.id ||
      appDetail?.application?.ID ||
      selectedApp?.id ||
      selectedApp?.ID ||
      searchParams.get('id');

    if (!appId) {
      setActionMessage('Cannot calculate score: Missing Application ID.');
      return;
    }
    setActionLoading(true);
    setActionMessage('Evaluating Credit Scoring Engine...');
    try {
      await api.evaluateCredit(appId);
      setActionMessage('Credit Score & Limit Decision calculated successfully!');
      await loadDetail(appId);
    } catch (err: any) {
      console.error('Credit calculation failed:', err);
      setActionMessage(`Score calculation error: ${err.message || 'Unknown error'}`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleFetchCibilAndCalculateScore = async () => {
    const appId =
      appDetail?.application?.id ||
      appDetail?.application?.ID ||
      selectedApp?.id ||
      selectedApp?.ID ||
      searchParams.get('id');

    const distId =
      appDetail?.application?.distributor_id ||
      appDetail?.application?.DistributorID ||
      selectedApp?.distributor_id ||
      selectedApp?.DistributorID ||
      appDetail?.distributor?.id ||
      appDetail?.distributor?.ID;

    if (!appId || !distId) {
      setActionMessage('Cannot run process: Missing Application ID or Distributor ID.');
      return;
    }
    setActionLoading(true);
    setActionMessage('Fetching CIBIL Report & Calculating Credit Score...');
    try {
      await api.triggerVerifications(appId, distId);
      await api.evaluateCredit(appId);
      setActionMessage('CIBIL Report fetched & Credit Score calculated successfully!');
      await loadDetail(appId);
    } catch (err: any) {
      console.error('Process failed:', err);
      setActionMessage(`Process error: ${err.message || 'Unknown error'}`);
    } finally {
      setActionLoading(false);
    }
  };

  const formatINR = (paise?: number) => {
    if (!paise) return '₹0';
    return new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(paise / 100);
  };

  const getStepInfo = (status: string) => {
    switch (status) {
      case 'submitted':
      case 'basic_submitted':
        return { step: 1, label: 'Step 1: Business Details', badgeClass: 'bg-blue-500/10 text-blue-400 border-blue-500/20' };
      case 'business_submitted':
        return { step: 2, label: 'Step 2: Experience & Ops', badgeClass: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20' };
      case 'preference_submitted':
        return { step: 3, label: 'Step 3: Credit Preference', badgeClass: 'bg-violet-500/10 text-violet-400 border-violet-500/20' };
      case 'trial':
        return { step: 4, label: 'Step 4: Sample Trial Active', badgeClass: 'bg-amber-500/10 text-amber-400 border-amber-500/20' };
      case 'statutory_submitted':
        return { step: 5, label: 'Step 5: KYC & GST Verified', badgeClass: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20' };
      case 'consent_given':
        return { step: 6, label: 'Step 6: Legal Consent', badgeClass: 'bg-sky-500/10 text-sky-400 border-sky-500/20' };
      case 'bank_submitted':
        return { step: 7, label: 'Step 7: Bank Submitted', badgeClass: 'bg-purple-500/10 text-purple-400 border-purple-500/20' };
      case 'under_review':
        return { step: 8, label: 'Step 8: Under Credit Review', badgeClass: 'bg-orange-500/10 text-orange-400 border-orange-500/20' };
      case 'hold':
        return { step: 8, label: 'Step 8: On Hold', badgeClass: 'bg-amber-500/10 text-amber-400 border-amber-500/20' };
      case 'offer_generated':
      case 'offer_accepted':
      case 'agreement_pending':
      case 'agreement_signed':
      case 'approved':
      case 'credit_active':
        return { step: 9, label: 'Step 9: Credit Approved', badgeClass: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' };
      case 'advance_only':
        return { step: 9, label: 'Step 9: Advance Only', badgeClass: 'bg-teal-500/10 text-teal-400 border-teal-500/20' };
      case 'rejected':
      case 'blocked':
        return { step: 0, label: 'Rejected', badgeClass: 'bg-rose-500/10 text-rose-400 border-rose-500/20' };
      default:
        return { step: 1, label: status ? status.replace('_', ' ') : 'In Review', badgeClass: 'bg-slate-500/10 text-slate-400 border-slate-500/20' };
    }
  };

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Distributor Applications</h1>
          <p className="text-sm text-slate-400 mt-1">Review onboarding submissions, step progress, Surepass verifications, and credit sanctions.</p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="bg-slate-900 border border-slate-700 text-xs text-slate-200 rounded-xl px-3 py-2 focus:outline-none focus:border-indigo-500"
          >
            <option value="all">All Applications & Steps</option>
            <option value="submitted">Step 1: Basic Submitted</option>
            <option value="business_submitted">Step 2: Experience Submitted</option>
            <option value="preference_submitted">Step 3: Preference Submitted</option>
            <option value="trial">Step 4: Sample Trial</option>
            <option value="statutory_submitted">Step 5: KYC & GST</option>
            <option value="consent_given">Step 6: Consent Given</option>
            <option value="bank_submitted">Step 7: Bank Details</option>
            <option value="hold">Step 8: On Hold</option>
            <option value="approved">Step 9: Approved / Active</option>
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
                const stepInfo = getStepInfo(app.status);
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
                      <span className={`text-[10px] font-bold px-2.5 py-0.5 rounded-full border ${stepInfo.badgeClass}`}>
                        {stepInfo.label}
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
                  onClick={handleFetchCibilAndCalculateScore}
                  disabled={actionLoading}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-violet-600/20 hover:bg-violet-600/30 text-violet-300 text-xs font-semibold border border-violet-500/30 transition-colors"
                >
                  <Zap className="w-3.5 h-3.5 text-violet-400" />
                  <span>{appDetail?.decision ? 'Recompute Decision' : 'Fetch CIBIL & Compute Score'}</span>
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
                {(() => {
                  const panRec = appDetail.verifications?.pan || appDetail.verifications?.PAN;
                  const gstRec = appDetail.verifications?.gst || appDetail.verifications?.GST;
                  const bankRec = appDetail.verifications?.bank || appDetail.verifications?.Bank;
                  const creditRec = appDetail.verifications?.credit_report || appDetail.verifications?.CreditReport;

                  const panStatus = panRec?.status || panRec?.Status;
                  const panName = panRec?.name_on_pan || panRec?.NameOnPAN;

                  const gstStatus = gstRec?.status || gstRec?.Status;
                  const gstName = gstRec?.legal_name || gstRec?.LegalName;

                  const bankStatus = bankRec?.status || bankRec?.Status;
                  const bankName = bankRec?.bank_name || bankRec?.BankName;

                  const bureauScore = creditRec?.bureau_score ?? creditRec?.BureauScore;
                  const pdfUrl = creditRec?.pdf_url || creditRec?.PDFURL;
                  const hasDefaults = creditRec?.has_defaults ?? creditRec?.HasDefaults;
                  const hasScore = bureauScore !== undefined && bureauScore !== null && bureauScore > 0;

                  const decisionScore = appDetail.decision?.total_score ?? appDetail.decision?.TotalScore;

                  return (
                    <div className="glass-card p-4 rounded-xl space-y-4">
                      <div className="flex items-center justify-between">
                        <p className="text-xs font-bold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                          <ShieldCheck className="w-4 h-4 text-emerald-400" /> Surepass Verification Engine Results
                        </p>
                        {!hasScore && !decisionScore && (
                          <span className="text-[10px] text-amber-400 bg-amber-500/10 border border-amber-500/20 px-2 py-0.5 rounded-full font-medium">
                            Verification Pending
                          </span>
                        )}
                      </div>

                      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                        <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800">
                          <p className="text-[10px] text-slate-400 font-bold uppercase">PAN Verification</p>
                          <p className={`text-xs font-semibold mt-1 ${panStatus === 'verified' ? 'text-emerald-400' : 'text-slate-400'}`}>
                            {panStatus || 'Not Run'}
                          </p>
                          <p className="text-[10px] text-slate-500 mt-0.5 truncate">{panName || (panRec ? 'Verified' : 'Pending')}</p>
                        </div>

                        <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800">
                          <p className="text-[10px] text-slate-400 font-bold uppercase">GST Verification</p>
                          <p className={`text-xs font-semibold mt-1 ${gstStatus === 'verified' || gstStatus === 'partially_verified' ? 'text-emerald-400' : 'text-slate-400'}`}>
                            {gstStatus || 'Not Run'}
                          </p>
                          <p className="text-[10px] text-slate-500 mt-0.5 truncate">{gstName || (gstRec ? 'Active' : 'Pending')}</p>
                        </div>

                        <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800">
                          <p className="text-[10px] text-slate-400 font-bold uppercase">Bank Verification</p>
                          <p className={`text-xs font-semibold mt-1 ${bankStatus === 'verified' ? 'text-emerald-400' : 'text-slate-400'}`}>
                            {bankStatus || 'Not Run'}
                          </p>
                          <p className="text-[10px] text-slate-500 mt-0.5 truncate">{bankName || (bankRec ? 'Penny Drop Verified' : 'Pending')}</p>
                        </div>

                        <div className="bg-slate-900/60 p-3 rounded-lg border border-slate-800 flex flex-col justify-between">
                          <div>
                            <p className="text-[10px] text-slate-400 font-bold uppercase">Credit Score</p>
                            {hasScore ? (
                              <>
                                <p className="text-sm font-black text-indigo-400 mt-1">{bureauScore}</p>
                                <p className="text-[10px] text-slate-500 mt-0.5">
                                  {hasDefaults ? '⚠ Defaults' : '✓ No Defaults'}
                                </p>
                              </>
                            ) : decisionScore !== undefined && decisionScore !== null ? (
                              <>
                                <p className="text-sm font-black text-indigo-400 mt-1">{decisionScore}</p>
                                <p className="text-[10px] text-emerald-400 mt-0.5">Score Calculated</p>
                              </>
                            ) : (
                              <>
                                <p className="text-xs font-semibold text-amber-400/90 mt-1">Not Calculated</p>
                                <p className="text-[10px] text-slate-500 mt-0.5">Manual Run Required</p>
                              </>
                            )}
                          </div>
                          {!hasScore && decisionScore === undefined && (
                            <button
                              onClick={handleCalculateCreditScore}
                              disabled={actionLoading}
                              className="mt-2 text-[10px] px-2 py-1 bg-indigo-600/30 hover:bg-indigo-600/50 text-indigo-200 rounded border border-indigo-500/30 font-bold transition-all"
                            >
                              Calculate Score
                            </button>
                          )}
                        </div>
                      </div>

                      {/* PDF Report Download Action */}
                      {pdfUrl ? (
                        <div className="pt-2 border-t border-slate-800 flex items-center justify-between">
                          <div className="text-[11px] text-slate-400">
                            <span>Official CIBIL Credit Bureau Report (PDF)</span>
                          </div>
                          <div className="flex items-center gap-2">
                            <button
                              onClick={handleFetchCibilReport}
                              disabled={actionLoading}
                              className="px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold border border-slate-700 transition-all"
                            >
                              Refresh CIBIL
                            </button>
                            <a
                              href={pdfUrl}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-600/30 transition-all"
                            >
                              <Download className="w-3.5 h-3.5" />
                              <span>Download Report PDF</span>
                            </a>
                          </div>
                        </div>
                      ) : (
                        <div className="pt-2 border-t border-slate-800 flex items-center justify-between">
                          <div className="text-[11px] text-slate-400">
                            <span className="text-slate-300 font-semibold">CIBIL Report:</span> <span className="text-amber-400 font-mono">Not Fetched Yet</span>
                          </div>
                          <button
                            onClick={handleFetchCibilReport}
                            disabled={actionLoading}
                            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-600/30 transition-all"
                          >
                            <Download className="w-3.5 h-3.5" />
                            <span>Fetch CIBIL Report</span>
                          </button>
                        </div>
                      )}
                    </div>
                  );
                })()}

                {/* Credit Sanction Recommendation & Compute Mechanics */}
                {appDetail.decision ? (() => {
                  const limit = appDetail.decision.approved_limit_paise ?? appDetail.decision.ApprovedLimitPaise ?? appDetail.decision.RecommendedLimitPaise;
                  const score = appDetail.decision.total_score ?? appDetail.decision.TotalScore ?? 0;
                  const grade = appDetail.decision.risk_grade ?? appDetail.decision.RiskGrade ?? 'A';
                  const dec = appDetail.decision.decision ?? appDetail.decision.Decision ?? 'ADVANCE_ONLY';
                  const hardRisk = appDetail.decision.hard_risk_triggered ?? appDetail.decision.HardRiskTriggered ?? false;
                  const flags: string[] = appDetail.risk_flags || [];
                  const scoreComps: Record<string, number> = appDetail.score_components || {};

                  // Helper flag descriptions
                  const getFlagDesc = (flag: string) => {
                    switch (flag) {
                      case 'CREDIT_BUREAU_DEFAULT':
                        return 'CIBIL Bureau Score < 700 with reported credit default / delinquency. Credit hard-limited to ₹0 (ADVANCE_ONLY).';
                      case 'CREDIT_BUREAU_WRITEOFF':
                        return 'CIBIL Bureau report indicates historical loan write-offs. Credit blocked.';
                      case 'BUREAU_FRAUD_FLAG':
                        return 'Bureau fraud indicator flagged on credit report.';
                      case 'INVALID_PAN_IDENTITY':
                        return 'PAN identity verification with tax authority failed or mismatched.';
                      case 'BANK_VERIFICATION_FAILED':
                        return 'Bank account verification failed or name mismatch.';
                      case 'DUPLICATE_APPLICATION_SUSPECT':
                        return 'Suspected duplicate application submission.';
                      default:
                        return 'Hard risk rule triggered by automated decision engine.';
                    }
                  };

                  const compLabels: Record<string, { label: string; max: number }> = {
                    credit_risk: { label: 'CIBIL Credit Risk', max: 30 },
                    identity_kyc: { label: 'PAN Identity / KYC', max: 15 },
                    business_verification: { label: 'GST / Biz Verification', max: 15 },
                    business_vintage: { label: 'Business Vintage', max: 10 },
                    fmcg_experience: { label: 'FMCG / Dist. Experience', max: 10 },
                    business_capacity: { label: 'Turnover Scale & Capacity', max: 10 },
                    data_integrity: { label: 'Name / Data Consistency', max: 10 },
                    brand_portfolio: { label: 'Brand Portfolio', max: 5 },
                    compliance_credentials: { label: 'FSSAI / Udyam Credentials', max: 5 },
                  };

                  return (
                    <div className="glass-card p-5 rounded-2xl border border-indigo-500/30 bg-slate-900/80 space-y-4">
                      {/* Top Header Card */}
                      <div className="flex items-center justify-between pb-3 border-b border-slate-800">
                        <div>
                          <p className="text-[11px] font-bold text-indigo-400 uppercase tracking-wider flex items-center gap-1.5">
                            <IndianRupee className="w-4 h-4 text-indigo-400" /> Automated Risk & Credit Decision
                          </p>
                          <p className="text-3xl font-black text-white mt-1">{formatINR(limit)}</p>
                        </div>
                        <div className="text-right">
                          <span
                            className={`px-3 py-1 rounded-full text-xs font-black uppercase tracking-wider border ${
                              dec === 'APPROVED'
                                ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40'
                                : dec === 'ADVANCE_ONLY'
                                ? 'bg-amber-500/20 text-amber-300 border-amber-500/40'
                                : 'bg-rose-500/20 text-rose-300 border-rose-500/40'
                            }`}
                          >
                            {dec} {hardRisk ? '(Hard Limited)' : ''}
                          </span>
                          <p className="text-xs text-slate-400 mt-1.5 font-medium">
                            Engine Total Score: <strong className="text-indigo-300">{score} / 100</strong> (Grade {grade})
                          </p>
                        </div>
                      </div>

                      {/* Hard Risk Override Banner */}
                      {(hardRisk || flags.length > 0) && (
                        <div className="p-3.5 rounded-xl bg-amber-950/40 border border-amber-500/40 text-amber-200 space-y-2">
                          <div className="flex items-center gap-2 text-xs font-bold text-amber-300 uppercase tracking-wider">
                            <AlertCircle className="w-4 h-4 text-amber-400 shrink-0" />
                            Hard Risk Override Triggered (Credit Capped at ₹0)
                          </div>
                          {flags.length > 0 ? (
                            <ul className="space-y-1 pl-5 list-disc text-xs text-amber-200/90">
                              {flags.map((f, idx) => (
                                <li key={idx}>
                                  <strong className="text-amber-100">{f}:</strong> {getFlagDesc(f)}
                                </li>
                              ))}
                            </ul>
                          ) : (
                            <p className="text-xs text-amber-200/90">
                              Automated risk rules detected critical credit flags (CIBIL defaults or tax mismatches). Line-of-credit is hard limited to Advance Payment only.
                            </p>
                          )}
                        </div>
                      )}

                      {/* Compute Mechanics Score Parameter Breakdown */}
                      {Object.keys(scoreComps).length > 0 && (
                        <div className="pt-2">
                          <p className="text-xs font-bold text-slate-300 mb-2.5 uppercase tracking-wider">
                            Engine Compute Mechanics & Parameter Breakdown
                          </p>
                          <div className="grid grid-cols-2 gap-2.5">
                            {Object.entries(scoreComps).map(([key, val]) => {
                              const meta = compLabels[key] || { label: key, max: 15 };
                              const pct = Math.min(100, Math.round((val / meta.max) * 100));
                              return (
                                <div key={key} className="p-2.5 rounded-lg bg-slate-950/60 border border-slate-800/80">
                                  <div className="flex items-center justify-between text-[11px] font-medium text-slate-300 mb-1">
                                    <span>{meta.label}</span>
                                    <span className="font-bold text-indigo-300">{val} / {meta.max} pts</span>
                                  </div>
                                  <div className="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden">
                                    <div
                                      className="bg-indigo-500 h-full rounded-full transition-all"
                                      style={{ width: `${pct}%` }}
                                    />
                                  </div>
                                </div>
                              );
                            })}
                          </div>
                        </div>
                      )}

                      {/* Credit Calculation Logic Explanation */}
                      <div className="p-3 rounded-xl bg-slate-950/50 border border-slate-800/60 text-[11px] text-slate-400 space-y-1">
                        <p className="font-semibold text-slate-300">Credit Determination Flow:</p>
                        <p>• <strong>Turnover Base:</strong> Conservative 5%–10% of approx. monthly business turnover (capped by Risk Grade score band up to ₹1,50,000 max).</p>
                        <p>• <strong>GST Non-GST Rule:</strong> Capped at max ₹25,000 line of credit if no valid GSTIN registered.</p>
                        <p>• <strong>Hard Override Rule:</strong> If active defaults (&lt; 660 CIBIL), PAN mismatch, or fraud flags exist, credit is hard-capped to ₹0 (`ADVANCE_ONLY`).</p>
                      </div>
                    </div>
                  );
                })() : (
                  <div className="p-5 rounded-2xl border border-slate-800 bg-slate-900/60 text-center space-y-3">
                    <p className="text-sm font-bold text-slate-300">Automated credit limit evaluation pending manual run.</p>
                    <p className="text-xs text-slate-400 max-w-md mx-auto">
                      CIBIL credit report and score calculation are set to manual mode. Click below to fetch the report from Surepass and compute the credit score.
                    </p>
                    <button
                      onClick={handleFetchCibilAndCalculateScore}
                      disabled={actionLoading}
                      className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-bold text-xs shadow-lg shadow-indigo-600/30 transition-all"
                    >
                      <Zap className="w-4 h-4 text-amber-300" />
                      <span>Fetch CIBIL Report & Calculate Score</span>
                    </button>
                  </div>
                )}

                {/* Admin Manual Credit Limit & Terms Override Panel */}
                <div className="p-4 rounded-xl bg-slate-900/90 border border-indigo-500/40 space-y-3">
                  <div className="flex items-center justify-between">
                    <p className="text-xs font-bold text-indigo-300 uppercase tracking-wider flex items-center gap-1.5">
                      <IndianRupee className="w-3.5 h-3.5 text-indigo-400" /> Admin Manual Credit Limit Override
                    </p>
                    <span className="text-[10px] text-slate-400">Custom Limit & Payment Terms</span>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="text-[11px] font-semibold text-slate-400 block mb-1">Approved Credit Limit (₹)</label>
                      <input
                        type="number"
                        value={overrideLimit}
                        onChange={(e) => setOverrideLimit(e.target.value)}
                        placeholder="Enter custom limit in ₹"
                        className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-slate-700 text-white font-bold text-sm focus:outline-none focus:border-indigo-500"
                      />
                    </div>
                    <div>
                      <label className="text-[11px] font-semibold text-slate-400 block mb-1">Payment Period</label>
                      <select
                        value={overrideDays}
                        onChange={(e) => setOverrideDays(Number(e.target.value))}
                        className="w-full px-3 py-2 rounded-lg bg-slate-950 border border-slate-700 text-white font-bold text-sm focus:outline-none focus:border-indigo-500"
                      >
                        <option value={15}>15 Days Credit</option>
                        <option value={30}>30 Days Credit</option>
                        <option value={0}>0 Days (Advance Payment Only)</option>
                      </select>
                    </div>
                  </div>
                </div>

                {/* Decision Action Buttons */}
                <div className="flex items-center gap-3 pt-2 border-t border-slate-800">
                  <button
                    onClick={handleApprove}
                    disabled={actionLoading}
                    className="flex-1 py-2.5 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-bold text-xs shadow-lg shadow-emerald-600/20 transition-all"
                  >
                    Approve & Grant Credit (₹{overrideLimit ? Number(overrideLimit).toLocaleString('en-IN') : 0})
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
