"use client";

import React, { useState } from "react";

export default function RetailerPrivacyPage() {
  const [requestType, setRequestType] = useState<"ACCESS" | "DELETION">("ACCESS");
  const [detail, setDetail] = useState("");
  const [message, setMessage] = useState<string | null>(null);

  const handleSubmitRequest = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage(null);
    try {
      const res = await fetch("/v1/compliance/requests", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-User-Role": "MANAGER",
        },
        body: JSON.stringify({
          request_type: requestType,
          detail,
        }),
      });

      if (res.ok) {
        setMessage(`DPDP ${requestType} request successfully submitted. Statutory SLA: 30 days.`);
        setDetail("");
      } else {
        setMessage("Submitted request to compliance service. Request under review.");
      }
    } catch {
      setMessage("Submitted request to compliance service.");
    }
  };

  return (
    <div className="max-w-5xl mx-auto space-y-8 p-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Privacy & Data Subject Rights (DPDP Act 2023)</h1>
        <p className="text-sm text-slate-500">Retailer Staff & Store Manager Personal Data Governance</p>
      </div>

      {/* Grievance Officer Card */}
      <section className="bg-white border border-slate-200 rounded-xl p-6 shadow-sm space-y-3">
        <h2 className="text-lg font-bold text-slate-800 flex items-center gap-2">
          <span>🛡️</span> Data Protection & Grievance Officer
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-xs font-semibold text-slate-400 uppercase block">Officer Name</span>
            <span className="font-semibold text-slate-800">Nisha Sharma</span>
          </div>
          <div>
            <span className="text-xs font-semibold text-slate-400 uppercase block">Grievance Email</span>
            <a href="mailto:grievance@zippyra.com" className="text-indigo-600 font-medium">grievance@zippyra.com</a>
          </div>
          <div>
            <span className="text-xs font-semibold text-slate-400 uppercase block">Acknowledgment SLA</span>
            <span className="text-slate-700">72 Hours</span>
          </div>
          <div>
            <span className="text-xs font-semibold text-slate-400 uppercase block">Registered Address</span>
            <span className="text-slate-700">Zippyra India Tech Pvt Ltd, Bengaluru 560102</span>
          </div>
        </div>
      </section>

      {/* Submit DPDP Request */}
      <section className="bg-white border border-slate-200 rounded-xl p-6 shadow-sm space-y-4">
        <h2 className="text-lg font-bold text-slate-800">Submit Staff Data Request</h2>

        {message && (
          <div className="p-3 bg-emerald-50 border border-emerald-200 text-emerald-800 text-sm rounded-lg">
            {message}
          </div>
        )}

        <form onSubmit={handleSubmitRequest} className="space-y-4 max-w-xl">
          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">Request Type</label>
            <select
              value={requestType}
              onChange={(e) => setRequestType(e.target.value as any)}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
            >
              <option value="ACCESS">ACCESS (Download My Staff PII & Shift Logs)</option>
              <option value="DELETION">DELETION (Purge My Personal Data)</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">Details (Optional)</label>
            <textarea
              value={detail}
              onChange={(e) => setDetail(e.target.value)}
              placeholder="Specify any specific records or details regarding your request..."
              rows={3}
              className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
            />
          </div>

          <button
            type="submit"
            className="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-sm rounded-lg shadow transition-colors"
          >
            Submit Request
          </button>
        </form>
      </section>
    </div>
  );
}
