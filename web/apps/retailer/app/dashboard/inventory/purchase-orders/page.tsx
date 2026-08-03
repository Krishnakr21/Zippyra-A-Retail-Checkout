'use client';

import React, { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { DataTable, Column, Badge } from '@zippyra/ui';
import { PurchaseOrder } from '@zippyra/types';
import { usePurchaseOrders } from '@zippyra/hooks';

function PurchaseOrdersContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { getPurchaseOrders, createPO } = usePurchaseOrders();
  const [pos, setPos] = useState<PurchaseOrder[]>([]);
  const [filterStatus, setFilterStatus] = useState<string>('ALL');
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);

  // New PO Form State
  const [vendorName, setVendorName] = useState('Acme Supplies');
  const [barcode, setBarcode] = useState(searchParams?.get('prefillBarcode') || '8901112223334');
  const [qty, setQty] = useState(Number(searchParams?.get('prefillQty')) || 50);

  useEffect(() => {
    if (searchParams?.get('prefillBarcode')) {
      setShowCreateModal(true);
    }
  }, [searchParams]);

  const loadPOs = () => {
    setLoading(true);
    const status = filterStatus === 'ALL' ? undefined : filterStatus;
    getPurchaseOrders('store-001', status)
      .then((res) => setPos(res.items || []))
      .catch(() => setPos([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadPOs();
  }, [filterStatus]);

  const handleCreatePO = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await createPO({
        store_id: 'store-001',
        vendor_name: vendorName,
        line_items: [
          {
            barcode,
            qty_ordered: qty,
            unit_cost_paise: 5000,
          },
        ],
      });
      setShowCreateModal(false);
      loadPOs();
    } catch (err: any) {
      alert(err.message || 'Failed to create PO');
    }
  };

  const columns: Column<PurchaseOrder>[] = [
    {
      header: 'PO ID',
      cell: (row) => <span className="font-mono text-xs">{row.id.substring(0, 8)}</span>,
    },
    { header: 'Vendor', accessorKey: 'vendor_name' },
    {
      header: 'Status',
      cell: (row) => <Badge status={row.status} />,
    },
    {
      header: 'Created Date',
      cell: (row) => new Date(row.created_at).toLocaleDateString(),
    },
    {
      header: 'Action',
      cell: (row) => (
        <button
          onClick={() => router.push(`/dashboard/inventory/purchase-orders/${row.id}`)}
          className="text-blue-600 hover:text-blue-800 text-xs font-semibold"
        >
          View Details →
        </button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Purchase Orders</h1>
          <p className="text-sm text-gray-500 mt-1">Vendor PO management and receipt tracking</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-semibold text-sm rounded-lg shadow-sm"
        >
          + New PO
        </button>
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2 border-b border-gray-200 pb-2">
        {['ALL', 'DRAFT', 'SUBMITTED', 'PARTIALLY_RECEIVED', 'RECEIVED', 'CANCELLED'].map((st) => (
          <button
            key={st}
            onClick={() => setFilterStatus(st)}
            className={`px-3 py-1.5 text-xs font-semibold rounded-md ${
              filterStatus === st ? 'bg-blue-600 text-white' : 'text-gray-600 hover:bg-gray-100'
            }`}
          >
            {st}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading purchase orders...</div>
      ) : (
        <DataTable
          columns={columns}
          data={pos}
          keyExtractor={(item) => item.id}
          emptyMessage="No purchase orders found."
        />
      )}

      {/* New PO Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-xl">
            <h3 className="text-lg font-bold text-gray-900 mb-4">Create Purchase Order</h3>
            <form onSubmit={handleCreatePO} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-gray-700 mb-1">Vendor Name</label>
                <input
                  type="text"
                  value={vendorName}
                  onChange={(e) => setVendorName(e.target.value)}
                  className="w-full px-3 py-2 border rounded-md text-sm"
                  required
                />
              </div>
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
                <label className="block text-xs font-semibold text-gray-700 mb-1">Quantity</label>
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
                  Create DRAFT PO
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default function PurchaseOrdersPage() {
  return (
    <Suspense fallback={<div className="p-8 text-center text-gray-500">Loading purchase orders...</div>}>
      <PurchaseOrdersContent />
    </Suspense>
  );
}
