'use client';

import React, { useState, useEffect } from 'react';
import { useStaff, useShiftHistory, StaffMember, ShiftRecord } from '@zippyra/hooks';
import { Badge, ConfirmDialog } from '@zippyra/ui';

const PHONE_REGEX = /^\+91[6-9]\d{9}$/;

export default function StaffPage() {
  const { getStaffList, createStaff, updateStaff, deactivateStaff } = useStaff();
  const { getShiftHistory } = useShiftHistory();

  // State
  const [activeTab, setActiveTab] = useState<'roster' | 'shifts'>('roster');
  const [staffList, setStaffList] = useState<StaffMember[]>([]);
  const [shiftList, setShiftList] = useState<ShiftRecord[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Filters
  const [roleFilter, setRoleFilter] = useState<string>('ALL');
  const [activeOnly, setActiveOnly] = useState<boolean>(true);
  const [dateFrom, setDateFrom] = useState<string>('');
  const [dateTo, setDateTo] = useState<string>('');

  // Modals
  const [isAddOpen, setIsAddOpen] = useState<boolean>(false);
  const [editMember, setEditMember] = useState<StaffMember | null>(null);
  const [deactivateMember, setDeactivateMember] = useState<StaffMember | null>(null);

  // Add Form State
  const [addName, setAddName] = useState<string>('');
  const [addPhone, setAddPhone] = useState<string>('');
  const [addRole, setAddRole] = useState<'CASHIER' | 'STOCK_ASSOCIATE' | 'SECURITY' | 'MANAGER'>('CASHIER');
  const [addInlineError, setAddInlineError] = useState<string | null>(null);
  const [addSubmitting, setAddSubmitting] = useState<boolean>(false);

  // Edit Form State
  const [editName, setEditName] = useState<string>('');
  const [editRole, setEditRole] = useState<'CASHIER' | 'STOCK_ASSOCIATE' | 'SECURITY' | 'MANAGER'>('CASHIER');
  const [editSubmitting, setEditSubmitting] = useState<boolean>(false);

  const storeId = 'store-mumbai-01'; // Default logged-in store context

  const fetchStaff = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await getStaffList(storeId, roleFilter, activeOnly);
      setStaffList(res.staff || []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load staff roster');
    } finally {
      setLoading(false);
    }
  };

  const fetchShifts = async () => {
    setLoading(true);
    try {
      const res = await getShiftHistory(storeId, dateFrom, dateTo);
      setShiftList(res.shifts || []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load shift history');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'roster') {
      fetchStaff();
    } else {
      fetchShifts();
    }
  }, [activeTab, roleFilter, activeOnly, dateFrom, dateTo]);

  // Handlers
  const handleAddSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setAddInlineError(null);

    const trimmedPhone = addPhone.trim();
    if (!PHONE_REGEX.test(trimmedPhone)) {
      setAddInlineError('Please enter a valid Indian mobile number (e.g. +919876543210)');
      return;
    }

    setAddSubmitting(true);
    try {
      await createStaff({
        name: addName.trim(),
        phone: trimmedPhone,
        role: addRole,
      });
      setIsAddOpen(false);
      setAddName('');
      setAddPhone('');
      setAddRole('CASHIER');
      fetchStaff();
    } catch (err: any) {
      if (err?.code === 'PHONE_ALREADY_STAFF' || err?.status === 409 || err?.message?.includes('already registered')) {
        setAddInlineError('This number is already registered as staff somewhere on the platform');
      } else {
        setAddInlineError(err?.message || 'Failed to add staff member');
      }
    } finally {
      setAddSubmitting(false);
    }
  };

  const handleEditSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editMember) return;
    setEditSubmitting(true);
    try {
      await updateStaff(editMember.id, {
        name: editName.trim(),
        role: editRole,
      });
      setEditMember(null);
      fetchStaff();
    } catch (err: any) {
      alert(err?.message || 'Failed to update staff member');
    } finally {
      setEditSubmitting(false);
    }
  };

  const handleConfirmDeactivate = async () => {
    if (!deactivateMember) return;
    try {
      await deactivateStaff(deactivateMember.id);
      setDeactivateMember(null);
      fetchStaff();
    } catch (err: any) {
      alert(err?.message || 'Failed to deactivate staff member');
    }
  };

  const maskPhone = (phone: string) => {
    if (!phone || phone.length < 4) return phone;
    const last4 = phone.slice(-4);
    return `+91XXXXXX${last4}`;
  };

  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case 'MANAGER':
        return 'purple';
      case 'SECURITY':
        return 'red';
      case 'STOCK_ASSOCIATE':
        return 'green';
      case 'CASHIER':
      default:
        return 'blue';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Staff Management</h1>
          <p className="text-sm text-gray-500 mt-1">Manage store staff roster, shift logs, and security PIN setup</p>
        </div>
        {activeTab === 'roster' && (
          <button
            onClick={() => setIsAddOpen(true)}
            className="inline-flex items-center justify-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-lg text-sm transition-colors shadow-sm"
          >
            + Add Staff
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('roster')}
            className={`py-3 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'roster'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Staff Roster
          </button>
          <button
            onClick={() => setActiveTab('shifts')}
            className={`py-3 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'shifts'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Shift History
          </button>
        </nav>
      </div>

      {/* Roster Tab Content */}
      {activeTab === 'roster' && (
        <div className="space-y-4">
          {/* Filters */}
          <div className="flex flex-wrap items-center justify-between gap-4 bg-white p-4 rounded-xl border border-gray-200 shadow-sm">
            <div className="flex items-center space-x-4">
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Filter by Role</label>
                <select
                  value={roleFilter}
                  onChange={(e) => setRoleFilter(e.target.value)}
                  className="px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                >
                  <option value="ALL">All Roles</option>
                  <option value="CASHIER">Cashier</option>
                  <option value="STOCK_ASSOCIATE">Stock Associate</option>
                  <option value="SECURITY">Security</option>
                  <option value="MANAGER">Manager</option>
                </select>
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="activeOnly"
                checked={activeOnly}
                onChange={(e) => setActiveOnly(e.target.checked)}
                className="h-4 w-4 text-indigo-600 rounded border-gray-300 focus:ring-indigo-500"
              />
              <label htmlFor="activeOnly" className="text-sm font-medium text-gray-700">
                Show Active Only
              </label>
            </div>
          </div>

          {/* DataTable */}
          <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
            {loading ? (
              <div className="p-8 text-center text-gray-500 text-sm">Loading staff roster...</div>
            ) : staffList.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">No staff members found matching criteria.</div>
            ) : (
              <table className="w-full text-left text-sm text-gray-600">
                <thead className="bg-gray-50 text-gray-500 font-semibold border-b border-gray-200 text-xs uppercase tracking-wider">
                  <tr>
                    <th className="px-6 py-3">Name</th>
                    <th className="px-6 py-3">Phone</th>
                    <th className="px-6 py-3">Role</th>
                    <th className="px-6 py-3">Status</th>
                    <th className="px-6 py-3">PIN Setup</th>
                    <th className="px-6 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {staffList.map((member) => (
                    <tr key={member.id} className="hover:bg-gray-50/50 transition-colors">
                      <td className="px-6 py-4 font-medium text-gray-900">{member.name}</td>
                      <td className="px-6 py-4 font-mono text-gray-600">{maskPhone(member.phone)}</td>
                      <td className="px-6 py-4">
                        <Badge status={member.role.replace('_', ' ')} />
                      </td>
                      <td className="px-6 py-4">
                        {member.is_active ? (
                          <Badge status="ACTIVE" />
                        ) : (
                          <Badge status="INACTIVE" />
                        )}
                      </td>
                      <td className="px-6 py-4">
                        {member.has_pin_set ? (
                          <span className="inline-flex items-center text-xs text-green-700 font-medium bg-green-50 px-2 py-0.5 rounded border border-green-200">
                            🔑 PIN Set
                          </span>
                        ) : (
                          <span className="inline-flex items-center text-xs text-gray-400 font-medium bg-gray-50 px-2 py-0.5 rounded border border-gray-200">
                            🔒 OTP Only
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4 text-right space-x-3">
                        <button
                          onClick={() => {
                            setEditMember(member);
                            setEditName(member.name);
                            setEditRole(member.role);
                          }}
                          className="text-indigo-600 hover:text-indigo-900 font-medium text-sm"
                        >
                          Edit
                        </button>
                        {member.is_active && (
                          <button
                            onClick={() => setDeactivateMember(member)}
                            className="text-red-600 hover:text-red-900 font-medium text-sm"
                          >
                            Deactivate
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* Shifts Tab Content */}
      {activeTab === 'shifts' && (
        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-4 bg-white p-4 rounded-xl border border-gray-200 shadow-sm">
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">From Date</label>
              <input
                type="date"
                value={dateFrom}
                onChange={(e) => setDateFrom(e.target.value)}
                className="px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">To Date</label>
              <input
                type="date"
                value={dateTo}
                onChange={(e) => setDateTo(e.target.value)}
                className="px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
              />
            </div>
          </div>

          <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
            {loading ? (
              <div className="p-8 text-center text-gray-500 text-sm">Loading shift history...</div>
            ) : shiftList.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">No shift records found for selected period.</div>
            ) : (
              <table className="w-full text-left text-sm text-gray-600">
                <thead className="bg-gray-50 text-gray-500 font-semibold border-b border-gray-200 text-xs uppercase tracking-wider">
                  <tr>
                    <th className="px-6 py-3">Staff Name</th>
                    <th className="px-6 py-3">Role</th>
                    <th className="px-6 py-3">Started At</th>
                    <th className="px-6 py-3">Ended At</th>
                    <th className="px-6 py-3">Duration</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {shiftList.map((shift) => (
                    <tr key={shift.id} className="hover:bg-gray-50/50 transition-colors">
                      <td className="px-6 py-4 font-medium text-gray-900">{shift.staff_name}</td>
                      <td className="px-6 py-4">
                        <Badge status={shift.role} />
                      </td>
                      <td className="px-6 py-4">{new Date(shift.started_at).toLocaleString()}</td>
                      <td className="px-6 py-4">
                        {shift.ended_at ? new Date(shift.ended_at).toLocaleString() : <span className="text-green-600 font-medium">Active Shift</span>}
                      </td>
                      <td className="px-6 py-4 font-mono text-xs">
                        {shift.duration_minutes ? `${Math.floor(shift.duration_minutes / 60)}h ${shift.duration_minutes % 60}m` : 'In Progress'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* Add Staff Modal */}
      {isAddOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-xl border border-gray-200 space-y-4">
            <h2 className="text-xl font-bold text-gray-900">Add Staff Member</h2>

            {addInlineError && (
              <div className="p-3 bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg" data-testid="inline-error">
                {addInlineError}
              </div>
            )}

            <form onSubmit={handleAddSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Full Name</label>
                <input
                  type="text"
                  required
                  value={addName}
                  onChange={(e) => setAddName(e.target.value)}
                  placeholder="e.g. Ramesh Kumar"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Mobile Phone Number</label>
                <input
                  type="text"
                  required
                  value={addPhone}
                  onChange={(e) => setAddPhone(e.target.value)}
                  placeholder="+919876543210"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Role Assignment</label>
                <select
                  value={addRole}
                  onChange={(e: any) => setAddRole(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                >
                  <option value="CASHIER">Cashier</option>
                  <option value="STOCK_ASSOCIATE">Stock Associate</option>
                  <option value="SECURITY">Security</option>
                  <option value="MANAGER">Manager</option>
                </select>
              </div>

              <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
                <button
                  type="button"
                  onClick={() => setIsAddOpen(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={addSubmitting}
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50"
                >
                  {addSubmitting ? 'Adding...' : 'Add Member'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Staff Modal */}
      {editMember && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-xl border border-gray-200 space-y-4">
            <h2 className="text-xl font-bold text-gray-900">Edit Staff Member</h2>

            <form onSubmit={handleEditSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Full Name</label>
                <input
                  type="text"
                  required
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-500 mb-1">Mobile Phone (Immutable)</label>
                <input
                  type="text"
                  disabled
                  value={editMember.phone}
                  className="w-full px-3 py-2 border border-gray-200 bg-gray-50 rounded-lg text-sm font-mono text-gray-500 cursor-not-allowed"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Role Assignment</label>
                <select
                  value={editRole}
                  onChange={(e: any) => setEditRole(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                >
                  <option value="CASHIER">Cashier</option>
                  <option value="STOCK_ASSOCIATE">Stock Associate</option>
                  <option value="SECURITY">Security</option>
                  <option value="MANAGER">Manager</option>
                </select>
              </div>

              <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
                <button
                  type="button"
                  onClick={() => setEditMember(null)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={editSubmitting}
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50"
                >
                  {editSubmitting ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Deactivate ConfirmDialog */}
      <ConfirmDialog
        isOpen={Boolean(deactivateMember)}
        title="Deactivate Staff Member"
        message="This will immediately log this staff member out everywhere and end any active shift. This cannot be undone without re-adding them."
        confirmLabel="Deactivate Staff"
        cancelLabel="Cancel"
        isDestructive={true}
        onConfirm={handleConfirmDeactivate}
        onCancel={() => setDeactivateMember(null)}
      />
    </div>
  );
}
