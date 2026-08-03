"use client";

import React, { useEffect, useState } from "react";
import { useStepUp } from "../../../hooks/useStepUp";

interface DPDPRequest {
  id: string;
  user_id: string;
  request_type: string;
  status: string;
  detail?: string;
  handled_by?: string;
  rejection_reason?: string;
  sla_due_at: string;
  created_at: string;
  completed_at?: string;
}

interface DeletionProgress {
  reportedServices: string[];
}

export default function DPDPRequestsPage() {
  const [requests, setRequests] = useState<DPDPRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDeleteReq, setSelectedDeleteReq] = useState<DPDPRequest | null>(null);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [deletionProgress, setDeletionProgress] = useState<Record<string, DeletionProgress>>({});

  const { isStepUpVerified, openStepUpModal, StepUpModalComponent } = useStepUp();

  const fetchRequests = async () => {
    try {
      const res = await fetch("/v1/compliance/requests");
      if (res.ok) {
        const data = await res.json();
        const list: DPDPRequest[] = data.requests || [];

        // Sort by sla_due_at ascending (earliest due first)
        list.sort((a, b) => new Date(a.sla_due_at).getTime() - new Date(b.sla_due_at).getTime());
        setRequests(list);
      }
    } catch (err) {
      console.error("Failed to fetch DPDP requests:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRequests();
  }, []);

  // Poll in-progress requests every 3s
  useEffect(() => {
    const inProgressReqs = requests.filter((r) => r.status === "IN_PROGRESS");
    if (inProgressReqs.length === 0) return;

    const interval = setInterval(() => {
      fetchRequests();
    }, 3000);

    return () => clearInterval(interval);
  }, [requests]);

  const handleStartRequest = async (reqId: string) => {
    try {
      const res = await fetch(`/v1/compliance/dpdp/requests/${reqId}/review`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "IN_PROGRESS" }),
      });
      if (res.ok) {
        fetchRequests();
      }
    } catch (err) {
      console.error("Failed to start request:", err);
    }
  };

  const triggerProcessDeletion = (req: DPDPRequest) => {
    setSelectedDeleteReq(req);
    setShowConfirmDialog(true);
  };

  const handleConfirmAndStepUp = () => {
    setShowConfirmDialog(false);
    if (!selectedDeleteReq) return;

    openStepUpModal(async () => {
      try {
        const res = await fetch(`/v1/compliance/requests/${selectedDeleteReq.id}/process-deletion`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ step_up_verified: true }),
        });
        if (res.ok) {
          fetchRequests();
        }
      } catch (err) {
        console.error("Failed to process deletion:", err);
      }
    });
  };

  const getSLAStatusClass = (slaDueStr: string, status: string) => {
    if (status === "COMPLETED") return "text-slate-400";

    const due = new Date(slaDueStr).getTime();
    const now = Date.now();
    const diffDays = (due - now) / (1000 * 3600 * 24);

    if (diffDays < 0) return "text-rose-400 font-bold bg-rose-500/10 p-1 rounded";
    if (diffDays <= 7) return "text-amber-400 font-semibold bg-amber-500/10 p-1 rounded";
    return "text-slate-300";
  };

  const expectedServices = ["auth-service", "order-service", "loyalty-service", "notification-service"];

  return (
    <div className="space-y-6 p-6 max-w-7xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold text-slate-100">DPDP Act 2023 Statutory Request Queue</h1>
        <p className="text-sm text-slate-400">
          Statutory Data Subject RightsQueue — 30-Day SLA Enforcement
        </p>
      </div>

      <StepUpModalComponent />

      {/* ConfirmDialog Modal */}
      {showConfirmDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="w-full max-w-lg bg-slate-900 border-2 border-rose-500/60 rounded-xl p-6 shadow-2xl space-y-4">
            <div className="flex items-center space-x-3 text-rose-400">
              <svg className="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <h3 className="text-lg font-bold text-rose-300">Permanent Irreversible Deletion Warning</h3>
            </div>

            <p className="text-sm text-slate-200 leading-relaxed" id="confirm-dialog-body">
              This will permanently and irreversibly delete this user's data across all services. This cannot be undone.
            </p>

            <div className="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                onClick={() => setShowConfirmDialog(false)}
                className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                id="confirm-deletion-btn"
                onClick={handleConfirmAndStepUp}
                className="px-4 py-2 text-xs font-bold bg-rose-600 hover:bg-rose-500 text-white rounded-lg transition-colors"
              >
                Confirm & Require Step-Up
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Queue Table */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-slate-300">
            <thead className="bg-slate-800/60 text-slate-400 text-xs uppercase">
              <tr>
                <th className="p-3">User ID</th>
                <th className="p-3">Type</th>
                <th className="p-3">Status</th>
                <th className="p-3">SLA Due At</th>
                <th className="p-3">Fan-Out Progress</th>
                <th className="p-3 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {loading ? (
                <tr>
                  <td colSpan={6} className="p-4 text-center text-slate-500">Loading DPDP queue...</td>
                </tr>
              ) : requests.length === 0 ? (
                <tr>
                  <td colSpan={6} className="p-4 text-center text-slate-500">No active DPDP requests in queue.</td>
                </tr>
              ) : (
                requests.map((req) => (
                  <tr key={req.id} className="hover:bg-slate-800/40">
                    <td className="p-3 font-mono text-xs text-slate-200">{req.user_id}</td>
                    <td className="p-3 font-semibold text-xs text-amber-400">{req.request_type}</td>
                    <td className="p-3">
                      <span
                        className={`inline-flex px-2 py-0.5 rounded text-xs font-semibold ${
                          req.status === "COMPLETED"
                            ? "bg-emerald-500/20 text-emerald-400"
                            : req.status === "IN_PROGRESS"
                            ? "bg-amber-500/20 text-amber-400"
                            : "bg-blue-500/20 text-blue-400"
                        }`}
                      >
                        {req.status}
                      </span>
                    </td>
                    <td className="p-3 text-xs">
                      <span className={getSLAStatusClass(req.sla_due_at, req.status)}>
                        {new Date(req.sla_due_at).toLocaleDateString()}
                      </span>
                    </td>
                    <td className="p-3 text-xs">
                      {req.request_type === "DELETION" && req.status === "IN_PROGRESS" ? (
                        <div className="flex items-center space-x-1">
                          {expectedServices.map((svc) => (
                            <span
                              key={svc}
                              className="px-1.5 py-0.5 text-[10px] rounded font-mono bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                              title={`${svc}: completed`}
                            >
                              {svc.split("-")[0]} ✓
                            </span>
                          ))}
                        </div>
                      ) : (
                        "-"
                      )}
                    </td>
                    <td className="p-3 text-right space-x-2">
                      {req.status === "RECEIVED" && (
                        <button
                          onClick={() => handleStartRequest(req.id)}
                          className="px-2.5 py-1 bg-blue-600/20 hover:bg-blue-600/30 text-blue-300 border border-blue-500/30 text-xs rounded transition-colors"
                        >
                          Start
                        </button>
                      )}

                      {req.request_type === "DELETION" && req.status !== "COMPLETED" && (
                        <button
                          id={`process-deletion-btn-${req.id}`}
                          onClick={() => triggerProcessDeletion(req)}
                          className="px-2.5 py-1 bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold rounded transition-colors"
                        >
                          Process Deletion
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
    </div>
  );
}
