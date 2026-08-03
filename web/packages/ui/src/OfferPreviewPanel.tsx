'use client';

import React, { useEffect, useState } from 'react';

export interface CompiledOfferRule {
  id: string;
  type: string;
  value: number;
  applies_to: string;
  target_ids?: string[];
  min_cart_value_paise: number;
  max_discount_paise?: number | null;
  active_from?: string;
  active_until?: string | null;
}

export interface OfferPreviewPanelProps {
  storeId: string;
}

export const OfferPreviewPanel: React.FC<OfferPreviewPanelProps> = ({ storeId }) => {
  const [rules, setRules] = useState<CompiledOfferRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchPreview = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/v1/cart/admin/offers/${storeId}/preview`);
      if (res.ok) {
        const data = await res.json();
        setRules(Array.isArray(data) ? data : []);
      } else {
        setError(`Failed to fetch compiled rules: HTTP ${res.status}`);
      }
    } catch (err: any) {
      setError(err?.message || 'Network error fetching live preview');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (storeId) {
      fetchPreview();
    }
  }, [storeId]);

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-4 shadow-lg">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-bold text-slate-100 flex items-center space-x-2">
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse"></span>
            <span>Live Compiled Rules Engine Preview</span>
          </h3>
          <p className="text-xs text-slate-400">
            Literal compiled ruleset served from Redis <code className="font-mono text-amber-400">offer_rules:{storeId}</code>
          </p>
        </div>

        <button
          onClick={fetchPreview}
          data-testid="refresh-preview-btn"
          className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs rounded-lg border border-slate-700 transition-colors"
        >
          Refresh Live Preview
        </button>
      </div>

      {loading ? (
        <div className="p-4 text-center text-xs text-slate-500">Loading compiled Redis rules...</div>
      ) : error ? (
        <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg text-rose-400 text-xs">
          {error}
        </div>
      ) : rules.length === 0 ? (
        <div className="p-4 bg-slate-950/60 border border-slate-800/80 rounded-lg text-center text-xs text-slate-500 font-mono">
          [ ] — Zero active compiled rules in Redis.
        </div>
      ) : (
        <div className="space-y-2">
          {rules.map((rule, idx) => (
            <div
              key={rule.id || idx}
              data-testid={`compiled-rule-row-${rule.id}`}
              className="p-3 bg-slate-950/80 border border-slate-800 rounded-lg flex items-center justify-between font-mono text-xs text-slate-300"
            >
              <div className="flex items-center space-x-3">
                <span className="text-amber-400 font-bold">#{idx + 1}</span>
                <span className="font-bold text-slate-100">{rule.type}</span>
                <span className="text-emerald-400">Value: {rule.value}</span>
                <span className="text-slate-400">Applies: {rule.applies_to}</span>
              </div>
              <div className="text-slate-400 text-[11px]">
                Min Cart: ₹{((rule.min_cart_value_paise || 0) / 100).toFixed(2)}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
