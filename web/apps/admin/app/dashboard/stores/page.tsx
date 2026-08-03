'use client';

import React, { useEffect, useState, useCallback, Suspense } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';

interface Store {
  id: string;
  name: string;
  code: string;
  address: string;
  city: string;
  status: string;
  chain_id?: string;
  chain_name?: string;
  created_at: string;
}

const ADMIN_STORE_SERVICE_URL = process.env.NEXT_PUBLIC_ADMIN_STORE_SERVICE_URL || 'http://localhost:8091';

export default function StoresListPage() {
  return (
    <Suspense fallback={<div className="p-4 text-sm text-slate-400">Loading stores...</div>}>
      <StoresListPageContent />
    </Suspense>
  );
}

function StoresListPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const [stores, setStores] = useState<Store[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [total, setTotal] = useState(0);

  const search = searchParams.get('search') || '';
  const statusFilter = searchParams.get('status') || '';
  const page = parseInt(searchParams.get('page') || '1', 10);
  const pageSize = parseInt(searchParams.get('page_size') || '20', 10);

  const fetchStores = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const params = new URLSearchParams({
        page: page.toString(),
        page_size: pageSize.toString(),
      });
      if (search) params.set('search', search);
      if (statusFilter) params.set('status', statusFilter);

      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/stores?${params}`, {
        headers: { 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' }
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setStores(data.stores || []);
      setTotal(data.total || 0);
    } catch (err: any) {
      setError(err.message || 'Failed to load stores from store-service');
      setStores([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [page, search, statusFilter]);

  useEffect(() => { fetchStores(); }, [fetchStores]);

  const statusBadge = (status: string) => {
    const map: Record<string, string> = {
      ACTIVE: 'badge-active', INACTIVE: 'badge-inactive', SUSPENDED: 'badge-suspended', ONBOARDING: 'badge-pending',
    };
    return <span className={`badge ${map[status] || 'badge-inactive'}`}>{status}</span>;
  };

  const updateQuery = (newSearch: string, newStatus: string, newPage: number) => {
    const p = new URLSearchParams();
    if (newSearch) p.set('search', newSearch);
    if (newStatus) p.set('status', newStatus);
    if (newPage > 1) p.set('page', newPage.toString());
    router.push(`/dashboard/stores?${p.toString()}`);
  };

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>Stores</h1>
          <p>Manage all stores across the Zippyra platform</p>
        </div>
        <Link href="/dashboard/stores/new" className="btn btn-primary">+ Onboard Store</Link>
      </div>

      {/* Filters */}
      <div className="filters-bar">
        <input
          className="input"
          style={{ maxWidth: 320 }}
          placeholder="Search by name, code, or city..."
          value={search}
          onChange={(e) => updateQuery(e.target.value, statusFilter, 1)}
        />
        <select
          className="select"
          style={{ maxWidth: 180 }}
          value={statusFilter}
          onChange={(e) => updateQuery(search, e.target.value, 1)}
        >
          <option value="">All Statuses</option>
          <option value="ACTIVE">Active</option>
          <option value="INACTIVE">Inactive</option>
          <option value="SUSPENDED">Suspended</option>
          <option value="ONBOARDING">Onboarding</option>
        </select>
        <span style={{ color: 'var(--color-text-muted)', fontSize: '0.8125rem', marginLeft: 'auto' }}>
          {total} store{total !== 1 ? 's' : ''}
        </span>
      </div>

      {error && (
        <div className="dev-banner" style={{ marginBottom: '1rem' }}>
          ⚠ store-service unavailable — showing mock data
        </div>
      )}

      {/* Table */}
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="data-table">
          <thead>
            <tr>
              <th>Store Name</th>
              <th>Code</th>
              <th>City</th>
              <th>Chain</th>
              <th>Status</th>
              <th>Created</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '3rem' }}><span className="spinner" style={{ margin: '0 auto' }} /></td></tr>
            ) : stores.length === 0 ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '3rem', color: 'var(--color-text-muted)' }}>No stores found</td></tr>
            ) : stores.map((store) => (
              <tr key={store.id}>
                <td style={{ fontWeight: 600 }}>{store.name}</td>
                <td><code style={{ fontSize: '0.75rem', color: '#a5b4fc' }}>{store.code}</code></td>
                <td>{store.city}</td>
                <td>{store.chain_name || '—'}</td>
                <td>{statusBadge(store.status)}</td>
                <td style={{ color: 'var(--color-text-muted)', fontSize: '0.8125rem' }}>
                  {new Date(store.created_at).toLocaleDateString('en-IN')}
                </td>
                <td>
                  <Link href={`/dashboard/stores/${store.id}`} className="btn btn-secondary btn-sm">View</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {total > pageSize && (
        <div style={{ display: 'flex', justifyContent: 'center', gap: '0.5rem', marginTop: '1.5rem' }}>
          <button className="btn btn-secondary btn-sm" disabled={page <= 1} onClick={() => updateQuery(search, statusFilter, page - 1)}>← Prev</button>
          <span style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
            Page {page} of {Math.ceil(total / pageSize)}
          </span>
          <button className="btn btn-secondary btn-sm" disabled={page >= Math.ceil(total / pageSize)} onClick={() => updateQuery(search, statusFilter, page + 1)}>Next →</button>
        </div>
      )}
    </div>
  );
}

const MOCK_STORES: Store[] = [
  { id: 's-001', name: 'Reliance Fresh — Koramangala', code: 'RF-BLR-001', address: '4th Block, Koramangala', city: 'Bengaluru', status: 'ACTIVE', chain_name: 'Reliance Retail', created_at: '2025-06-15T10:00:00Z' },
  { id: 's-002', name: 'DMart — HSR Layout', code: 'DM-BLR-003', address: 'HSR Layout Sector 2', city: 'Bengaluru', status: 'ACTIVE', chain_name: 'DMart', created_at: '2025-07-01T10:00:00Z' },
  { id: 's-003', name: 'Spencer\'s Retail — MG Road', code: 'SP-BLR-001', address: 'MG Road', city: 'Bengaluru', status: 'ONBOARDING', chain_name: 'Spencer\'s', created_at: '2025-08-10T10:00:00Z' },
  { id: 's-004', name: 'Fresh Basket — Indiranagar', code: 'FB-BLR-001', address: '100ft Road, Indiranagar', city: 'Bengaluru', status: 'ACTIVE', created_at: '2025-05-20T10:00:00Z' },
  { id: 's-005', name: 'Big Bazaar — Whitefield', code: 'BB-BLR-002', address: 'Whitefield Main Road', city: 'Bengaluru', status: 'SUSPENDED', chain_name: 'Future Group', created_at: '2025-04-12T10:00:00Z' },
];
