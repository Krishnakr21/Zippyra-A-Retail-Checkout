'use client';

import React, { useState, useEffect, useCallback } from 'react';

interface ProductItem {
  id: string;
  store_id: string;
  chain_id: string;
  barcode: string;
  name: string;
  description?: string;
  price_paise: number;
  mrp_paise: number;
  hsn_code: string;
  is_active: boolean;
  image_url?: string;
  thumbnail_url?: string;
  image_processing_status?: 'PENDING' | 'PROCESSED' | 'FAILED' | string;
}

const CATALOG_SERVICE_URL = process.env.NEXT_PUBLIC_CATALOG_SERVICE_URL || 'http://localhost:8083';

export default function AdminProductsPage() {
  const [products, setProducts] = useState<ProductItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [storeId, setStoreId] = useState('store-001');

  // New Product Modal State
  const [showModal, setShowModal] = useState(false);
  const [name, setName] = useState('');
  const [barcode, setBarcode] = useState('8901030000018');
  const [price, setPrice] = useState('150');
  const [mrp, setMrp] = useState('200');
  const [hsnCode, setHsnCode] = useState('1001');
  const [imageUrl, setImageUrl] = useState('raw/product_image_001.png');
  const [creating, setCreating] = useState(false);

  const fetchProducts = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await fetch(`${CATALOG_SERVICE_URL}/v1/catalog/admin/products?store_id=${storeId}&page=1&page_size=50`, {
        headers: { 'X-User-Type': 'ADMIN', 'X-User-Role': 'ADMIN', 'X-Chain-ID': 'chain-001' }
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setProducts(data.products || []);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch catalog products');
      setProducts([]);
    } finally {
      setLoading(false);
    }
  }, [storeId]);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  const handleCreateProduct = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !barcode || !price) return;
    setCreating(true);
    try {
      const pricePaise = Math.round(parseFloat(price) * 100);
      const mrpPaise = Math.round(parseFloat(mrp) * 100);

      const res = await fetch(`${CATALOG_SERVICE_URL}/v1/catalog/admin/products`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-Type': 'ADMIN',
          'X-User-Role': 'ADMIN',
          'X-Chain-ID': 'chain-001'
        },
        body: JSON.stringify({
          store_id: storeId,
          chain_id: 'chain-001',
          name,
          barcode,
          price_paise: pricePaise,
          mrp_paise: mrpPaise,
          hsn_code: hsnCode,
          image_url: imageUrl,
        }),
      });

      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      setShowModal(false);
      setName('');
      setImageUrl('raw/product_image_' + Math.floor(Math.random() * 1000) + '.png');
      fetchProducts();
    } catch (err: any) {
      alert(err.message || 'Failed to create product');
    } finally {
      setCreating(false);
    }
  };

  const renderImageStatus = (prod: ProductItem) => {
    const status = prod.image_processing_status || 'PROCESSED';
    if (status === 'PENDING') {
      return <span className="badge badge-pending">⏳ Processing image...</span>;
    }
    if (status === 'FAILED') {
      return <span className="badge badge-suspended">⚠️ Image Failed</span>;
    }
    if (prod.thumbnail_url || prod.image_url) {
      return (
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <img
            src={prod.thumbnail_url || prod.image_url}
            alt={prod.name}
            style={{ width: 32, height: 32, borderRadius: 4, objectFit: 'cover' }}
            onError={(e) => { (e.target as HTMLElement).style.display = 'none'; }}
          />
          <span className="badge badge-active">PROCESSED</span>
        </div>
      );
    }
    return <span className="badge badge-inactive">No Image</span>;
  };

  return (
    <div>
      <div className="page-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>Product Catalog</h1>
          <p>Manage store inventory, barcodes, and automated S3 image processing status</p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          + Add Product
        </button>
      </div>

      {error && (
        <div className="alert alert-error" style={{ marginBottom: '1.5rem' }}>
          {error}
        </div>
      )}

      {/* Filter / Search Bar */}
      <div className="card" style={{ marginBottom: '1.5rem', display: 'flex', gap: '1rem', alignItems: 'center' }}>
        <div style={{ flex: 1 }}>
          <label className="input-label">Filter by Store ID</label>
          <input
            type="text"
            className="input"
            value={storeId}
            onChange={(e) => setStoreId(e.target.value)}
            placeholder="e.g. store-001"
          />
        </div>
        <button className="btn btn-secondary" onClick={fetchProducts} style={{ marginTop: '1.5rem' }}>
          Refresh
        </button>
      </div>

      {/* Products Table */}
      <div className="card" style={{ padding: 0 }}>
        {loading ? (
          <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--color-text-muted)' }}>
            Loading products...
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Product</th>
                <th>Barcode</th>
                <th>Price / MRP</th>
                <th>HSN Code</th>
                <th>Image Processing Status</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {products.length === 0 ? (
                <tr>
                  <td colSpan={6} style={{ textAlign: 'center', padding: '2rem', color: 'var(--color-text-muted)' }}>
                    No products found for this store.
                  </td>
                </tr>
              ) : (
                products.map((prod) => (
                  <tr key={prod.id}>
                    <td>
                      <div style={{ fontWeight: 600 }}>{prod.name}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>ID: {prod.id}</div>
                    </td>
                    <td style={{ fontFamily: 'monospace' }}>{prod.barcode}</td>
                    <td>
                      ₹{(prod.price_paise / 100).toFixed(2)}{' '}
                      <span style={{ fontSize: '0.75rem', textDecoration: 'line-through', color: 'var(--color-text-muted)' }}>
                        ₹{(prod.mrp_paise / 100).toFixed(2)}
                      </span>
                    </td>
                    <td style={{ fontFamily: 'monospace' }}>{prod.hsn_code}</td>
                    <td>{renderImageStatus(prod)}</td>
                    <td>
                      <span className={`badge ${prod.is_active ? 'badge-active' : 'badge-inactive'}`}>
                        {prod.is_active ? 'ACTIVE' : 'INACTIVE'}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        )}
      </div>

      {/* Add Product Modal */}
      {showModal && (
        <div className="modal-overlay">
          <div className="modal">
            <h2>Add New Product</h2>
            <form onSubmit={handleCreateProduct}>
              <div className="input-group">
                <label className="input-label">Product Name</label>
                <input type="text" className="input" value={name} onChange={(e) => setName(e.target.value)} required />
              </div>
              <div className="input-group">
                <label className="input-label">Barcode (EAN-13)</label>
                <input type="text" className="input" value={barcode} onChange={(e) => setBarcode(e.target.value)} required />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                <div className="input-group">
                  <label className="input-label">Price (₹)</label>
                  <input type="number" step="0.01" className="input" value={price} onChange={(e) => setPrice(e.target.value)} required />
                </div>
                <div className="input-group">
                  <label className="input-label">MRP (₹)</label>
                  <input type="number" step="0.01" className="input" value={mrp} onChange={(e) => setMrp(e.target.value)} required />
                </div>
              </div>
              <div className="input-group">
                <label className="input-label">HSN Code</label>
                <input type="text" className="input" value={hsnCode} onChange={(e) => setHsnCode(e.target.value)} required />
              </div>
              <div className="input-group">
                <label className="input-label">S3 Image Key (Raw Upload)</label>
                <input type="text" className="input" value={imageUrl} onChange={(e) => setImageUrl(e.target.value)} placeholder="raw/filename.png" />
                <span style={{ fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
                  Keys starting with raw/ will trigger S3 Lambda processing and show ⏳ Processing image...
                </span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1.5rem' }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary" disabled={creating}>
                  {creating ? 'Saving...' : 'Create Product'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
