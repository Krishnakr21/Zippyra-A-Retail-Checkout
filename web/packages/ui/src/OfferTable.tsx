'use client';

import React from 'react';
import { Badge } from './Badge';

export interface OfferTableItem {
  id: string;
  chain_id: string;
  store_id?: string | null;
  type: 'PERCENT_OFF' | 'FLAT_OFF' | 'BOGO' | 'CATEGORY_PERCENT_OFF' | (string & {});
  applies_to: string;
  target_ids?: string[];
  rule_config: Record<string, any>;
  min_cart_value_paise: number;
  max_discount_paise?: number | null;
  priority: number;
  active_from: string;
  active_until?: string | null;
  is_active: boolean;
  scope?: 'CHAIN_WIDE' | 'STORE_SPECIFIC';
}

export interface OfferTableProps {
  offers: OfferTableItem[];
  showScope?: boolean;
  userRole?: string; // 'MANAGER' | 'OWNER' | 'FINANCE' | 'OPERATIONS'
  onEdit?: (offer: OfferTableItem) => void;
  onToggle?: (offer: OfferTableItem, isActive: boolean) => void;
  onDelete?: (offer: OfferTableItem) => void;
}

export const OfferTable: React.FC<OfferTableProps> = ({
  offers,
  showScope = false,
  userRole = 'MANAGER',
  onEdit,
  onToggle,
  onDelete,
}) => {
  const computeStatus = (offer: OfferTableItem) => {
    if (!offer.is_active) return { label: 'Paused', variant: 'warning' as const };

    const now = Date.now();
    const from = new Date(offer.active_from).getTime();
    const until = offer.active_until ? new Date(offer.active_until).getTime() : null;

    if (from > now) return { label: 'Scheduled', variant: 'info' as const };
    if (until && until < now) return { label: 'Expired', variant: 'error' as const };
    return { label: 'Active', variant: 'success' as const };
  };

  const isEditable = (offer: OfferTableItem) => {
    const isChainWide = !offer.store_id || offer.scope === 'CHAIN_WIDE';
    if (isChainWide && userRole !== 'OWNER') {
      return { allowed: false, reason: "Managed by your chain's HQ team." };
    }
    if (userRole === 'FINANCE' || userRole === 'OPERATIONS') {
      return { allowed: false, reason: 'Only OWNER role can modify offers.' };
    }
    return { allowed: true, reason: '' };
  };

  const formatOfferRuleDetail = (offer: OfferTableItem) => {
    const cfg = offer.rule_config || {};
    switch (offer.type) {
      case 'PERCENT_OFF':
      case 'CATEGORY_PERCENT_OFF':
        return `${cfg.percent || 0}% Off`;
      case 'FLAT_OFF':
        return `₹${((cfg.flat_paise || 0) / 100).toFixed(2)} Flat Off`;
      case 'BOGO':
        return `Buy ${cfg.buy_qty || 2} Get ${cfg.get_qty || 1} Free`;
      default:
        return offer.type;
    }
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm text-slate-300">
          <thead className="bg-slate-800/60 text-slate-400 text-xs uppercase">
            <tr>
              <th className="p-3">Type & Details</th>
              <th className="p-3">Applies To</th>
              {showScope && <th className="p-3">Scope</th>}
              <th className="p-3">Priority</th>
              <th className="p-3">Min Cart</th>
              <th className="p-3">Status</th>
              <th className="p-3">Active State</th>
              <th className="p-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {offers.length === 0 ? (
              <tr>
                <td colSpan={showScope ? 8 : 7} className="p-6 text-center text-slate-500">
                  No promotional offers found.
                </td>
              </tr>
            ) : (
              offers.map((offer) => {
                const status = computeStatus(offer);
                const editPerm = isEditable(offer);
                const isChainWide = !offer.store_id || offer.scope === 'CHAIN_WIDE';

                return (
                  <tr key={offer.id} className="hover:bg-slate-800/40 transition-colors">
                    <td className="p-3">
                      <div className="font-semibold text-slate-100">{formatOfferRuleDetail(offer)}</div>
                      <div className="text-[11px] font-mono text-slate-400">{offer.type}</div>
                    </td>

                    <td className="p-3 text-xs">
                      <span className="font-mono text-slate-200">{offer.applies_to}</span>
                      {offer.target_ids && offer.target_ids.length > 0 && (
                        <div className="text-[10px] text-slate-400 truncate max-w-[150px]">
                          {offer.target_ids.join(', ')}
                        </div>
                      )}
                    </td>

                    {showScope && (
                      <td className="p-3">
                        <span
                          className={`inline-flex px-2 py-0.5 rounded text-[10px] font-bold ${
                            isChainWide
                              ? 'bg-purple-500/20 text-purple-300 border border-purple-500/30'
                              : 'bg-blue-500/20 text-blue-300 border border-blue-500/30'
                          }`}
                        >
                          {isChainWide ? 'CHAIN_WIDE' : 'STORE_SPECIFIC'}
                        </span>
                      </td>
                    )}

                    <td className="p-3 text-xs font-mono">{offer.priority}</td>

                    <td className="p-3 text-xs font-mono">
                      ₹{(offer.min_cart_value_paise / 100).toLocaleString('en-IN')}
                    </td>

                    <td className="p-3">
                      <Badge status={status.label} />
                    </td>

                    <td className="p-3">
                      <label className={`inline-flex items-center cursor-pointer ${!editPerm.allowed ? 'opacity-50 cursor-not-allowed' : ''}`}>
                        <input
                          type="checkbox"
                          checked={offer.is_active}
                          disabled={!editPerm.allowed}
                          onChange={(e) => onToggle && editPerm.allowed && onToggle(offer, e.target.checked)}
                          className="rounded border-slate-700 bg-slate-800 text-amber-500 focus:ring-amber-500"
                        />
                        <span className="ml-2 text-xs text-slate-400">{offer.is_active ? 'On' : 'Off'}</span>
                      </label>
                    </td>

                    <td className="p-3 text-right space-x-2">
                      <button
                        data-testid={`edit-offer-btn-${offer.id}`}
                        disabled={!editPerm.allowed}
                        title={editPerm.allowed ? 'Edit offer' : editPerm.reason}
                        onClick={() => onEdit && editPerm.allowed && onEdit(offer)}
                        className="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs rounded border border-slate-700 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                      >
                        Edit
                      </button>

                      <button
                        data-testid={`delete-offer-btn-${offer.id}`}
                        disabled={!editPerm.allowed}
                        title={editPerm.allowed ? 'Delete offer' : editPerm.reason}
                        onClick={() => onDelete && editPerm.allowed && onDelete(offer)}
                        className="px-2.5 py-1 bg-rose-600/20 hover:bg-rose-600/30 text-rose-400 border border-rose-500/30 text-xs rounded transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
