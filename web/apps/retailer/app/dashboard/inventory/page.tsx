'use client';

import React from 'react';
import Link from 'next/link';

export default function InventoryOverviewPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Inventory Management</h1>
        <p className="text-sm text-gray-500 mt-1">Manage stock levels, purchase orders, goods receipts, and inter-store transfers</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Link href="/dashboard/inventory/low-stock" className="p-6 bg-white rounded-xl border border-gray-200 shadow-sm hover:border-blue-500 transition-colors">
          <h3 className="font-bold text-lg text-gray-900">Low Stock List</h3>
          <p className="text-xs text-gray-500 mt-1">View items below reorder point & trigger POs</p>
        </Link>
        <Link href="/dashboard/inventory/purchase-orders" className="p-6 bg-white rounded-xl border border-gray-200 shadow-sm hover:border-blue-500 transition-colors">
          <h3 className="font-bold text-lg text-gray-900">Purchase Orders</h3>
          <p className="text-xs text-gray-500 mt-1">Create, submit, and track vendor POs</p>
        </Link>
        <Link href="/dashboard/inventory/grn" className="p-6 bg-white rounded-xl border border-gray-200 shadow-sm hover:border-blue-500 transition-colors">
          <h3 className="font-bold text-lg text-gray-900">GRN & QC</h3>
          <p className="text-xs text-gray-500 mt-1">Receive warehouse deliveries & complete QC</p>
        </Link>
        <Link href="/dashboard/inventory/transfers" className="p-6 bg-white rounded-xl border border-gray-200 shadow-sm hover:border-blue-500 transition-colors">
          <h3 className="font-bold text-lg text-gray-900">Inter-Store Transfers</h3>
          <p className="text-xs text-gray-500 mt-1">Manage outgoing and incoming stock transfers</p>
        </Link>
      </div>
    </div>
  );
}
