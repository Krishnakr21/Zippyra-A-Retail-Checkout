'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useGrn } from '@zippyra/hooks';

export default function GrnListPage() {
  const router = useRouter();
  const { createGRN } = useGrn();
  const [showAdHocModal, setShowAdHocModal] = useState(false);
  const [barcode, setBarcode] = useState('8901112223334');
  const [qty, setQty] = useState(10);

  const handleCreateAdHoc = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await createGRN({
        store_id: 'store-001',
        line_items: [
          {
            id: 'item-1',
            barcode,
            qty_received: qty,
            unit_cost_paise: 1000,
            qc_status: 'PENDING',
          },
        ],
      });
      setShowAdHocModal(false);
      router.push(`/dashboard/inventory/grn/${res.id}`);
    } catch (err: any) {
      alert(err.message || 'Failed to create GRN');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Goods Received Notes (GRN)</h1>
          <p className="text-sm text-gray-500 mt-1">Process warehouse deliveries and complete Quality Control reviews</p>
        </div>
        <button
          onClick={() => setShowAdHocModal(true)}
          className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white font-semibold text-sm rounded-lg shadow-sm"
        >
          + New Ad-hoc GRN
        </button>
      </div>

      <div className="p-8 bg-white rounded-xl border border-gray-200 shadow-sm text-center">
        <h3 className="font-bold text-gray-900 text-lg">Receive Deliveries</h3>
        <p className="text-sm text-gray-500 mt-1 max-w-md mx-auto">
          Click "+ New Ad-hoc GRN" to create an un-linked delivery receipt or process a Purchase Order delivery.
        </p>
      </div>

      {showAdHocModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-xl">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Create Ad-hoc GRN</h3>
            <form onSubmit={handleCreateAdHoc} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-1">Item Barcode</label>
                <input
                  type="text"
                  value={barcode}
                  onChange={(e) => setBarcode(e.target.value)}
                  className="w-full px-3 py-2 border rounded-md text-sm"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-1">Qty Received</label>
                <input
                  type="number"
                  value={qty}
                  onChange={(e) => setQty(Number(e.target.value))}
                  className="w-full px-3 py-2 border rounded-md text-sm"
                  required
                />
              </div>
              <div className="flex justify-end gap-3 mt-6">
                <button
                  type="button"
                  onClick={() => setShowAdHocModal(false)}
                  className="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-md"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-purple-600 text-white text-sm font-semibold rounded-md"
                >
                  Create GRN
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
