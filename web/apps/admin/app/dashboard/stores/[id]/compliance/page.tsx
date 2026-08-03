"use client";

import React, { useEffect, useState } from "react";
import { useParams } from "next/navigation";

interface IRNRecord {
  id: string;
  order_id: string;
  store_id: string;
  status: string;
  irn?: string;
  ack_no?: string;
  failure_reason?: string;
  retry_count: number;
  created_at: string;
}

interface MerchantKYC {
  store_id: string;
  gstin?: string;
  gstin_verified: boolean;
  pan_number?: string;
  pan_verified: boolean;
  bank_account_last4?: string;
  razorpay_account_id?: string;
  kyc_status: string;
}

interface VelocityAlert {
  id: string;
  store_id: string;
  alert_type: string;
  detail: string;
  created_at: string;
}

export default function StoreCompliancePage() {
  const params = useParams();
  const storeId = (params?.id as string) || "store-1";

  const [irnRecords, setIrnRecords] = useState<IRNRecord[]>([]);
  const [kyc, setKyc] = useState<MerchantKYC>({
    store_id: storeId,
    gstin_verified: false,
    pan_verified: false,
    kyc_status: "PENDING",
  });
  const [velocityAlerts, setVelocityAlerts] = useState<VelocityAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [savingKyc, setSavingKyc] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  const fetchComplianceData = async () => {
    setLoading(true);
    try {
      // 1. Fetch IRN records
      const irnRes = await fetch(`/v1/compliance/irn?store_id=${storeId}`);
      if (irnRes.ok) {
        const data = await irnRes.json();
        setIrnRecords(data.records || []);
      }

      // 2. Fetch KYC
      const kycRes = await fetch(`/v1/compliance/kyc?store_id=${storeId}`);
      if (kycRes.ok) {
        const data = await kycRes.json();
        setKyc(data);
      }

      // 3. Fetch Velocity Alerts
      const velocityRes = await fetch(`/v1/compliance/velocity-alerts?store_id=${storeId}&resolved=false`);
      if (velocityRes.ok) {
        const data = await velocityRes.json();
        setVelocityAlerts(data.alerts || []);
      }
    } catch (err) {
      console.error("Failed to load compliance data:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchComplianceData();
  }, [storeId]);

  const handleRetryIRN = async (orderId: string) => {
    try {
      const res = await fetch(`/v1/compliance/irn/${orderId}/retry`, { method: "POST" });
      if (res.ok) {
        setMessage(`Retry triggered for order ${orderId}`);
        fetchComplianceData();
      }
    } catch (err) {
      console.error("Failed to retry IRN:", err);
    }
  };

  const handleSaveKYC = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingKyc(true);
    setMessage(null);
    try {
      const res = await fetch(`/v1/compliance/kyc/${storeId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(kyc),
      });
      if (res.ok) {
        const updated = await res.json();
        setKyc(updated);
        setMessage("Merchant KYC updated successfully!");
      }
    } catch (err) {
      console.error("Failed to save KYC:", err);
    } finally {
      setSavingKyc(false);
    }
  };

  const handleResolveAlert = async (alertId: string) => {
    try {
      const res = await fetch(`/v1/compliance/velocity-alerts/${alertId}/resolve`, { method: "POST" });
      if (res.ok) {
        setVelocityAlerts((prev) => prev.filter((a) => a.id !== alertId));
      }
    } catch (err) {
      console.error("Failed to resolve alert:", err);
    }
  };

  const failedCeilingRecords = irnRecords.filter((r) => r.status === "FAILED" && r.retry_count >= 3);

  return (
    <div className="space-y-8 p-6 max-w-7xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold text-slate-100">Store Statutory Compliance & Operations</h1>
        <p className="text-sm text-slate-400">Store ID: {storeId}</p>
      </div>

      {message && (
        <div className="p-3 bg-emerald-500/10 border border-emerald-500/30 rounded-lg text-emerald-400 text-sm">
          {message}
        </div>
      )}

      {/* KYC Mandatory Payment Warning Banner */}
      {kyc.kyc_status !== "VERIFIED" && (
        <div id="kyc-unverified-banner" className="p-4 bg-rose-500/15 border-2 border-rose-500/60 rounded-xl flex items-center space-x-3 text-rose-200 shadow-lg">
          <svg className="w-6 h-6 text-rose-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div>
            <h4 className="font-bold text-rose-300">Statutory Payment Processing Blocked</h4>
            <p className="text-sm text-rose-200">
              This store cannot legally process real payments until KYC is verified.
            </p>
          </div>
        </div>
      )}

      {/* 1. Merchant KYC Form */}
      <section className="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-4">
        <h2 className="text-lg font-semibold text-slate-100 flex items-center space-x-2">
          <span>Merchant KYC & Statutory Credentials</span>
        </h2>

        <form onSubmit={handleSaveKYC} className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">PAN Number</label>
            <input
              type="text"
              id="pan-number-input"
              value={kyc.pan_number || ""}
              onChange={(e) => setKyc({ ...kyc, pan_number: e.target.value })}
              placeholder="ABCDE1234F"
              className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-100 text-sm focus:outline-none focus:border-amber-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">Bank Account Last 4</label>
            <input
              type="text"
              id="bank-account-input"
              value={kyc.bank_account_last4 || ""}
              onChange={(e) => setKyc({ ...kyc, bank_account_last4: e.target.value })}
              placeholder="9876"
              className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-100 text-sm focus:outline-none focus:border-amber-500"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-slate-400 mb-1">KYC Status</label>
            <select
              id="kyc-status-select"
              value={kyc.kyc_status}
              onChange={(e) => setKyc({ ...kyc, kyc_status: e.target.value })}
              className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-100 text-sm focus:outline-none focus:border-amber-500"
            >
              <option value="PENDING">PENDING</option>
              <option value="VERIFIED">VERIFIED</option>
              <option value="REJECTED">REJECTED</option>
            </select>
          </div>

          <div className="flex items-center space-x-6 pt-6">
            <label className="flex items-center space-x-2 text-sm text-slate-300 cursor-pointer">
              <input
                type="checkbox"
                id="gstin-verified-checkbox"
                checked={kyc.gstin_verified}
                onChange={(e) => setKyc({ ...kyc, gstin_verified: e.target.checked })}
                className="rounded border-slate-700 bg-slate-800 text-amber-500 focus:ring-amber-500"
              />
              <span>GSTIN Verified</span>
            </label>

            <label className="flex items-center space-x-2 text-sm text-slate-300 cursor-pointer">
              <input
                type="checkbox"
                id="pan-verified-checkbox"
                checked={kyc.pan_verified}
                onChange={(e) => setKyc({ ...kyc, pan_verified: e.target.checked })}
                className="rounded border-slate-700 bg-slate-800 text-amber-500 focus:ring-amber-500"
              />
              <span>PAN Verified</span>
            </label>
          </div>

          <div className="md:col-span-2 flex justify-end">
            <button
              type="submit"
              id="save-kyc-btn"
              disabled={savingKyc}
              className="px-4 py-2 bg-amber-500 hover:bg-amber-400 text-slate-950 font-semibold text-sm rounded-lg transition-colors disabled:opacity-50"
            >
              {savingKyc ? "Saving..." : "Update Merchant KYC"}
            </button>
          </div>
        </form>
      </section>

      {/* 2. GST E-Invoice IRN Section */}
      <section className="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-4">
        <h2 className="text-lg font-semibold text-slate-100">GST E-Invoice (IRN) Submissions</h2>

        {/* Failed Ceiling Callout */}
        {failedCeilingRecords.length > 0 && (
          <div className="p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl space-y-2">
            <h4 className="font-bold text-amber-400 text-sm">
              {failedCeilingRecords.length} Invoice(s) Exceeded Automatic Retry Ceiling
            </h4>
            <p className="text-xs text-amber-200/80">
              These orders failed automatic background IRP portal submission 3 times. Manual retry can be triggered below.
            </p>
          </div>
        )}

        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-slate-300">
            <thead className="bg-slate-800/60 text-slate-400 text-xs uppercase">
              <tr>
                <th className="p-3">Order ID</th>
                <th className="p-3">Status</th>
                <th className="p-3">IRN</th>
                <th className="p-3">Retries</th>
                <th className="p-3 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {irnRecords.length === 0 ? (
                <tr>
                  <td colSpan={5} className="p-4 text-center text-slate-500">No IRN records found for this store.</td>
                </tr>
              ) : (
                irnRecords.map((rec) => (
                  <tr key={rec.id} className="hover:bg-slate-800/40">
                    <td className="p-3 font-mono text-xs text-slate-200">{rec.order_id}</td>
                    <td className="p-3">
                      <span
                        className={`inline-flex px-2 py-0.5 rounded text-xs font-semibold ${
                          rec.status === "ISSUED"
                            ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                            : rec.status === "FAILED"
                            ? "bg-rose-500/20 text-rose-400 border border-rose-500/40"
                            : "bg-amber-500/20 text-amber-400 border border-amber-500/40"
                        }`}
                      >
                        {rec.status}
                      </span>
                    </td>
                    <td className="p-3 font-mono text-xs truncate max-w-xs text-slate-400">
                      {rec.irn || rec.failure_reason || "-"}
                    </td>
                    <td className="p-3 text-xs">{rec.retry_count} / 3</td>
                    <td className="p-3 text-right">
                      {rec.status === "FAILED" && (
                        <button
                          onClick={() => handleRetryIRN(rec.order_id)}
                          className="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs rounded transition-colors"
                        >
                          Retry Now
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </section>

      {/* 3. Velocity Alerts Section */}
      <section className="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-4">
        <h2 className="text-lg font-semibold text-slate-100">Unresolved Velocity Alerts</h2>

        <div className="space-y-3">
          {velocityAlerts.length === 0 ? (
            <p className="text-sm text-slate-500">No unresolved velocity alerts for this store.</p>
          ) : (
            velocityAlerts.map((alert) => (
              <div key={alert.id} className="p-4 bg-slate-800/60 border border-slate-700/60 rounded-lg flex items-center justify-between">
                <div>
                  <span className="inline-block px-2 py-0.5 text-xs font-bold text-amber-400 bg-amber-500/10 border border-amber-500/30 rounded mb-1">
                    {alert.alert_type}
                  </span>
                  <p className="text-xs font-mono text-slate-300">{alert.detail}</p>
                </div>

                <button
                  onClick={() => handleResolveAlert(alert.id)}
                  className="px-3 py-1 bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-400 border border-emerald-500/30 rounded text-xs transition-colors"
                >
                  Resolve Alert
                </button>
              </div>
            ))
          )}
        </div>
      </section>
    </div>
  );
}
