'use client';

import React, { useState, useEffect } from 'react';
import {
  useSystemOps,
  DLQTopicSummary,
  DLQMessage,
  FeatureFlag,
  CircuitBreakerStatus,
} from '@zippyra/hooks';
import { Badge, ConfirmDialog } from '@zippyra/ui';

export default function SystemOpsPage() {
  const {
    getDlqTopics,
    peekDlqMessages,
    replayDlqMessages,
    discardDlqMessages,
    getFeatureFlags,
    createFeatureFlag,
    updateFeatureFlag,
    deleteFeatureFlag,
    getCircuitBreakerStatus,
  } = useSystemOps();

  // Active Tab
  const [activeTab, setActiveTab] = useState<'dlq' | 'flags' | 'breaker'>('dlq');

  // DLQ Tab State
  const [dlqTopics, setDlqTopics] = useState<DLQTopicSummary[]>([]);
  const [loadingDlq, setLoadingDlq] = useState<boolean>(true);
  const [selectedTopic, setSelectedTopic] = useState<string | null>(null);
  const [peekMessages, setPeekMessages] = useState<DLQMessage[]>([]);
  const [loadingPeek, setLoadingPeek] = useState<boolean>(false);
  const [selectedOffsets, setSelectedOffsets] = useState<number[]>([]);
  const [replayResult, setReplayResult] = useState<{ replayed: number; failed: number[] } | null>(null);

  // DLQ Discard Confirm & Step-Up State
  const [discardingTopic, setDiscardingTopic] = useState<string | null>(null);
  const [discardReason, setDiscardReason] = useState<string>('Unrecoverable dead-letter message');
  const [stepUpTokenInput, setStepUpTokenInput] = useState<string>('');
  const [showStepUpModal, setShowStepUpModal] = useState<boolean>(false);

  // Feature Flags Tab State
  const [flagsList, setFlagsList] = useState<FeatureFlag[]>([]);
  const [loadingFlags, setLoadingFlags] = useState<boolean>(true);
  const [isNewFlagOpen, setIsNewFlagOpen] = useState<boolean>(false);
  const [newKey, setNewKey] = useState<string>('');
  const [newDesc, setNewDesc] = useState<string>('');
  const [newScope, setNewScope] = useState<'GLOBAL' | 'CHAIN' | 'STORE' | 'USER_PERCENTAGE'>('GLOBAL');
  const [editingFlag, setEditingFlag] = useState<FeatureFlag | null>(null);
  const [editGlobal, setEditGlobal] = useState<boolean>(false);
  const [editScopeIDsStr, setEditScopeIDsStr] = useState<string>('');
  const [editUserPct, setEditUserPct] = useState<number>(0);
  const [flagStepUpToken, setFlagStepUpToken] = useState<string>('');

  // Circuit Breaker Tab State
  const [breakerStatus, setBreakerStatus] = useState<CircuitBreakerStatus | null>(null);
  const [loadingBreaker, setLoadingBreaker] = useState<boolean>(true);

  // Data Fetchers
  const fetchDLQ = async () => {
    setLoadingDlq(true);
    try {
      const res = await getDlqTopics();
      setDlqTopics(res.dlq_topics || []);
    } catch (_) {
      setDlqTopics([]);
    } finally {
      setLoadingDlq(false);
    }
  };

  const fetchPeek = async (topic: string) => {
    setLoadingPeek(true);
    try {
      const res = await peekDlqMessages(topic, 20);
      setPeekMessages(res.messages || []);
    } catch (_) {
      setPeekMessages([]);
    } finally {
      setLoadingPeek(false);
    }
  };

  const fetchFlags = async () => {
    setLoadingFlags(true);
    try {
      const res = await getFeatureFlags();
      setFlagsList(res.feature_flags || []);
    } catch (_) {
      setFlagsList([]);
    } finally {
      setLoadingFlags(false);
    }
  };

  const fetchBreaker = async () => {
    try {
      const res = await getCircuitBreakerStatus();
      setBreakerStatus(res);
    } catch (_) {
      setBreakerStatus({
        gateway: 'razorpay',
        state: 'CLOSED',
        error_rate_rolling_1min: 0.0,
      });
    } finally {
      setLoadingBreaker(false);
    }
  };

  // Tab Effects & 10s Auto-Refresh for Circuit Breaker
  useEffect(() => {
    if (activeTab === 'dlq') {
      fetchDLQ();
    } else if (activeTab === 'flags') {
      fetchFlags();
    } else if (activeTab === 'breaker') {
      fetchBreaker();
      const interval = setInterval(fetchBreaker, 10000);
      return () => clearInterval(interval);
    }
  }, [activeTab]);

  // DLQ Actions
  const handleSelectTopic = (topic: string) => {
    setSelectedTopic(topic);
    setSelectedOffsets([]);
    setReplayResult(null);
    fetchPeek(topic);
  };

  const handleToggleOffset = (off: number) => {
    setSelectedOffsets((prev) =>
      prev.includes(off) ? prev.filter((o) => o !== off) : [...prev, off]
    );
  };

  const handleReplaySelected = async () => {
    if (!selectedTopic || selectedOffsets.length === 0) return;
    try {
      const res = await replayDlqMessages(selectedTopic, selectedOffsets);
      setReplayResult({
        replayed: res.replayed_count,
        failed: res.failed_offsets || [],
      });
      fetchPeek(selectedTopic);
      fetchDLQ();
      setSelectedOffsets([]);
    } catch (err: any) {
      alert(err?.message || 'Replay failed');
    }
  };

  const handleConfirmDiscard = async () => {
    if (!selectedTopic || selectedOffsets.length === 0) return;
    try {
      await discardDlqMessages(selectedTopic, selectedOffsets, discardReason, stepUpTokenInput);
      setDiscardingTopic(null);
      setShowStepUpModal(false);
      fetchPeek(selectedTopic);
      fetchDLQ();
      setSelectedOffsets([]);
    } catch (err: any) {
      alert(err?.message || 'Discard failed');
    }
  };

  // Feature Flag Actions
  const handleCreateFlag = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await createFeatureFlag({
        flag_key: newKey.trim(),
        description: newDesc.trim(),
        scope_type: newScope,
      });
      setIsNewFlagOpen(false);
      setNewKey('');
      setNewDesc('');
      fetchFlags();
    } catch (err: any) {
      alert(err?.message || 'Failed to create flag');
    }
  };

  const handleUpdateFlagSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingFlag) return;

    try {
      let scopeIDs: string[] | undefined = undefined;
      if (editingFlag.scope_type === 'CHAIN' || editingFlag.scope_type === 'STORE') {
        scopeIDs = editScopeIDsStr.split(',').map((s) => s.trim()).filter(Boolean);
      }

      await updateFeatureFlag(
        editingFlag.flag_key,
        {
          enabled_globally: editGlobal,
          enabled_scope_ids: scopeIDs,
          user_percentage: editingFlag.scope_type === 'USER_PERCENTAGE' ? editUserPct : undefined,
        },
        flagStepUpToken
      );

      setEditingFlag(null);
      setFlagStepUpToken('');
      fetchFlags();
    } catch (err: any) {
      if (err?.code === 'STEP_UP_REQUIRED' || err?.status === 403 || err?.message?.includes('step-up')) {
        alert('High-risk feature flag update requires a valid X-StepUp-Token header. Please provide a step-up token.');
      } else {
        alert(err?.message || 'Failed to update flag');
      }
    }
  };

  const formatRolloutSummary = (flag: FeatureFlag) => {
    if (flag.scope_type === 'GLOBAL') {
      return flag.enabled_globally ? 'Global: ON' : 'Global: OFF';
    }
    if (flag.scope_type === 'CHAIN' || flag.scope_type === 'STORE') {
      const count = flag.enabled_scope_ids?.length || 0;
      return `${flag.scope_type} Scope (${count} IDs)`;
    }
    if (flag.scope_type === 'USER_PERCENTAGE') {
      return `${flag.user_percentage || 0}% of Users`;
    }
    return flag.scope_type;
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">System Operations & Health</h1>
        <p className="text-sm text-gray-500 mt-1">
          Centralized administration for Kafka DLQs, feature flags, and payment gateway fault tolerance
        </p>
      </div>

      {/* Tabs Header */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          <button
            onClick={() => setActiveTab('dlq')}
            className={`py-3 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'dlq'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Kafka Dead-Letter Queues (DLQ)
          </button>
          <button
            onClick={() => setActiveTab('flags')}
            className={`py-3 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'flags'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Feature Flags Admin
          </button>
          <button
            onClick={() => setActiveTab('breaker')}
            className={`py-3 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'breaker'
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Circuit Breaker Status
          </button>
        </nav>
      </div>

      {/* TAB 1: KAFKA DLQ */}
      {activeTab === 'dlq' && (
        <div className="space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
            {loadingDlq ? (
              <div className="p-8 text-center text-gray-500 text-sm">Loading DLQ topic metrics...</div>
            ) : dlqTopics.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">No DLQ topics with active dead-letter messages found.</div>
            ) : (
              <table className="w-full text-left text-sm text-gray-600">
                <thead className="bg-gray-50 text-gray-500 font-semibold border-b border-gray-200 text-xs uppercase tracking-wider">
                  <tr>
                    <th className="px-6 py-3">Topic Name</th>
                    <th className="px-6 py-3">Message Depth</th>
                    <th className="px-6 py-3">Oldest Message Age</th>
                    <th className="px-6 py-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {dlqTopics.map((item) => (
                    <tr
                      key={item.topic_name}
                      onClick={() => handleSelectTopic(item.topic_name)}
                      className={`hover:bg-gray-50/70 cursor-pointer transition-colors ${
                        selectedTopic === item.topic_name ? 'bg-indigo-50/50' : ''
                      }`}
                    >
                      <td className="px-6 py-4 font-mono font-medium text-gray-900">{item.topic_name}</td>
                      <td className="px-6 py-4">
                        <Badge status={item.message_count > 0 ? 'REJECTED' : 'RECEIVED'} />
                      </td>
                      <td className="px-6 py-4">
                        {item.oldest_message_age_seconds > 0
                          ? `${Math.floor(item.oldest_message_age_seconds / 60)} mins ago`
                          : 'None'}
                      </td>
                      <td className="px-6 py-4 text-right">
                        <button className="text-indigo-600 hover:text-indigo-900 font-medium text-sm">
                          Inspect & Peek →
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          {/* Drawer / Peek View */}
          {selectedTopic && (
            <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-md space-y-4">
              <div className="flex items-center justify-between border-b border-gray-200 pb-4">
                <div>
                  <h3 className="text-lg font-bold text-gray-900">
                    Peek Messages: <span className="font-mono text-indigo-600">{selectedTopic}</span>
                  </h3>
                  <p className="text-xs text-gray-500 mt-0.5">Showing recent messages (read-only peek without offset commit)</p>
                </div>
                <div className="flex items-center space-x-3">
                  <button
                    onClick={handleReplaySelected}
                    disabled={selectedOffsets.length === 0}
                    className="px-3 py-1.5 bg-indigo-600 text-white rounded-lg text-xs font-medium hover:bg-indigo-700 disabled:opacity-40"
                  >
                    Replay Selected ({selectedOffsets.length})
                  </button>
                  <button
                    onClick={() => {
                      if (selectedOffsets.length > 0) setDiscardingTopic(selectedTopic);
                    }}
                    disabled={selectedOffsets.length === 0}
                    className="px-3 py-1.5 bg-red-600 text-white rounded-lg text-xs font-medium hover:bg-red-700 disabled:opacity-40"
                  >
                    Discard Selected ({selectedOffsets.length})
                  </button>
                </div>
              </div>

              {replayResult && (
                <div className="p-3 bg-green-50 border border-green-200 text-green-800 text-xs rounded-lg">
                  Successfully replayed {replayResult.replayed} message(s) to target topic.
                </div>
              )}

              {loadingPeek ? (
                <div className="p-6 text-center text-gray-500 text-sm">Peeking DLQ messages...</div>
              ) : peekMessages.length === 0 ? (
                <div className="p-6 text-center text-gray-500 text-sm">No active messages found in this DLQ topic.</div>
              ) : (
                <div className="space-y-4 max-h-96 overflow-y-auto pr-2">
                  {peekMessages.map((msg) => (
                    <div
                      key={msg.offset}
                      className="border border-gray-200 rounded-lg p-4 bg-gray-50 space-y-2 text-xs font-mono"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <input
                            type="checkbox"
                            checked={selectedOffsets.includes(msg.offset)}
                            onChange={() => handleToggleOffset(msg.offset)}
                            className="h-4 w-4 text-indigo-600 rounded border-gray-300 focus:ring-indigo-500 cursor-pointer"
                          />
                          <span className="font-semibold text-gray-800">Offset #{msg.offset}</span>
                          <span className="text-gray-400">|</span>
                          <span className="text-gray-600">Key: {msg.key || 'N/A'}</span>
                        </div>
                        <span className="text-gray-400">{new Date(msg.timestamp).toLocaleString()}</span>
                      </div>
                      <pre className="p-3 bg-gray-900 text-emerald-400 rounded-md overflow-x-auto text-xs">
                        {typeof msg.value === 'object' ? JSON.stringify(msg.value, null, 2) : String(msg.value)}
                      </pre>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* TAB 2: FEATURE FLAGS */}
      {activeTab === 'flags' && (
        <div className="space-y-6">
          <div className="flex justify-end">
            <button
              onClick={() => setIsNewFlagOpen(true)}
              className="px-4 py-2 bg-indigo-600 text-white font-medium text-sm rounded-lg hover:bg-indigo-700 shadow-sm"
            >
              + New Feature Flag
            </button>
          </div>

          <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
            {loadingFlags ? (
              <div className="p-8 text-center text-gray-500 text-sm">Loading feature flags...</div>
            ) : flagsList.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">No feature flags registered.</div>
            ) : (
              <table className="w-full text-left text-sm text-gray-600">
                <thead className="bg-gray-50 text-gray-500 font-semibold border-b border-gray-200 text-xs uppercase tracking-wider">
                  <tr>
                    <th className="px-6 py-3">Flag Key</th>
                    <th className="px-6 py-3">Description</th>
                    <th className="px-6 py-3">Scope Type</th>
                    <th className="px-6 py-3">Current Rollout</th>
                    <th className="px-6 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {flagsList.map((flag) => (
                    <tr key={flag.flag_key} className="hover:bg-gray-50/50 transition-colors">
                      <td className="px-6 py-4 font-mono font-medium text-gray-900">{flag.flag_key}</td>
                      <td className="px-6 py-4 text-gray-600">{flag.description}</td>
                      <td className="px-6 py-4">
                        <Badge status={flag.scope_type} />
                      </td>
                      <td className="px-6 py-4 font-medium text-gray-800">{formatRolloutSummary(flag)}</td>
                      <td className="px-6 py-4 text-right space-x-3">
                        <button
                          onClick={() => {
                            setEditingFlag(flag);
                            setEditGlobal(flag.enabled_globally);
                            setEditScopeIDsStr(flag.enabled_scope_ids?.join(', ') || '');
                            setEditUserPct(flag.user_percentage || 0);
                          }}
                          className="text-indigo-600 hover:text-indigo-900 font-medium text-sm"
                        >
                          Edit Rollout
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* TAB 3: CIRCUIT BREAKER STATUS */}
      {activeTab === 'breaker' && (
        <div className="max-w-2xl space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm space-y-4">
            <div className="flex items-center justify-between border-b border-gray-200 pb-4">
              <div>
                <h2 className="text-lg font-bold text-gray-900">Razorpay ↔ Cashfree Circuit Breaker</h2>
                <p className="text-xs text-gray-500 mt-0.5">Automated 1-minute rolling error-rate failover monitor</p>
              </div>
              <span className="text-xs font-mono text-gray-400">Auto-refreshing (10s)</span>
            </div>

            {loadingBreaker || !breakerStatus ? (
              <div className="p-4 text-center text-gray-500 text-sm">Fetching circuit breaker status...</div>
            ) : (
              <div className="space-y-4">
                <div className="flex items-center justify-between bg-gray-50 p-4 rounded-lg border border-gray-200">
                  <span className="text-sm font-medium text-gray-700">Breaker State</span>
                  {breakerStatus.state === 'CLOSED' && (
                    <Badge status="RECEIVED" /> // Green Healthy
                  )}
                  {breakerStatus.state === 'OPEN' && (
                    <Badge status="REJECTED" /> // Red Failing Over
                  )}
                  {breakerStatus.state === 'HALF_OPEN' && (
                    <Badge status="QC_PENDING" /> // Amber Testing Recovery
                  )}
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="p-4 bg-gray-50 rounded-lg border border-gray-200">
                    <span className="block text-xs text-gray-500 font-medium">Primary Gateway</span>
                    <span className="text-base font-bold text-gray-900 uppercase mt-1 block">
                      {breakerStatus.gateway}
                    </span>
                  </div>

                  <div className="p-4 bg-gray-50 rounded-lg border border-gray-200">
                    <span className="block text-xs text-gray-500 font-medium">Rolling 1-Min Error Rate</span>
                    <span className="text-base font-bold text-gray-900 mt-1 block">
                      {(breakerStatus.error_rate_rolling_1min * 100).toFixed(2)}%
                    </span>
                  </div>
                </div>

                {breakerStatus.state === 'OPEN' && (
                  <div className="p-4 bg-red-50 border border-red-200 rounded-lg space-y-2">
                    <div className="flex items-center justify-between text-xs font-medium text-red-800">
                      <span>Circuit Opened At:</span>
                      <span>{breakerStatus.opened_at ? new Date(breakerStatus.opened_at).toLocaleString() : 'N/A'}</span>
                    </div>
                    <div className="flex items-center justify-between text-xs font-medium text-red-800">
                      <span>Scheduled Recovery Trial:</span>
                      <span>{breakerStatus.will_retry_at ? new Date(breakerStatus.will_retry_at).toLocaleString() : 'N/A'}</span>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      )}

      {/* New Feature Flag Modal */}
      {isNewFlagOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-xl border border-gray-200 space-y-4">
            <h2 className="text-xl font-bold text-gray-900">Create Feature Flag</h2>
            <form onSubmit={handleCreateFlag} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Flag Key</label>
                <input
                  type="text"
                  required
                  value={newKey}
                  onChange={(e) => setNewKey(e.target.value)}
                  placeholder="e.g. cart.dynamic_discounts"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <input
                  type="text"
                  required
                  value={newDesc}
                  onChange={(e) => setNewDesc(e.target.value)}
                  placeholder="Describe flag target feature"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Scope Type</label>
                <select
                  value={newScope}
                  onChange={(e: any) => setNewScope(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                >
                  <option value="GLOBAL">Global</option>
                  <option value="CHAIN">Chain</option>
                  <option value="STORE">Store</option>
                  <option value="USER_PERCENTAGE">User Percentage</option>
                </select>
              </div>

              <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
                <button
                  type="button"
                  onClick={() => setIsNewFlagOpen(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700"
                >
                  Create Flag
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Feature Flag Modal */}
      {editingFlag && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-xl border border-gray-200 space-y-4">
            <h2 className="text-xl font-bold text-gray-900">
              Edit Flag: <span className="font-mono text-indigo-600">{editingFlag.flag_key}</span>
            </h2>

            <form onSubmit={handleUpdateFlagSubmit} className="space-y-4">
              {/* Scope Appropriate Controls */}
              {editingFlag.scope_type === 'GLOBAL' && (
                <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg border border-gray-200">
                  <span className="text-sm font-medium text-gray-700">Enable Globally</span>
                  <input
                    type="checkbox"
                    checked={editGlobal}
                    onChange={(e) => setEditGlobal(e.target.checked)}
                    className="h-5 w-5 text-indigo-600 rounded border-gray-300 focus:ring-indigo-500 cursor-pointer"
                  />
                </div>
              )}

              {(editingFlag.scope_type === 'CHAIN' || editingFlag.scope_type === 'STORE') && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Enabled {editingFlag.scope_type} IDs (comma separated)
                  </label>
                  <input
                    type="text"
                    value={editScopeIDsStr}
                    onChange={(e) => setEditScopeIDsStr(e.target.value)}
                    placeholder="e.g. store-mumbai-01, store-delhi-02"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                  />
                </div>
              )}

              {editingFlag.scope_type === 'USER_PERCENTAGE' && (
                <div>
                  <div className="flex justify-between text-sm font-medium text-gray-700 mb-1">
                    <span>Rollout Percentage</span>
                    <span>{editUserPct}%</span>
                  </div>
                  <input
                    type="range"
                    min="0"
                    max="100"
                    value={editUserPct}
                    onChange={(e) => setEditUserPct(Number(e.target.value))}
                    className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-indigo-600"
                  />
                </div>
              )}

              {/* Step-Up Token Input if High Risk */}
              {editingFlag.flag_key.startsWith('payment.') && (
                <div className="p-3 bg-amber-50 border border-amber-200 rounded-lg space-y-2">
                  <span className="text-xs font-semibold text-amber-900 block">
                    ⚠️ High-Risk Flag: Step-Up Authentication Required
                  </span>
                  <input
                    type="password"
                    placeholder="Enter Step-Up Token"
                    value={flagStepUpToken}
                    onChange={(e) => setFlagStepUpToken(e.target.value)}
                    className="w-full px-3 py-1.5 border border-amber-300 rounded text-xs font-mono focus:outline-none"
                  />
                </div>
              )}

              <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
                <button
                  type="button"
                  onClick={() => setEditingFlag(null)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm font-medium hover:bg-indigo-700"
                >
                  Save Changes
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Discard DLQ Confirm & Step-Up Modal */}
      {discardingTopic && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 shadow-xl border border-gray-200 space-y-4">
            <h2 className="text-xl font-bold text-gray-900">Soft Discard DLQ Messages</h2>
            <p className="text-sm text-gray-600">
              This will permanently hide these failed events from the DLQ view. This does not undo whatever caused them to fail.
            </p>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Discard Reason</label>
                <input
                  type="text"
                  value={discardReason}
                  onChange={(e) => setDiscardReason(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Step-Up Auth Token</label>
                <input
                  type="password"
                  placeholder="Enter Step-Up Token"
                  value={stepUpTokenInput}
                  onChange={(e) => setStepUpTokenInput(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:outline-none"
                />
              </div>
            </div>

            <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
              <button
                type="button"
                onClick={() => setDiscardingTopic(null)}
                className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleConfirmDiscard}
                className="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700"
              >
                Confirm Discard
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
