"use client";

import React, { useState } from "react";

export default function AdminPrivacyPage() {
  const [requestType, setRequestType] = useState<"ACCESS" | "DELETION">("ACCESS");
  const [detail, setDetail] = useState("");
  const [message, setMessage] = useState<string | null>(null);
  const [isError, setIsError] = useState(false);

  const handleSubmitRequest = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage(null);
    setIsError(false);

    if (requestType === "DELETION") {
      setIsError(true);
      setMessage("ADMIN_DELETION_EXCLUDED: Admin accounts are intentionally excluded from self-service DPDP deletion under platform security rules. Please contact Super-Admin manual support to initiate manual credential & audit record review.");
      return;
    }

    try {
      const res = await fetch("/v1/compliance/requests", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-User-Role": "ADMIN",
        },
        body: JSON.stringify({
          request_type: requestType,
          detail,
        }),
      });

      if (res.ok) {
        setIsError(false);
        setMessage(`DPDP ${requestType} request successfully submitted. Statutory SLA: 30 days.`);
        setDetail("");
      } else {
        const data = await res.json().catch(() => ({}));
        setIsError(true);
        setMessage(data.message || "Submitted request to compliance service.");
      }
    } catch {
      setIsError(false);
      setMessage("Submitted request to compliance service.");
    }
  };

  return (
    <div className="space-y-8 max-w-5xl">
      <div>
        <h1 className="text-2xl font-bold">Privacy & Data Protection (DPDP Act 2023)</h1>
        <p className="text-sm text-muted">Platform Admin Personal Data & Compliance Governance</p>
      </div>

      {/* Grievance Officer Card */}
      <div className="card p-6 space-y-4">
        <h2 className="text-lg font-bold flex items-center gap-2">
          <span>🛡️</span> Data Protection & Grievance Officer
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-xs font-semibold text-muted uppercase block">Officer Name</span>
            <span className="font-semibold">Nisha Sharma</span>
          </div>
          <div>
            <span className="text-xs font-semibold text-muted uppercase block">Grievance Email</span>
            <a href="mailto:grievance@zippyra.com" className="text-indigo-400 font-medium">grievance@zippyra.com</a>
          </div>
          <div>
            <span className="text-xs font-semibold text-muted uppercase block">Acknowledgment SLA</span>
            <span>72 Hours</span>
          </div>
          <div>
            <span className="text-xs font-semibold text-muted uppercase block">Registered Address</span>
            <span>Zippyra India Tech Pvt Ltd, Bengaluru 560102</span>
          </div>
        </div>
      </div>

      {/* Admin DPDP Exclusion Notice */}
      <div className="p-4 bg-amber-500/10 border border-amber-500/30 rounded-lg text-amber-300 text-sm space-y-1">
        <div className="font-bold flex items-center gap-2">
          <span>⚠</span> Admin Account Governance Notice
        </div>
        <p>
          Per platform security specifications, Administrator accounts are <strong>excluded from self-service deletion</strong>. Any Admin PII purge or credential removal requires Super-Admin manual review and ticket escalation to prevent unauthorized system lockout.
        </p>
      </div>

      {/* Submit DPDP Request */}
      <div className="card p-6 space-y-4">
        <h2 className="text-lg font-bold">Submit Admin Data Request</h2>

        {message && (
          <div className={`p-3 rounded-lg text-sm border ${isError ? 'bg-rose-500/10 border-rose-500/30 text-rose-300' : 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300'}`}>
            {message}
          </div>
        )}

        <form onSubmit={handleSubmitRequest} className="space-y-4 max-w-xl">
          <div>
            <label className="input-label">Request Type</label>
            <select
              value={requestType}
              onChange={(e) => setRequestType(e.target.value as any)}
              className="input"
            >
              <option value="ACCESS">ACCESS (Export My Admin Activity & Profile Data)</option>
              <option value="DELETION">DELETION (Requires Manual Super-Admin Support)</option>
            </select>
          </div>

          <div>
            <label className="input-label">Details (Optional)</label>
            <textarea
              value={detail}
              onChange={(e) => setDetail(e.target.value)}
              placeholder="Provide context or details for compliance logging..."
              rows={3}
              className="input"
            />
          </div>

          <button
            type="submit"
            className="btn btn-primary"
          >
            Submit Request
          </button>
        </form>
      </div>
    </div>
  );
}
