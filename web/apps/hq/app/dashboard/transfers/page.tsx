'use client';

import React, { useEffect, useState } from 'react';

export default function InterStoreTransfersPage() {
  const [transfers, setTransfers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchTransfers() {
      try {
        const token = localStorage.getItem('hq_access_token');
        const res = await fetch('http://localhost:8016/v1/chain-hq/transfers', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (res.ok) {
          const json = await res.json();
          setTransfers(json.transfers || []);
        } else {
          throw new Error('Failed to fetch transfers');
        }
      } catch (err) {
        // Mock fallback data for testing
        setTransfers([
          { id: 'tr-101', source_store_id: 'Reliance Digital Flagship (Mumbai)', dest_store_id: 'Reliance Digital Express (Delhi)', status: 'IN_TRANSIT', items_count: 15, created_at: new Date().toISOString() },
          { id: 'tr-102', source_store_id: 'Reliance Smart Superstore (Bengaluru)', dest_store_id: 'Reliance Digital Flagship (Mumbai)', status: 'RECEIVED', items_count: 42, created_at: new Date().toISOString() },
        ]);
      } finally {
        setLoading(false);
      }
    }

    fetchTransfers();
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Inter-Store Stock Transfers</h1>
          <p className="text-sm text-slate-400 mt-1">Read-only chain-wide visibility into stock movements</p>
        </div>

        <div>
          <select className="px-3.5 py-2 rounded-xl bg-slate-900 border border-slate-800 text-xs text-slate-300 focus:outline-none">
            <option value="">All Statuses</option>
            <option value="REQUESTED">REQUESTED</option>
            <option value="IN_TRANSIT">IN_TRANSIT</option>
            <option value="RECEIVED">RECEIVED</option>
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
                <th className="px-6 py-4">Transfer ID</th>
                <th className="px-6 py-4">Source Store</th>
                <th className="px-6 py-4">Destination Store</th>
                <th className="px-6 py-4">Items Count</th>
                <th className="px-6 py-4">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {transfers.map((t) => (
                <tr key={t.id} className="hover:bg-slate-800/40 transition-colors">
                  <td className="px-6 py-4 font-mono font-semibold text-indigo-400">{t.id}</td>
                  <td className="px-6 py-4 text-white font-medium">{t.source_store_id}</td>
                  <td className="px-6 py-4 text-white font-medium">{t.dest_store_id}</td>
                  <td className="px-6 py-4 text-slate-300">{t.items_count || 10} SKUs</td>
                  <td className="px-6 py-4">
                    <span className={`px-3 py-1 rounded-full text-xs font-bold ${
                      t.status === 'RECEIVED'
                        ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                        : t.status === 'IN_TRANSIT'
                        ? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20'
                        : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                    }`}>
                      {t.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
