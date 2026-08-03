'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { DataTable, Column, Badge } from '@zippyra/ui';
import { Order } from '@zippyra/types';
import { useOrders } from '@zippyra/hooks';

export default function OrdersPage() {
  const router = useRouter();
  const { getStoreOrders } = useOrders();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getStoreOrders('store-001')
      .then((res) => setOrders(res.orders || []))
      .catch(() => setOrders([]))
      .finally(() => setLoading(false));
  }, []);

  const columns: Column<Order>[] = [
    {
      header: 'Order ID',
      cell: (row) => <span className="font-mono text-xs">{row.id.substring(0, 8)}</span>,
    },
    {
      header: 'Total Amount',
      cell: (row) => <span className="font-bold text-gray-900">₹{(row.total_paise / 100.0).toFixed(2)}</span>,
    },
    { header: 'Payment Method', accessorKey: 'payment_method' },
    {
      header: 'Status',
      cell: (row) => <Badge status={row.status} />,
    },
    {
      header: 'Date',
      cell: (row) => new Date(row.created_at).toLocaleString(),
    },
    {
      header: 'Action',
      cell: (row) => (
        <button
          onClick={() => router.push(`/dashboard/orders/${row.id}`)}
          className="text-blue-600 hover:text-blue-800 text-xs font-semibold"
        >
          View Details →
        </button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Store Orders</h1>
        <p className="text-sm text-gray-500 mt-1">History of customer self-checkout orders and return requests</p>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading store orders...</div>
      ) : (
        <DataTable
          columns={columns}
          data={orders}
          keyExtractor={(item) => item.id}
          emptyMessage="No store orders found."
        />
      )}
    </div>
  );
}
