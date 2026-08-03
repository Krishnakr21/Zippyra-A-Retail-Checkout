'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function LoginPage() {
  const router = useRouter();
  const [authMode, setAuthMode] = useState<'OTP' | 'PIN'>('OTP');

  // OTP State
  const [otpStep, setOtpStep] = useState<'SEND' | 'VERIFY'>('SEND');
  const [channel, setChannel] = useState<'phone' | 'email'>('phone');
  const [identifier, setIdentifier] = useState('');
  const [otp, setOtp] = useState('');

  // PIN State
  const [storeId, setStoreId] = useState('store-001');
  const [pin, setPin] = useState('');

  // Status & Errors
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const apiBase = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8088';

  const handleSendOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!identifier.trim()) {
      setError('Phone number or email is required');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/v1/retailer-auth/otp/send`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ channel, identifier: identifier.trim() }),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.message || 'Failed to send OTP');
      }

      setOtpStep('VERIFY');
      setSuccessMsg('OTP sent successfully. Please enter code.');
    } catch (err: any) {
      setError(err?.message || 'Failed to send OTP');
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!otp.trim()) {
      setError('OTP code is required');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/v1/retailer-auth/otp/verify`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ channel, identifier: identifier.trim(), otp: otp.trim() }),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.message || 'Invalid or expired OTP code');
      }

      const data = await res.json();
      localStorage.setItem('token', data.access_token || 'mock-retailer-token');
      localStorage.setItem('role', data.role || 'STORE_MANAGER');
      localStorage.setItem('store_id', data.store_id || 'store-001');
      document.cookie = `token=${data.access_token || 'mock-retailer-token'}; path=/; max-age=86400; SameSite=Lax`;

      router.push('/dashboard');
    } catch (err: any) {
      setError(err?.message || 'Verification failed');
    } finally {
      setLoading(false);
    }
  };

  const handlePinLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!storeId.trim() || !pin.trim()) {
      setError('Store ID and PIN are required');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/v1/retailer-auth/pin/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ store_id: storeId.trim(), pin: pin.trim() }),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.message || 'Invalid Store ID or Staff PIN');
      }

      const data = await res.json();
      localStorage.setItem('token', data.access_token || 'mock-staff-pin-token');
      localStorage.setItem('role', data.role || 'STORE_STAFF');
      localStorage.setItem('store_id', data.store_id || storeId.trim());
      document.cookie = `token=${data.access_token || 'mock-staff-pin-token'}; path=/; max-age=86400; SameSite=Lax`;

      router.push('/dashboard');
    } catch (err: any) {
      setError(err?.message || 'PIN login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col justify-center py-12 sm:px-6 lg:px-8 text-slate-100">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center">
        <div className="mx-auto w-12 h-12 rounded-xl bg-indigo-600/20 border border-indigo-500/40 flex items-center justify-center text-indigo-400 font-extrabold text-2xl shadow-lg shadow-indigo-500/10">
          Z
        </div>
        <h2 className="mt-4 text-3xl font-extrabold tracking-tight text-white">Zippyra Retailer Dashboard</h2>
        <p className="mt-2 text-sm text-slate-400">Store Manager & Staff Operations Console</p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-slate-900 border border-slate-800 py-8 px-4 shadow-2xl sm:rounded-2xl sm:px-10">
          {/* Mode Selector */}
          <div className="grid grid-cols-2 gap-2 bg-slate-950 p-1 rounded-xl mb-6 border border-slate-800">
            <button
              type="button"
              onClick={() => {
                setAuthMode('OTP');
                setError(null);
              }}
              data-testid="mode-otp-btn"
              className={`py-2 text-xs font-bold rounded-lg transition-all ${
                authMode === 'OTP' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-white'
              }`}
            >
              Phone / Email OTP
            </button>
            <button
              type="button"
              onClick={() => {
                setAuthMode('PIN');
                setError(null);
              }}
              data-testid="mode-pin-btn"
              className={`py-2 text-xs font-bold rounded-lg transition-all ${
                authMode === 'PIN' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-white'
              }`}
            >
              Quick Staff PIN
            </button>
          </div>

          {error && (
            <div data-testid="login-error-banner" className="mb-4 p-3 bg-rose-500/10 border border-rose-500/30 rounded-xl text-rose-400 text-xs">
              {error}
            </div>
          )}

          {successMsg && (
            <div className="mb-4 p-3 bg-emerald-500/10 border border-emerald-500/30 rounded-xl text-emerald-400 text-xs">
              {successMsg}
            </div>
          )}

          {authMode === 'OTP' ? (
            otpStep === 'SEND' ? (
              <form onSubmit={handleSendOTP} className="space-y-4">
                <div>
                  <label htmlFor="channel-select" className="block text-xs font-semibold text-slate-400 mb-1">
                    Channel
                  </label>
                  <select
                    id="channel-select"
                    value={channel}
                    onChange={(e) => setChannel(e.target.value as any)}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  >
                    <option value="phone">Phone Number (+91)</option>
                    <option value="email">Email Address</option>
                  </select>
                </div>

                <div>
                  <label htmlFor="identifier-input" className="block text-xs font-semibold text-slate-400 mb-1">
                    {channel === 'phone' ? 'Phone Number' : 'Email Address'}
                  </label>
                  <input
                    type="text"
                    id="identifier-input"
                    data-testid="identifier-input"
                    value={identifier}
                    onChange={(e) => setIdentifier(e.target.value)}
                    placeholder={channel === 'phone' ? '+919876543210' : 'manager@store.com'}
                    aria-label="Identifier"
                    className="w-full px-3 py-2.5 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    autoFocus
                  />
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  data-testid="send-otp-btn"
                  className="w-full py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-sm rounded-xl transition-all shadow-lg shadow-indigo-600/30 disabled:opacity-50"
                >
                  {loading ? 'Sending OTP...' : 'Send Login OTP'}
                </button>
              </form>
            ) : (
              <form onSubmit={handleVerifyOTP} className="space-y-4">
                <div>
                  <label htmlFor="otp-input" className="block text-xs font-semibold text-slate-400 mb-1">
                    6-Digit OTP Code
                  </label>
                  <input
                    type="text"
                    id="otp-input"
                    data-testid="otp-input"
                    value={otp}
                    onChange={(e) => setOtp(e.target.value)}
                    placeholder="123456"
                    aria-label="OTP Code"
                    className="w-full px-3 py-2.5 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm font-mono text-center tracking-widest text-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    autoFocus
                  />
                </div>

                <div className="flex items-center justify-between text-xs">
                  <button
                    type="button"
                    onClick={() => setOtpStep('SEND')}
                    className="text-slate-400 hover:text-white"
                  >
                    ← Change Phone/Email
                  </button>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  data-testid="verify-otp-btn"
                  className="w-full py-3 bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-sm rounded-xl transition-all shadow-lg shadow-emerald-600/30 disabled:opacity-50"
                >
                  {loading ? 'Verifying...' : 'Verify & Enter Dashboard'}
                </button>
              </form>
            )
          ) : (
            <form onSubmit={handlePinLogin} className="space-y-4">
              <div>
                <label htmlFor="store-id-input" className="block text-xs font-semibold text-slate-400 mb-1">
                  Store Location ID
                </label>
                <input
                  type="text"
                  id="store-id-input"
                  data-testid="store-id-input"
                  value={storeId}
                  onChange={(e) => setStoreId(e.target.value)}
                  placeholder="store-001"
                  aria-label="Store Location ID"
                  className="w-full px-3 py-2.5 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>

              <div>
                <label htmlFor="pin-input" className="block text-xs font-semibold text-slate-400 mb-1">
                  6-Digit Staff Quick PIN
                </label>
                <input
                  type="password"
                  id="pin-input"
                  data-testid="pin-input"
                  value={pin}
                  onChange={(e) => setPin(e.target.value)}
                  placeholder="******"
                  aria-label="6-Digit Staff Quick PIN"
                  className="w-full px-3 py-2.5 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm font-mono tracking-widest focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  autoFocus
                />
              </div>

              <button
                type="submit"
                disabled={loading}
                data-testid="pin-login-btn"
                className="w-full py-3 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-sm rounded-xl transition-all shadow-lg shadow-indigo-600/30 disabled:opacity-50"
              >
                {loading ? 'Authenticating...' : 'Sign In with Quick PIN'}
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
