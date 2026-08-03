'use client';

import React, { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Badge } from '@zippyra/ui';
import { PurchaseOrder } from '@zippyra/types';
import { usePurchaseOrders } from '@zippyra/hooks';

export default function PODetailPage() {
  const params = useParams();
  const router = useRouter();
  const poId = params.id as string;
  const { getPODetail, submitPO } = usePurchaseOrders();
  const [po, setPo] = useState<PurchaseOrder | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState('');

  const loadPO = () => {
    setLoading(true);
    getPODetail(poId)
      .then((data) => setPo(data))
      .catch((err) => setMessage(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (poId) loadPO();
  }, [poId]);

  const handleSubmitPO = async () => {
    setSubmitting(true);
    setMessage('');
    try {
      await submitPO(poId);
      loadPO();
    } catch (err: any) {
      if (err.code === 'PO_ALREADY_SUBMITTED') {
        setMessage('Purchase Order is already submitted.');
      } else {
        setMessage(err.message || 'Failed to submit PO');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <div className="p-8 text-center text-gray-500">Loading PO details...</div>;
  if (!po) return <div className="p-8 text-center text-red-500">{message || 'PO not found'}</div>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <button onClick={() => router.back()} className="text-xs text-blue-600 hover:underline mb-1">
            ← Back to PO List
          </button>
          <h1 className="text-2xl font-bold text-gray-900">PO #{po.id.substring(0, 8)}</h1>
          <p className="text-sm text-gray-500 mt-1">Vendor: {po.vendor_name}</p>
        </div>
        <div className="flex items-center gap-4">
          <Badge status={po.status} />
          {po.status === 'DRAFT' && (
            <button
              onClick={handleSubmitPO}
              disabled={submitting}
              className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-sm rounded-lg shadow-sm disabled:opacity-50"
            >
              {submitting ? 'Submitting...' : 'Submit PO'}
            </button>
          )}
        </div>
      </div>

      {message && (
        <div className="p-4 bg-amber-50 text-amber-800 rounded-lg text-sm border border-amber-200">
          {message}
        </div>
      )}

      {/* Progress Card */}
      <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4">
        <h3 className="font-bold text-gray-900">Item Fulfillment Progress</h3>
        {(po.line_items || []).map((item, idx) => {
          const ordered = item.qty_ordered || 1;
          const received = item.qty_received || 0;
          const pct = Math.min(100, Math.round((received / ordered) * 100));

          return (
            <div key={idx} className="p-4 bg-gray-50 rounded-lg space-y-2">
              <div className="flex justify-between text-sm">
                <span className="font-semibold text-gray-900">Barcode: {item.barcode}</span>
                <span className="text-gray-600 font-medium">
                  {received} / {ordered} received ({pct}%)
                </span>
              </div>
              <div className="w-full bg-gray-200 h-2.5 rounded-full overflow-hidden">
                <div className="bg-blue-600 h-2.5 rounded-full" style={{ width: `${pct}%` }} />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
