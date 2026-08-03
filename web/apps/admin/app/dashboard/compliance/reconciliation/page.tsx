"use client";

import React, { useEffect, useState } from "react";

interface SettlementReport {
  id: string;
  report_date: string;
  total_transactions: number;
  total_amount_paise: number;
  total_settled_amount_paise: number;
  discrepancy_count: number;
  discrepancies: string;
  status: string;
  generated_at: string;
}

export default function SettlementReconciliationPage() {
  const [reports, setReports] = useState<SettlementReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const fetchReports = async () => {
    try {
      const res = await fetch("/v1/compliance/reconciliation-reports");
      if (res.ok) {
        const data = await res.json();
        setReports(data.reports || []);
      }
    } catch (err) {
      console.error("Failed to fetch settlement reports:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchReports();
  }, []);

  const parseDiscrepancies = (discrepanciesStr: string): any[] => {
    try {
      return JSON.parse(discrepanciesStr || "[]");
    } catch {
      return [];
    }
  };

  return (
    <div className="space-y-6 p-6 max-w-7xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold text-slate-100">Settlement Reconciliation Engine</h1>
        <p className="text-sm text-slate-400">
          Daily Razorpay Gateway vs Internal Captured Payments Audit
        </p>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-slate-300">
            <thead className="bg-slate-800/60 text-slate-400 text-xs uppercase">
              <tr>
                <th className="p-3">Report Date</th>
                <th className="p-3">Transactions</th>
                <th className="p-3">Internal Total</th>
                <th className="p-3">Settled Total</th>
                <th className="p-3">Discrepancies</th>
                <th className="p-3">Status</th>
                <th className="p-3 text-right">Details</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {loading ? (
                <tr>
                  <td colSpan={7} className="p-4 text-center text-slate-500">Loading settlement reports...</td>
                </tr>
              ) : reports.length === 0 ? (
                <tr>
                  <td colSpan={7} className="p-4 text-center text-slate-500">No reconciliation reports generated yet.</td>
                </tr>
              ) : (
                reports.map((r) => {
                  const isExpanded = expandedId === r.id;
                  const discrepanciesList = parseDiscrepancies(r.discrepancies);
                  const isIncomplete = r.status === "INCOMPLETE";

                  return (
                    <React.Fragment key={r.id}>
                      <tr className={`hover:bg-slate-800/40 ${isIncomplete ? "bg-rose-500/10" : ""}`}>
                        <td className="p-3 font-semibold text-slate-100">{r.report_date}</td>
                        <td className="p-3 text-xs">{r.total_transactions}</td>
                        <td className="p-3 text-xs font-mono">₹{(r.total_amount_paise / 100).toLocaleString("en-IN")}</td>
                        <td className="p-3 text-xs font-mono">₹{(r.total_settled_amount_paise / 100).toLocaleString("en-IN")}</td>
                        <td className="p-3">
                          <span
                            className={`inline-flex px-2 py-0.5 rounded text-xs font-bold ${
                              r.discrepancy_count > 0
                                ? "bg-rose-500/20 text-rose-400 border border-rose-500/30"
                                : "bg-emerald-500/20 text-emerald-400"
                            }`}
                          >
                            {r.discrepancy_count}
                          </span>
                        </td>
                        <td className="p-3">
                          {isIncomplete ? (
                            <span className="inline-flex items-center space-x-1 px-2.5 py-1 bg-rose-600 text-white font-bold text-xs rounded shadow">
                              <span>⚠️ INCOMPLETE</span>
                            </span>
                          ) : (
                            <span className="text-xs text-emerald-400 font-semibold">COMPLETED</span>
                          )}
                        </td>
                        <td className="p-3 text-right">
                          <button
                            onClick={() => setExpandedId(isExpanded ? null : r.id)}
                            className="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs rounded transition-colors"
                          >
                            {isExpanded ? "Hide Discrepancies" : "View Discrepancies"}
                          </button>
                        </td>
                      </tr>

                      {/* Expandable Discrepancy Row */}
                      {isExpanded && (
                        <tr className="bg-slate-950/80">
                          <td colSpan={7} className="p-4 border-l-4 border-amber-500">
                            {isIncomplete && (
                              <div className="mb-3 p-3 bg-rose-500/20 border border-rose-500/40 rounded text-xs text-rose-300 font-semibold">
                                Razorpay Gateway API was unreachable on this day. Manual audit required before marking report clean.
                              </div>
                            )}

                            <h4 className="text-xs font-bold text-slate-300 mb-2 uppercase">Discrepancy Breakdown</h4>
                            {discrepanciesList.length === 0 ? (
                              <p className="text-xs text-emerald-400">Zero discrepancies detected. 100% matched.</p>
                            ) : (
                              <div className="space-y-2">
                                {discrepanciesList.map((disc, idx) => (
                                  <div key={idx} className="p-2.5 bg-slate-900 border border-slate-800 rounded text-xs font-mono flex items-center justify-between text-slate-300">
                                    <span>Payment ID: {disc.payment_id}</span>
                                    <span>Expected: ₹{((disc.expected || 0) / 100).toFixed(2)}</span>
                                    <span>Settled: ₹{((disc.settled || 0) / 100).toFixed(2)}</span>
                                    <span className="text-rose-400 font-bold">{disc.reason || "MISMATCH"}</span>
                                  </div>
                                ))}
                              </div>
                            )}
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
