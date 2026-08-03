'use client';

import React, { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useGrn } from '@zippyra/hooks';

export default function GRNDetailPage() {
  const params = useParams();
  const router = useRouter();
  const grnId = params.id as string;
  const { updateQC, completeGRN } = useGrn();

  const [qcDecisions, setQcDecisions] = useState<Record<string, 'PASSED' | 'REJECTED'>>({
    'item-1': 'PASSED',
  });
  const [status, setStatus] = useState<'QC_PENDING' | 'COMPLETED'>('QC_PENDING');
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  const handleComplete = async () => {
    setLoading(true);
    setErrorMsg('');
    try {
      // 1. Submit QC updates
      const updates = Object.entries(qcDecisions).map(([itemId, qcStatus]) => ({
        grn_line_item_id: itemId,
        qc_status: qcStatus,
      }));
      await updateQC(grnId, updates);

      // 2. Complete GRN
      await completeGRN(grnId);
      setStatus('COMPLETED');
    } catch (err: any) {
      if (err.code === 'QC_INCOMPLETE') {
        setErrorMsg('QC Incomplete: Please set QC decision for all line items before completing.');
      } else if (err.code === 'GRN_ALREADY_COMPLETED') {
        setStatus('COMPLETED');
        setErrorMsg('GRN was already completed by another staff member.');
      } else {
        setErrorMsg(err.message || 'Failed to complete GRN');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <button onClick={() => router.back()} className="text-xs text-blue-600 hover:underline mb-1">
            ← Back to GRN List
          </button>
          <h1 className="text-2xl font-bold text-gray-900">GRN #{grnId.substring(0, 8)}</h1>
          <p className="text-sm text-gray-500 mt-1">Status: {status}</p>
        </div>
        {status !== 'COMPLETED' && (
          <button
            onClick={handleComplete}
            disabled={loading}
            className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white font-semibold text-sm rounded-lg shadow-sm disabled:opacity-50"
          >
            {loading ? 'Completing...' : 'Complete GRN'}
          </button>
        )}
      </div>

      {errorMsg && (
        <div className="p-4 bg-amber-50 text-amber-800 rounded-lg text-sm border border-amber-200">
          {errorMsg}
        </div>
      )}

      {/* QC Review Items */}
      <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4">
        <h3 className="font-bold text-gray-900">Quality Control (QC) Decisions</h3>
        <div className="p-4 border rounded-lg flex items-center justify-between">
          <div>
            <p className="font-semibold text-gray-900">Item Barcode: 8901112223334</p>
            <p className="text-xs text-gray-500">Qty Received: 10 units</p>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setQcDecisions((prev) => ({ ...prev, 'item-1': 'PASSED' }))}
              className={`px-3 py-1.5 rounded text-xs font-bold ${
                qcDecisions['item-1'] === 'PASSED' ? 'bg-emerald-600 text-white' : 'bg-gray-100 text-gray-700'
              }`}
            >
              PASSED
            </button>
            <button
              onClick={() => setQcDecisions((prev) => ({ ...prev, 'item-1': 'REJECTED' }))}
              className={`px-3 py-1.5 rounded text-xs font-bold ${
                qcDecisions['item-1'] === 'REJECTED' ? 'bg-rose-600 text-white' : 'bg-gray-100 text-gray-700'
              }`}
            >
              REJECTED
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
