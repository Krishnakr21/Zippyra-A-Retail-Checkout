'use client';

import React from 'react';

export interface NavItem {
  label: string;
  href: string;
  icon?: React.ReactNode;
  allowedRoles?: string[];
}

export interface SidebarProps {
  items: NavItem[];
  currentPath: string;
  userRole?: string;
  onNavigate: (href: string) => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ items, currentPath, userRole = 'MANAGER', onNavigate }) => {
  const visibleItems = items.filter((item) => {
    if (!item.allowedRoles) return true;
    return item.allowedRoles.includes(userRole);
  });

  return (
    <aside className="w-64 bg-slate-900 text-slate-300 min-h-screen p-4 flex flex-col justify-between">
      <div>
        <div className="flex items-center gap-3 px-3 py-4 mb-6 border-b border-slate-800">
          <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white font-black">
            Z
          </div>
          <div>
            <h1 className="font-bold text-white text-base leading-tight">Zippyra Retailer</h1>
            <span className="text-xs text-slate-400">Store Operator Portal</span>
          </div>
        </div>

        <nav className="space-y-1">
          {visibleItems.map((item) => {
            const isActive = currentPath === item.href || (item.href !== '/dashboard' && currentPath.startsWith(item.href));
            return (
              <button
                key={item.href}
                onClick={() => onNavigate(item.href)}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-blue-600 text-white'
                    : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
                }`}
              >
                {item.icon}
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>
      </div>

      <div className="p-3 border-t border-slate-800 text-xs text-slate-500">
        Zippyra Retailer Platform v1.0
      </div>
    </aside>
  );
};
