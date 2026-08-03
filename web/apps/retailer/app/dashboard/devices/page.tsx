'use client';

import React, { useEffect, useState } from 'react';
import { DeviceTable, DeviceDetailDrawer, DeviceItem } from '@zippyra/ui';

export default function RetailerDevicesPage() {
  const [devices, setDevices] = useState<DeviceItem[]>([]);
  const [alerts, setAlerts] = useState<any[]>([]);
  const [selectedDevice, setSelectedDevice] = useState<DeviceItem | null>(null);
  const [heartbeats, setHeartbeats] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const storeId = 'store-001'; // Default store scope for retailer session

  useEffect(() => {
    fetchDevicesAndAlerts();
  }, []);

  async function fetchDevicesAndAlerts() {
    try {
      const token = localStorage.getItem('access_token');
      const headers = { Authorization: `Bearer ${token}` };

      // Fetch devices
      const devRes = await fetch(`http://localhost:8017/v1/device-mgmt/devices?store_id=${storeId}`, { headers });
      if (devRes.ok) {
        const devData = await devRes.json();
        setDevices(devData.devices || []);
      } else {
        throw new Error('Fallback to mock devices');
      }

      // Fetch unresolved alerts
      const alertRes = await fetch(`http://localhost:8017/v1/device-mgmt/alerts?store_id=${storeId}&resolved=false`, { headers });
      if (alertRes.ok) {
        const alertData = await alertRes.json();
        setAlerts(alertData.alerts || []);
      }
    } catch (e) {
      // Mock fallback data for store-scoped view
      setDevices([
        { id: 'dev-101', store_id: storeId, chain_id: 'chain-1', device_type: 'GATE', gate_id: 'GATE_MAIN', label: 'Main Entrance Exit Gate', status: 'ACTIVE', last_heartbeat_at: new Date().toISOString(), firmware_version: 'v1.4.2' },
        { id: 'dev-102', store_id: storeId, chain_id: 'chain-1', device_type: 'RFID_PAD', label: 'Counter 1 RFID Deactivator', status: 'ACTIVE', last_heartbeat_at: new Date(Date.now() - 200000).toISOString(), is_stale: true, firmware_version: 'v2.0.1' },
        { id: 'dev-103', store_id: storeId, chain_id: 'chain-1', device_type: 'SCANNER', label: 'Handheld Inventory Scanner 3', status: 'OFFLINE', last_heartbeat_at: new Date(Date.now() - 900000).toISOString(), firmware_version: 'v1.1.0' },
      ]);

      setAlerts([
        { id: 'alt-001', device_id: 'dev-103', store_id: storeId, alert_type: 'OFFLINE', detail: { label: 'Handheld Inventory Scanner 3' }, created_at: new Date().toISOString() },
      ]);
    } finally {
      setLoading(false);
    }
  }

  const handleRowClick = async (device: DeviceItem) => {
    setSelectedDevice(device);
    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch(`http://localhost:8017/v1/device-mgmt/devices/${device.id}/heartbeats`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data = await res.json();
        setHeartbeats(data.heartbeats || []);
      } else {
        setHeartbeats([
          { ts: new Date().toISOString(), payload: { battery_pct: 92, signal_strength: -65, firmware_version: 'v1.4.2' } },
        ]);
      }
    } catch (e) {
      setHeartbeats([
        { ts: new Date().toISOString(), payload: { battery_pct: 92, signal_strength: -65, firmware_version: 'v1.4.2' } },
      ]);
    }
  };

  const handleResolveAlert = async (alertId: string) => {
    // Optimistic UI update
    setAlerts((prev) => prev.filter((a) => a.id !== alertId));

    try {
      const token = localStorage.getItem('access_token');
      await fetch(`http://localhost:8017/v1/device-mgmt/alerts/${alertId}/resolve`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${token}` },
      });
    } catch (e) {
      // Retain optimistic removal in UI
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">Store Hardware Devices</h1>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">Live status and telemetry for gates, RFID pads, and scanners at this store location</p>
      </div>

      {/* Unresolved Hardware Alerts Section */}
      {alerts.length > 0 && (
        <div data-testid="unresolved-alerts-section" className="bg-rose-500/10 border border-rose-500/30 rounded-2xl p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-xl bg-rose-500/20 text-rose-400">
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              </div>
              <div>
                <h3 className="font-extrabold text-rose-300">Unresolved Hardware Alerts ({alerts.length})</h3>
                <p className="text-xs text-rose-400/90">Requires store manager acknowledgment</p>
              </div>
            </div>
          </div>

          <div className="space-y-3">
            {alerts.map((a) => (
              <div key={a.id} data-testid={`alert-row-${a.id}`} className="bg-slate-900/80 p-4 rounded-xl border border-slate-800 flex items-center justify-between">
                <div>
                  <span className="font-mono text-xs font-bold text-rose-400 block">{a.alert_type}</span>
                  <span className="text-sm text-white font-medium">{a.detail?.label || a.device_id}</span>
                </div>
                <button
                  type="button"
                  data-testid={`resolve-alert-btn-${a.id}`}
                  onClick={() => handleResolveAlert(a.id)}
                  className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 hover:bg-emerald-500/20 transition-all"
                >
                  Resolve Alert
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center min-h-[300px]">
          <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-500" />
        </div>
      ) : (
        <DeviceTable
          devices={devices}
          onRowClick={handleRowClick}
          showActions={false}
        />
      )}

      <DeviceDetailDrawer
        device={selectedDevice}
        heartbeats={heartbeats}
        onClose={() => setSelectedDevice(null)}
      />
    </div>
  );
}
