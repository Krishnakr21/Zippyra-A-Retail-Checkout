'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';

type WizardStep = 'basics' | 'location' | 'hours' | 'payment' | 'review';

const STEPS: { key: WizardStep; label: string }[] = [
  { key: 'basics', label: 'Basic Info' },
  { key: 'location', label: 'Location & Geofence' },
  { key: 'hours', label: 'Operating Hours' },
  { key: 'payment', label: 'Payment Setup' },
  { key: 'review', label: 'Review & Create' },
];

const ADMIN_STORE_SERVICE_URL = process.env.NEXT_PUBLIC_ADMIN_STORE_SERVICE_URL || 'http://localhost:8091';

interface StoreForm {
  name: string;
  code: string;
  chain_id: string;
  address: string;
  city: string;
  state: string;
  pincode: string;
  lat: string;
  lng: string;
  geofence_radius: string;
  opening_time: string;
  closing_time: string;
  max_concurrent_orders: string;
  razorpay_account_id: string;
}

const initialForm: StoreForm = {
  name: '', code: '', chain_id: '', address: '', city: '', state: '', pincode: '',
  lat: '', lng: '', geofence_radius: '500', opening_time: '08:00', closing_time: '22:00',
  max_concurrent_orders: '10', razorpay_account_id: '',
};

