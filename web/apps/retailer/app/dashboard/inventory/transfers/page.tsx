'use client';

import React, { useState, useEffect } from 'react';
import { DataTable, Column, Badge } from '@zippyra/ui';
import { TransferOrder } from '@zippyra/types';

export default function TransfersPage() {
  const [transfers, setTransfers] = useState<TransferOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<'OUTGOING' | 'INCOMING'>('OUTGOING');
  const [errorMsg, setErrorMsg] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);

  // Form states
  const [destStoreId, setDestStoreId] = useState('44444444-4444-4444-4444-444444444444');
  const [barcode, setBarcode] = useState('8901030012345');
  const [qty, setQty] = useState(20);

  const currentStoreId = '33333333-3333-3333-3333-333333333333';

  const fetchTransfers = async () => {
    setLoading(true);
    setErrorMsg('');
    try {
      const res = await fetch(`/api/inventory/transfers?store_id=${currentStoreId}`);
      if (res.ok) {
        const data = await res.json();
        setTransfers(data.transfers || []);
      } else {
        setErrorMsg('Failed to load inter-store transfers');
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'Error loading transfers');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTransfers();
  }, []);

  const handleCreateTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg('');
    try {
      const res = await fetch('/api/inventory/transfers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          source_store_id: currentStoreId,
          dest_store_id: destStoreId,
          line_items: [{ barcode, qty_requested: qty }],
        }),
      });

      if (!res.ok) {
        throw new Error('Failed to submit transfer request');
      }

      setShowCreateModal(false);
      fetchTransfers();
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to create transfer');
    }
  };

  const handleApprove = async (id: string) => {
    try {
      await fetch(`/api/warehouse/transfer/${id}/approve`, { method: 'PUT' });
      fetchTransfers();
    } catch (err: any) {
      setErrorMsg('Failed to approve transfer');
    }
  };

  const handleShip = async (id: string) => {
    try {
      await fetch(`/api/warehouse/transfer/${id}/ship`, { method: 'PUT' });
      fetchTransfers();
    } catch (err: any) {
      setErrorMsg('Failed to ship transfer');
    }
  };

  const handleReceive = async (id: string) => {
    try {
      await fetch(`/api/warehouse/transfer/${id}/receive`, { method: 'PUT' });
      fetchTransfers();
    } catch (err: any) {
      setErrorMsg('Failed to receive transfer');
    }
  };

  const filteredTransfers = transfers.filter((t) => {
    if (tab === 'OUTGOING') return t.source_store_id === currentStoreId;
    return t.dest_store_id === currentStoreId;
  });

  const columns: Column<TransferOrder>[] = [
    {
      header: 'Transfer ID',
      cell: (row) => <span className="font-mono text-xs">{row.id.substring(0, 8)}</span>,
    },
    {
      header: 'Direction',
      cell: (row) => (
        <span className="text-xs font-semibold px-2 py-0.5 rounded bg-gray-100 text-gray-800">
          {row.source_store_id === currentStoreId ? 'OUTGOING' : 'INCOMING'}
        </span>
      ),
    },
    { header: 'Dest Store', cell: (row) => <span className="font-mono text-xs">{row.dest_store_id.substring(0, 8)}</span> },
    {
      header: 'Status',
      cell: (row) => <Badge status={row.status} />,
    },
    {
      header: 'Actions',
      cell: (row) => {
        return (
          <div className="flex gap-2">
            {row.status === 'REQUESTED' && (
              <button
                onClick={() => handleApprove(row.id)}
                className="px-2.5 py-1 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold rounded"
              >
                Approve
              </button>
            )}
            {row.status === 'APPROVED' && (
              <button
                onClick={() => handleShip(row.id)}
                className="px-2.5 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded"
              >
                Ship Stock
              </button>
            )}
            {row.status === 'IN_TRANSIT' && (
              <button
                onClick={() => handleReceive(row.id)}
                className="px-2.5 py-1 bg-purple-600 hover:bg-purple-700 text-white text-xs font-semibold rounded"
              >
                Receive Stock
              </button>
            )}
          </div>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Inter-Store Transfers</h1>
          <p className="text-sm text-gray-500 mt-1">Manage stock transfer requests between chain stores</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-semibold text-sm rounded-lg shadow-sm"
        >
          + Request Transfer
        </button>
      </div>

      {errorMsg && (
        <div className="p-4 bg-rose-50 text-rose-800 rounded-lg text-sm border border-rose-200">
          {errorMsg}
        </div>
      )}

      {/* Filter Tabs */}
      <div className="flex gap-2 border-b border-gray-200 pb-2">
        <button
          onClick={() => setTab('OUTGOING')}
          className={`px-3 py-1.5 text-xs font-semibold rounded-md ${
            tab === 'OUTGOING' ? 'bg-blue-600 text-white' : 'text-gray-600 hover:bg-gray-100'
          }`}
        >
          Outgoing Transfers ({transfers.filter((t) => t.source_store_id === currentStoreId).length})
        </button>
        <button
          onClick={() => setTab('INCOMING')}
          className={`px-3 py-1.5 text-xs font-semibold rounded-md ${
            tab === 'INCOMING' ? 'bg-blue-600 text-white' : 'text-gray-600 hover:bg-gray-100'
          }`}
        >
          Incoming Transfers ({transfers.filter((t) => t.dest_store_id === currentStoreId).length})
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading transfers...</div>
      ) : (
        <DataTable
          columns={columns}
          data={filteredTransfers}
          keyExtractor={(item) => item.id}
          emptyMessage="No inter-store transfers found."
        />
      )}

      {/* Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-xl">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Request Inter-Store Transfer</h3>
            <form onSubmit={handleCreateTransfer} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-1">Destination Store ID</label>
                <input
                  type="text"
                  value={destStoreId}
                  onChange={(e) => setDestStoreId(e.target.value)}
                  className="w-full px-3 py-2 border rounded-md text-sm font-mono"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-1">Item Barcode</label>
                <input
                  type="text"
                  value={barcode}
                  onChange={(e) => setBarcode(e.target.value)}
                  className="w-full px-3 py-2 border rounded-md text-sm font-mono"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-1">Qty Requested</label>
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
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-md"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white text-sm font-semibold rounded-md"
                >
                  Submit Request
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
