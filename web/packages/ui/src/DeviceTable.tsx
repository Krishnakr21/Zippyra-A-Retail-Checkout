'use client';

import React from 'react';
import { DeviceStatusBadge } from './DeviceStatusBadge';

export interface DeviceItem {
  id: string;
  store_id: string;
  chain_id: string;
  device_type: string;
  gate_id?: string;
  label: string;
  status: string;
  last_heartbeat_at?: string;
  is_stale?: boolean;
  firmware_version?: string;
}

export interface DeviceTableProps {
  devices: DeviceItem[];
  onRowClick?: (device: DeviceItem) => void;
  showActions?: boolean;
  onDecommission?: (device: DeviceItem) => void;
  onRotateCert?: (device: DeviceItem) => void;
}

function getRelativeTime(isoStr?: string): string {
  if (!isoStr) return 'Never';
  const diffSec = Math.floor((Date.now() - new Date(isoStr).getTime()) / 1000);
  if (diffSec < 60) return `${diffSec}s ago`;
  if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
  if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`;
  return `${Math.floor(diffSec / 86400)}d ago`;
}

function getDeviceTypeIcon(type: string): string {
  switch ((type || '').toUpperCase()) {
    case 'GATE':
      return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5';
    case 'RFID_PAD':
      return 'M13 10V3L4 14h7v7l9-11h-7z';
    case 'SCANNER':
      return 'M12 4v16m8-8H4';
    case 'KIOSK':
      return 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z';
    case 'PRINTER':
      return 'M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z';
    default:
      return 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2z';
  }
}

export const DeviceTable: React.FC<DeviceTableProps> = ({
  devices,
  onRowClick,
  showActions = false,
  onDecommission,
  onRotateCert,
}) => {
  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl overflow-hidden shadow-2xl">
      <table className="w-full text-left text-sm text-slate-300">
        <thead className="bg-slate-950 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-800">
          <tr>
            <th className="px-6 py-4">Device Label</th>
            <th className="px-6 py-4">Type & Hardware</th>
            <th className="px-6 py-4">Status</th>
            <th className="px-6 py-4">Last Heartbeat</th>
            {showActions && <th className="px-6 py-4 text-right">Actions</th>}
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-800/60">
          {devices.map((d) => (
            <tr
              key={d.id}
              onClick={() => onRowClick && onRowClick(d)}
              className="hover:bg-slate-800/40 cursor-pointer transition-colors"
            >
              <td className="px-6 py-4 font-semibold text-white">
                <div className="flex items-center gap-2">
                  <span>{d.label}</span>
                  {d.gate_id && (
                    <span className="px-2 py-0.5 rounded text-[10px] bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 font-mono">
                      Gate: {d.gate_id}
                    </span>
                  )}
                </div>
              </td>
              <td className="px-6 py-4">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-lg bg-slate-800 text-indigo-400">
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.75} d={getDeviceTypeIcon(d.device_type)} />
                    </svg>
                  </div>
                  <span className="font-mono text-xs text-slate-300">{d.device_type}</span>
                </div>
              </td>
              <td className="px-6 py-4">
                <DeviceStatusBadge status={d.status} />
              </td>
              <td className="px-6 py-4">
                <div className="flex items-center gap-2 text-xs">
                  <span className="text-slate-400 font-mono">{getRelativeTime(d.last_heartbeat_at)}</span>
                  {d.is_stale && (
                    <span
                      title="Heartbeat is aging toward offline threshold (>180s)"
                      className="text-amber-400 bg-amber-500/10 p-1 rounded border border-amber-500/20"
                    >
                      <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                      </svg>
                    </span>
                  )}
                </div>
              </td>
              {showActions && (
                <td className="px-6 py-4 text-right" onClick={(e) => e.stopPropagation()}>
                  {d.status !== 'DECOMMISSIONED' && (
                    <div className="flex items-center justify-end gap-2">
                      <button
                        type="button"
                        onClick={() => onRotateCert && onRotateCert(d)}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-indigo-500/10 text-indigo-300 border border-indigo-500/20 hover:bg-indigo-500/20 transition-all"
                      >
                        Rotate Cert
                      </button>
                      <button
                        type="button"
                        onClick={() => onDecommission && onDecommission(d)}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20 hover:bg-rose-500/20 transition-all"
                      >
                        Decommission
                      </button>
                    </div>
                  )}
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
