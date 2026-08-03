'use client';

import React, { useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Sidebar, TopNav, NavItem } from '@zippyra/ui';

const navItems: NavItem[] = [
  { label: 'Overview', href: '/dashboard' },
  { label: 'Inventory', href: '/dashboard/inventory' },
  { label: 'Orders', href: '/dashboard/orders' },
  { label: 'Catalog', href: '/dashboard/catalog' },
  { label: 'Settings', href: '/dashboard/settings', allowedRoles: ['MANAGER', 'ADMIN'] },
  { label: 'Staff Management', href: '/dashboard/staff', allowedRoles: ['MANAGER', 'ADMIN'] },
  { label: 'Analytics', href: '/dashboard/analytics' },
  { label: 'Device Management', href: '/dashboard/devices' },
  { label: 'Offers & Rules', href: '/dashboard/offers' },
  { label: 'Coupons', href: '/dashboard/coupons' },
  { label: 'Privacy & DPDP', href: '/dashboard/privacy' },
];

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [currentStoreId, setCurrentStoreId] = useState('store-001');

  const stores = [
    { id: 'store-001', name: 'Downtown Superstore' },
    { id: 'store-002', name: 'Metro Plaza Express' },
  ];

  const handleLogout = async () => {
    await fetch('/api/auth/logout', { method: 'POST' });
    router.push('/login');
  };

  return (
    <div className="flex min-h-screen bg-slate-50">
      <Sidebar
        items={navItems}
        currentPath={pathname}
        userRole="MANAGER"
        onNavigate={(href) => router.push(href)}
      />
      <div className="flex-1 flex flex-col min-w-0">
        <TopNav
          staffName="Store Manager"
          role="MANAGER"
          storeName="Downtown Superstore"
          stores={stores}
          currentStoreId={currentStoreId}
          onStoreChange={(id) => setCurrentStoreId(id)}
          onLogout={handleLogout}
        />
        <main className="p-8 flex-1 overflow-y-auto">{children}</main>
      </div>
    </div>
  );
}
