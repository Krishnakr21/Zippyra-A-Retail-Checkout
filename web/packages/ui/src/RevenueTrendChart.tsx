'use client';

import React from 'react';
import {
  ComposedChart,
  Bar,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

export interface SalesPeriodData {
  period: string;
  revenuePaise: number;
  orderCount: number;
  discountPaise?: number;
}

export interface RevenueTrendChartProps {
  data: SalesPeriodData[];
  granularity: 'day' | 'week' | 'month';
  onGranularityChange?: (g: 'day' | 'week' | 'month') => void;
  loading?: boolean;
  title?: string;
}

export function RevenueTrendChart({
  data = [],
  granularity = 'day',
  onGranularityChange,
  loading = false,
  title = 'Revenue & Order Trends',
}: RevenueTrendChartProps) {
  const chartData = data.map((item) => ({
    ...item,
    revenueRupees: Number((item.revenuePaise / 100).toFixed(2)),
  }));

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4" data-testid="revenue-trend-chart">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h3 className="text-lg font-bold text-white">{title}</h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Revenue (₹) and completed order volumes over time
          </p>
        </div>
        <div className="inline-flex rounded-lg bg-slate-800 p-1 border border-slate-700/60 self-start sm:self-auto" data-testid="granularity-toggle">
          {(['day', 'week', 'month'] as const).map((g) => (
            <button
              key={g}
              data-testid={`granularity-${g}`}
              onClick={() => onGranularityChange && onGranularityChange(g)}
              className={`px-3 py-1.5 text-xs font-semibold rounded-md transition-all ${
                granularity === g
                  ? 'bg-indigo-600 text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700/50'
              }`}
            >
              {g.charAt(0).toUpperCase() + g.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className="h-72 flex items-center justify-center text-slate-400 text-sm">
          <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-indigo-500 mr-3" />
          Loading trend data...
        </div>
      ) : chartData.length === 0 ? (
        <div className="h-72 flex items-center justify-center text-slate-500 text-sm border border-dashed border-slate-800 rounded-xl">
          No sales data recorded for the selected date range.
        </div>
      ) : (
        <div className="h-72 w-full pt-2">
          <ResponsiveContainer width="100%" height="100%">
            <ComposedChart data={chartData} margin={{ top: 10, right: 20, left: 10, bottom: 20 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
              <XAxis dataKey="period" stroke="#64748b" fontSize={12} tickLine={false} />
              <YAxis
                yAxisId="left"
                orientation="left"
                stroke="#818cf8"
                fontSize={12}
                tickFormatter={(val) => `₹${val.toLocaleString()}`}
                tickLine={false}
              />
              <YAxis
                yAxisId="right"
                orientation="right"
                stroke="#34d399"
                fontSize={12}
                tickLine={false}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: '#0f172a',
                  borderColor: '#334155',
                  borderRadius: '0.75rem',
                  color: '#f8fafc',
                }}
                formatter={(value: any, name: string) => {
                  if (name === 'Revenue (₹)') return [`₹${Number(value).toLocaleString()}`, name];
                  return [value, name];
                }}
              />
              <Legend wrapperStyle={{ paddingTop: '10px' }} />
              <Bar
                yAxisId="left"
                dataKey="revenueRupees"
                name="Revenue (₹)"
                fill="#6366f1"
                radius={[6, 6, 0, 0]}
                barSize={32}
              />
              <Line
                yAxisId="right"
                type="monotone"
                dataKey="orderCount"
                name="Order Count"
                stroke="#10b981"
                strokeWidth={3}
                dot={{ fill: '#10b981', r: 4 }}
              />
            </ComposedChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}
