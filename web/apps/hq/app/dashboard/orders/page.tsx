'use client';

import React, { useEffect, useState } from 'react';

export default function ChainOrdersPage() {
  const [orders, setOrders] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedOrder, setSelectedOrder] = useState<any>(null);

  useEffect(() => {
    async function fetchOrders() {
      try {
        const token = localStorage.getItem('hq_access_token');
        const res = await fetch('http://localhost:8016/v1/chain-hq/orders', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const json = await res.json();
          setOrders(json.orders || []);
        } else {
          throw new Error('Failed to fetch orders');
        }
      } catch (err) {
        // Mock fallback data for testing
        setOrders([
          { id: 'ord-8801', store_id: 'store-001', store_name: 'Reliance Digital Flagship', total_paise: 4599900, status: 'COMPLETED', created_at: new Date().toISOString() },
          { id: 'ord-8802', store_id: 'store-002', store_name: 'Reliance Digital Express', total_paise: 129900, status: 'COMPLETED', created_at: new Date().toISOString() },
          { id: 'ord-8803', store_id: 'store-001', store_name: 'Reliance Digital Flagship', total_paise: 899000, status: 'RETURNED', created_at: new Date().toISOString() },
        ]);
      } finally {
        setLoading(false);
      }
    }

    fetchOrders();
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Chain Orders</h1>
          <p className="text-sm text-slate-400 mt-1">Itemized orders across all stores under your chain</p>
        </div>

        <div className="flex gap-3">
          <input
            type="date"
            className="px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-800 text-xs text-slate-300 focus:outline-none"
          />
          <select className="px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-800 text-xs text-slate-300 focus:outline-none">
            <option value="">All Stores</option>
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
                <th className="px-6 py-4">Order ID</th>
                <th className="px-6 py-4">Store</th>
                <th className="px-6 py-4">Amount</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4">Date</th>
                <th className="px-6 py-4">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {orders.map((o) => (
                <tr key={o.id} className="hover:bg-slate-800/40 transition-colors">
                  <td className="px-6 py-4 font-mono font-semibold text-indigo-400">{o.id}</td>
                  <td className="px-6 py-4 text-white font-medium">{o.store_name || o.store_id}</td>
                  <td className="px-6 py-4 font-bold text-white">₹{((o.total_paise || 0) / 100).toLocaleString('en-IN')}</td>
                  <td className="px-6 py-4">
                    <span className={`px-3 py-1 rounded-full text-xs font-bold ${
                      o.status === 'COMPLETED'
                        ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                        : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                    }`}>
                      {o.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-xs text-slate-400">{new Date(o.created_at).toLocaleDateString()}</td>
                  <td className="px-6 py-4">
                    <button
                      onClick={() => setSelectedOrder(o)}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 text-slate-300 hover:bg-slate-700 transition-all"
                    >
                      Drill Down
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Order Detail Modal */}
      {selectedOrder && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="glass-panel w-full max-w-lg rounded-2xl p-6 shadow-2xl relative border border-slate-700/60">
            <button
              onClick={() => setSelectedOrder(null)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white text-xl"
            >
              &times;
            </button>
            <h2 className="text-xl font-bold text-white mb-1">Order Details: {selectedOrder.id}</h2>
            <p className="text-xs text-slate-400 mb-6">Cross-Store Order Audit</p>

            <div className="space-y-4 text-sm">
              <div className="bg-slate-900/60 p-3.5 rounded-xl border border-slate-800 flex justify-between">
                <span className="text-xs text-slate-500">Store</span>
                <span className="text-slate-200 font-medium">{selectedOrder.store_name || selectedOrder.store_id}</span>
              </div>
              <div className="bg-slate-900/60 p-3.5 rounded-xl border border-slate-800 flex justify-between">
                <span className="text-xs text-slate-500">Total Amount</span>
                <span className="text-emerald-400 font-bold">₹{((selectedOrder.total_paise || 0) / 100).toLocaleString('en-IN')}</span>
              </div>
              <div className="bg-slate-900/60 p-3.5 rounded-xl border border-slate-800 flex justify-between">
                <span className="text-xs text-slate-500">Status</span>
                <span className="text-slate-200 font-semibold">{selectedOrder.status}</span>
              </div>
            </div>

            <div className="mt-6 flex justify-end">
              <button
                onClick={() => setSelectedOrder(null)}
                className="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 hover:bg-slate-700 text-sm font-semibold"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
