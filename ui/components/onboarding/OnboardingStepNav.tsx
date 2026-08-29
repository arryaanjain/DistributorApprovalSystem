import React from "react";
import { Check, ChevronRight, Lock } from "lucide-react";
import { OnboardingStep, STEP_LABELS } from "@/types/onboarding";

interface OnboardingStepNavProps {
  step: OnboardingStep;
  maxAllowedStepIndex?: number;
  onStepClick?: (step: OnboardingStep) => void;
}

export const OnboardingStepNav: React.FC<OnboardingStepNavProps> = ({
  step,
  maxAllowedStepIndex = 9,
  onStepClick,
}) => {
  const stepKeys: OnboardingStep[] = [
    "step1_business_det",
    "step2_business_exp",
    "step3_credit_pref",
    "step4_order_req",
    "step5_kyc_gst",
    "step6_auth",
    "step7_bank",
    "step8_approval",
    "step9_dashboard",
  ];

  const getStepProgressIndex = () => {
    return stepKeys.indexOf(step);
  };

  const currentProgressIdx = getStepProgressIndex() + 1;
  const currentStepInfo = Object.values(STEP_LABELS).find((s) => s.num === currentProgressIdx);
  const progressPercent = Math.round((currentProgressIdx / 9) * 100);

  const handleStepClick = (num: number) => {
    if (onStepClick && num <= maxAllowedStepIndex) {
      onStepClick(stepKeys[num - 1]);
    }
  };

  return (
    <div className="mb-6 sm:mb-8 space-y-3">
      {/* MOBILE COMPACT STEP INDICATOR HEADER (< lg) */}
      <div className="block lg:hidden bg-slate-900/90 border border-slate-800 rounded-2xl p-4 space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="px-2.5 py-1 bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 rounded-full text-xs font-bold font-mono">
              Step {currentProgressIdx} of 9
            </span>
            <h3 className="text-sm font-bold text-white tracking-tight">
              {currentStepInfo?.title || "Onboarding Step"}
            </h3>
          </div>
          <span className="text-xs font-bold text-emerald-400 font-mono">
            {progressPercent}%
          </span>
        </div>

        {/* Scrollable Pills Carousel on Mobile */}
        <div className="flex items-center gap-1.5 overflow-x-auto scrollbar-none touch-pan-x -mx-1 px-1 py-1">
          {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((num) => {
            const isDone = num < currentProgressIdx;
            const isCurrent = num === currentProgressIdx;
            const isLocked = num > maxAllowedStepIndex;
            const stepTitle = Object.values(STEP_LABELS).find((s) => s.num === num)?.title;

            return (
              <button
                key={num}
                onClick={() => handleStepClick(num)}
                disabled={!onStepClick || isLocked}
                type="button"
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-[11px] font-bold whitespace-nowrap transition-all touch-manipulation min-h-[34px] ${
                  isDone
                    ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                    : isCurrent
                    ? "bg-indigo-600 text-white shadow-md shadow-indigo-600/30"
                    : isLocked
                    ? "bg-slate-900/60 text-slate-600 border border-slate-800/80 cursor-not-allowed opacity-60"
                    : "bg-slate-800/60 text-slate-400 border border-slate-700/50 hover:bg-slate-800 hover:text-white"
                }`}
              >
                {isDone ? (
                  <Check className="w-3 h-3 text-emerald-400 shrink-0" />
                ) : isLocked ? (
                  <Lock className="w-3 h-3 text-slate-600 shrink-0" />
                ) : (
                  <span className="font-mono text-[10px]">{num}.</span>
                )}
                <span>{stepTitle}</span>
              </button>
            );
          })}
        </div>
      </div>

      {/* DESKTOP FULL PIPELINE STEP BAR (>= lg) */}
      <div className="hidden lg:flex items-center justify-between overflow-x-auto py-2">
        {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((num) => {
          const isDone = num < currentProgressIdx;
          const isCurrent = num === currentProgressIdx;
          const isLocked = num > maxAllowedStepIndex;
          const isClickable = Boolean(onStepClick) && !isLocked;

          return (
            <div
              key={num}
              onClick={() => handleStepClick(num)}
              title={
                isLocked
                  ? `Complete Step ${maxAllowedStepIndex} first to unlock`
                  : isClickable
                  ? `Jump to ${Object.values(STEP_LABELS).find((s) => s.num === num)?.title}`
                  : undefined
              }
              className={`flex items-center gap-2 min-w-max ${
                isClickable ? "cursor-pointer group select-none" : isLocked ? "cursor-not-allowed opacity-60" : ""
              }`}
            >
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-all ${
                  isDone
                    ? "bg-emerald-500 text-slate-950 font-extrabold group-hover:bg-emerald-400 group-hover:scale-105"
                    : isCurrent
                    ? "bg-indigo-600 text-white shadow-lg shadow-indigo-600/40 ring-4 ring-indigo-500/20 group-hover:bg-indigo-500"
                    : isLocked
                    ? "bg-slate-900 text-slate-600 border border-slate-800"
                    : "bg-slate-800 text-slate-400 group-hover:bg-slate-700 group-hover:text-slate-200"
                }`}
              >
                {isDone ? (
                  <Check className="w-4 h-4" />
                ) : isLocked ? (
                  <Lock className="w-3.5 h-3.5 text-slate-600" />
                ) : (
                  num
                )}
              </div>
              <span
                className={`text-xs font-medium transition-colors ${
                  isCurrent
                    ? "text-indigo-400 font-semibold"
                    : isDone
                    ? "text-slate-300 group-hover:text-white font-medium"
                    : isLocked
                    ? "text-slate-600"
                    : "text-slate-500 group-hover:text-slate-300"
                }`}
              >
                {Object.values(STEP_LABELS).find((s) => s.num === num)?.title}
              </span>
              {num < 9 && <ChevronRight className="w-4 h-4 text-slate-700" />}
            </div>
          );
        })}
      </div>

      {/* HORIZONTAL PROGRESS BAR */}
      <div className="w-full bg-slate-800/80 h-2 rounded-full overflow-hidden">
        <div
          className="bg-gradient-to-r from-indigo-500 to-emerald-400 h-full transition-all duration-500"
          style={{ width: `${currentProgressIdx * 11.11}%` }}
        />
      </div>
    </div>
  );
};
