"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { Badge, ConfirmDialog, ERPCredentialModal } from "@zippyra/ui";

export interface ERPConnectionItem {
  id: string;
  chain_id: string;
  erp_type: "SAP" | "TALLY" | "BUSY";
  integration_mode: "DIRECT" | "AGENT_POLLED";
  display_name: string;
  enabled_outbound_events: string[];
  status: "PENDING_SETUP" | "ACTIVE" | "PAUSED" | "ERROR";
  last_inbound_at?: string | null;
  last_outbound_at?: string | null;
  last_agent_poll_at?: string | null;
  created_at: string;
}

export default function HQERPPage() {
  const chainId = "chain-001";
  const [userRole, setUserRole] = useState<string>("OWNER"); // Default to OWNER, can switch for testing/role gating

  const [connections, setConnections] = useState<ERPConnectionItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // New Connection Form Modal
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [erpType, setErpType] = useState<"SAP" | "TALLY" | "BUSY">("SAP");
  const [mode, setMode] = useState<"DIRECT" | "AGENT_POLLED">("DIRECT");
  const [displayName, setDisplayName] = useState("");
  const [enabledEvents, setEnabledEvents] = useState<string[]>(["order.completed"]);
  const [baseURL, setBaseURL] = useState("");
  const [authType, setAuthType] = useState("BASIC");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [apiKey, setApiKey] = useState("");

  // One-time credential modal
  const [credentialModalData, setCredentialModalData] = useState<{
    connectionId: string;
    mode: "DIRECT" | "AGENT_POLLED";
    secret: string;
    agentKey?: string | null;
    note?: string;
  } | null>(null);

  // Confirm dialog for Rotate Secret
  const [rotateConfirmConn, setRotateConfirmConn] = useState<ERPConnectionItem | null>(null);

  const fetchConnections = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/v1/integration/connections?chain_id=${chainId}`, {
        headers: {
          "X-User-Role": userRole,
          "X-Chain-ID": chainId,
        },
      });
      if (res.ok) {
        const data = await res.json();
        setConnections(data.connections || []);
      } else {
        setError(`Failed to fetch ERP connections: HTTP ${res.status}`);
      }
    } catch (err: any) {
      setError(err?.message || "Network error fetching connections");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConnections();
  }, [userRole]);

  // Handle ERP Type change auto-suggestion for mode
  const handleErpTypeChange = (type: "SAP" | "TALLY" | "BUSY") => {
    setErpType(type);
    if (type === "TALLY" || type === "BUSY") {
      setMode("AGENT_POLLED");
    } else {
      setMode("DIRECT");
    }
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const payload: any = {
      chain_id: chainId,
      erp_type: erpType,
      integration_mode: mode,
      display_name: displayName || `${erpType} Integration`,
      enabled_outbound_events: enabledEvents,
    };

    if (mode === "DIRECT") {
      payload.outbound_config = {
        base_url: baseURL,
        auth_type: authType,
        username,
        password,
        api_key: apiKey,
      };
    }

    try {
      const res = await fetch("/v1/integration/connections", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-User-Role": userRole,
          "X-Chain-ID": chainId,
          "X-User-ID": "usr-owner-1",
        },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.message || `Failed to create connection: HTTP ${res.status}`);
      }

      const data = await res.json();
      setShowCreateModal(false);

      // Open credential download modal
      setCredentialModalData({
        connectionId: data.connection.id,
        mode: data.connection.integration_mode,
        secret: data.webhook_secret,
        agentKey: data.agent_api_key,
        note: data.connector_setup_note,
      });

      fetchConnections();
    } catch (err: any) {
      setError(err?.message || "Failed to create connection");
    }
  };

  const handleToggleStatus = async (conn: ERPConnectionItem) => {
    const newStatus = conn.status === "PAUSED" ? "ACTIVE" : "PAUSED";
    try {
      const res = await fetch(`/v1/integration/connections/${conn.id}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-User-Role": userRole,
          "X-Chain-ID": chainId,
        },
        body: JSON.stringify({ status: newStatus }),
      });
      if (res.ok) {
        fetchConnections();
      }
    } catch (err) {
      console.error("Failed to toggle status:", err);
    }
  };

  const handleRotateSecret = async (conn: ERPConnectionItem) => {
    try {
      const res = await fetch(`/v1/integration/connections/${conn.id}/rotate-secret`, {
        method: "POST",
        headers: {
          "X-User-Role": userRole,
          "X-Chain-ID": chainId,
        },
      });

      if (res.ok) {
        const data = await res.json();
        setRotateConfirmConn(null);
        setCredentialModalData({
          connectionId: conn.id,
          mode: conn.integration_mode,
          secret: data.webhook_secret,
          agentKey: data.agent_api_key,
          note: "Secret rotated. Old secret remains valid for a 5-minute grace window.",
        });
        fetchConnections();
      }
    } catch (err) {
      console.error("Failed to rotate secret:", err);
    }
  };

  const isStaleAgent = (conn: ERPConnectionItem) => {
    if (conn.integration_mode !== "AGENT_POLLED") return false;
    if (!conn.last_agent_poll_at) return true;
    const diffMs = Date.now() - new Date(conn.last_agent_poll_at).getTime();
    return diffMs > 5 * 60 * 1000; // 5 minutes staleness threshold
  };

  const statusVariant = (status: string) => {
    switch (status) {
      case "ACTIVE":
        return "success";
      case "PENDING_SETUP":
        return "warning";
      case "PAUSED":
        return "info";
      default:
        return "error";
    }
  };

  return (
    <div className="space-y-8 p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">Enterprise ERP Integrations</h1>
          <p className="text-sm text-slate-400">SAP OData Cloud Direct & On-Premise Tally / Busy Connector Management</p>
        </div>

        <div className="flex items-center space-x-4">
          {/* Role simulation selector for testing/verifying gating */}
          <div className="flex items-center space-x-2 text-xs text-slate-400">
            <span>Simulate Role:</span>
            <select
              value={userRole}
              onChange={(e) => setUserRole(e.target.value)}
              className="px-2 py-1 bg-slate-800 border border-slate-700 rounded text-slate-200 text-xs"
            >
              <option value="OWNER">OWNER</option>
              <option value="FINANCE">FINANCE</option>
              <option value="OPERATIONS">OPERATIONS</option>
            </select>
          </div>

          {/* New Connection Button - OWNER role only */}
          {userRole === "OWNER" && (
            <button
              onClick={() => setShowCreateModal(true)}
              data-testid="new-connection-btn"
              className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white font-bold text-sm rounded-lg transition-colors shadow flex items-center space-x-2"
            >
              <span>+ New ERP Connection</span>
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg text-rose-400 text-sm">
          {error}
        </div>
      )}

      {/* Connection List */}
      <section className="space-y-4">
        {loading ? (
          <div className="p-8 text-center text-slate-500">Loading ERP connections...</div>
        ) : connections.length === 0 ? (
          <div className="p-12 bg-slate-900 border border-slate-800 rounded-xl text-center space-y-3">
            <div className="text-slate-400 text-sm font-semibold">No ERP Connections Configured</div>
            <p className="text-xs text-slate-500 max-w-md mx-auto">
              Connect your SAP Cloud OData endpoints or install on-premise Tally / Busy connector agents to sync inventory, GRNs, and completed orders.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {connections.map((conn) => {
              const stale = isStaleAgent(conn);

              return (
                <div
                  key={conn.id}
                  data-testid={`erp-connection-card-${conn.id}`}
                  className="bg-slate-900 border border-slate-800 rounded-xl p-5 shadow-lg flex flex-col justify-between space-y-4 hover:border-slate-700 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="space-y-1">
                      <div className="flex items-center space-x-2">
                        <span className="font-extrabold text-white text-base">{conn.display_name}</span>
                        <span className="px-2 py-0.5 text-[10px] font-mono font-bold bg-slate-800 text-purple-300 border border-slate-700 rounded">
                          {conn.erp_type}
                        </span>
                      </div>
                      <div className="text-xs text-slate-400 font-mono">ID: {conn.id}</div>
                    </div>

                    <Badge status={conn.status} />
                  </div>

                  {/* Mode & Tooltip */}
                  <div className="flex items-center space-x-3 text-xs">
                    <span
                      className={`inline-flex items-center px-2.5 py-1 rounded-md font-bold text-[11px] ${
                        conn.integration_mode === "DIRECT"
                          ? "bg-emerald-500/10 text-emerald-300 border border-emerald-500/30"
                          : "bg-purple-500/10 text-purple-300 border border-purple-500/30"
                      }`}
                    >
                      {conn.integration_mode}
                    </span>

                    {conn.integration_mode === "AGENT_POLLED" && (
                      <span
                        className="text-[11px] text-slate-400 hover:text-slate-200 underline cursor-help"
                        title="Requires installing the Zippyra Connector on the machine running Tally/Busy"
                      >
                        Requires Connector Agent
                      </span>
                    )}
                  </div>

                  {/* Staleness warning banner */}
                  {stale && (
                    <div
                      data-testid={`staleness-warning-${conn.id}`}
                      className="p-2.5 bg-amber-500/10 border border-amber-500/30 rounded-lg text-amber-300 text-xs flex items-center space-x-2"
                    >
                      <span className="w-2 h-2 rounded-full bg-amber-400 animate-ping"></span>
                      <span className="font-semibold">Connector May Be Offline</span>
                      <span className="text-[10px] text-amber-400/80">(No poll check-in within last 5 minutes)</span>
                    </div>
                  )}

                  {/* Timestamps */}
                  <div className="grid grid-cols-2 gap-2 text-[11px] text-slate-400 bg-slate-950/60 p-3 rounded-lg border border-slate-800 font-mono">
                    <div>
                      <span className="text-slate-500 block">Last Inbound</span>
                      {conn.last_inbound_at ? new Date(conn.last_inbound_at).toLocaleTimeString() : "Never"}
                    </div>
                    <div>
                      <span className="text-slate-500 block">Last Outbound</span>
                      {conn.last_outbound_at ? new Date(conn.last_outbound_at).toLocaleTimeString() : "Never"}
                    </div>
                  </div>

                  {/* Action Controls */}
                  <div className="flex items-center justify-between pt-2 border-t border-slate-800">
                    <Link
                      href={`/dashboard/erp/${conn.id}`}
                      data-testid={`view-detail-btn-${conn.id}`}
                      className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold rounded-lg transition-colors border border-slate-700"
                    >
                      View Logs & Sync Jobs &rarr;
                    </Link>

                    {userRole === "OWNER" && (
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={() => handleToggleStatus(conn)}
                          data-testid={`toggle-status-btn-${conn.id}`}
                          className="px-2.5 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-lg transition-colors border border-slate-700"
                        >
                          {conn.status === "PAUSED" ? "Resume" : "Pause"}
                        </button>

                        <button
                          onClick={() => setRotateConfirmConn(conn)}
                          data-testid={`rotate-secret-btn-${conn.id}`}
                          className="px-2.5 py-1.5 bg-amber-500/10 hover:bg-amber-500/20 text-amber-300 border border-amber-500/30 text-xs font-semibold rounded-lg transition-colors"
                        >
                          Rotate Secret
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>

      {/* New Connection Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4 overflow-y-auto">
          <div className="bg-slate-900 border border-slate-800 w-full max-w-xl rounded-2xl p-6 shadow-2xl space-y-5 text-slate-200 my-8">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <h3 className="text-lg font-bold text-white">Create New ERP Connection</h3>
              <button
                onClick={() => setShowCreateModal(false)}
                className="text-slate-400 hover:text-white text-xs font-bold"
              >
                &times; Close
              </button>
            </div>

            <form onSubmit={handleCreateSubmit} className="space-y-4 text-xs">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-slate-400 mb-1 font-semibold">ERP System</label>
                  <select
                    value={erpType}
                    onChange={(e) => handleErpTypeChange(e.target.value as any)}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm"
                  >
                    <option value="SAP">SAP (Cloud / OData Direct)</option>
                    <option value="TALLY">Tally (On-Premise Connector Agent)</option>
                    <option value="BUSY">Busy (On-Premise Connector Agent)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-semibold">Integration Mode</label>
                  <select
                    value={mode}
                    onChange={(e) => setMode(e.target.value as any)}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm"
                  >
                    <option value="DIRECT">DIRECT (Outbound Push via SAP OData)</option>
                    <option value="AGENT_POLLED">AGENT_POLLED (On-Premise Polling Agent)</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-slate-400 mb-1 font-semibold">Display Name</label>
                <input
                  type="text"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="e.g. Primary Store Tally ERP"
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm"
                  required
                />
              </div>

              {/* Checkboxes for Outbound Events */}
              <div>
                <label className="block text-slate-400 mb-2 font-semibold">Enabled Outbound Push Events</label>
                <div className="space-y-2 bg-slate-950 p-3 rounded-lg border border-slate-800">
                  {["order.completed", "inventory.stock_updated", "warehouse.grn_completed"].map((ev) => (
                    <label key={ev} className="flex items-center space-x-2 text-slate-300 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={enabledEvents.includes(ev)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setEnabledEvents([...enabledEvents, ev]);
                          } else {
                            setEnabledEvents(enabledEvents.filter((x) => x !== ev));
                          }
                        }}
                        className="rounded border-slate-700 bg-slate-800 text-purple-600 focus:ring-purple-500"
                      />
                      <span className="font-mono">{ev}</span>
                    </label>
                  ))}
                </div>
              </div>

              {/* DIRECT Mode Outbound Config */}
              {mode === "DIRECT" && (
                <div className="p-4 bg-slate-950 rounded-xl border border-slate-800 space-y-3">
                  <h4 className="font-bold text-slate-200">SAP OData Outbound Endpoint Config</h4>
                  <div>
                    <label className="block text-slate-400 mb-1">Base OData URL</label>
                    <input
                      type="url"
                      value={baseURL}
                      onChange={(e) => setBaseURL(e.target.value)}
                      placeholder="https://sap-cloud.example.com/sap/bc/odata"
                      className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white font-mono"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-slate-400 mb-1">Auth Type</label>
                      <select
                        value={authType}
                        onChange={(e) => setAuthType(e.target.value)}
                        className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white"
                      >
                        <option value="BASIC">Basic Auth (Username / Password)</option>
                        <option value="API_KEY">API Key (X-API-Key Header)</option>
                      </select>
                    </div>
                    {authType === "BASIC" ? (
                      <div>
                        <label className="block text-slate-400 mb-1">Username</label>
                        <input
                          type="text"
                          value={username}
                          onChange={(e) => setUsername(e.target.value)}
                          className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white"
                        />
                      </div>
                    ) : (
                      <div>
                        <label className="block text-slate-400 mb-1">API Key</label>
                        <input
                          type="password"
                          value={apiKey}
                          onChange={(e) => setApiKey(e.target.value)}
                          className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white"
                        />
                      </div>
                    )}
                  </div>
                  {authType === "BASIC" && (
                    <div>
                      <label className="block text-slate-400 mb-1">Password</label>
                      <input
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white"
                      />
                    </div>
                  )}
                </div>
              )}

              <div className="flex justify-end space-x-3 pt-3 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 text-slate-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  data-testid="submit-connection-btn"
                  className="px-5 py-2 bg-purple-600 hover:bg-purple-500 text-white font-bold rounded-lg shadow"
                >
                  Create Connection
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Confirm Rotate Secret Dialog */}
      {rotateConfirmConn && (
        <ConfirmDialog
          isOpen={true}
          title="Rotate Integration Secret Credentials?"
          message="Rotating secret credentials generates a new HMAC Webhook Secret (and Agent API Key). Old secret remains valid for a 5-minute grace window before full invalidation."
          confirmText="Rotate Secret Now"
          onConfirm={() => handleRotateSecret(rotateConfirmConn)}
          onCancel={() => setRotateConfirmConn(null)}
        />
      )}

      {/* One-Time Credential Display Modal */}
      {credentialModalData && (
        <ERPCredentialModal
          connectionId={credentialModalData.connectionId}
          integrationMode={credentialModalData.mode}
          webhookSecret={credentialModalData.secret}
          agentApiKey={credentialModalData.agentKey}
          connectorSetupNote={credentialModalData.note}
          onConfirmed={() => setCredentialModalData(null)}
        />
      )}
    </div>
  );
}
