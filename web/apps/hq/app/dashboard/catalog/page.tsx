'use client';

import React, { useEffect, useState } from 'react';

export default function CatalogDirectoryPage() {
  const [products, setProducts] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchCatalog() {
      try {
        const token = localStorage.getItem('hq_access_token');
        const res = await fetch('http://localhost:8016/v1/chain-hq/catalog', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const json = await res.json();
          setProducts(json.products || []);
        } else {
          throw new Error('Failed to fetch catalog');
        }
      } catch (err) {
        // Mock fallback data for testing
        setProducts([
          { barcode: '8901234567890', name: 'Samsung Galaxy S24 Ultra 5G', category: 'Smartphones', price_paise: 12999900, hsn_code: '85171200' },
          { barcode: '8901234567891', name: 'Sony WH-1000XM5 Wireless Headphones', category: 'Audio', price_paise: 2999000, hsn_code: '85183000' },
          { barcode: '8901234567892', name: 'LG C3 55" 4K OLED Smart TV', category: 'Television', price_paise: 11499000, hsn_code: '85287200' },
        ]);
      } finally {
        setLoading(false);
      }
    }

    fetchCatalog();
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Chain Catalog Directory</h1>
          <p className="text-sm text-slate-400 mt-1">Read-only master catalog across all chain stores</p>
        </div>

        <div>
          <select className="px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-800 text-xs text-slate-300 focus:outline-none">
            <option value="">All Stores (Chain Master)</option>
            <option value="store-001">Reliance Digital Flagship</option>
            <option value="store-002">Reliance Digital Express</option>
          </select>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center min-h-[300px]">
          <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-500" />
        </div>
      ) : (
        <div className="glass-panel rounded-2xl overflow-hidden shadow-2xl">
          <table className="w-full text-left text-sm text-slate-300">
            <thead className="bg-slate-900/80 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-800">
              <tr>
                <th className="px-6 py-4">Barcode</th>
                <th className="px-6 py-4">Product Name</th>
                <th className="px-6 py-4">Category</th>
                <th className="px-6 py-4">HSN Code</th>
                <th className="px-6 py-4">Price</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {products.map((p) => (
                <tr key={p.barcode} className="hover:bg-slate-800/40 transition-colors">
                  <td className="px-6 py-4 font-mono text-slate-400">{p.barcode}</td>
                  <td className="px-6 py-4 font-semibold text-white">{p.name}</td>
                  <td className="px-6 py-4 text-slate-300">{p.category}</td>
                  <td className="px-6 py-4 font-mono text-xs text-slate-400">{p.hsn_code}</td>
                  <td className="px-6 py-4 font-bold text-emerald-400">₹{((p.price_paise || 0) / 100).toLocaleString('en-IN')}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
