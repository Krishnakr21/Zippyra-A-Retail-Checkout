'use client';

import React, { useState, useEffect, useRef } from 'react';

export interface CredentialBundle {
  certPem?: string;
  privateKeyPem?: string;
  rootCaPem?: string;
  deviceJwt?: string;
}

export interface CredentialDownloadModalProps {
  bundle: CredentialBundle;
  onConfirmed: () => void;
}

export const CredentialDownloadModal: React.FC<CredentialDownloadModalProps> = ({
  bundle,
  onConfirmed,
}) => {
  const [downloaded, setDownloaded] = useState(false);
  const [confirmed, setConfirmed] = useState(false);
  const modalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    modalRef.current?.focus();
  }, []);

  const handleDownload = () => {
    const bundleText = `=== AWS IOT DEVICE CREDENTIAL BUNDLE ===
Generated on: ${new Date().toISOString()}

--- 1. DEVICE CERTIFICATE (device_cert.pem) ---
${bundle.certPem || ''}

--- 2. PRIVATE KEY (private_key.pem) ---
${bundle.privateKeyPem || ''}

--- 3. AMAZON ROOT CA (root_ca.pem) ---
${bundle.rootCaPem || ''}

--- 4. DEVICE JWT (1-Year TTL) ---
${bundle.deviceJwt || ''}
`;

    const blob = new Blob([bundleText], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `device_credentials_${Date.now()}.txt`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    setDownloaded(true);
  };

  return (
    <div
      tabIndex={-1}
      ref={modalRef}
      role="dialog"
      aria-modal="true"
      aria-labelledby="credential-modal-title"
      aria-describedby="credential-modal-desc"
      className="fixed inset-0 z-50 bg-slate-950/85 backdrop-blur-md flex items-center justify-center p-4 focus:outline-none"
    >
      <div className="bg-slate-900 border border-slate-700/80 w-full max-w-xl rounded-2xl p-6 shadow-2xl space-y-6 text-slate-200">
        <div className="flex items-start gap-4 bg-amber-500/10 border border-amber-500/30 p-4 rounded-xl text-amber-300">
          <div className="p-2 rounded-lg bg-amber-500/20 text-amber-400 shrink-0">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <h3 id="credential-modal-title" className="font-extrabold text-amber-200 text-lg">CRITICAL: One-Time Credential Access</h3>
            <p id="credential-modal-desc" className="text-xs text-amber-300/90 mt-1 leading-relaxed">
              These cryptographic credentials (X.509 Certificate, Private Key, Root CA, and 1-Year Device JWT) <strong>will never be shown again</strong>. Download and securely transfer them to the physical hardware device now.
            </p>
          </div>
        </div>

        <div className="space-y-3 text-xs">
          <div className="bg-slate-950 p-3 rounded-lg border border-slate-800 font-mono text-slate-400 overflow-x-auto max-h-28">
            <span className="text-slate-500 block mb-1">X.509 Certificate PEM</span>
            {bundle.certPem ? bundle.certPem.substring(0, 70) + '...' : '(Included in Bundle)'}
          </div>
          <div className="bg-slate-950 p-3 rounded-lg border border-slate-800 font-mono text-slate-400 overflow-x-auto">
            <span className="text-slate-500 block mb-1">RSA Private Key PEM</span>
            ********************* (Hidden for Security) *********************
          </div>
          <div className="bg-slate-950 p-3 rounded-lg border border-slate-800 font-mono text-slate-400 overflow-x-auto">
            <span className="text-slate-500 block mb-1">1-Year Device JWT</span>
            {bundle.deviceJwt ? bundle.deviceJwt.substring(0, 50) + '...' : '(Included in Bundle)'}
          </div>
        </div>

        <div className="pt-2 flex flex-col gap-4">
          <button
            type="button"
            onClick={handleDownload}
            aria-label="Download Credential Bundle"
            data-testid="download-bundle-btn"
            className="w-full py-3 px-4 rounded-xl font-bold text-white bg-indigo-600 hover:bg-indigo-500 transition-all shadow-lg shadow-indigo-600/30 flex items-center justify-center gap-2"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Download Credential Bundle
          </button>

          <label htmlFor="confirm-download-checkbox" className="flex items-center gap-3 text-xs text-slate-300 cursor-pointer bg-slate-800/50 p-3 rounded-xl border border-slate-700">
            <input
              type="checkbox"
              id="confirm-download-checkbox"
              data-testid="confirm-download-checkbox"
              checked={confirmed}
              onChange={(e) => setConfirmed(e.target.checked)}
              aria-label="I have downloaded and stored these credentials securely."
              className="rounded border-slate-700 bg-slate-900 text-indigo-600 focus:ring-indigo-500 w-4 h-4"
            />
            <span>I have downloaded and stored these credentials securely.</span>
          </label>
        </div>

        <div className="flex justify-end pt-2">
          <button
            type="button"
            disabled={!confirmed}
            data-testid="close-credential-modal-btn"
            onClick={onConfirmed}
            aria-label="Done & Dismiss"
            className="px-6 py-2.5 rounded-xl font-semibold text-xs text-white bg-emerald-600 hover:bg-emerald-500 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
          >
            Done & Dismiss
          </button>
        </div>
      </div>
    </div>
  );
};
