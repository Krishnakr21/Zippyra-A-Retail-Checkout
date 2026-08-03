'use client';

import React, { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useCatalog } from '@zippyra/hooks';

export default function ProductEditPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;
  const isNew = id === 'new';
  const { createProduct, updateProduct } = useCatalog();

  const [name, setName] = useState('Sample Product');
  const [barcode, setBarcode] = useState('8901112223334');
  const [priceRupees, setPriceRupees] = useState(150);
  const [mrpRupees, setMrpRupees] = useState(199);
  const [hsnCode, setHsnCode] = useState('1001');
  const [categoryId, setCategoryId] = useState('cat-grocery');
  const [imageUrl, setImageUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMsg('');

    const payload = {
      store_id: 'store-001',
      chain_id: 'chain-1',
      name,
      barcode,
      price_paise: Math.round(priceRupees * 100),
      mrp_paise: Math.round(mrpRupees * 100),
      hsn_code: hsnCode,
      category_id: categoryId,
      image_url: imageUrl,
      is_active: true,
      is_returnable: true,
    };

    try {
      if (isNew) {
        await createProduct(payload);
      } else {
        await updateProduct(id, payload);
      }
      setMsg('Product saved successfully.');
      router.push('/dashboard/catalog');
    } catch (err: any) {
      setMsg(err.message || 'Failed to save product');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <button onClick={() => router.back()} className="text-xs text-blue-600 hover:underline mb-1">
          ← Back to Catalog
        </button>
        <h1 className="text-2xl font-bold text-gray-900">{isNew ? 'Add New Product' : 'Edit Product'}</h1>
      </div>

      {msg && (
        <div className="p-4 bg-blue-50 text-blue-800 rounded-lg text-sm border border-blue-200">
          {msg}
        </div>
      )}

      <form onSubmit={handleSubmit} className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4">
        <div>
          <label className="block text-xs font-semibold text-gray-700 mb-1">Product Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 border rounded-md text-sm"
            required
          />
        </div>
        <div>
          <label className="block text-xs font-semibold text-gray-700 mb-1">Barcode</label>
          <input
            type="text"
            value={barcode}
            onChange={(e) => setBarcode(e.target.value)}
            className="w-full px-3 py-2 border rounded-md text-sm"
            required
          />
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold text-gray-700 mb-1">Selling Price (₹)</label>
            <input
              type="number"
              value={priceRupees}
              onChange={(e) => setPriceRupees(Number(e.target.value))}
              className="w-full px-3 py-2 border rounded-md text-sm"
              required
            />
          </div>
          <div>
            <label className="block text-xs font-semibold text-gray-700 mb-1">MRP (₹)</label>
            <input
              type="number"
              value={mrpRupees}
              onChange={(e) => setMrpRupees(Number(e.target.value))}
              className="w-full px-3 py-2 border rounded-md text-sm"
              required
            />
          </div>
        </div>
        <div>
          <label className="block text-xs font-semibold text-gray-700 mb-1">HSN Code</label>
          <input
            type="text"
            value={hsnCode}
            onChange={(e) => setHsnCode(e.target.value)}
            className="w-full px-3 py-2 border rounded-md text-sm"
            required
          />
        </div>

        <div className="pt-4 flex justify-end gap-3">
          <button
            type="button"
            onClick={() => router.back()}
            className="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-md"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-md disabled:opacity-50"
          >
            {loading ? 'Saving...' : 'Save Product'}
          </button>
        </div>
      </form>
    </div>
  );
}
