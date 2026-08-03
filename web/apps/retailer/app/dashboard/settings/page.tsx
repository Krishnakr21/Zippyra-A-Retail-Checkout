'use client';

import React, { useState } from 'react';

export default function StoreSettingsPage() {
  const [openingTime, setOpeningTime] = useState('08:00');
  const [closingTime, setClosingTime] = useState('22:00');
  const [timezone, setTimezone] = useState('Asia/Kolkata');
  const [capacityMax, setCapacityMax] = useState(50);
  const [loadingHours, setLoadingHours] = useState(false);
  const [loadingCapacity, setLoadingCapacity] = useState(false);
  const [msg, setMsg] = useState('');

  const handleUpdateHours = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoadingHours(true);
    setMsg('');
    try {
      const res = await fetch('/api/store/hours', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          store_id: 'store-001',
          opening_time: openingTime,
          closing_time: closingTime,
          timezone,
        }),
      });
      if (!res.ok) throw new Error('Failed to update hours');
      setMsg('Store operating hours updated successfully.');
    } catch (err: any) {
      setMsg(err.message || 'Error updating hours');
    } finally {
      setLoadingHours(false);
    }
  };

  const handleUpdateCapacity = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoadingCapacity(true);
    setMsg('');
    try {
      const res = await fetch('/api/store/capacity', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          store_id: 'store-001',
          capacity_max: capacityMax,
        }),
      });
      if (!res.ok) throw new Error('Failed to update capacity');
      setMsg('Maximum in-store customer capacity updated successfully.');
    } catch (err: any) {
      setMsg(err.message || 'Error updating capacity');
    } finally {
      setLoadingCapacity(false);
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Store Settings</h1>
        <p className="text-sm text-gray-500 mt-1">Manage operating hours and maximum occupancy capacity for Downtown Superstore</p>
      </div>

      {msg && (
        <div className="p-4 bg-emerald-50 text-emerald-800 rounded-lg text-sm border border-emerald-200">
          {msg}
        </div>
      )}

      {/* Hours Form */}
      <form onSubmit={handleUpdateHours} className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4">
        <h3 className="font-bold text-gray-900 text-lg">Operating Hours</h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold text-gray-700 mb-1">Opening Time</label>
            <input
              type="time"
              value={openingTime}
              onChange={(e) => setOpeningTime(e.target.value)}
              className="w-full px-3 py-2 border rounded-md text-sm"
              required
            />
          </div>
          <div>
            <label className="block text-xs font-semibold text-gray-700 mb-1">Closing Time</label>
            <input
              type="time"
              value={closingTime}
              onChange={(e) => setClosingTime(e.target.value)}
              className="w-full px-3 py-2 border rounded-md text-sm"
              required
            />
          </div>
        </div>
        <div>
          <label className="block text-xs font-semibold text-gray-700 mb-1">Timezone</label>
          <input
            type="text"
            value={timezone}
            onChange={(e) => setTimezone(e.target.value)}
            className="w-full px-3 py-2 border rounded-md text-sm"
            required
          />
        </div>
        <div className="pt-2 flex justify-end">
          <button
            type="submit"
            disabled={loadingHours}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-md disabled:opacity-50"
          >
            {loadingHours ? 'Saving...' : 'Update Operating Hours'}
          </button>
        </div>
      </form>

      {/* Capacity Form */}
      <form onSubmit={handleUpdateCapacity} className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4">
        <h3 className="font-bold text-gray-900 text-lg">In-Store Capacity</h3>
        <div>
          <label className="block text-xs font-semibold text-gray-700 mb-1">Maximum Simultaneous Customers</label>
          <input
            type="number"
            value={capacityMax}
            onChange={(e) => setCapacityMax(Number(e.target.value))}
            className="w-full px-3 py-2 border rounded-md text-sm"
            min={1}
            required
          />
        </div>
        <div className="pt-2 flex justify-end">
          <button
            type="submit"
            disabled={loadingCapacity}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-md disabled:opacity-50"
          >
            {loadingCapacity ? 'Saving...' : 'Update Max Capacity'}
          </button>
        </div>
      </form>
    </div>
  );
}
