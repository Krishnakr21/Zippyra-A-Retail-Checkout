'use client';

import React from 'react';

export interface FunnelStageItem {
  stage: string;
  sessionCount: number;
  conversionFromPreviousPercent: number;
}

export interface FunnelChartProps {
  stages: FunnelStageItem[];
  loading?: boolean;
  title?: string;
}

const STAGE_CONFIG: Record<string, { label: string; icon: string; color: string }> = {
  SESSION_STARTED: { label: '1. Session Started', icon: '🏪', color: 'from-blue-600 to-indigo-600' },
  CHECKOUT_INITIATED: { label: '2. Checkout Initiated', icon: '🛒', color: 'from-indigo-600 to-violet-600' },
  PAYMENT_CONFIRMED: { label: '3. Payment Confirmed', icon: '💳', color: 'from-violet-600 to-purple-600' },
  ORDER_COMPLETED: { label: '4. Order Completed', icon: '📦', color: 'from-purple-600 to-emerald-600' },
  EXIT_VALIDATED: { label: '5. Exit Validated', icon: '🚪', color: 'from-emerald-600 to-teal-500' },
};

const REQUIRED_STAGES = [
  'SESSION_STARTED',
  'CHECKOUT_INITIATED',
  'PAYMENT_CONFIRMED',
  'ORDER_COMPLETED',
  'EXIT_VALIDATED',
];

export function FunnelChart({
  stages = [],
  loading = false,
  title = 'Customer Conversion Funnel',
}: FunnelChartProps) {
  // Ensure all 5 stages exist in order even if input list is incomplete
  const stageMap = new Map(stages.map((s) => [s.stage, s]));
  const normalizedStages: FunnelStageItem[] = REQUIRED_STAGES.map((stageKey) => {
    const existing = stageMap.get(stageKey);
    return (
      existing || {
        stage: stageKey,
        sessionCount: 0,
        conversionFromPreviousPercent: 0,
      }
    );
  });

  const maxSessions = Math.max(...normalizedStages.map((s) => s.sessionCount), 1);

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-6" data-testid="funnel-chart">
      <div>
        <h3 className="text-lg font-bold text-white">{title}</h3>
        <p className="text-xs text-slate-400 mt-0.5">
          End-to-end customer journey from store check-in to gate exit validation
        </p>
      </div>

      {loading ? (
        <div className="h-64 flex items-center justify-center text-slate-400 text-sm">
          <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-indigo-500 mr-3" />
          Loading funnel data...
        </div>
      ) : (
        <div className="space-y-4" data-testid="funnel-stages-list">
          {normalizedStages.map((item, idx) => {
            const config = STAGE_CONFIG[item.stage] || {
              label: item.stage,
              icon: '📊',
              color: 'from-indigo-600 to-purple-600',
            };
            const widthPct = Math.max(Math.round((item.sessionCount / maxSessions) * 100), 4);
            const isFirst = idx === 0;

            return (
              <div key={item.stage} className="space-y-1.5" data-testid={`funnel-stage-${item.stage}`}>
                <div className="flex items-center justify-between text-xs font-semibold text-slate-300">
                  <span className="flex items-center gap-2">
                    <span>{config.icon}</span>
                    <span className="text-sm font-bold text-white">{config.label}</span>
                  </span>
                  <div className="flex items-center gap-3">
                    <span className="font-mono text-slate-200 bg-slate-800 px-2.5 py-0.5 rounded-md border border-slate-700">
                      {item.sessionCount.toLocaleString()} sessions
                    </span>
                    {!isFirst && (
                      <span
                        className={`font-mono font-bold px-2 py-0.5 rounded-md text-xs ${
                          item.conversionFromPreviousPercent > 75
                            ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                            : item.conversionFromPreviousPercent > 30
                            ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                            : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                        }`}
                      >
                        {item.conversionFromPreviousPercent.toFixed(1)}% conv.
                      </span>
                    )}
                  </div>
                </div>

                <div className="h-6 w-full bg-slate-800/80 rounded-lg overflow-hidden p-0.5 border border-slate-700/50">
                  <div
                    className={`h-full rounded-md bg-gradient-to-r ${config.color} transition-all duration-500 flex items-center justify-end pr-2`}
                    style={{ width: `${widthPct}%` }}
                  >
                    {widthPct > 20 && (
                      <span className="text-[10px] font-bold text-white drop-shadow">
                        {widthPct}%
                      </span>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
