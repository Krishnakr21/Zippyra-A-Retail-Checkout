'use client';

import React from 'react';
import { DeviceItem } from './DeviceTable';
import { DeviceStatusBadge } from './DeviceStatusBadge';

export interface DeviceDetailDrawerProps {
  device: DeviceItem | null;
  heartbeats?: any[];
  onClose: () => void;
}

export const DeviceDetailDrawer: React.FC<DeviceDetailDrawerProps> = ({
  device,
  heartbeats = [],
  onClose,
}) => {
  if (!device) return null;

  // Extract generic telemetry keys from heartbeats
  const telemetryKeys = new Set<string>();
  heartbeats.forEach((hb) => {
    if (hb.payload && typeof hb.payload === 'object') {
      Object.keys(hb.payload).forEach((k) => {
        if (typeof hb.payload[k] === 'number') {
          telemetryKeys.add(k);
        }
      });
    }
  });

  const keysList = Array.from(telemetryKeys);

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex justify-end">
      <div className="bg-slate-900 border-l border-slate-800 w-full max-w-xl h-full p-6 shadow-2xl overflow-y-auto flex flex-col justify-between text-slate-200">
        <div className="space-y-6">
          <div className="flex items-center justify-between border-b border-slate-800 pb-4">
            <div>
              <h2 className="text-xl font-extrabold text-white">{device.label}</h2>
              <p className="text-xs text-slate-400 font-mono mt-0.5">{device.id}</p>
            </div>
            <button
              onClick={onClose}
              className="text-slate-400 hover:text-white text-2xl font-bold p-1"
            >
              &times;
            </button>
          </div>

          <div className="flex items-center justify-between bg-slate-950/80 p-4 rounded-xl border border-slate-800">
            <span className="text-xs text-slate-400 uppercase font-semibold">Device Status</span>
            <DeviceStatusBadge status={device.status} />
          </div>

          <div className="grid grid-cols-2 gap-4 text-xs">
            <div className="bg-slate-950/60 p-3.5 rounded-xl border border-slate-800">
              <span className="text-slate-500 block mb-1">Device Type</span>
              <span className="font-semibold text-white font-mono">{device.device_type}</span>
            </div>
            <div className="bg-slate-950/60 p-3.5 rounded-xl border border-slate-800">
              <span className="text-slate-500 block mb-1">Gate ID</span>
              <span className="font-semibold text-indigo-400 font-mono">{device.gate_id || 'N/A'}</span>
            </div>
            <div className="bg-slate-950/60 p-3.5 rounded-xl border border-slate-800">
              <span className="text-slate-500 block mb-1">AWS IoT Thing Name</span>
              <span className="font-mono text-slate-300 truncate block">{device.id ? `thing-${device.device_type.toLowerCase()}-${device.id.substring(0, 8)}` : 'N/A'}</span>
            </div>
            <div className="bg-slate-950/60 p-3.5 rounded-xl border border-slate-800">
              <span className="text-slate-500 block mb-1">Firmware Version</span>
              <span className="font-mono text-emerald-400">{device.firmware_version || 'v1.0.0'}</span>
            </div>
          </div>

          {/* Telemetry History Section */}
          <div className="space-y-4 pt-4 border-t border-slate-800">
            <h3 className="font-bold text-white text-sm">24h Telemetry & Heartbeat Log</h3>

            {heartbeats.length === 0 ? (
              <div className="bg-slate-950/40 border border-slate-800/80 rounded-xl p-6 text-center text-slate-500 text-xs">
                No telemetry heartbeats recorded in the last 24 hours.
              </div>
            ) : (
              <div className="space-y-3">
                {/* Generic numeric telemetry metrics breakdown */}
                {keysList.map((key) => {
                  const latestVal = heartbeats[heartbeats.length - 1]?.payload?.[key];
                  return (
                    <div key={key} className="bg-slate-950/80 p-3.5 rounded-xl border border-slate-800 flex items-center justify-between">
                      <span className="text-xs font-semibold text-slate-400 capitalize">{key.replace(/_/g, ' ')}</span>
                      <span className="font-mono font-bold text-indigo-400 text-sm">{latestVal ?? 'N/A'}</span>
                    </div>
                  );
                })}

                <div className="bg-slate-950 p-4 rounded-xl border border-slate-800 max-h-48 overflow-y-auto space-y-2 text-[11px] font-mono">
                  {heartbeats.map((hb, idx) => (
                    <div key={idx} className="flex items-center justify-between text-slate-400 border-b border-slate-900 pb-1">
                      <span>{new Date(hb.ts || Date.now()).toLocaleTimeString()}</span>
                      <span className="text-slate-300">{JSON.stringify(hb.payload)}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="pt-6 border-t border-slate-800 flex justify-end">
          <button
            onClick={onClose}
            className="px-5 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold"
          >
            Close Details
          </button>
        </div>
      </div>
    </div>
  );
};
