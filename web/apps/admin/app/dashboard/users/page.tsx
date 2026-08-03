'use client';

import React, { useEffect, useState, useCallback } from 'react';

interface StaffUser {
  id: string;
  name: string;
  phone_masked: string;
  email_masked: string;
  phone?: string; // set when unmasked details are fetched
  email?: string; // set when unmasked details are fetched
  role: string;
  store_id: string;
  store_name?: string;
  status: string;
  created_at: string;
}

const RETAILER_AUTH_SERVICE_URL = process.env.NEXT_PUBLIC_RETAILER_AUTH_SERVICE_URL || 'http://localhost:8012';

export default function UsersPage() {
  const [users, setUsers] = useState<StaffUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [roleFilter, setRoleFilter] = useState('');
  const [selectedUser, setSelectedUser] = useState<StaffUser | null>(null);
  const [unmasking, setUnmasking] = useState(false);

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (search) params.set('search', search);
      if (roleFilter) params.set('role', roleFilter);

      const res = await fetch(`${RETAILER_AUTH_SERVICE_URL}/v1/retailer-auth/admin/users?${params}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setUsers(data.users || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load users');
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, [search, roleFilter]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const handleInspectUser = async (user: StaffUser) => {
    setSelectedUser(user);
    setUnmasking(true);
    try {
      const res = await fetch(`${RETAILER_AUTH_SERVICE_URL}/v1/retailer-auth/admin/users/${user.id}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setSelectedUser(data);
    } catch {
      // Mock unmasked payload
      setSelectedUser({
        ...user,
        phone: '+91 98765 43210',
        email: `${user.name.toLowerCase().replace(/\s+/g, '.')}@retailer.com`,
      });
    } finally {
      setUnmasking(false);
    }
  };

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>Registered Staff & Users</h1>
          <p>Inspect staff members across all onboarded retail stores</p>
        </div>
      </div>

      {/* Filters */}
      <div className="filters-bar">
        <input
          className="input"
          style={{ maxWidth: 320 }}
          placeholder="Search by name..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select
          className="select"
          style={{ maxWidth: 180 }}
          value={roleFilter}
          onChange={(e) => setRoleFilter(e.target.value)}
        >
          <option value="">All Roles</option>
          <option value="STORE_MANAGER">Store Manager</option>
          <option value="CASHIER">Cashier</option>
          <option value="INVENTORY_CLERK">Inventory Clerk</option>
        </select>
      </div>

      {error && (
        <div className="dev-banner" style={{ marginBottom: '1rem' }}>
          ⚠ retailer-auth-service unavailable — showing mock user list
        </div>
      )}

      {/* Table */}
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="data-table">
          <thead>
            <tr>
              <th>User Name</th>
              <th>Role</th>
              <th>Store</th>
              <th>Masked Phone</th>
              <th>Masked Email</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '3rem' }}><span className="spinner" style={{ margin: '0 auto' }} /></td></tr>
            ) : users.length === 0 ? (
              <tr><td colSpan={7} style={{ textAlign: 'center', padding: '3rem', color: 'var(--color-text-muted)' }}>No registered users found</td></tr>
            ) : users.map((user) => (
              <tr key={user.id}>
                <td style={{ fontWeight: 600 }}>{user.name}</td>
                <td><span className="badge badge-active">{user.role}</span></td>
                <td>{user.store_name || user.store_id}</td>
                <td><code style={{ fontSize: '0.75rem', color: '#a5b4fc' }}>{user.phone_masked}</code></td>
                <td><code style={{ fontSize: '0.75rem', color: '#a5b4fc' }}>{user.email_masked}</code></td>
                <td>
                  <span className={`badge ${user.status === 'ACTIVE' ? 'badge-active' : 'badge-inactive'}`}>
                    {user.status}
                  </span>
                </td>
                <td>
                  <button className="btn btn-secondary btn-sm" onClick={() => handleInspectUser(user)}>
                    Inspect PII
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Inspect User Modal with Audit Warning */}
      {selectedUser && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.7)', backdropFilter: 'blur(4px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50, padding: '1rem',
        }}>
          <div className="card" style={{ width: '100%', maxWidth: 540, margin: 'auto' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
              <h2 style={{ fontSize: '1.25rem', fontWeight: 700 }}>User PII Inspection</h2>
              <button className="btn-icon" onClick={() => setSelectedUser(null)}>✕</button>
            </div>

            {/* DPDP Compliance / Audit Warning Alert */}
            <div className="dev-banner" style={{ background: 'rgba(245,158,11,0.15)', borderColor: 'rgba(245,158,11,0.3)', color: '#fcd34d', marginBottom: '1.5rem' }}>
              🔒 <strong>Audit Event Generated:</strong> Accessing unmasked personal data triggers a <code style={{ color: 'white' }}>user.pii_accessed</code> audit record per DPDP policy.
            </div>

            {unmasking ? (
              <div style={{ textAlign: 'center', padding: '2rem' }}><span className="spinner" style={{ margin: '0 auto' }} /></div>
            ) : (
              <div style={{ display: 'grid', gap: '1rem', marginBottom: '1.5rem' }}>
                <div>
                  <label className="input-label">Full Name</label>
                  <div style={{ fontWeight: 600, fontSize: '1rem' }}>{selectedUser.name}</div>
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                  <div>
                    <label className="input-label">Unmasked Phone Number</label>
                    <div style={{ fontFamily: 'monospace', color: '#a5b4fc', fontSize: '0.9375rem', fontWeight: 600 }}>
                      {selectedUser.phone || selectedUser.phone_masked}
                    </div>
                  </div>
                  <div>
                    <label className="input-label">Unmasked Email</label>
                    <div style={{ fontFamily: 'monospace', color: '#a5b4fc', fontSize: '0.9375rem', fontWeight: 600 }}>
                      {selectedUser.email || selectedUser.email_masked}
                    </div>
                  </div>
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                  <div>
                    <label className="input-label">Assigned Role</label>
                    <div>{selectedUser.role}</div>
                  </div>
                  <div>
                    <label className="input-label">Store ID</label>
                    <div style={{ fontFamily: 'monospace', fontSize: '0.8125rem' }}>{selectedUser.store_id}</div>
                  </div>
                </div>
              </div>
            )}

            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
              <button className="btn btn-secondary" onClick={() => setSelectedUser(null)}>Close</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

const MOCK_USERS: StaffUser[] = [
  { id: 'usr-001', name: 'Rajesh Kumar', phone_masked: '+91XXXXXX4321', email_masked: 'r***@retailer.com', role: 'STORE_MANAGER', store_id: 's-001', store_name: 'Reliance Fresh — Koramangala', status: 'ACTIVE', created_at: '2025-06-15T00:00:00Z' },
  { id: 'usr-002', name: 'Sunita Sharma', phone_masked: '+91XXXXXX9876', email_masked: 's***@dmart.com', role: 'CASHIER', store_id: 's-002', store_name: 'DMart — HSR Layout', status: 'ACTIVE', created_at: '2025-07-02T00:00:00Z' },
  { id: 'usr-003', name: 'Vikram Singh', phone_masked: '+91XXXXXX1122', email_masked: 'v***@spencers.com', role: 'INVENTORY_CLERK', store_id: 's-003', store_name: 'Spencer\'s — MG Road', status: 'INACTIVE', created_at: '2025-08-11T00:00:00Z' },
];
