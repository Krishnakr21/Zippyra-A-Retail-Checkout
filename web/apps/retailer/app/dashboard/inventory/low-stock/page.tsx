'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { DataTable, Column } from '@zippyra/ui';
import { LowStockItem } from '@zippyra/types';
import { useInventory, useCatalog } from '@zippyra/hooks';

export default function LowStockPage() {
  const router = useRouter();
  const { getLowStockItems } = useInventory();
  const { getProductByBarcode } = useCatalog();
  const [items, setItems] = useState<LowStockItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [names, setNames] = useState<Record<string, string>>({});

  useEffect(() => {
    getLowStockItems('store-001')
      .then((res) => {
        setItems(res.items);
        // Lazily resolve product names per barcode
        res.items.forEach((item) => {
          getProductByBarcode('store-001', item.barcode)
            .then((p) => {
              if (p && p.name) {
                setNames((prev) => ({ ...prev, [item.barcode]: p.name }));
              }
            })
            .catch(() => {});
        });
      })
      .finally(() => setLoading(false));
  }, []);

  const columns: Column<LowStockItem>[] = [
    { header: 'Barcode', accessorKey: 'barcode' },
    {
      header: 'Product Name',
      cell: (row) => names[row.barcode] || row.product_name || row.barcode,
    },
    {
      header: 'On-Hand Qty',
      cell: (row) => <span className="font-bold text-red-600">{row.on_hand_qty}</span>,
    },
    { header: 'Reorder Point', accessorKey: 'reorder_point' },
    { header: 'Reorder Qty', accessorKey: 'reorder_qty' },
    {
      header: 'Action',
      cell: (row) => (
        <button
          onClick={() =>
            router.push(`/dashboard/inventory/purchase-orders?prefillBarcode=${row.barcode}&prefillQty=${row.reorder_qty}`)
          }
          className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-md"
        >
          Create PO
        </button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Low Stock List</h1>
        <p className="text-sm text-gray-500 mt-1">Items currently below their configured reorder point threshold</p>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading low-stock items...</div>
      ) : (
        <DataTable
          columns={columns}
          data={items}
          keyExtractor={(item) => item.barcode}
          emptyMessage="No low-stock items right now."
        />
      )}
    </div>
  );
}