export default function NewStorePage() {
  const router = useRouter();
  const [step, setStep] = useState<WizardStep>('basics');
  const [form, setForm] = useState<StoreForm>(initialForm);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const stepIndex = STEPS.findIndex(s => s.key === step);
  const update = (field: keyof StoreForm) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
    setForm(prev => ({ ...prev, [field]: e.target.value }));

  const nextStep = () => {
    const next = STEPS[stepIndex + 1];
    if (next) setStep(next.key);
  };
  const prevStep = () => {
    const prev = STEPS[stepIndex - 1];
    if (prev) setStep(prev.key);
  };

  const handleSubmit = async () => {
    setSubmitting(true);
    setError('');
    try {
      const payload = {
        name: form.name,
        code: form.code,
        chain_id: form.chain_id || undefined,
        address: form.address,
        city: form.city,
        state: form.state,
        pincode: form.pincode,
        location: { lat: parseFloat(form.lat) || 0, lng: parseFloat(form.lng) || 0 },
        geofence_radius_meters: parseInt(form.geofence_radius) || 500,
        operating_hours: { open: form.opening_time, close: form.closing_time },
        max_concurrent_orders: parseInt(form.max_concurrent_orders) || 10,
        razorpay_account_id: form.razorpay_account_id || undefined,
      };
      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/stores`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `HTTP ${res.status}`);
      }
      router.push('/dashboard/stores');
    } catch (err: any) {
      setError(err.message || 'Failed to create store');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>Onboard New Store</h1>
          <p>Multi-step wizard to create and configure a store</p>
        </div>
      </div>

      {/* Wizard Steps */}
      <div className="wizard-steps">
        {STEPS.map((s, i) => (
          <div
            key={s.key}
            className={`wizard-step ${s.key === step ? 'active' : ''} ${i < stepIndex ? 'completed' : ''}`}
            onClick={() => i <= stepIndex && setStep(s.key)}
          >
            <span className="step-num">{i < stepIndex ? '✓' : i + 1}</span>
            {s.label}
          </div>
        ))}
      </div>

      {error && (
        <div style={{
          padding: '0.75rem 1rem', marginBottom: '1rem', borderRadius: '0.5rem',
          background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.3)',
          color: '#fca5a5', fontSize: '0.8125rem',
        }}>{error}</div>
      )}

      <div className="card" style={{ maxWidth: 680 }}>
        {/* Step: Basics */}
        {step === 'basics' && (
          <div>
            <h2 style={{ fontSize: '1.125rem', fontWeight: 700, marginBottom: '1.5rem' }}>Basic Information</h2>
            <div style={{ display: 'grid', gap: '1rem' }}>
              <div>
                <label className="input-label">Store Name *</label>
                <input className="input" value={form.name} onChange={update('name')} placeholder="e.g. Reliance Fresh — Koramangala" />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div>
                  <label className="input-label">Store Code *</label>
                  <input className="input" value={form.code} onChange={update('code')} placeholder="RF-BLR-001" />
                </div>
                <div>
                  <label className="input-label">Chain (optional)</label>
                  <select className="select" value={form.chain_id} onChange={update('chain_id')}>
                    <option value="">Standalone Store</option>
                    <option value="chain-001">Reliance Retail</option>
                    <option value="chain-002">DMart</option>
                  </select>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Step: Location */}
        {step === 'location' && (
          <div>
            <h2 style={{ fontSize: '1.125rem', fontWeight: 700, marginBottom: '1.5rem' }}>Location & Geofence</h2>
            <div style={{ display: 'grid', gap: '1rem' }}>
              <div>
                <label className="input-label">Full Address *</label>
                <input className="input" value={form.address} onChange={update('address')} placeholder="4th Block, Koramangala" />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem' }}>
                <div><label className="input-label">City *</label><input className="input" value={form.city} onChange={update('city')} /></div>
                <div><label className="input-label">State *</label><input className="input" value={form.state} onChange={update('state')} /></div>
                <div><label className="input-label">PIN Code *</label><input className="input" value={form.pincode} onChange={update('pincode')} maxLength={6} /></div>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem' }}>
                <div><label className="input-label">Latitude</label><input className="input" value={form.lat} onChange={update('lat')} placeholder="12.9352" /></div>
                <div><label className="input-label">Longitude</label><input className="input" value={form.lng} onChange={update('lng')} placeholder="77.6245" /></div>
                <div><label className="input-label">Geofence Radius (m)</label><input className="input" type="number" value={form.geofence_radius} onChange={update('geofence_radius')} /></div>
              </div>
              <div className="map-container">🗺 Map preview — requires Google Maps API key</div>
            </div>
          </div>
        )}

        {/* Step: Hours */}
        {step === 'hours' && (
          <div>
            <h2 style={{ fontSize: '1.125rem', fontWeight: 700, marginBottom: '1.5rem' }}>Operating Hours & Capacity</h2>
            <div style={{ display: 'grid', gap: '1rem' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div><label className="input-label">Opening Time</label><input className="input" type="time" value={form.opening_time} onChange={update('opening_time')} /></div>
                <div><label className="input-label">Closing Time</label><input className="input" type="time" value={form.closing_time} onChange={update('closing_time')} /></div>
              </div>
              <div>
                <label className="input-label">Max Concurrent Orders</label>
                <input className="input" type="number" value={form.max_concurrent_orders} onChange={update('max_concurrent_orders')} />
                <p style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)', marginTop: '0.25rem' }}>
                  New orders will be throttled once this limit is reached
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Step: Payment */}
        {step === 'payment' && (
          <div>
            <h2 style={{ fontSize: '1.125rem', fontWeight: 700, marginBottom: '1.5rem' }}>Payment Configuration</h2>
            <div style={{ display: 'grid', gap: '1rem' }}>
              <div>
                <label className="input-label">Razorpay Sub-Merchant Account ID</label>
                <input className="input" value={form.razorpay_account_id} onChange={update('razorpay_account_id')} placeholder="acc_XXXXXXXXXXXXXX" />
                <p style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)', marginTop: '0.25rem' }}>
                  Leave blank to configure later. KYC status will be tracked by store-service.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Step: Review */}
        {step === 'review' && (
          <div>
            <h2 style={{ fontSize: '1.125rem', fontWeight: 700, marginBottom: '1.5rem' }}>Review & Create</h2>
            <div className="json-viewer">
              {JSON.stringify({
                name: form.name, code: form.code, city: form.city, state: form.state,
                geofence_radius: `${form.geofence_radius}m`, hours: `${form.opening_time} – ${form.closing_time}`,
                max_orders: form.max_concurrent_orders, razorpay: form.razorpay_account_id || 'Not configured',
              }, null, 2)}
            </div>
          </div>
        )}

        {/* Navigation */}
        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '2rem', paddingTop: '1.5rem', borderTop: '1px solid var(--color-border)' }}>
          {stepIndex > 0 ? (
            <button className="btn btn-secondary" onClick={prevStep}>← Back</button>
          ) : <div />}
          {step === 'review' ? (
            <button className="btn btn-primary" onClick={handleSubmit} disabled={submitting}>
              {submitting ? <span className="spinner" /> : 'Create Store →'}
            </button>
          ) : (
            <button className="btn btn-primary" onClick={nextStep}>Continue →</button>
          )}
        </div>
      </div>
    </div>
  );
}
