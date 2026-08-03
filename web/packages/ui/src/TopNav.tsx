'use client';

import React from 'react';

export interface TopNavProps {
  staffName?: string;
  role?: string;
  storeName?: string;
  stores?: { id: string; name: string }[];
  currentStoreId?: string;
  onStoreChange?: (storeId: string) => void;
  onLogout?: () => void;
}

export const TopNav: React.FC<TopNavProps> = ({
  staffName = 'Store Manager',
  role = 'MANAGER',
  storeName = 'Downtown Superstore',
  stores = [],
  currentStoreId,
  onStoreChange,
  onLogout,
}) => {
  return (
    <header className="h-16 bg-white border-b border-gray-200 px-6 flex items-center justify-between shadow-sm">
      <div className="flex items-center gap-4">
        {stores.length > 1 ? (
          <select
            value={currentStoreId}
            onChange={(e) => onStoreChange?.(e.target.value)}
            className="text-sm font-semibold text-gray-900 border border-gray-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            {stores.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
        ) : (
          <span className="text-base font-bold text-gray-900">{storeName}</span>
        )}
      </div>

      <div className="flex items-center gap-4">
        <div className="text-right">
          <p className="text-sm font-semibold text-gray-900 leading-tight">{staffName}</p>
          <span className="text-xs font-semibold px-2 py-0.5 rounded bg-blue-100 text-blue-800">
            {role}
          </span>
        </div>

        {onLogout && (
          <button
            onClick={onLogout}
            className="px-3 py-1.5 text-xs font-medium text-red-600 hover:text-red-800 border border-red-200 hover:border-red-300 rounded-md transition-colors"
          >
            Sign Out
          </button>
        )}
      </div>
    </header>
  );
};
