import { Fragment } from 'react'
import { Check } from 'lucide-react'

export interface WizardStep {
  id: number
  title: string
  subtitle: string
}

export interface WizardStepperProps {
  currentStep: number
  steps: WizardStep[]
  onStepClick?: (stepId: number) => void
}

export default function WizardStepper({ currentStep, steps, onStepClick }: WizardStepperProps) {
  return (
    <div className="w-full mb-8">
      <div className="flex items-center justify-between relative">
        {steps.map((step, idx) => {
          const isCompleted = step.id < currentStep
          const isActive = step.id === currentStep
          const isClickable = Boolean(onStepClick && (isCompleted || step.id <= currentStep))

          return (
            <Fragment key={step.id}>
              {/* Step Item */}
              <div
                onClick={() => {
                  if (isClickable && onStepClick) {
                    onStepClick(step.id)
                  }
                }}
                className={`flex items-center gap-3 relative z-10 ${
                  isClickable ? 'cursor-pointer group' : 'cursor-default'
                }`}
              >
                {/* Step Circle/Badge */}
                <div
                  className={`w-9 h-9 sm:w-10 sm:h-10 rounded-full flex items-center justify-center font-bold text-sm transition-all duration-200 shrink-0 ${
                    isCompleted
                      ? 'bg-emerald-500/20 border-2 border-emerald-500 text-emerald-400 shadow-[0_0_10px_rgba(16,185,129,0.25)]'
                      : isActive
                      ? 'bg-[var(--color-primary)] border-2 border-blue-400 text-white shadow-[0_0_14px_rgba(59,130,246,0.5)]'
                      : 'bg-[var(--color-input)] border border-[var(--color-border)] text-[var(--color-muted)]'
                  }`}
                >
                  {isCompleted ? <Check size={18} className="stroke-[2.5]" /> : step.id}
                </div>

                {/* Step Text */}
                <div className="hidden sm:block">
                  <div
                    className={`text-xs font-bold uppercase tracking-wider transition-colors duration-150 ${
                      isActive
                        ? 'text-white'
                        : isCompleted
                        ? 'text-emerald-400'
                        : 'text-[var(--color-muted)]'
                    }`}
                  >
                    {step.title}
                  </div>
                  <div className="text-[0.75rem] text-[var(--color-muted)] font-normal">
                    {step.subtitle}
                  </div>
                </div>
              </div>

              {/* Connecting line */}
              {idx < steps.length - 1 && (
                <div className="flex-1 mx-2 sm:mx-4 h-[2px] bg-[var(--color-border)] rounded-full overflow-hidden">
                  <div
                    className={`h-full transition-all duration-300 ${
                      step.id < currentStep
                        ? 'bg-emerald-500 w-full'
                        : 'bg-transparent w-0'
                    }`}
                  />
                </div>
              )}
            </Fragment>
          )
        })}
      </div>
    </div>
  )
}
