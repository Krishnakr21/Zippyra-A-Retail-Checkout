'use client';

import React, { useState } from 'react';

export interface OfferFormValues {
  id?: string;
  chain_id?: string;
  store_id?: string | null;
  type: 'PERCENT_OFF' | 'FLAT_OFF' | 'BOGO' | 'CATEGORY_PERCENT_OFF';
  applies_to: 'ALL' | 'CATEGORY' | 'BARCODE_LIST';
  target_ids: string[];
  rule_config: Record<string, any>;
  min_cart_value_paise: number;
  max_discount_paise?: number | null;
  priority: number;
  active_from?: string;
  active_until?: string | null;
}

export interface OfferFormProps {
  mode: 'create' | 'edit';
  scope: 'STORE' | 'CHAIN_WIDE';
  initialValues?: Partial<OfferFormValues>;
  onSubmit: (values: OfferFormValues) => Promise<void>;
  onCancel: () => void;
}

export const OfferForm: React.FC<OfferFormProps> = ({
  mode,
  scope,
  initialValues,
  onSubmit,
  onCancel,
}) => {
  const [type, setType] = useState<OfferFormValues['type']>(initialValues?.type || 'PERCENT_OFF');
  const [appliesTo, setAppliesTo] = useState<OfferFormValues['applies_to']>(initialValues?.applies_to || 'ALL');
  const [targetIdsStr, setTargetIdsStr] = useState<string>(initialValues?.target_ids?.join(', ') || '');
  
  // Rule Config inputs
  const [percent, setPercent] = useState<number>(initialValues?.rule_config?.percent || 10);
  const [flatAmountRs, setFlatAmountRs] = useState<number>((initialValues?.rule_config?.flat_paise || 1000) / 100);
  const [buyQty, setBuyQty] = useState<number>(initialValues?.rule_config?.buy_qty || 2);
  const [getQty, setGetQty] = useState<number>(initialValues?.rule_config?.get_qty || 1);

  // Cart & Schedule inputs
  const [minCartValueRs, setMinCartValueRs] = useState<number>((initialValues?.min_cart_value_paise || 0) / 100);
  const [maxDiscountRs, setMaxDiscountRs] = useState<string>(
    initialValues?.max_discount_paise ? (initialValues.max_discount_paise / 100).toString() : ''
  );
  const [priority, setPriority] = useState<number>(initialValues?.priority ?? 100);
  const [activeFrom, setActiveFrom] = useState<string>(
    initialValues?.active_from ? initialValues.active_from.slice(0, 16) : new Date().toISOString().slice(0, 16)
  );
  const [activeUntil, setActiveUntil] = useState<string>(
    initialValues?.active_until ? initialValues.active_until.slice(0, 16) : ''
  );

  const [validationError, setValidationError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);

    // 1. Client-Side Validation
    if ((type === 'PERCENT_OFF' || type === 'CATEGORY_PERCENT_OFF') && (percent < 1 || percent > 90)) {
      setValidationError('Percent off must be between 1% and 90%');
      return;
    }

    if (type === 'BOGO' && getQty > buyQty) {
      setValidationError('Get quantity cannot exceed Buy quantity');
      return;
    }

    if (activeUntil && new Date(activeUntil) <= new Date(activeFrom)) {
      setValidationError('Active until date must be strictly after active from date');
      return;
    }

    if (appliesTo !== 'ALL' && !targetIdsStr.trim()) {
      setValidationError(`Target IDs cannot be empty when applies_to is ${appliesTo}`);
      return;
    }

    // Build Rule Config JSON
    let ruleConfig: Record<string, any> = {};
    if (type === 'PERCENT_OFF' || type === 'CATEGORY_PERCENT_OFF') {
      ruleConfig = { percent: Number(percent) };
    } else if (type === 'FLAT_OFF') {
      ruleConfig = { flat_paise: Math.round(Number(flatAmountRs) * 100) };
    } else if (type === 'BOGO') {
      ruleConfig = { buy_qty: Number(buyQty), get_qty: Number(getQty) };
    }

    const targetIds = targetIdsStr
      .split(',')
      .map((s) => s.trim())
      .filter((s) => s.length > 0);

    const payload: OfferFormValues = {
      type,
      applies_to: appliesTo,
      target_ids: targetIds,
      rule_config: ruleConfig,
      min_cart_value_paise: Math.round(Number(minCartValueRs) * 100),
      max_discount_paise: maxDiscountRs ? Math.round(Number(maxDiscountRs) * 100) : null,
      priority: Number(priority),
      active_from: new Date(activeFrom).toISOString(),
      active_until: activeUntil ? new Date(activeUntil).toISOString() : null,
    };

    setSubmitting(true);
    try {
      await onSubmit(payload);
    } catch (err: any) {
      setValidationError(err?.message || 'Failed to submit offer');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4 bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-xl">
      <div className="flex items-center justify-between border-b border-slate-800 pb-3">
        <h3 className="text-lg font-bold text-slate-100">
          {mode === 'create' ? 'Create New Offer' : 'Edit Offer'} ({scope === 'CHAIN_WIDE' ? 'Chain-Wide' : 'Store-Specific'})
        </h3>
        <span className="text-xs px-2.5 py-1 bg-amber-500/10 text-amber-400 border border-amber-500/30 rounded-full font-semibold">
          {scope}
        </span>
      </div>

      {validationError && (
        <div data-testid="offer-form-error" className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg text-rose-400 text-xs">
          {validationError}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Offer Type */}
        <div>
          <label htmlFor="offer-type-select" className="block text-xs font-semibold text-slate-400 mb-1">Offer Type</label>
          <select
            id="offer-type-select"
            aria-label="Offer Type"
            value={type}
            onChange={(e) => setType(e.target.value as any)}
            className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          >
            <option value="PERCENT_OFF">PERCENT_OFF (% Off Total Cart)</option>
            <option value="FLAT_OFF">FLAT_OFF (Flat ₹ Amount Off)</option>
            <option value="BOGO">BOGO (Buy X Get Y Free)</option>
            <option value="CATEGORY_PERCENT_OFF">CATEGORY_PERCENT_OFF (% Off Category)</option>
          </select>
        </div>

        {/* Applies To */}
        <div>
          <label htmlFor="applies-to-select" className="block text-xs font-semibold text-slate-400 mb-1">Applies To</label>
          <select
            id="applies-to-select"
            aria-label="Applies To"
            value={appliesTo}
            onChange={(e) => setAppliesTo(e.target.value as any)}
            className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          >
            <option value="ALL">ALL (Entire Store Catalog)</option>
            <option value="CATEGORY">CATEGORY (Specific Category IDs)</option>
            <option value="BARCODE_LIST">BARCODE_LIST (Specific Item Barcodes)</option>
          </select>
        </div>

        {/* Dynamic Rule Config Inputs */}
        {(type === 'PERCENT_OFF' || type === 'CATEGORY_PERCENT_OFF') && (
          <div>
            <label htmlFor="percent-input" className="block text-xs font-semibold text-slate-400 mb-1">Discount Percent (%)</label>
            <input
              type="number"
              id="percent-input"
              aria-label="Discount Percent (%)"
              value={percent}
              onChange={(e) => setPercent(Number(e.target.value))}
              min={1}
              max={90}
              placeholder="e.g. 15"
              className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
            />
          </div>
        )}

        {type === 'FLAT_OFF' && (
          <div>
            <label htmlFor="flat-amount-input" className="block text-xs font-semibold text-slate-400 mb-1">Flat Off Amount (₹)</label>
            <input
              type="number"
              id="flat-amount-input"
              aria-label="Flat Off Amount (₹)"
              value={flatAmountRs}
              onChange={(e) => setFlatAmountRs(Number(e.target.value))}
              min={1}
              placeholder="e.g. 50"
              className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
            />
          </div>
        )}

        {type === 'BOGO' && (
          <div className="flex space-x-3">
            <div className="flex-1">
              <label htmlFor="buy-qty-input" className="block text-xs font-semibold text-slate-400 mb-1">Buy Quantity</label>
              <input
                type="number"
                id="buy-qty-input"
                aria-label="Buy Quantity"
                value={buyQty}
                onChange={(e) => setBuyQty(Number(e.target.value))}
                min={1}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
              />
            </div>
            <div className="flex-1">
              <label htmlFor="get-qty-input" className="block text-xs font-semibold text-slate-400 mb-1">Get Free Quantity</label>
              <input
                type="number"
                id="get-qty-input"
                aria-label="Get Free Quantity"
                value={getQty}
                onChange={(e) => setGetQty(Number(e.target.value))}
                min={1}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
              />
            </div>
          </div>
        )}

        {/* Target IDs input */}
        {appliesTo !== 'ALL' && (
          <div className="md:col-span-2">
            <label htmlFor="target-ids-input" className="block text-xs font-semibold text-slate-400 mb-1">
              Target IDs ({appliesTo === 'CATEGORY' ? 'Comma-separated Category IDs' : 'Comma-separated Barcodes'})
            </label>
            <input
              type="text"
              id="target-ids-input"
              aria-label="Target IDs"
              value={targetIdsStr}
              onChange={(e) => setTargetIdsStr(e.target.value)}
              placeholder={appliesTo === 'CATEGORY' ? 'e.g. cat-beverages, cat-snacks' : 'e.g. 8901030300011, 4006381333931'}
              className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
            />
          </div>
        )}

        {/* Min Cart Value & Max Discount */}
        <div>
          <label htmlFor="min-cart-input" className="block text-xs font-semibold text-slate-400 mb-1">Min Cart Subtotal (₹)</label>
          <input
            type="number"
            id="min-cart-input"
            aria-label="Min Cart Subtotal (₹)"
            value={minCartValueRs}
            onChange={(e) => setMinCartValueRs(Number(e.target.value))}
            min={0}
            className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          />
        </div>

        <div>
          <label htmlFor="max-discount-input" className="block text-xs font-semibold text-slate-400 mb-1">Max Discount Cap (₹, optional)</label>
          <input
            type="number"
            id="max-discount-input"
            aria-label="Max Discount Cap (₹, optional)"
            value={maxDiscountRs}
            onChange={(e) => setMaxDiscountRs(e.target.value)}
            placeholder="Uncapped if blank"
            className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          />
        </div>

        {/* Priority */}
        <div>
          <label htmlFor="priority-input" className="block text-xs font-semibold text-slate-400 mb-1 flex items-center justify-between">
            <span>Priority</span>
            <span className="text-[10px] text-amber-400 font-normal">Lower number = evaluated first</span>
          </label>
          <input
            type="number"
            id="priority-input"
            aria-label="Priority"
            value={priority}
            onChange={(e) => setPriority(Number(e.target.value))}
            min={1}
            className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          />
        </div>

        {/* Schedule */}
        <div>
          <label htmlFor="active-from-input" className="block text-xs font-semibold text-slate-400 mb-1">Active From</label>
          <input
            type="datetime-local"
            id="active-from-input"
            aria-label="Active From"
            value={activeFrom}
            onChange={(e) => setActiveFrom(e.target.value)}
            className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          />
        </div>

        <div>
          <label htmlFor="active-until-input" className="block text-xs font-semibold text-slate-400 mb-1">Active Until (Optional)</label>
          <input
            type="datetime-local"
            id="active-until-input"
            aria-label="Active Until (Optional)"
            value={activeUntil}
            onChange={(e) => setActiveUntil(e.target.value)}
            className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
          />
        </div>
      </div>

      <div className="flex justify-end space-x-3 pt-4 border-t border-slate-800">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white rounded-lg transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          id="submit-offer-btn"
          disabled={submitting}
          className="px-4 py-2 text-xs font-bold bg-amber-500 hover:bg-amber-400 text-slate-950 rounded-lg transition-colors disabled:opacity-50"
        >
          {submitting ? 'Saving...' : mode === 'create' ? 'Create Offer' : 'Save Changes'}
        </button>
      </div>
    </form>
  );
};
