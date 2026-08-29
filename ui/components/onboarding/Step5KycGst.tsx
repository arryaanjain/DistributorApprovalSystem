import React from "react";
import { FileCheck, ArrowLeft, ArrowRight, CheckCircle2, AlertTriangle, ShieldCheck, UserCheck, Building2, Check, Sparkles } from "lucide-react";
import { Step5Data } from "@/types/onboarding";

interface VerificationResults {
  panVerified?: boolean;
  gstVerified?: boolean;
  panHolderName?: string;
  gstLegalName?: string;
}

interface Step5KycGstProps {
  step5: Step5Data;
  setStep5: React.Dispatch<React.SetStateAction<Step5Data>>;
  loading: boolean;
  onBack: () => void;
  onSubmit: (e: React.FormEvent) => void;
  verificationWarnings?: string[];
  verificationResults?: VerificationResults;
  step1Name?: string;
  step1BusinessName?: string;
  onGoToStep1?: () => void;
}

function levenshtein(a: string, b: string): number {
  const matrix: number[][] = [];
  for (let i = 0; i <= b.length; i++) matrix[i] = [i];
  for (let j = 0; j <= a.length; j++) matrix[0][j] = j;
  for (let i = 1; i <= b.length; i++) {
    for (let j = 1; j <= a.length; j++) {
      if (b.charAt(i - 1) === a.charAt(j - 1)) {
        matrix[i][j] = matrix[i - 1][j - 1];
      } else {
        matrix[i][j] = Math.min(
          matrix[i - 1][j - 1] + 1,
          matrix[i][j - 1] + 1,
          matrix[i - 1][j] + 1
        );
      }
    }
  }
  return matrix[b.length][a.length];
}

// Case-insensitive & token/fuzzy edit distance matching helper
function checkNamesMatch(a?: string, b?: string): boolean {
  if (!a || !b) return false;
  const cleanA = a.toUpperCase().trim().replace(/[^A-Z0-9\s]/g, "");
  const cleanB = b.toUpperCase().trim().replace(/[^A-Z0-9\s]/g, "");
  if (cleanA === cleanB || cleanA.includes(cleanB) || cleanB.includes(cleanA)) return true;

  const wordsA = cleanA.split(/\s+/).filter(Boolean);
  const wordsB = cleanB.split(/\s+/).filter(Boolean);
  if (wordsA.length === 0 || wordsB.length === 0) return false;

  const fullWordsA = wordsA.filter((w) => w.length > 1);
  const fullWordsB = wordsB.filter((w) => w.length > 1);

  if (fullWordsA.length > 0 && fullWordsB.length > 0) {
    const smaller = fullWordsA.length <= fullWordsB.length ? fullWordsA : fullWordsB;
    const larger = fullWordsA.length > fullWordsB.length ? fullWordsA : fullWordsB;

    let matches = 0;
    for (const sw of smaller) {
      for (const lw of larger) {
        if (sw === lw || lw.startsWith(sw) || sw.startsWith(lw) || (sw.length >= 4 && lw.length >= 4 && levenshtein(sw, lw) <= 2)) {
          matches++;
          break;
        }
      }
    }
    const required = Math.min(2, smaller.length);
    if (matches >= required) return true;
  }
  return false;
}

