'use client';

import React, { useEffect, useState } from 'react';

interface PlatformMetrics {
  totalStores: number;
  activeStores: number;
  totalChains: number;
  pendingOnboarding: number;
  totalUsers: number;
  auditEventsToday: number;
}

const MOCK_METRICS: PlatformMetrics = {
  totalStores: 247,
  activeStores: 231,
  totalChains: 14,
  pendingOnboarding: 6,
  totalUsers: 1842,
  auditEventsToday: 389,
};

const recentActivity = [
  { time: '2 min ago', actor: 'Priya S.', action: 'Created store "Reliance Fresh — Koramangala"', type: 'store.created' },
  { time: '12 min ago', actor: 'Rahul M.', action: 'Rotated QR token for store-089', type: 'store.qr_rotated' },
  { time: '28 min ago', actor: 'System', action: 'Chain "DMart" config synced (14 stores)', type: 'chain.synced' },
  { time: '45 min ago', actor: 'Priya S.', action: 'Onboarded chain "Spencer\'s Retail"', type: 'chain.created' },
  { time: '1h ago', actor: 'Amit K.', action: 'CSV import: 342 products for store-156', type: 'catalog.imported' },
  { time: '1h 20m ago', actor: 'System', action: 'HSN verification: 12 products flagged', type: 'catalog.hsn_check' },
];

const onboardingChecklist = [
  { label: 'store-service', ok: true },
  { label: 'catalog-service', ok: true },
  { label: 'retailer-auth-service', ok: true },
  { label: 'audit-service', ok: true },
  { label: 'admin-auth-service', ok: false },
  { label: 'device-mgmt-service', ok: false },
  { label: 'chain-hq-service', ok: false },
  { label: 'compliance-service', ok: false },
];

const ADMIN_STORE_SERVICE_URL = process.env.NEXT_PUBLIC_ADMIN_STORE_SERVICE_URL || 'http://localhost:8097';
const ANALYTICS_SERVICE_URL = process.env.NEXT_PUBLIC_ANALYTICS_SERVICE_URL || 'http://localhost:8092';

export default function DashboardOverview() {
  const [metrics, setMetrics] = useState<PlatformMetrics>({
    totalStores: 0, activeStores: 0, totalChains: 0, pendingOnboarding: 0, totalUsers: 0, auditEventsToday: 0,
  });
  const [animatedMetrics, setAnimatedMetrics] = useState<PlatformMetrics>({
    totalStores: 0, activeStores: 0, totalChains: 0, pendingOnboarding: 0, totalUsers: 0, auditEventsToday: 0,
  });

  useEffect(() => {
    async function loadMetrics() {
      try {
        const [storesRes, chainsRes] = await Promise.all([
          fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/stores`, { headers: { 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' } }).catch(() => null),
          fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/chains`, { headers: { 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' } }).catch(() => null),
        ]);

        let stores: any[] = [];
        let totalChains = 0;
        if (storesRes && storesRes.ok) {
          const data = await storesRes.json();
          stores = data.stores || [];
        }
        if (chainsRes && chainsRes.ok) {
          const data = await chainsRes.json();
          totalChains = (data.chains || []).length;
        }

        const totalStores = stores.length;
        const activeStores = stores.filter(s => s.status === 'ACTIVE').length;
        const pendingOnboarding = stores.filter(s => s.status === 'ONBOARDING').length;

        setMetrics({
          totalStores,
          activeStores,
          totalChains,
          pendingOnboarding,
          totalUsers: 48,
          auditEventsToday: 142,
        });
      } catch {
        // Leave initial zeros if offline
      }
    }
    loadMetrics();
  }, []);

  useEffect(() => {
    const duration = 800;
    const steps = 30;
    const interval = duration / steps;
    let step = 0;
    const timer = setInterval(() => {
      step++;
      const progress = Math.min(step / steps, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      setAnimatedMetrics({
        totalStores: Math.round(metrics.totalStores * eased),
        activeStores: Math.round(metrics.activeStores * eased),
        totalChains: Math.round(metrics.totalChains * eased),
        pendingOnboarding: Math.round(metrics.pendingOnboarding * eased),
        totalUsers: Math.round(metrics.totalUsers * eased),
        auditEventsToday: Math.round(metrics.auditEventsToday * eased),
      });
      if (step >= steps) clearInterval(timer);
    }, interval);
    return () => clearInterval(timer);
  }, [metrics]);

  const metricCards = [
    { label: 'Total Stores', value: animatedMetrics.totalStores, icon: '🏪', accent: '#6366f1' },
    { label: 'Active Stores', value: animatedMetrics.activeStores, icon: '✅', accent: '#22c55e' },
    { label: 'Chains', value: animatedMetrics.totalChains, icon: '🔗', accent: '#a855f7' },
    { label: 'Pending Onboarding', value: animatedMetrics.pendingOnboarding, icon: '⏳', accent: '#f59e0b' },
    { label: 'Registered Users', value: animatedMetrics.totalUsers.toLocaleString(), icon: '👤', accent: '#06b6d4' },
    { label: 'Audit Events Today', value: animatedMetrics.auditEventsToday, icon: '📋', accent: '#ec4899' },
  ];

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>Platform Overview</h1>
          <p>Real-time operational health across all Zippyra services</p>
        </div>
      </div>

      {/* Metrics Grid */}
      <div className="metric-grid">
        {metricCards.map((m) => (
          <div key={m.label} className="card metric-card" style={{ '--accent': m.accent } as React.CSSProperties}>
            <div className="metric-value">{m.value}</div>
            <div className="metric-label">{m.label}</div>
            <div className="metric-icon" style={{ background: `${m.accent}15`, fontSize: '1.25rem' }}>
              {m.icon}
            </div>
          </div>
        ))}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 340px', gap: '1.5rem' }}>
        {/* Recent Activity */}
        <div className="card" style={{ padding: 0 }}>
          <div style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--color-border)' }}>
            <h2 style={{ fontSize: '1rem', fontWeight: 700 }}>Recent Activity</h2>
          </div>
          <div>
            {recentActivity.map((item, i) => (
              <div key={i} style={{
                display: 'flex', alignItems: 'flex-start', gap: '0.75rem',
                padding: '0.875rem 1.5rem', borderBottom: i < recentActivity.length - 1 ? '1px solid rgba(51,65,85,0.5)' : 'none',
              }}>
                <span style={{
                  fontSize: '0.6875rem', color: 'var(--color-text-muted)', minWidth: 72, paddingTop: 2,
                }}>{item.time}</span>
                <div>
                  <span style={{ fontWeight: 600, fontSize: '0.875rem' }}>{item.actor}</span>
                  <span style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}> — {item.action}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Service Health Checklist */}
        <div className="card" style={{ padding: 0 }}>
          <div style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--color-border)' }}>
            <h2 style={{ fontSize: '1rem', fontWeight: 700 }}>Service Health</h2>
          </div>
          {onboardingChecklist.map((svc) => (
            <div key={svc.label} className="checklist-item">
              <span className={`checklist-dot ${svc.ok ? 'green' : 'red'}`} />
              <span style={{ fontSize: '0.8125rem', flex: 1 }}>{svc.label}</span>
              <span className={`badge ${svc.ok ? 'badge-active' : 'badge-inactive'}`}>
                {svc.ok ? 'Live' : 'Not Built'}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
