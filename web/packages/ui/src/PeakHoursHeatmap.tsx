'use client';

import React, { useState } from 'react';

export interface PeakHourCell {
  dayOfWeek: number; // 0=Sunday, 6=Saturday
  hour: number; // 0..23
  avgTransactionsPerWeek: number;
  recommendedStaff: number;
}

export interface PeakHoursHeatmapProps {
  grid: PeakHourCell[];
  loading?: boolean;
  title?: string;
}

const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const HOURS = Array.from({ length: 24 }, (_, i) => i);

export function PeakHoursHeatmap({
  grid = [],
  loading = false,
  title = 'Peak Hours Heatmap & Staffing Suggestions',
}: PeakHoursHeatmapProps) {
  const [hoveredCell, setHoveredCell] = useState<{
    day: number;
    hour: number;
    avgTx: number;
    staff: number;
    x: number;
    y: number;
  } | null>(null);

  // Map lookup by `${dayOfWeek}_${hour}`
  const gridMap = new Map<string, PeakHourCell>();
  let maxVolume = 1;
  grid.forEach((cell) => {
    gridMap.set(`${cell.dayOfWeek}_${cell.hour}`, cell);
    if (cell.avgTransactionsPerWeek > maxVolume) {
      maxVolume = cell.avgTransactionsPerWeek;
    }
  });

  const getHeatColor = (avgTx: number) => {
    if (avgTx === 0) return 'bg-slate-800/40 border-slate-800 text-slate-600';
    const ratio = avgTx / maxVolume;
    if (ratio < 0.2) return 'bg-indigo-950/60 border-indigo-900/50 text-indigo-300';
    if (ratio < 0.4) return 'bg-indigo-900/80 border-indigo-700/60 text-indigo-200';
    if (ratio < 0.7) return 'bg-indigo-600 border-indigo-500 text-white';
    return 'bg-purple-600 border-purple-400 text-white font-bold animate-pulse';
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4 relative" data-testid="peak-hours-heatmap">
      <div>
        <h3 className="text-lg font-bold text-white">{title}</h3>
        <p className="text-xs text-slate-400 mt-0.5">
          Hourly transaction density across days of the week with recommended staffing levels
        </p>
      </div>

      {loading ? (
        <div className="h-64 flex items-center justify-center text-slate-400 text-sm">
          <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-indigo-500 mr-3" />
          Loading peak hours heatmap...
        </div>
      ) : (
        <div className="overflow-x-auto pb-2">
          <div className="min-w-[700px] space-y-1">
            {/* Hours Header */}
            <div className="grid grid-cols-[60px_repeat(24,_1fr)] gap-1 text-center text-[10px] font-mono text-slate-400 mb-2">
              <div />
              {HOURS.map((h) => (
                <div key={h}>{h}h</div>
              ))}
            </div>

            {/* Heatmap Rows */}
            {DAYS.map((dayName, dayIdx) => (
              <div key={dayName} className="grid grid-cols-[60px_repeat(24,_1fr)] gap-1 items-center">
                <div className="text-xs font-semibold text-slate-300 font-mono text-right pr-2">
                  {dayName}
                </div>
                {HOURS.map((hour) => {
                  const cellData = gridMap.get(`${dayIdx}_${hour}`) || {
                    dayOfWeek: dayIdx,
                    hour,
                    avgTransactionsPerWeek: 0,
                    recommendedStaff: 1,
                  };
                  const colorClass = getHeatColor(cellData.avgTransactionsPerWeek);

                  return (
                    <div
                      key={hour}
                      data-testid={`heatmap-cell-${dayIdx}-${hour}`}
                      onMouseEnter={(e) => {
                        const rect = e.currentTarget.getBoundingClientRect();
                        setHoveredCell({
                          day: dayIdx,
                          hour,
                          avgTx: cellData.avgTransactionsPerWeek,
                          staff: cellData.recommendedStaff,
                          x: rect.left + rect.width / 2,
                          y: rect.top,
                        });
                      }}
                      onMouseLeave={() => setHoveredCell(null)}
                      className={`h-8 rounded border transition-all cursor-pointer flex items-center justify-center text-[10px] ${colorClass}`}
                    >
                      {cellData.avgTransactionsPerWeek > 0 ? cellData.recommendedStaff : ''}
                    </div>
                  );
                })}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Floating Hover Tooltip */}
      {hoveredCell && (
        <div
          data-testid="heatmap-tooltip"
          className="fixed z-50 transform -translate-x-1/2 -translate-y-full mb-2 bg-slate-950 border border-slate-700 text-white rounded-xl p-3 shadow-2xl pointer-events-none text-xs space-y-1"
          style={{ left: hoveredCell.x, top: hoveredCell.y }}
        >
          <div className="font-bold text-indigo-300">
            {DAYS[hoveredCell.day]} @ {hoveredCell.hour}:00 - {hoveredCell.hour + 1}:00
          </div>
          <div>
            Avg Transactions/wk: <span className="font-mono text-emerald-400 font-bold">{hoveredCell.avgTx.toFixed(1)}</span>
          </div>
          <div>
            Recommended Staff: <span className="font-mono text-amber-400 font-bold">{hoveredCell.staff}</span>
          </div>
        </div>
      )}

      {/* Explicit Staffing Formula Caveat Caption */}
      <div className="text-xs text-slate-400 bg-slate-800/60 border border-slate-800 p-3 rounded-xl flex items-start gap-2" data-testid="staffing-caveat-caption">
        <svg className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>
          Staffing suggestion based on a simple throughput estimate — use judgment for final scheduling.
        </span>
      </div>
    </div>
  );
}
