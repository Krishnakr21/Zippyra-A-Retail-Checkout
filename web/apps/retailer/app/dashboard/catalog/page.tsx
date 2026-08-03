'use client';

import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { DataTable, Column, Badge } from '@zippyra/ui';
import { Product } from '@zippyra/types';
import { useCatalog } from '@zippyra/hooks';

export default function CatalogPage() {
  const router = useRouter();
  const { getProducts } = useCatalog();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getProducts('store-001')
      .then((res) => setProducts(res.products || []))
      .catch(() => setProducts([]))
      .finally(() => setLoading(false));
  }, []);

  const columns: Column<Product>[] = [
    { header: 'Barcode', accessorKey: 'barcode' },
    { header: 'Product Name', accessorKey: 'name' },
    {
      header: 'Price',
      cell: (row) => <span className="font-bold">₹{(row.price_paise / 100.0).toFixed(2)}</span>,
    },
    {
      header: 'MRP',
      cell: (row) => <span className="text-gray-500 line-through">₹{(row.mrp_paise / 100.0).toFixed(2)}</span>,
    },
    { header: 'HSN Code', accessorKey: 'hsn_code' },
    {
      header: 'Status',
      cell: (row) => <Badge status={row.is_active ? 'ACTIVE' : 'INACTIVE'} />,
    },
    {
      header: 'Action',
      cell: (row) => (
        <button
          onClick={() => router.push(`/dashboard/catalog/${row.id}`)}
          className="text-blue-600 hover:text-blue-800 text-xs font-semibold"
        >
          Edit
        </button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Product Catalog</h1>
          <p className="text-sm text-gray-500 mt-1">Manage store inventory catalog items and prices</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => router.push('/dashboard/catalog/import')}
            className="px-4 py-2 bg-slate-800 hover:bg-slate-900 text-white font-semibold text-sm rounded-lg shadow-sm"
          >
            📥 Bulk CSV Import
          </button>
          <button
            onClick={() => router.push('/dashboard/catalog/new')}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-semibold text-sm rounded-lg shadow-sm"
          >
            + Add Product
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading catalog items...</div>
      ) : (
        <DataTable
          columns={columns}
          data={products}
          keyExtractor={(item) => item.id}
          emptyMessage="No catalog items found for this store."
        />
      )}
    </div>
  );
}
