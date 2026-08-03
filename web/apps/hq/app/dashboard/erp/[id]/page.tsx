"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { Badge } from "@zippyra/ui";

export interface ConnectionDetail {
  id: string;
  chain_id: string;
  erp_type: string;
  integration_mode: "DIRECT" | "AGENT_POLLED";
  display_name: string;
  enabled_outbound_events: string[];
  status: string;
  last_inbound_at?: string | null;
  last_outbound_at?: string | null;
  last_agent_poll_at?: string | null;
}

export interface SyncJobItem {
  id: string;
  connection_id: string;
  direction: string;
  source_event_type: string;
  source_event_id: string;
  payload: any;
  status: "PENDING" | "DELIVERED" | "ACKNOWLEDGED" | "FAILED";
  attempt_count: number;
  failure_reason?: string | null;
  created_at: string;
}

export interface WebhookEventItem {
  id: string;
  connection_id: string;
  event_id: string;
  event_type: string;
  raw_payload: any;
  processing_result: "PENDING" | "APPLIED" | "FAILED" | "REJECTED";
  failure_reason?: string | null;
  created_at: string;
}

export default function ERPConnectionDetailPage({ params }: { params: { id: string } }) {
  const connectionId = params.id;
  const [conn, setConn] = useState<ConnectionDetail | null>(null);
  const [loadingConn, setLoadingConn] = useState(true);
  const [activeTab, setActiveTab] = useState<"sync-jobs" | "webhook-events">("sync-jobs");

  // Sync Jobs state
  const [syncJobs, setSyncJobs] = useState<SyncJobItem[]>([]);
  const [syncJobStatusFilter, setSyncJobStatusFilter] = useState<string>("");
  const [loadingJobs, setLoadingJobs] = useState(false);

  // Webhook Events state
  const [webhookEvents, setWebhookEvents] = useState<WebhookEventItem[]>([]);
  const [webhookResultFilter, setWebhookResultFilter] = useState<string>("");
  const [loadingEvents, setLoadingEvents] = useState(false);

  const fetchConnection = async () => {
    setLoadingConn(true);
    try {
      const res = await fetch(`/v1/integration/connections/${connectionId}`, {
        headers: { "X-User-Role": "OWNER" },
      });
      if (res.ok) {
        const data = await res.json();
        setConn(data);
      }
    } catch (err) {
      console.error("Failed to fetch connection detail:", err);
    } finally {
      setLoadingConn(false);
    }
  };

  const fetchSyncJobs = async () => {
    setLoadingJobs(true);
    try {
      let url = `/v1/integration/connections/${connectionId}/sync-jobs?direction=OUTBOUND`;
      if (syncJobStatusFilter) {
        url += `&status=${syncJobStatusFilter}`;
      }
      const res = await fetch(url, { headers: { "X-User-Role": "OWNER" } });
      if (res.ok) {
        const data = await res.json();
        setSyncJobs(data.jobs || []);
      }
    } catch (err) {
      console.error("Failed to fetch sync jobs:", err);
    } finally {
      setLoadingJobs(false);
    }
  };

  const fetchWebhookEvents = async () => {
    setLoadingEvents(true);
    try {
      let url = `/v1/integration/connections/${connectionId}/webhook-events`;
      if (webhookResultFilter) {
        url += `?result=${webhookResultFilter}`;
      }
      const res = await fetch(url, { headers: { "X-User-Role": "OWNER" } });
      if (res.ok) {
        const data = await res.json();
        setWebhookEvents(data.events || []);
      }
    } catch (err) {
      console.error("Failed to fetch webhook events:", err);
    } finally {
      setLoadingEvents(false);
    }
  };

  useEffect(() => {
    fetchConnection();
  }, [connectionId]);

  useEffect(() => {
    if (activeTab === "sync-jobs") {
      fetchSyncJobs();
    } else {
      fetchWebhookEvents();
    }
  }, [activeTab, syncJobStatusFilter, webhookResultFilter, connectionId]);

  const handleRetryJob = async (jobId: string) => {
    try {
      const res = await fetch(`/v1/integration/connections/${connectionId}/sync-jobs/${jobId}/retry`, {
        method: "POST",
        headers: { "X-User-Role": "OWNER" },
      });
      if (res.ok) {
        fetchSyncJobs();
      }
    } catch (err) {
      console.error("Failed to retry sync job:", err);
    }
  };

  if (loadingConn) {
    return <div className="p-8 text-center text-slate-500">Loading connection detail...</div>;
  }

  if (!conn) {
    return (
      <div className="p-8 text-center space-y-4">
        <div className="text-rose-400 font-bold">ERP Connection Not Found</div>
        <Link href="/dashboard/erp" className="text-purple-400 text-xs underline">
          &larr; Back to ERP Connections
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6 max-w-7xl mx-auto">
      <div>
        <Link href="/dashboard/erp" className="text-purple-400 text-xs font-semibold hover:underline">
          &larr; Back to ERP Connections List
        </Link>
      </div>

      {/* Header Info Card */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-lg space-y-4">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold text-slate-100">{conn.display_name}</h1>
            <div className="text-xs font-mono text-slate-400 mt-1">Connection ID: {conn.id}</div>
          </div>

          <div className="flex items-center space-x-3">
            <span className="px-2.5 py-1 text-xs font-mono font-bold bg-slate-800 text-purple-300 border border-slate-700 rounded-lg">
              {conn.erp_type}
            </span>
            <span
              className={`px-2.5 py-1 text-xs font-bold rounded-lg ${
                conn.integration_mode === "DIRECT"
                  ? "bg-emerald-500/10 text-emerald-300 border border-emerald-500/30"
                  : "bg-purple-500/10 text-purple-300 border border-purple-500/30"
              }`}
            >
              {conn.integration_mode}
            </span>
            <Badge status={conn.status} />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs text-slate-300 bg-slate-950 p-4 rounded-xl border border-slate-800 font-mono">
          <div>
            <span className="text-slate-500 block">Last Inbound Event</span>
            {conn.last_inbound_at ? new Date(conn.last_inbound_at).toLocaleString() : "Never"}
          </div>
          <div>
            <span className="text-slate-500 block">Last Outbound Push</span>
            {conn.last_outbound_at ? new Date(conn.last_outbound_at).toLocaleString() : "Never"}
          </div>
          <div>
            <span className="text-slate-500 block">Last Agent Poll Check-in</span>
            {conn.last_agent_poll_at ? new Date(conn.last_agent_poll_at).toLocaleString() : "N/A (Direct)"}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-slate-800 flex space-x-6 text-sm font-semibold text-slate-400">
        <button
          onClick={() => setActiveTab("sync-jobs")}
          data-testid="tab-sync-jobs"
          className={`pb-3 transition-colors ${
            activeTab === "sync-jobs" ? "text-purple-400 border-b-2 border-purple-500" : "hover:text-slate-200"
          }`}
        >
          Outbound Sync Jobs Queue & Log
        </button>

        <button
          onClick={() => setActiveTab("webhook-events")}
          data-testid="tab-webhook-events"
          className={`pb-3 transition-colors ${
            activeTab === "webhook-events" ? "text-purple-400 border-b-2 border-purple-500" : "hover:text-slate-200"
          }`}
        >
          Inbound Webhook Events Log
        </button>
      </div>

      {/* Tab 1: Sync Jobs */}
      {activeTab === "sync-jobs" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-bold text-slate-200">Outbound Sync Jobs</h2>

            <div className="flex items-center space-x-2 text-xs">
              <span className="text-slate-400">Status Filter:</span>
              <select
                value={syncJobStatusFilter}
                onChange={(e) => setSyncJobStatusFilter(e.target.value)}
                className="px-2.5 py-1 bg-slate-800 border border-slate-700 rounded-lg text-slate-200"
              >
                <option value="">All Statuses</option>
                <option value="PENDING">PENDING</option>
                <option value="DELIVERED">DELIVERED</option>
                <option value="ACKNOWLEDGED">ACKNOWLEDGED</option>
                <option value="FAILED">FAILED</option>
              </select>
            </div>
          </div>

          <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
            <table className="w-full text-left text-xs text-slate-300">
              <thead className="bg-slate-800/60 text-slate-400 uppercase font-semibold">
                <tr>
                  <th className="p-3">Source Event</th>
                  <th className="p-3">Event ID</th>
                  <th className="p-3">Attempts</th>
                  <th className="p-3">Status</th>
                  <th className="p-3">Created At</th>
                  <th className="p-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800 font-mono">
                {loadingJobs ? (
                  <tr>
                    <td colSpan={6} className="p-6 text-center text-slate-500">Loading sync jobs...</td>
                  </tr>
                ) : syncJobs.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="p-6 text-center text-slate-500">No outbound sync jobs found.</td>
                  </tr>
                ) : (
                  syncJobs.map((job) => (
                    <tr key={job.id} data-testid={`sync-job-row-${job.id}`} className="hover:bg-slate-800/40">
                      <td className="p-3 font-semibold text-slate-100">{job.source_event_type}</td>
                      <td className="p-3 text-slate-400">{job.source_event_id}</td>
                      <td className="p-3 text-amber-400">{job.attempt_count}</td>
                      <td className="p-3">
                        <Badge status={job.status} />
                      </td>
                      <td className="p-3 text-slate-400">{new Date(job.created_at).toLocaleTimeString()}</td>
                      <td className="p-3 text-right">
                        {conn.integration_mode === "DIRECT" && job.status === "FAILED" && (
                          <button
                            onClick={() => handleRetryJob(job.id)}
                            data-testid={`retry-job-btn-${job.id}`}
                            className="px-2.5 py-1 bg-amber-500/20 hover:bg-amber-500/30 text-amber-300 border border-amber-500/30 rounded text-[11px] font-bold"
                          >
                            Retry Push
                          </button>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 2: Webhook Events Log */}
      {activeTab === "webhook-events" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-bold text-slate-200">Inbound Webhook Troubleshooting Log</h2>

            <div className="flex items-center space-x-2 text-xs">
              <span className="text-slate-400">Result Filter:</span>
              <select
                value={webhookResultFilter}
                onChange={(e) => setWebhookResultFilter(e.target.value)}
                className="px-2.5 py-1 bg-slate-800 border border-slate-700 rounded-lg text-slate-200"
              >
                <option value="">All Results</option>
                <option value="APPLIED">APPLIED</option>
                <option value="FAILED">FAILED</option>
                <option value="PENDING">PENDING</option>
                <option value="REJECTED">REJECTED</option>
              </select>
            </div>
          </div>

          <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
            <table className="w-full text-left text-xs text-slate-300">
              <thead className="bg-slate-800/60 text-slate-400 uppercase font-semibold">
                <tr>
                  <th className="p-3">Event Type</th>
                  <th className="p-3">Event ID</th>
                  <th className="p-3">Result</th>
                  <th className="p-3">Failure Reason</th>
                  <th className="p-3 text-right">Received At</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800 font-mono">
                {loadingEvents ? (
                  <tr>
                    <td colSpan={5} className="p-6 text-center text-slate-500">Loading webhook events...</td>
                  </tr>
                ) : webhookEvents.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="p-6 text-center text-slate-500">No inbound webhook events logged.</td>
                  </tr>
                ) : (
                  webhookEvents.map((ev) => (
                    <tr key={ev.id} data-testid={`webhook-event-row-${ev.id}`} className="hover:bg-slate-800/40">
                      <td className="p-3 font-semibold text-slate-100">{ev.event_type}</td>
                      <td className="p-3 text-slate-400">{ev.event_id}</td>
                      <td className="p-3">
                        <Badge status={ev.processing_result} />
                      </td>
                      <td className="p-3 text-rose-400 truncate max-w-[200px]">{ev.failure_reason || "-"}</td>
                      <td className="p-3 text-right text-slate-400">{new Date(ev.created_at).toLocaleTimeString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
