'use client';

import React from 'react';
import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';

const navSections = [
  {
    label: 'Platform',
    items: [
      { label: 'Dashboard', href: '/dashboard', icon: '📊' },
      { label: 'Chains', href: '/dashboard/chains', icon: '🔗' },
      { label: 'Stores', href: '/dashboard/stores', icon: '🏪' },
      { label: 'Products', href: '/dashboard/products', icon: '📦' },
    ],
  },
  {
    label: 'Operations',
    items: [
      { label: 'Users', href: '/dashboard/users', icon: '👤' },
      { label: 'Audit Log', href: '/dashboard/audit', icon: '📋' },
      { label: 'Privacy & DPDP', href: '/dashboard/privacy', icon: '🔒' },
      { label: 'System', href: '/dashboard/system', icon: '⚙️' },
    ],
  },
];

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();

  const handleLogout = async () => {
    document.cookie = 'admin_session=; path=/; max-age=0';
    router.push('/login');
  };

  return (
    <div className="app-shell">
      {/* Sidebar */}
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="sidebar-brand-icon">Z</div>
          <div>
            <h1>Zippyra</h1>
            <span>Admin Console</span>
          </div>
        </div>

        <nav className="sidebar-nav">
          {navSections.map((section) => (
            <React.Fragment key={section.label}>
              <div className="sidebar-section-label">{section.label}</div>
              {section.items.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  className={pathname === item.href || (item.href !== '/dashboard' && pathname?.startsWith(item.href)) ? 'active' : ''}
                >
                  <span className="icon">{item.icon}</span>
                  {item.label}
                </Link>
              ))}
            </React.Fragment>
          ))}
        </nav>

        <div className="sidebar-footer">
          <button className="btn btn-secondary btn-sm" onClick={handleLogout} style={{ width: '100%', justifyContent: 'center' }}>
            Sign Out
          </button>
        </div>
      </aside>

      {/* Main content area */}
      <div className="main-content">
        <header className="topnav">
          <div className="topnav-left">
            <div className="dev-banner" style={{ margin: 0, padding: '0.375rem 0.75rem', fontSize: '0.75rem' }}>
              ⚠ Mock Auth Active
            </div>
          </div>
          <div className="topnav-right">
            <span style={{ fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>Dev Admin</span>
            <div className="topnav-avatar">DA</div>
          </div>
        </header>

        <div className="page-content">
          {children}
        </div>
      </div>
    </div>
  );
}
