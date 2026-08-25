import React from "react";
import { Check, ChevronRight } from "lucide-react";
import { OnboardingStep, STEP_LABELS } from "@/types/onboarding";

interface OnboardingStepNavProps {
  step: OnboardingStep;
  onStepClick?: (step: OnboardingStep) => void;
}

export const OnboardingStepNav: React.FC<OnboardingStepNavProps> = ({ step, onStepClick }) => {
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

  const handleStepClick = (num: number) => {
    if (onStepClick && num <= stepKeys.length) {
      onStepClick(stepKeys[num - 1]);
    }
  };

  return (
    <div className="mb-8">
      {/* Step Progress Bar */}
      <div className="flex items-center justify-between mb-4 overflow-x-auto py-2">
        {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((num) => {
          const isDone = num < currentProgressIdx;
          const isCurrent = num === currentProgressIdx;
          const isClickable = Boolean(onStepClick);

          return (
            <div
              key={num}
              onClick={() => handleStepClick(num)}
              title={isClickable ? `Jump to ${Object.values(STEP_LABELS).find((s) => s.num === num)?.title}` : undefined}
              className={`flex items-center gap-2 min-w-max ${
                isClickable ? "cursor-pointer group select-none" : ""
              }`}
            >
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-all ${
                  isDone
                    ? "bg-emerald-500 text-slate-950 font-extrabold group-hover:bg-emerald-400 group-hover:scale-105"
                    : isCurrent
                    ? "bg-indigo-600 text-white shadow-lg shadow-indigo-600/40 ring-4 ring-indigo-500/20 group-hover:bg-indigo-500"
                    : "bg-slate-800 text-slate-500 group-hover:bg-slate-700 group-hover:text-slate-300"
                }`}
              >
                {isDone ? <Check className="w-4 h-4" /> : num}
              </div>
              <span
                className={`text-xs font-medium transition-colors ${
                  isCurrent
                    ? "text-indigo-400 font-semibold"
                    : isDone
                    ? "text-slate-300 group-hover:text-white font-medium"
                    : "text-slate-600 group-hover:text-slate-400"
                }`}
              >
                {Object.values(STEP_LABELS).find((s) => s.num === num)?.title}
              </span>
              {num < 9 && <ChevronRight className="w-4 h-4 text-slate-700" />}
            </div>
          );
        })}
      </div>

      <div className="w-full bg-slate-800/80 h-2 rounded-full overflow-hidden">
        <div
          className="bg-gradient-to-r from-indigo-500 to-violet-500 h-full transition-all duration-500"
          style={{ width: `${currentProgressIdx * 11.11}%` }}
        />
      </div>
    </div>
  );
};
