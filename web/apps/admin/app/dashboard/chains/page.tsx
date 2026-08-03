'use client';

import React, { useEffect, useState, useCallback } from 'react';

interface Chain {
  id: string;
  name: string;
  legal_entity_name?: string;
  default_gstin_prefix?: string;
  status: string;
  store_count: number;
  created_at: string;
}

const ADMIN_STORE_SERVICE_URL = process.env.NEXT_PUBLIC_ADMIN_STORE_SERVICE_URL || 'http://localhost:8091';

export default function ChainsPage() {
  const [chains, setChains] = useState<Chain[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showModal, setShowModal] = useState(false);

  // New chain form state
  const [name, setName] = useState('');
  const [legalEntityName, setLegalEntityName] = useState('');
  const [gstinPrefix, setGstinPrefix] = useState('');
  const [creating, setCreating] = useState(false);

  const fetchChains = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/chains`, {
        headers: { 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' }
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setChains(data.chains || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load chains');
      setChains([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchChains();
  }, [fetchChains]);

  const handleCreateChain = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) return;
    setCreating(true);
    try {
      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/chains`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' },
        body: JSON.stringify({
          name,
          legal_entity_name: legalEntityName,
          default_gstin_prefix: gstinPrefix,
        }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      setShowModal(false);
      setName('');
      setLegalEntityName('');
      setGstinPrefix('');
      fetchChains();
    } catch (err: any) {
      alert(err.message || 'Failed to create chain');
    } finally {
      setCreating(false);
    }
  };

  const handleToggleStatus = async (chain: Chain) => {
    const nextStatus = chain.status === 'ACTIVE' ? 'SUSPENDED' : 'ACTIVE';
    try {
      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/chains/${chain.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' },
        body: JSON.stringify({ status: nextStatus }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      fetchChains();
    } catch (err: any) {
      alert(err.message || 'Failed to update chain status');
    }
  };

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>Retail Chains</h1>
          <p>Manage multi-store retail organizations and legal entities</p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          + Create Chain
        </button>
      </div>

      {error && (
        <div className="dev-banner" style={{ marginBottom: '1rem' }}>
          ⚠ store-service unavailable — showing offline mock chains
        </div>
      )}

      {/* Table */}
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="data-table">
          <thead>
            <tr>
              <th>Chain Name</th>
              <th>Legal Entity Name</th>
              <th>GSTIN Prefix</th>
              <th>Stores Count</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={6} style={{ textAlign: 'center', padding: '3rem' }}><span className="spinner" style={{ margin: '0 auto' }} /></td></tr>
            ) : chains.length === 0 ? (
              <tr><td colSpan={6} style={{ textAlign: 'center', padding: '3rem', color: 'var(--color-text-muted)' }}>No chains registered</td></tr>
            ) : chains.map((chain) => (
              <tr key={chain.id}>
                <td style={{ fontWeight: 600 }}>{chain.name}</td>
                <td>{chain.legal_entity_name || '—'}</td>
                <td><code style={{ fontSize: '0.75rem', color: '#a5b4fc' }}>{chain.default_gstin_prefix || '—'}</code></td>
                <td style={{ fontWeight: 700 }}>{chain.store_count} stores</td>
                <td>
                  <span className={`badge ${chain.status === 'ACTIVE' ? 'badge-active' : 'badge-suspended'}`}>
                    {chain.status}
                  </span>
                </td>
                <td>
                  <button
                    className={`btn btn-sm ${chain.status === 'ACTIVE' ? 'btn-danger' : 'btn-secondary'}`}
                    onClick={() => handleToggleStatus(chain)}
                  >
                    {chain.status === 'ACTIVE' ? 'Suspend' : 'Activate'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Create Modal */}
      {showModal && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.7)', backdropFilter: 'blur(4px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50, padding: '1rem',
        }}>
          <div className="card" style={{ width: '100%', maxWidth: 500, margin: 'auto' }}>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 700, marginBottom: '1.5rem' }}>Create Retail Chain</h2>
            <form onSubmit={handleCreateChain}>
              <div style={{ display: 'grid', gap: '1rem', marginBottom: '1.5rem' }}>
                <div>
                  <label className="input-label">Chain Display Name *</label>
                  <input className="input" value={name} onChange={e => setName(e.target.value)} placeholder="e.g. Reliance Retail" required />
                </div>
                <div>
                  <label className="input-label">Legal Entity Name</label>
                  <input className="input" value={legalEntityName} onChange={e => setLegalEntityName(e.target.value)} placeholder="Reliance Retail Ltd." />
                </div>
                <div>
                  <label className="input-label">Default GSTIN Prefix (2-digit state code)</label>
                  <input className="input" value={gstinPrefix} onChange={e => setGstinPrefix(e.target.value)} maxLength={2} placeholder="29" />
                </div>
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Cancel</button>
                <button type="submit" className="btn btn-primary" disabled={creating}>
                  {creating ? <span className="spinner" /> : 'Create Chain'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

const MOCK_CHAINS: Chain[] = [
  { id: 'chain-001', name: 'Reliance Retail', legal_entity_name: 'Reliance Retail Limited', default_gstin_prefix: '29', status: 'ACTIVE', store_count: 8, created_at: '2025-01-10T00:00:00Z' },
  { id: 'chain-002', name: 'DMart', legal_entity_name: 'Avenue Supermarts Ltd.', default_gstin_prefix: '27', status: 'ACTIVE', store_count: 5, created_at: '2025-02-15T00:00:00Z' },
  { id: 'chain-003', name: 'Spencer\'s Retail', legal_entity_name: 'Spencer\'s Retail Limited', default_gstin_prefix: '19', status: 'SUSPENDED', store_count: 1, created_at: '2025-03-01T00:00:00Z' },
];
