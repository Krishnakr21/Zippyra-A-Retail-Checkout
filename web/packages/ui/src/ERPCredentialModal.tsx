'use client';

import React, { useState } from 'react';

export interface ERPCredentialModalProps {
  connectionId?: string;
  integrationMode?: 'DIRECT' | 'AGENT_POLLED';
  webhookSecret: string;
  agentApiKey?: string | null;
  connectorSetupNote?: string;
  onConfirmed: () => void;
}

export const ERPCredentialModal: React.FC<ERPCredentialModalProps> = ({
  connectionId,
  integrationMode,
  webhookSecret,
  agentApiKey,
  connectorSetupNote,
  onConfirmed,
}) => {
  const [downloaded, setDownloaded] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  const handleDownload = () => {
    const text = `=== ZIPPYRA ERP INTEGRATION CREDENTIALS ===
Generated on: ${new Date().toISOString()}
Connection ID: ${connectionId || 'N/A'}
Integration Mode: ${integrationMode || 'N/A'}

--- 1. INBOUND WEBHOOK SECRET ---
${webhookSecret}

${
  agentApiKey
    ? `--- 2. AGENT API KEY (Bearer Token for Connector Agent) ---
${agentApiKey}
`
    : ''
}
${connectorSetupNote ? `--- SETUP INSTRUCTIONS ---\n${connectorSetupNote}\n` : ''}
`;

    const blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `erp_credentials_${connectionId || Date.now()}.txt`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    setDownloaded(true);
  };

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/85 backdrop-blur-md flex items-center justify-center p-4 overflow-y-auto">
      <div className="bg-slate-900 border border-slate-700/80 w-full max-w-xl rounded-2xl p-6 shadow-2xl space-y-6 text-slate-200 my-8">
        <div className="flex items-start gap-4 bg-amber-500/10 border border-amber-500/30 p-4 rounded-xl text-amber-300">
          <div className="p-2 rounded-lg bg-amber-500/20 text-amber-400 shrink-0">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <h3 className="font-extrabold text-amber-200 text-lg">CRITICAL: One-Time Secret Credentials</h3>
            <p className="text-xs text-amber-300/90 mt-1 leading-relaxed">
              These secret credentials (HMAC Webhook Secret {agentApiKey ? '& Agent API Key' : ''}) <strong>will never be shown again</strong>. Download and securely store them now.
            </p>
          </div>
        </div>

        <div className="space-y-3 text-xs">
          <div className="bg-slate-950 p-3 rounded-lg border border-slate-800 font-mono text-slate-400">
            <span className="text-slate-500 block mb-1">Inbound Webhook HMAC Secret</span>
            <span className="text-emerald-400 font-bold select-all">{webhookSecret}</span>
          </div>

          {agentApiKey && (
            <div className="bg-slate-950 p-3 rounded-lg border border-slate-800 font-mono text-slate-400">
              <span className="text-slate-500 block mb-1">Agent API Key (Bearer Token)</span>
              <span className="text-purple-400 font-bold select-all">{agentApiKey}</span>
            </div>
          )}

          {integrationMode === 'AGENT_POLLED' && (
            <div className="bg-purple-500/10 border border-purple-500/30 p-4 rounded-xl space-y-2">
              <h4 className="text-xs font-bold text-purple-300 flex items-center space-x-2">
                <span>Next Steps: On-Premise Connector Agent Setup</span>
              </h4>
              <p className="text-[11px] text-purple-200/80 leading-relaxed">
                Download the Zippyra Connector Agent binary on the store machine running Tally or Busy. Configure the agent with Connection ID <code className="font-mono text-amber-300">{connectionId}</code> and the Agent API Key above.
              </p>
              <div className="text-[10px] text-purple-400 font-semibold italic">
                (Installer Download: Connector binary v1.0 available via Zippyra HQ downloads)
              </div>
            </div>
          )}
        </div>

        <div className="pt-2 flex flex-col gap-4">
          <button
            type="button"
            onClick={handleDownload}
            data-testid="download-erp-credentials-btn"
            className="w-full py-3 px-4 rounded-xl font-bold text-white bg-purple-600 hover:bg-purple-500 transition-all shadow-lg shadow-purple-600/30 flex items-center justify-center gap-2"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Download Credentials TXT
          </button>

          <label className="flex items-center gap-3 text-xs text-slate-300 cursor-pointer bg-slate-800/50 p-3 rounded-xl border border-slate-700">
            <input
              type="checkbox"
              data-testid="confirm-download-checkbox"
              checked={confirmed}
              onChange={(e) => setConfirmed(e.target.checked)}
              className="rounded border-slate-700 bg-slate-900 text-purple-600 focus:ring-purple-500 w-4 h-4"
            />
            <span>I have saved these credentials in a secure credential store.</span>
          </label>
        </div>

        <div className="flex justify-end pt-2">
          <button
            type="button"
            disabled={!confirmed}
            data-testid="close-erp-credential-modal-btn"
            onClick={onConfirmed}
            className="px-6 py-2.5 rounded-xl font-semibold text-xs text-white bg-emerald-600 hover:bg-emerald-500 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
          >
            Done & Dismiss
          </button>
        </div>
      </div>
    </div>
  );
};