export const Step5KycGst: React.FC<Step5KycGstProps> = ({
  step5,
  setStep5,
  loading,
  onBack,
  onSubmit,
  verificationWarnings,
  verificationResults,
  step1Name,
  step1BusinessName,
  onGoToStep1,
}) => {
  const isPanVerified = verificationResults?.panVerified;
  const isGstVerified = verificationResults?.gstVerified;
  const panHolderName = verificationResults?.panHolderName;
  const gstLegalName = verificationResults?.gstLegalName;

  const hasFetchedResults = Boolean(panHolderName || gstLegalName);
  const isPanNameMatched = checkNamesMatch(step1Name, panHolderName);
  const isGstNameMatched = checkNamesMatch(step1BusinessName, gstLegalName) || checkNamesMatch(step1Name, gstLegalName);

  const hasNameMismatch = hasFetchedResults && (
    (panHolderName && !isPanNameMatched) || 
    (gstLegalName && step5.has_gst && !isGstNameMatched)
  );
  const hasWarnings = Boolean(verificationWarnings && verificationWarnings.length > 0);

  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-8 shadow-2xl backdrop-blur-xl max-w-3xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-indigo-500/10 text-indigo-400 rounded-xl flex items-center justify-center border border-indigo-500/20">
            <FileCheck className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">Step 5: KYC & GST Verification</h2>
            <p className="text-xs text-slate-400">Real-time Surepass statutory verification with tax & GST registry</p>
          </div>
        </div>

        {(isPanVerified || isGstVerified) && (
          <div className="flex items-center gap-1.5 px-3 py-1.5 bg-emerald-500/10 border border-emerald-500/30 rounded-lg text-emerald-400 text-xs font-semibold">
            <ShieldCheck className="w-4 h-4" />
            <span>Verification Record Active</span>
          </div>
        )}
      </div>

      {/* Real-time Surepass Fetched Comparison Cards */}
      {(panHolderName || gstLegalName) && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* PAN Comparison Card */}
          {panHolderName && (
            <div className={`p-4 rounded-xl border ${isPanVerified || isPanNameMatched ? "bg-emerald-950/20 border-emerald-500/30" : "bg-amber-950/20 border-amber-500/30"}`}>
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
                  <UserCheck className="w-4 h-4 text-indigo-400" />
                  <span>PAN Registry (Surepass)</span>
                </div>
                {isPanNameMatched ? (
                  <span className="text-[10px] font-semibold px-2 py-0.5 bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded flex items-center gap-1">
                    <CheckCircle2 className="w-3 h-3" /> Name Matched
                  </span>
                ) : (
                  <span className="text-[10px] font-semibold px-2 py-0.5 bg-amber-500/20 text-amber-400 border border-amber-500/30 rounded flex items-center gap-1">
                    <AlertTriangle className="w-3 h-3" /> Name Mismatch
                  </span>
                )}
              </div>
              <div className="space-y-2 text-xs">
                <div className="flex justify-between items-center bg-slate-800/60 p-2 rounded-lg">
                  <span className="text-slate-400">Tax PAN Name:</span>
                  <span className="font-semibold text-white font-mono">{panHolderName}</span>
                </div>
                {step1Name && (
                  <div className="flex justify-between items-center bg-slate-800/40 p-2 rounded-lg">
                    <span className="text-slate-400">Step 1 Registered:</span>
                    <span className="font-semibold text-slate-300 font-mono">{step1Name}</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* GST Comparison Card */}
          {gstLegalName && (
            <div className={`p-4 rounded-xl border ${isGstVerified || isGstNameMatched ? "bg-emerald-950/20 border-emerald-500/30" : "bg-amber-950/20 border-amber-500/30"}`}>
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
                  <Building2 className="w-4 h-4 text-indigo-400" />
                  <span>GST Registry (Surepass)</span>
                </div>
                {isGstNameMatched ? (
                  <span className="text-[10px] font-semibold px-2 py-0.5 bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded flex items-center gap-1">
                    <CheckCircle2 className="w-3 h-3" /> Business Matched
                  </span>
                ) : (
                  <span className="text-[10px] font-semibold px-2 py-0.5 bg-amber-500/20 text-amber-400 border border-amber-500/30 rounded flex items-center gap-1">
                    <AlertTriangle className="w-3 h-3" /> Review Needed
                  </span>
                )}
              </div>
              <div className="space-y-2 text-xs">
                <div className="flex justify-between items-center bg-slate-800/60 p-2 rounded-lg">
                  <span className="text-slate-400">GST Legal Name:</span>
                  <span className="font-semibold text-white font-mono">{gstLegalName}</span>
                </div>
                {step1BusinessName && (
                  <div className="flex justify-between items-center bg-slate-800/40 p-2 rounded-lg">
                    <span className="text-slate-400">Step 2 Business:</span>
                    <span className="font-semibold text-slate-300 font-mono">{step1BusinessName}</span>
                  </div>
                )}
                {step1Name && (
                  <div className="flex justify-between items-center bg-slate-800/40 p-2 rounded-lg">
                    <span className="text-slate-400">Step 1 Registered:</span>
                    <span className="font-semibold text-slate-300 font-mono">{step1Name}</span>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Mismatch Warning Alert Box & Go To Step 1 Action */}
      {(hasWarnings || hasNameMismatch) && (
        <div className="p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl space-y-3">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-amber-400" />
            <h4 className="text-xs font-bold uppercase text-amber-400 tracking-wider">Verification Mismatch Warnings</h4>
          </div>
          {verificationWarnings && verificationWarnings.length > 0 && (
            <ul className="text-xs text-amber-200 space-y-1 list-disc pl-5">
              {verificationWarnings.map((w, i) => (
                <li key={i}>{w}</li>
              ))}
            </ul>
          )}
          {onGoToStep1 && hasNameMismatch && (
            <div className="pt-1">
              <button
                type="button"
                onClick={onGoToStep1}
                className="px-3.5 py-2 bg-amber-500/20 hover:bg-amber-500/30 text-amber-300 border border-amber-500/40 rounded-lg text-xs font-semibold flex items-center gap-2 transition-colors"
              >
                <ArrowLeft className="w-3.5 h-3.5" />
                <span>Go to Step 1 to Update Details / Fix Name Typos</span>
              </button>
            </div>
          )}
        </div>
      )}

      <form onSubmit={onSubmit} className="space-y-6">
        <div>
          <div className="flex items-center justify-between mb-1">
            <label className="block text-xs font-semibold uppercase text-slate-400">
              PAN Number (10 Digits) *
            </label>
            {panHolderName && (
              <span className="text-[11px] text-emerald-400 font-medium flex items-center gap-1">
                <Check className="w-3 h-3" /> Verified: {panHolderName}
              </span>
            )}
          </div>
          <input
            type="text"
            value={step5.pan}
            onChange={(e) => setStep5({ ...step5, pan: e.target.value.toUpperCase() })}
            required
            maxLength={10}
            placeholder="ABCDE1234F"
            className="w-full px-4 py-3 bg-slate-800/80 border border-slate-700 rounded-xl text-sm font-mono text-slate-100 focus:outline-none focus:border-indigo-500 uppercase tracking-widest"
          />
          {panHolderName && step1Name && (
            <p className="mt-1 text-[11px] text-slate-400 flex items-center justify-between">
              <span>Step 1 Name: <strong className="text-slate-200 font-mono">{step1Name}</strong></span>
              <span className={isPanNameMatched ? "text-emerald-400 font-semibold" : "text-amber-400 font-semibold"}>
                {isPanNameMatched ? "✓ Name Cross-Check Passed" : "⚠ Name Discrepancy Flagged"}
              </span>
            </p>
          )}
        </div>

        <div className="p-4 bg-slate-800/50 border border-slate-700/60 rounded-xl space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm font-semibold text-white">Does your business have a GST Registration?</span>
            <input
              type="checkbox"
              checked={step5.has_gst}
              onChange={(e) => setStep5({ ...step5, has_gst: e.target.checked })}
              className="w-5 h-5 text-indigo-600 rounded bg-slate-900 border-slate-700"
            />
          </div>

          {step5.has_gst && (
            <div>
              <div className="flex items-center justify-between mb-1">
                <label className="block text-xs font-semibold uppercase text-slate-400">
                  GSTIN (15 Digits)
                </label>
                {gstLegalName && (
                  <span className="text-[11px] text-emerald-400 font-medium flex items-center gap-1">
                    <Check className="w-3 h-3" /> Verified: {gstLegalName}
                  </span>
                )}
              </div>
              <input
                type="text"
                value={step5.gst_number}
                onChange={(e) => setStep5({ ...step5, gst_number: e.target.value.toUpperCase() })}
                maxLength={15}
                placeholder="24ABCDE1234F1Z5"
                className="w-full px-4 py-3 bg-slate-800 border border-slate-700 rounded-xl text-sm font-mono text-slate-100 focus:outline-none focus:border-indigo-500 uppercase tracking-widest"
              />
              {gstLegalName && step1BusinessName && (
                <p className="mt-1 text-[11px] text-slate-400 flex items-center justify-between">
                  <span>Business Name: <strong className="text-slate-200 font-mono">{step1BusinessName}</strong></span>
                  <span className={isGstNameMatched ? "text-emerald-400 font-semibold" : "text-amber-400 font-semibold"}>
                    {isGstNameMatched ? "✓ Business Cross-Check Passed" : "⚠ Business Discrepancy Flagged"}
                  </span>
                </p>
              )}
            </div>
          )}
        </div>

        {/* Optional Compliance Credentials with Weightage Mechanics */}
        <div className="p-5 bg-slate-800/40 border border-slate-700/80 rounded-2xl space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ShieldCheck className="w-4 h-4 text-emerald-400" />
              <h3 className="text-xs font-bold uppercase text-slate-200 tracking-wider">
                Optional Compliance & Regulatory Credentials
              </h3>
            </div>
            <span className="text-[11px] font-semibold text-emerald-400 bg-emerald-500/10 border border-emerald-500/30 px-2.5 py-1 rounded-lg flex items-center gap-1">
              <Sparkles className="w-3 h-3" /> +5 Max Compliance Weightage Boost
            </span>
          </div>

          <p className="text-xs text-slate-400">
            Providing optional regulatory numbers (FSSAI / Udyam MSME) directly boosts your automated credit evaluation score and increases pre-approved credit limits.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-1">
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label className="block text-xs font-semibold uppercase text-slate-300">
                  FSSAI Food License (Optional)
                </label>
                <span className={`text-[10px] font-semibold px-2 py-0.5 rounded border transition-all ${
                  step5.fssai_number && step5.fssai_number.trim().length > 0
                    ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40"
                    : "bg-slate-800 text-slate-400 border-slate-700"
                }`}>
                  {step5.fssai_number && step5.fssai_number.trim().length > 0 ? "✓ +3 Pts Earned" : "+3 Pts Weightage"}
                </span>
              </div>
              <input
                type="text"
                value={step5.fssai_number}
                onChange={(e) => setStep5({ ...step5, fssai_number: e.target.value })}
                placeholder="10020021000123"
                className="w-full px-4 py-3 bg-slate-900/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500 placeholder:text-slate-500"
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label className="block text-xs font-semibold uppercase text-slate-300">
                  Udyam / MSME Registration (Optional)
                </label>
                <span className={`text-[10px] font-semibold px-2 py-0.5 rounded border transition-all ${
                  step5.udyam_number && step5.udyam_number.trim().length > 0
                    ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40"
                    : "bg-slate-800 text-slate-400 border-slate-700"
                }`}>
                  {step5.udyam_number && step5.udyam_number.trim().length > 0 ? "✓ +3 Pts Earned" : "+3 Pts Weightage"}
                </span>
              </div>
              <input
                type="text"
                value={step5.udyam_number}
                onChange={(e) => setStep5({ ...step5, udyam_number: e.target.value })}
                placeholder="UDYAM-GJ-01-0001234"
                className="w-full px-4 py-3 bg-slate-900/80 border border-slate-700 rounded-xl text-sm text-slate-100 focus:outline-none focus:border-indigo-500 placeholder:text-slate-500"
              />
            </div>
          </div>
        </div>

        <div className="flex justify-between pt-4">
          <button
            type="button"
            onClick={onBack}
            className="px-4 py-2.5 bg-slate-800 text-slate-300 text-sm rounded-xl hover:bg-slate-700 flex items-center gap-2"
          >
            <ArrowLeft className="w-4 h-4" /> Back
          </button>
          {(() => {
            const isFullyVerified = Boolean(isPanVerified && (!step5.has_gst || isGstVerified) && (!verificationWarnings || verificationWarnings.length === 0));
            const hasDiscrepancy = Boolean((verificationWarnings && verificationWarnings.length > 0) || (!isPanVerified && panHolderName));

            return (
              <button
                type="submit"
                disabled={loading}
                className={`px-6 py-3 font-medium rounded-xl shadow-lg transition-all flex items-center gap-2 disabled:opacity-50 ${
                  isFullyVerified
                    ? "bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-600/30"
                    : hasDiscrepancy
                    ? "bg-amber-600 hover:bg-amber-500 text-white shadow-amber-600/30"
                    : "bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-600/30"
                }`}
              >
                <span>
                  {loading
                    ? "Verifying with Surepass..."
                    : isFullyVerified
                    ? "Continue to Authorization"
                    : hasDiscrepancy
                    ? "Re-verify Statutory Details"
                    : "Submit KYC & Verify"}
                </span>
                <ArrowRight className="w-4 h-4" />
              </button>
            );
          })()}
        </div>
      </form>
    </div>
  );
};
