'use client';

import React, { useEffect, useState, useCallback } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';

interface StoreDetail {
  id: string;
  name: string;
  code: string;
  chain_id?: string;
  chain_name?: string;
  address: string;
  city: string;
  state: string;
  pincode: string;
  location?: { lat: number; lng: number };
  geofence_radius_meters: number;
  operating_hours?: { open: string; close: string };
  max_concurrent_orders: number;
  status: string;
  razorpay_account_id?: string;
  razorpay_kyc_status?: string;
  qr_token?: string;
}

interface HSNCheckResult {
  store_id: string;
  is_ready: boolean;
  total_products: number;
  unmapped_hsn_codes: string[];
}

const ADMIN_STORE_SERVICE_URL = process.env.NEXT_PUBLIC_ADMIN_STORE_SERVICE_URL || 'http://localhost:8091';
const CATALOG_SERVICE_URL = process.env.NEXT_PUBLIC_CATALOG_SERVICE_URL || 'http://localhost:8011';

export default function StoreDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [store, setStore] = useState<StoreDetail | null>(null);
  const [hsnCheck, setHsnCheck] = useState<HSNCheckResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [hsnLoading, setHsnLoading] = useState(false);
  const [error, setError] = useState('');
  const [qrMessage, setQrMessage] = useState('');

  // Form states for editable fields
  const [geofenceRadius, setGeofenceRadius] = useState<number>(500);
  const [openTime, setOpenTime] = useState('08:00');
  const [closeTime, setCloseTime] = useState('22:00');
  const [updatingHours, setUpdatingHours] = useState(false);
  const [updatingGeofence, setUpdatingGeofence] = useState(false);

  const fetchStoreDetail = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/stores/${id}`, {
        headers: { 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' }
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setStore(data);
      setGeofenceRadius(data.geofence_radius_meters || 500);
      if (data.operating_hours) {
        setOpenTime(data.operating_hours.open || '08:00');
        setCloseTime(data.operating_hours.close || '22:00');
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load store details');
      setStore(null);
    } finally {
      setLoading(false);
    }
  }, [id]);

  const checkHSNReadiness = useCallback(async () => {
    setHsnLoading(true);
    try {
      const res = await fetch(`${CATALOG_SERVICE_URL}/v1/catalog/admin/hsn-check?store_id=${id}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setHsnCheck(data);
    } catch {
      setHsnCheck(null);
    } finally {
      setHsnLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchStoreDetail();
    checkHSNReadiness();
  }, [fetchStoreDetail, checkHSNReadiness]);

  const handleRotateQR = async () => {
    setQrMessage('');
    try {
      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/stores/${id}/qr-tokens/rotate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' },
        body: JSON.stringify({ gate_ids: ['gate-001'] })
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setStore(prev => prev ? { ...prev, qr_token: data.qr_token } : null);
      setQrMessage('QR token successfully rotated!');
    } catch (err: any) {
      setQrMessage('Failed to rotate QR token: ' + (err.message || 'Server error'));
    }
  };

  const handleUpdateGeofence = async () => {
    setUpdatingGeofence(true);
    try {
      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/stores/${id}/geofence`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' },
        body: JSON.stringify({ radius_meters: Number(geofenceRadius) }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      alert('Geofence radius updated');
    } catch (err: any) {
      alert('Failed to update geofence: ' + (err.message || 'Server error'));
    } finally {
      setUpdatingGeofence(false);
    }
  };

  const handleUpdateHours = async () => {
    setUpdatingHours(true);
    try {
      const res = await fetch(`${ADMIN_STORE_SERVICE_URL}/v1/admin-store/stores/${id}/hours`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN' },
        body: JSON.stringify({ opening_time: openTime, closing_time: closeTime, timezone: 'Asia/Kolkata' }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      alert('Operating hours updated');
    } catch (err: any) {
      alert('Failed to update operating hours: ' + (err.message || 'Server error'));
    } finally {
      setUpdatingHours(false);
    }
  };

  if (loading) {
    return <div style={{ textAlign: 'center', padding: '5rem' }}><span className="spinner" style={{ margin: '0 auto' }} /></div>;
  }

  if (!store) {
    return <div>Store not found</div>;
  }

  return (
    <div>
      <div className="page-header">
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <Link href="/dashboard/stores" className="btn btn-secondary btn-sm">← Stores</Link>
            <h1>{store.name}</h1>
            <span className="badge badge-active">{store.status}</span>
          </div>
          <p style={{ marginTop: '0.25rem' }}>Code: <code style={{ color: '#a5b4fc' }}>{store.code}</code> | ID: {store.id}</p>
        </div>
        <button className="btn btn-danger" onClick={handleRotateQR}>
          🔄 Rotate QR Token
        </button>
      </div>

      {qrMessage && (
        <div className="dev-banner" style={{ background: 'rgba(34,197,94,0.15)', borderColor: 'rgba(34,197,94,0.3)', color: '#86efac', marginBottom: '1.5rem' }}>
          ✓ {qrMessage}
        </div>
      )}

      {error && (
        <div className="dev-banner" style={{ marginBottom: '1.5rem' }}>
          ⚠ store-service unavailable — showing offline mock details
        </div>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1.5rem' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>

          {/* Store Info & Geofence Card */}
          <div className="card">
            <h2 style={{ fontSize: '1.125rem', fontWeight: 700, marginBottom: '1rem' }}>Location & Geofence</h2>
            <p style={{ fontSize: '0.875rem', color: 'var(--color-text-muted)', marginBottom: '1rem' }}>
              Address: {store.address}, {store.city}, {store.state} - {store.pincode}
            </p>
            <div style={{ display: 'flex', alignItems: 'flex-end', gap: '1rem' }}>
              <div style={{ flex: 1 }}>
                <label className="input-label">Geofence Radius (Meters)</label>
                <input
                  type="number"
                  className="input"
                  value={geofenceRadius}
                  onChange={(e) => setGeofenceRadius(Number(e.target.value))}
                />
              </div>
              <button className="btn btn-primary" onClick={handleUpdateGeofence} disabled={updatingGeofence}>
                {updatingGeofence ? <span className="spinner" /> : 'Save Geofence'}
              </button>
            </div>
          </div>

          {/* Operating Hours Card */}
          <div className="card">
            <h2 style={{ fontSize: '1.125rem', fontWeight: 700, marginBottom: '1rem' }}>Operating Hours</h2>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr auto', gap: '1rem', alignItems: 'flex-end' }}>
              <div>
                <label className="input-label">Opening Time</label>
                <input type="time" className="input" value={openTime} onChange={(e) => setOpenTime(e.target.value)} />
              </div>
              <div>
                <label className="input-label">Closing Time</label>
                <input type="time" className="input" value={closeTime} onChange={(e) => setCloseTime(e.target.value)} />
              </div>
              <button className="btn btn-primary" onClick={handleUpdateHours} disabled={updatingHours}>
                {updatingHours ? <span className="spinner" /> : 'Save Hours'}
              </button>
            </div>
          </div>

          {/* HSN Readiness Check Card */}
          <div className="card">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
              <div>
                <h2 style={{ fontSize: '1.125rem', fontWeight: 700 }}>Catalog HSN Compliance Check</h2>
                <p style={{ fontSize: '0.8125rem', color: 'var(--color-text-muted)' }}>Validates HSN tax codes against standard GST rate map</p>
              </div>
              <button className="btn btn-secondary btn-sm" onClick={checkHSNReadiness} disabled={hsnLoading}>
                {hsnLoading ? <span className="spinner" /> : 'Re-Check'}
              </button>
            </div>

            {hsnCheck ? (
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1rem' }}>
                  <span className={`badge ${hsnCheck.is_ready ? 'badge-active' : 'badge-suspended'}`}>
                    {hsnCheck.is_ready ? 'HSN Compliant & Ready' : 'Unmapped HSN Codes Found'}
                  </span>
                  <span style={{ fontSize: '0.875rem', color: 'var(--color-text-muted)' }}>
                    Total Products: {hsnCheck.total_products}
                  </span>
                </div>

                {!hsnCheck.is_ready && (
                  <div style={{ background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.2)', padding: '1rem', borderRadius: '0.5rem' }}>
                    <p style={{ fontSize: '0.8125rem', color: '#fca5a5', fontWeight: 600, marginBottom: '0.5rem' }}>
                      The following HSN codes lack valid GST rate mappings:
                    </p>
                    <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                      {hsnCheck.unmapped_hsn_codes.map(code => (
                        <span key={code} style={{ background: 'rgba(239,68,68,0.2)', color: '#fca5a5', padding: '0.25rem 0.5rem', borderRadius: '0.25rem', fontSize: '0.75rem', fontFamily: 'monospace' }}>
                          {code}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div style={{ color: 'var(--color-text-muted)', fontSize: '0.875rem' }}>Loading HSN verification status...</div>
            )}
          </div>
        </div>

        {/* Sidebar Info Cards */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          <div className="card">
            <h3 style={{ fontSize: '0.875rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--color-text-muted)', marginBottom: '1rem' }}>
              QR Token & Terminal
            </h3>
            <div style={{ background: 'var(--color-surface)', padding: '0.875rem', borderRadius: '0.5rem', border: '1px solid var(--color-border)', wordBreak: 'break-all', fontFamily: 'monospace', fontSize: '0.75rem', color: '#a5b4fc', marginBottom: '0.75rem' }}>
              {store.qr_token || 'qr_tok_live_8f912a77b82c'}
            </div>
            <p style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
              Scan at store terminal to pair staff devices. Rotating invalidates active QR posters.
            </p>
          </div>

          <div className="card">
            <h3 style={{ fontSize: '0.875rem', textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--color-text-muted)', marginBottom: '1rem' }}>
              Payment Integration
            </h3>
            <div style={{ fontSize: '0.875rem', marginBottom: '0.5rem' }}>
              <span style={{ color: 'var(--color-text-muted)' }}>Razorpay Account ID:</span><br />
              <code style={{ fontSize: '0.8125rem', color: '#a5b4fc' }}>{store.razorpay_account_id || 'acc_mock_8812739'}</code>
            </div>
            <div style={{ fontSize: '0.875rem' }}>
              <span style={{ color: 'var(--color-text-muted)' }}>KYC Status:</span><br />
              <span className="badge badge-active" style={{ marginTop: '0.25rem' }}>{store.razorpay_kyc_status || 'VERIFIED'}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

const MOCK_STORES: StoreDetail[] = [
  {
    id: 's-001',
    name: 'Reliance Fresh — Koramangala',
    code: 'RF-BLR-001',
    address: '4th Block, Koramangala',
    city: 'Bengaluru',
    state: 'Karnataka',
    pincode: '560034',
    geofence_radius_meters: 500,
    operating_hours: { open: '08:00', close: '22:00' },
    max_concurrent_orders: 10,
    status: 'ACTIVE',
    chain_name: 'Reliance Retail',
    razorpay_account_id: 'acc_rel_001928',
    razorpay_kyc_status: 'VERIFIED',
    qr_token: 'qr_tok_rel_koramangala_v1',
  },
];
