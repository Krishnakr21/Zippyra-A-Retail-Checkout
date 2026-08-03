'use client';

import React, { useEffect, useState } from 'react';

export default function StoresPage() {
  const [stores, setStores] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedStore, setSelectedStore] = useState<any>(null);

  useEffect(() => {
    async function fetchStores() {
      try {
        const token = localStorage.getItem('hq_access_token');
        const res = await fetch('http://localhost:8016/v1/chain-hq/stores', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const json = await res.json();
          setStores(json.stores || []);
        } else {
          throw new Error('Failed to fetch stores');
        }
      } catch (err) {
        // Mock fallback data for testing
        setStores([
          { id: 'store-001', name: 'Reliance Digital Flagship', city: 'Mumbai', state: 'Maharashtra', status: 'ACTIVE', capacity_max: 500, opening_time: '09:00', closing_time: '22:00' },
          { id: 'store-002', name: 'Reliance Digital Express', city: 'Delhi', state: 'Delhi', status: 'ACTIVE', capacity_max: 300, opening_time: '10:00', closing_time: '21:30' },
          { id: 'store-003', name: 'Reliance Smart Superstore', city: 'Bengaluru', state: 'Karnataka', status: 'UNDER_MAINTENANCE', capacity_max: 800, opening_time: '08:00', closing_time: '23:00' },
        ]);
      } finally {
        setLoading(false);
      }
    }

    fetchStores();
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-extrabold text-white tracking-tight">Chain Stores Directory</h1>
        <p className="text-sm text-slate-400 mt-1">Read-only view of all stores registered under your chain</p>
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
                <th className="px-6 py-4">Store Name</th>
                <th className="px-6 py-4">Location</th>
                <th className="px-6 py-4">Capacity</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {stores.map((s) => (
                <tr key={s.id} className="hover:bg-slate-800/40 transition-colors">
                  <td className="px-6 py-4 font-semibold text-white">{s.name}</td>
                  <td className="px-6 py-4 text-slate-300">{s.city}, {s.state}</td>
                  <td className="px-6 py-4 text-slate-400">{s.capacity_max || 500} shoppers</td>
                  <td className="px-6 py-4">
                    <span className={`px-3 py-1 rounded-full text-xs font-bold ${
                      s.status === 'ACTIVE'
                        ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                        : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                    }`}>
                      {s.status}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <button
                      onClick={() => setSelectedStore(s)}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-indigo-600/20 text-indigo-300 border border-indigo-500/30 hover:bg-indigo-600/30 transition-all"
                    >
                      View Details
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Read-Only Store Detail Drawer */}
      {selectedStore && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="glass-panel w-full max-w-lg rounded-2xl p-6 shadow-2xl relative border border-slate-700/60">
            <button
              onClick={() => setSelectedStore(null)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white text-xl"
            >
              &times;
            </button>
            <h2 className="text-xl font-bold text-white mb-1">{selectedStore.name}</h2>
            <p className="text-xs text-slate-400 mb-6">Read-Only Chain HQ Store View</p>

            <div className="space-y-4 text-sm">
              <div className="bg-slate-900/60 p-3.5 rounded-xl border border-slate-800">
                <span className="text-xs text-slate-500 block">Store ID</span>
                <span className="font-mono text-slate-200">{selectedStore.id}</span>
              </div>
              <div className="bg-slate-900/60 p-3.5 rounded-xl border border-slate-800">
                <span className="text-xs text-slate-500 block">Address & City</span>
                <span className="text-slate-200">{selectedStore.city}, {selectedStore.state}</span>
              </div>
              <div className="bg-slate-900/60 p-3.5 rounded-xl border border-slate-800">
                <span className="text-xs text-slate-500 block">Hours of Operation</span>
                <span className="text-slate-200">{selectedStore.opening_time || '09:00'} - {selectedStore.closing_time || '22:00'}</span>
              </div>
              <div className="bg-slate-900/60 p-3.5 rounded-xl border border-slate-800">
                <span className="text-xs text-slate-500 block">Max Capacity</span>
                <span className="text-slate-200">{selectedStore.capacity_max || 500} Max Shoppers</span>
              </div>
            </div>

            <div className="mt-6 flex justify-end">
              <button
                onClick={() => setSelectedStore(null)}
                className="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 hover:bg-slate-700 text-sm font-semibold"
              >
                Close View
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
