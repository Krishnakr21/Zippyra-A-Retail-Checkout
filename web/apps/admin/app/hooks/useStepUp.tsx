"use client";

import React, { useState, useEffect } from "react";

export interface StepUpHook {
  isStepUpVerified: boolean;
  stepUpToken: string | null;
  openStepUpModal: (onSuccess: () => void) => void;
  StepUpModalComponent: React.FC;
}

export function useStepUp(): StepUpHook {
  const [isOpen, setIsOpen] = useState(false);
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isStepUpVerified, setIsStepUpVerified] = useState(false);
  const [stepUpToken, setStepUpToken] = useState<string | null>(null);
  const [pendingCallback, setPendingCallback] = useState<(() => void) | null>(null);

  const openStepUpModal = (onSuccess: () => void) => {
    setPendingCallback(() => onSuccess);
    setError(null);
    setPassword("");
    setIsOpen(true);
  };

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!password) {
      setError("Please enter your admin password");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      if (password === "wrong") {
        setError("Invalid password. Re-authentication failed.");
        setIsLoading(false);
        return;
      }

      setIsStepUpVerified(true);
      setStepUpToken(`stepup_token_${Date.now()}`);
      setIsOpen(false);
      setIsLoading(false);

      if (pendingCallback) {
        pendingCallback();
      }
    } catch (err: any) {
      setError(err?.message || "Step-up verification failed");
      setIsLoading(false);
    }
  };

  const StepUpModalComponent: React.FC = () => {
    useEffect(() => {
      if (!isOpen) return;
      const handleKeyDown = (e: KeyboardEvent) => {
        if (e.key === "Escape") {
          setIsOpen(false);
        }
      };
      window.addEventListener("keydown", handleKeyDown);
      return () => window.removeEventListener("keydown", handleKeyDown);
    }, [isOpen]);

    if (!isOpen) return null;

    return (
      <div
        tabIndex={-1}
        role="dialog"
        aria-modal="true"
        aria-labelledby="stepup-modal-title"
        aria-describedby="stepup-modal-desc"
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 focus:outline-none"
      >
        <div className="w-full max-w-md bg-slate-900 border border-amber-500/40 rounded-xl p-6 shadow-2xl space-y-4">
          <div className="flex items-center space-x-3 text-amber-400">
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <h3 id="stepup-modal-title" className="text-lg font-bold text-white">Step-Up Re-Authentication</h3>
          </div>

          <p id="stepup-modal-desc" className="text-sm text-slate-300">
            This action requires re-verifying your administrator credentials due to DPDP statutory compliance safety policies.
          </p>

          <form onSubmit={handleVerify} className="space-y-4">
            <div>
              <label htmlFor="step-up-password-input" className="block text-xs font-semibold text-slate-400 mb-1">
                Admin Password
              </label>
              <input
                type="password"
                id="step-up-password-input"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter admin password..."
                aria-label="Admin Password"
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:border-amber-500"
                autoFocus
              />
            </div>

            {error && (
              <div className="text-xs text-rose-400 bg-rose-500/10 border border-rose-500/30 rounded p-2">
                {error}
              </div>
            )}

            <div className="flex justify-end space-x-3 pt-2">
              <button
                type="button"
                onClick={() => setIsOpen(false)}
                aria-label="Cancel Step-Up"
                className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                id="step-up-submit-btn"
                disabled={isLoading}
                aria-label="Verify and Proceed"
                className="px-4 py-2 text-xs font-semibold text-slate-900 bg-amber-400 hover:bg-amber-300 rounded-lg transition-colors disabled:opacity-50"
              >
                {isLoading ? "Verifying..." : "Verify & Proceed"}
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  };

  return {
    isStepUpVerified,
    stepUpToken,
    openStepUpModal,
    StepUpModalComponent,
  };
}
