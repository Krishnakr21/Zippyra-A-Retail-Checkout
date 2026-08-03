'use client';

import React, { useEffect, useState, useCallback } from 'react';

interface AuditAction {
  id: string;
  actor_id: string;
  actor_name?: string;
  action_type: string;
  target_type: string;
  target_id: string;
  payload: any;
  source_service: string;
  request_id: string;
  created_at: string;
}

const AUDIT_SERVICE_URL = process.env.NEXT_PUBLIC_AUDIT_SERVICE_URL || 'http://localhost:8015';

export default function AuditLogPage() {
  const [actions, setActions] = useState<AuditAction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [expandedRow, setExpandedRow] = useState<string | null>(null);

  // Filters
  const [serviceFilter, setServiceFilter] = useState('');
  const [actionTypeFilter, setActionTypeFilter] = useState('');
  const [actorIdFilter, setActorIdFilter] = useState('');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;

  const fetchAuditLog = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: page.toString(),
        page_size: pageSize.toString(),
      });
      if (serviceFilter) params.set('service', serviceFilter);
      if (actionTypeFilter) params.set('action_type', actionTypeFilter);
      if (actorIdFilter) params.set('actor_id', actorIdFilter);

      const res = await fetch(`${AUDIT_SERVICE_URL}/v1/audit/actions?${params}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setActions(data.actions || []);
      setTotal(data.total || 0);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch audit trail');
      setActions([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [page, serviceFilter, actionTypeFilter, actorIdFilter]);

  useEffect(() => {
    fetchAuditLog();
  }, [fetchAuditLog]);

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>Audit Trail</h1>
          <p>Centralized immutable log of administrative and system events across Zippyra</p>
        </div>
      </div>

      {/* Filters Bar */}
      <div className="filters-bar">
        <select
          className="select"
          style={{ maxWidth: 180 }}
          value={serviceFilter}
          onChange={(e) => { setServiceFilter(e.target.value); setPage(1); }}
        >
          <option value="">All Services</option>
          <option value="store-service">store-service</option>
          <option value="catalog-service">catalog-service</option>
          <option value="retailer-auth-service">retailer-auth-service</option>
          <option value="audit-service">audit-service</option>
        </select>

        <input
          className="input"
          style={{ maxWidth: 200 }}
          placeholder="Action type (e.g. store.created)"
          value={actionTypeFilter}
          onChange={(e) => { setActionTypeFilter(e.target.value); setPage(1); }}
        />

        <input
          className="input"
          style={{ maxWidth: 180 }}
          placeholder="Actor ID"
          value={actorIdFilter}
          onChange={(e) => { setActorIdFilter(e.target.value); setPage(1); }}
        />

        <span style={{ color: 'var(--color-text-muted)', fontSize: '0.8125rem', marginLeft: 'auto' }}>
          Total Events: {total}
        </span>
      </div>

      {error && (
        <div className="dev-banner" style={{ marginBottom: '1rem' }}>
          ⚠ audit-service unavailable — displaying local audit records
        </div>
      )}

      {/* Table */}
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="data-table">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Service</th>
              <th>Action Type</th>
              <th>Actor</th>
              <th>Target</th>
              <th>Request ID</th>
              <th>Payload</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '3rem' }}><span className="spinner" style={{ margin: '0 auto' }} /></td></tr>
            ) : actions.length === 0 ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '3rem', color: 'var(--color-text-muted)' }}>No audit events found matching filters</td></tr>
            ) : actions.map((act) => (
              <React.Fragment key={act.id}>
                <tr>
                  <td style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)', whiteSpace: 'nowrap' }}>
                    {new Date(act.created_at).toLocaleString('en-IN')}
                  </td>
                  <td>
                    <span className="badge badge-active" style={{ fontSize: '0.6875rem' }}>
                      {act.source_service}
                    </span>
                  </td>
                  <td style={{ fontWeight: 600, color: '#a5b4fc' }}>{act.action_type}</td>
                  <td>{act.actor_name || act.actor_id}</td>
                  <td style={{ fontSize: '0.8125rem' }}>
                    <span style={{ color: 'var(--color-text-muted)' }}>{act.target_type}:</span> {act.target_id}
                  </td>
                  <td><code style={{ fontSize: '0.6875rem' }}>{act.request_id.substring(0, 8)}...</code></td>
                  <td>
                    <button
                      className="btn btn-secondary btn-sm"
                      onClick={() => setExpandedRow(expandedRow === act.id ? null : act.id)}
                    >
                      {expandedRow === act.id ? 'Hide' : 'View Payload'}
                    </button>
                  </td>
                </tr>
                {expandedRow === act.id && (
                  <tr>
                    <td colSpan={7} style={{ background: 'rgba(15,23,42,0.8)', padding: '1rem 1.5rem' }}>
                      <div className="json-viewer">
                        {JSON.stringify(act.payload, null, 2)}
                      </div>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {total > pageSize && (
        <div style={{ display: 'flex', justifyContent: 'center', gap: '0.5rem', marginTop: '1.5rem' }}>
          <button className="btn btn-secondary btn-sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>← Prev</button>
          <span style={{ padding: '0.5rem 1rem', fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
            Page {page} of {Math.ceil(total / pageSize)}
          </span>
          <button className="btn btn-secondary btn-sm" disabled={page >= Math.ceil(total / pageSize)} onClick={() => setPage(p => p + 1)}>Next →</button>
        </div>
      )}
    </div>
  );
}

const MOCK_AUDIT_ACTIONS: AuditAction[] = [
  {
    id: 'aud-001',
    actor_id: 'admin-001',
    actor_name: 'Jane Doe',
    action_type: 'store.created',
    target_type: 'store',
    target_id: 's-001',
    payload: { name: 'Reliance Fresh — Koramangala', city: 'Bengaluru', code: 'RF-BLR-001' },
    source_service: 'store-service',
    request_id: 'req-9812739812739812',
    created_at: '2026-07-31T19:40:00Z',
  },
  {
    id: 'aud-002',
    actor_id: 'admin-001',
    actor_name: 'Jane Doe',
    action_type: 'user.pii_accessed',
    target_type: 'staff_user',
    target_id: 'usr-001',
    payload: { inspected_fields: ['phone', 'email'], reason: 'Support investigation' },
    source_service: 'retailer-auth-service',
    request_id: 'req-1238912389123891',
    created_at: '2026-07-31T19:42:15Z',
  },
  {
    id: 'aud-003',
    actor_id: 'admin-002',
    actor_name: 'Rahul M.',
    action_type: 'store.qr_rotated',
    target_type: 'store',
    target_id: 's-002',
    payload: { previous_token_invalidated: true },
    source_service: 'store-service',
    request_id: 'req-4412390123901293',
    created_at: '2026-07-31T19:50:00Z',
  },
];
