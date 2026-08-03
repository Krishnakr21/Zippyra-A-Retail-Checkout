import React from 'react';

export interface DeviceStatusBadgeProps {
  status: string;
}

export const DeviceStatusBadge: React.FC<DeviceStatusBadgeProps> = ({ status }) => {
  const normalized = (status || 'PROVISIONING').toUpperCase();

  let styleClass = 'bg-slate-100 text-slate-700 border-slate-200';
  if (normalized === 'ACTIVE') {
    styleClass = 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20';
  } else if (normalized === 'OFFLINE') {
    styleClass = 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20';
  } else if (normalized === 'PROVISIONING') {
    styleClass = 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-500/10 dark:text-amber-400 dark:border-amber-500/20';
  } else if (normalized === 'DECOMMISSIONED') {
    styleClass = 'bg-slate-100 text-slate-500 border-slate-200 dark:bg-slate-800 dark:text-slate-500 dark:border-slate-700';
  }

  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${styleClass}`}>
      <span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${
        normalized === 'ACTIVE' ? 'bg-emerald-500 animate-pulse' :
        normalized === 'OFFLINE' ? 'bg-rose-500' :
        normalized === 'PROVISIONING' ? 'bg-amber-500 animate-ping' : 'bg-slate-400'
      }`} />
      {normalized}
    </span>
  );
};
