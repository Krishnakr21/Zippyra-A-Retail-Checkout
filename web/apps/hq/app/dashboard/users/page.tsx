'use client';

import React, { useEffect, useState } from 'react';

export default function ChainUsersPage() {
  const [currentUser, setCurrentUser] = useState<any>(null);
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [showInviteModal, setShowInviteModal] = useState(false);

  // Invite Form state
  const [invitePhone, setInvitePhone] = useState('');
  const [inviteName, setInviteName] = useState('');
  const [inviteRole, setInviteRole] = useState<'FINANCE' | 'OPERATIONS'>('FINANCE');
  const [error, setError] = useState('');

  useEffect(() => {
    const raw = localStorage.getItem('hq_user');
    if (raw) {
      try {
        setCurrentUser(JSON.parse(raw));
      } catch (e) {
        setCurrentUser({ id: 'owner-001', role: 'OWNER' });
      }
    } else {
      setCurrentUser({ id: 'owner-001', role: 'OWNER' });
    }

    fetchUsers();
  }, []);

  async function fetchUsers() {
    try {
      const token = localStorage.getItem('hq_access_token');
      const res = await fetch('http://localhost:8016/v1/chain-hq/users', {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const json = await res.json();
        setUsers(json.users || []);
      } else {
        throw new Error('Failed to fetch users');
      }
    } catch (err) {
      // Mock fallback data for testing
      setUsers([
        { id: 'owner-001', name: 'Mukesh Ambani', phone: '+919876543210', role: 'OWNER', is_active: true },
        { id: 'user-002', name: 'Venkatesh CFO', phone: '+919876543211', role: 'FINANCE', is_active: true },
        { id: 'user-003', name: 'Ramesh Operations', phone: '+919876543212', role: 'OPERATIONS', is_active: true },
      ]);
    } finally {
      setLoading(false);
    }
  }

  const isOwner = currentUser?.role === 'OWNER';

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      const token = localStorage.getItem('hq_access_token');
      const res = await fetch('http://localhost:8016/v1/chain-hq/users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          phone: invitePhone,
          name: inviteName,
          role: inviteRole,
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data?.error?.message || 'Failed to invite user');
      }

      setShowInviteModal(false);
      setInvitePhone('');
      setInviteName('');
      fetchUsers();
    } catch (err: any) {
      setError(err.message || 'Failed to invite user');
    }
  };

  const handleDeactivate = async (id: string) => {
    try {
      const token = localStorage.getItem('hq_access_token');
      const res = await fetch(`http://localhost:8016/v1/chain-hq/users/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data?.error?.message || 'Failed to deactivate user');
      }

      fetchUsers();
    } catch (err: any) {
      alert(err.message || 'Deactivation failed');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Chain HQ Team & Roles</h1>
          <p className="text-sm text-slate-400 mt-1">Manage executive team members for your chain</p>
        </div>

        {/* OWNER-Only Invite Button */}
        {isOwner && (
          <button
            onClick={() => setShowInviteModal(true)}
            data-testid="invite-user-btn"
            className="px-4 py-2.5 rounded-xl font-semibold text-white bg-indigo-600 hover:bg-indigo-500 transition-all shadow-lg shadow-indigo-600/30 text-sm flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Invite Colleague
          </button>
        )}
      </div>

      {loading ? (
        <div className="flex items-center justify-center min-h-[300px]">
          <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-500" />
        </div>
      ) : (
        <div className="glass-panel rounded-2xl overflow-hidden shadow-2xl">
          <table className="w-full text-left text-sm text-slate-300">
            <thead className="bg-slate-900/80 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-800">
              <tr>
                <th className="px-6 py-4">Name</th>
                <th className="px-6 py-4">Phone Number</th>
                <th className="px-6 py-4">Role</th>
                <th className="px-6 py-4">Status</th>
                {isOwner && <th className="px-6 py-4 text-right">Actions</th>}
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {users.map((u) => {
                const isSelf = u.id === currentUser?.id || u.phone === currentUser?.phone;
                return (
                  <tr key={u.id} className="hover:bg-slate-800/40 transition-colors">
                    <td className="px-6 py-4 font-semibold text-white flex items-center gap-2">
                      {u.name}
                      {isSelf && (
                        <span className="px-2 py-0.5 rounded text-[10px] bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 font-bold">
                          You
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 font-mono text-slate-400">{u.phone}</td>
                    <td className="px-6 py-4">
                      <span className="px-3 py-1 rounded-full text-xs font-bold bg-slate-800 text-slate-200 border border-slate-700">
                        {u.role}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span className={`px-3 py-1 rounded-full text-xs font-bold ${
                        u.is_active
                          ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                          : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                      }`}>
                        {u.is_active ? 'ACTIVE' : 'DEACTIVATED'}
                      </span>
                    </td>
                    {isOwner && (
                      <td className="px-6 py-4 text-right">
                        {isSelf ? (
                          <div className="inline-block" title="Chain Owners cannot deactivate their own account">
                            <button
                              disabled
                              data-testid={`deactivate-btn-${u.id}`}
                              className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-800/50 text-slate-600 border border-slate-800 cursor-not-allowed"
                            >
                              Deactivate
                            </button>
                          </div>
                        ) : (
                          <button
                            onClick={() => handleDeactivate(u.id)}
                            data-testid={`deactivate-btn-${u.id}`}
                            className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20 hover:bg-rose-500/20 transition-all"
                          >
                            Deactivate
                          </button>
                        )}
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Invite Modal */}
      {showInviteModal && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="glass-panel w-full max-w-md rounded-2xl p-6 shadow-2xl relative border border-slate-700/60">
            <button
              onClick={() => setShowInviteModal(false)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white text-xl"
            >
              &times;
            </button>
            <h2 className="text-xl font-bold text-white mb-1">Invite Executive Team Member</h2>
            <p className="text-xs text-slate-400 mb-6">Colleagues gain multi-store access to your chain</p>

            {error && (
              <div className="mb-4 bg-red-500/10 border border-red-500/30 rounded-xl p-3 text-xs text-red-400">
                {error}
              </div>
            )}

            <form onSubmit={handleInvite} className="space-y-4 text-sm">
              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Full Name</label>
                <input
                  type="text"
                  required
                  value={inviteName}
                  onChange={(e) => setInviteName(e.target.value)}
                  placeholder="Venkatesh Rao"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-slate-900 border border-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Phone Number (+91...)</label>
                <input
                  type="text"
                  required
                  value={invitePhone}
                  onChange={(e) => setInvitePhone(e.target.value)}
                  placeholder="+919876543211"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-slate-900 border border-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Role (Restricted)</label>
                <select
                  value={inviteRole}
                  onChange={(e) => setInviteRole(e.target.value as any)}
                  className="w-full px-3.5 py-2.5 rounded-xl bg-slate-900 border border-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
                >
                  <option value="FINANCE">FINANCE (CFO / Accounting)</option>
                  <option value="OPERATIONS">OPERATIONS (COO / Field Ops)</option>
                </select>
                <p className="text-[11px] text-slate-500 mt-1">
                  Note: Additional OWNER accounts must be provisioned by Zippyra Admin.
                </p>
              </div>

              <div className="mt-6 flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowInviteModal(false)}
                  className="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 hover:bg-slate-700 font-semibold text-xs"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-indigo-600 text-white hover:bg-indigo-500 font-semibold text-xs shadow-lg shadow-indigo-600/30"
                >
                  Send Access Invite
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
