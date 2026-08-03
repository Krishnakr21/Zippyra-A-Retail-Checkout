'use client';

import React, { useState } from 'react';

/**
 * ⚠ MOCK AUTH — admin-auth-service NOT YET BUILT
 *
 * This login page assumes admin-auth-service's contract (email/SSO login + mandatory
 * 2FA TOTP) which does NOT exist yet. It is wired against a LOCAL MOCK API route
 * (app/api/mock-admin-auth/route.ts) that accepts any email ending in the
 * ALLOWED_ADMIN_EMAIL_DOMAIN and any 6-digit TOTP code.
 *
 * DO NOT deploy to production until admin-auth-service is built.
 */

type AuthStep = 'email' | 'totp';

export default function LoginPage() {
  const [step, setStep] = useState<AuthStep>('email');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await fetch('/api/mock-admin-auth', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ step: 'email', email, password }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || 'Authentication failed');
      } else {
        setStep('totp');
      }
    } catch {
      setError('Network error');
    } finally {
      setLoading(false);
    }
  };

  const handleTotpSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await fetch('/api/mock-admin-auth', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ step: 'totp', email, totp_code: totpCode }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || 'TOTP verification failed');
      } else {
        window.location.href = '/dashboard';
      }
    } catch {
      setError('Network error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e1b4b 50%, #0f172a 100%)',
      padding: '1rem',
    }}>
      <div style={{ width: '100%', maxWidth: 420 }}>
        {/* Dev banner */}
        <div className="dev-banner" style={{ marginBottom: '1.5rem' }}>
          ⚠ Using mock auth — admin-auth-service not yet built
        </div>

        <div className="card" style={{ padding: '2.5rem' }}>
          <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
            <div style={{
              width: 56, height: 56, margin: '0 auto 1rem',
              background: 'linear-gradient(135deg, #6366f1, #a855f7)',
              borderRadius: 14, display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontWeight: 800, fontSize: 20, color: 'white',
            }}>Z</div>
            <h1 style={{
              fontSize: '1.5rem', fontWeight: 800,
              background: 'linear-gradient(135deg, #f1f5f9, #cbd5e1)',
              WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
            }}>Admin Console</h1>
            <p style={{ fontSize: '0.875rem', color: 'var(--color-text-muted)', marginTop: '0.25rem' }}>
              Zippyra Internal Operations
            </p>
          </div>

          {error && (
            <div style={{
              padding: '0.75rem 1rem', marginBottom: '1rem', borderRadius: '0.5rem',
              background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.3)',
              color: '#fca5a5', fontSize: '0.8125rem',
            }}>{error}</div>
          )}

          {step === 'email' ? (
            <form onSubmit={handleEmailSubmit}>
              <div style={{ marginBottom: '1rem' }}>
                <label className="input-label">Work Email</label>
                <input
                  className="input"
                  type="email"
                  placeholder="you@zippyra.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  autoFocus
                />
              </div>
              <div style={{ marginBottom: '1.5rem' }}>
                <label className="input-label">Password</label>
                <input
                  className="input"
                  type="password"
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              <button className="btn btn-primary" type="submit" disabled={loading}
                style={{ width: '100%', justifyContent: 'center', padding: '0.75rem' }}>
                {loading ? <span className="spinner" /> : 'Continue to 2FA'}
              </button>
            </form>
          ) : (
            <form onSubmit={handleTotpSubmit}>
              <p style={{ fontSize: '0.875rem', color: 'var(--color-text-muted)', marginBottom: '1rem' }}>
                Enter the 6-digit code from your authenticator app.
              </p>
              <div style={{ marginBottom: '1.5rem' }}>
                <label className="input-label">TOTP Code</label>
                <input
                  className="input"
                  type="text"
                  placeholder="000000"
                  maxLength={6}
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  required
                  autoFocus
                  style={{ letterSpacing: '0.5em', textAlign: 'center', fontSize: '1.5rem', fontWeight: 700 }}
                />
              </div>
              <button className="btn btn-primary" type="submit" disabled={loading}
                style={{ width: '100%', justifyContent: 'center', padding: '0.75rem' }}>
                {loading ? <span className="spinner" /> : 'Verify & Sign In'}
              </button>
              <button type="button" className="btn btn-secondary" onClick={() => setStep('email')}
                style={{ width: '100%', justifyContent: 'center', marginTop: '0.75rem' }}>
                ← Back to email
              </button>
            </form>
          )}
          <div style={{ marginTop: '1.5rem', textAlign: 'center', fontSize: '0.75rem', color: '#94a3b8' }}>
            Lost access to your phone or authenticator? <br />
            Contact your Lead Platform Administrator to request a credential reset.
          </div>
        </div>
      </div>
    </div>
  );
}
