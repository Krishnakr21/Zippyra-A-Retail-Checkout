'use client';

import React, { useEffect, useState } from 'react';
import { StatCard } from '@zippyra/ui';
import { useInventory } from '@zippyra/hooks';

export default function DashboardPage() {
  const { getLowStockItems } = useInventory();
  const [lowStockCount, setLowStockCount] = useState<number | null>(null);

  useEffect(() => {
    getLowStockItems('store-001')
      .then((res) => setLowStockCount(res.items.length))
      .catch(() => setLowStockCount(0));
  }, []);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard Overview</h1>
        <p className="text-sm text-gray-500 mt-1">Real-time status of store operations and inventory</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Low Stock Alerts"
          value={lowStockCount !== null ? lowStockCount : '...'}
          subtitle="Items below reorder point"
        />
        <StatCard
          title="Pending GRN Receipts"
          value={3}
          subtitle="Goods received notes awaiting receipt/QC"
        />
        <StatCard
          title="Today's Orders"
          value={42}
          subtitle="Orders processed today"
        />
        <StatCard
          title="Sales Volume Trend"
          value=""
          isComingSoon={true}
          subtitle="analytics-service dashboard integration"
        />
      </div>

      <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
        <h2 className="text-lg font-bold text-gray-900 mb-2">Quick Navigation</h2>
        <p className="text-sm text-gray-500 mb-6">Access key operational features for Downtown Superstore</p>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <a
            href="/dashboard/inventory/low-stock"
            className="p-4 rounded-lg bg-blue-50 border border-blue-100 hover:bg-blue-100 transition-colors"
          >
            <h3 className="font-semibold text-blue-900">Low Stock Reorders</h3>
            <p className="text-xs text-blue-700 mt-1">Review items needing PO creation</p>
          </a>
          <a
            href="/dashboard/inventory/grn"
            className="p-4 rounded-lg bg-purple-50 border border-purple-100 hover:bg-purple-100 transition-colors"
          >
            <h3 className="font-semibold text-purple-900">GRN Receive & QC</h3>
            <p className="text-xs text-purple-700 mt-1">Process warehouse deliveries</p>
          </a>
          <a
            href="/dashboard/orders"
            className="p-4 rounded-lg bg-emerald-50 border border-emerald-100 hover:bg-emerald-100 transition-colors"
          >
            <h3 className="font-semibold text-emerald-900">Store Orders</h3>
            <p className="text-xs text-emerald-700 mt-1">View order history and return requests</p>
          </a>
        </div>
      </div>
    </div>
  );
}
