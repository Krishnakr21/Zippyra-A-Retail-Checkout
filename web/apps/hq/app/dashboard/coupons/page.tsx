"use client";

import React, { useEffect, useState } from "react";

interface CouponItem {
  id: string;
  chain_id: string;
  store_id?: string;
  code: string;
  discount_type: "PERCENT_OFF" | "FLAT_OFF";
  discount_value: number;
  min_cart_value_paise: number;
  max_uses?: number;
  max_uses_per_customer: number;
  current_use_count: number;
  is_active: boolean;
  active_from: string;
  active_until?: string;
}

export default function ChainCouponsPage() {
  const chainId = "chain-default-001";
  const [coupons, setCoupons] = useState<CouponItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showFormModal, setShowFormModal] = useState(false);
  const [editingCoupon, setEditingCoupon] = useState<CouponItem | null>(null);

  // Form State
  const [code, setCode] = useState("");
  const [discountType, setDiscountType] = useState<"PERCENT_OFF" | "FLAT_OFF">("PERCENT_OFF");
  const [discountValue, setDiscountValue] = useState(15);
  const [minCartValuePaise, setMinCartValuePaise] = useState(0);
  const [maxUses, setMaxUses] = useState<string>("");
  const [maxUsesPerCustomer, setMaxUsesPerCustomer] = useState(1);
  const [targetScope, setTargetScope] = useState<"CHAIN_WIDE" | "SPECIFIC_STORE">("CHAIN_WIDE");
  const [specificStoreId, setSpecificStoreId] = useState("store-001");
  const [formError, setFormError] = useState<string | null>(null);

  const fetchCoupons = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/v1/cart/admin/coupons?chain_id=${chainId}&include_inactive=true`);
      if (res.ok) {
        const data = await res.json();
        setCoupons(data.coupons || []);
      } else {
        setError(`Failed to fetch chain coupons: HTTP ${res.status}`);
      }
    } catch (err: any) {
      setError(err?.message || "Network error fetching coupons");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCoupons();
  }, []);

  const openCreateModal = () => {
    setEditingCoupon(null);
    setCode("");
    setDiscountType("PERCENT_OFF");
    setDiscountValue(15);
    setMinCartValuePaise(0);
    setMaxUses("");
    setMaxUsesPerCustomer(1);
    setTargetScope("CHAIN_WIDE");
    setSpecificStoreId("store-001");
    setFormError(null);
    setShowFormModal(true);
  };

  const openEditModal = (coupon: CouponItem) => {
    setEditingCoupon(coupon);
    setCode(coupon.code);
    setDiscountType(coupon.discount_type);
    setDiscountValue(coupon.discount_value);
    setMinCartValuePaise(coupon.min_cart_value_paise);
    setMaxUses(coupon.max_uses !== undefined && coupon.max_uses !== null ? String(coupon.max_uses) : "");
    setMaxUsesPerCustomer(coupon.max_uses_per_customer);
    if (coupon.store_id) {
      setTargetScope("SPECIFIC_STORE");
      setSpecificStoreId(coupon.store_id);
    } else {
      setTargetScope("CHAIN_WIDE");
    }
    setFormError(null);
    setShowFormModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (discountType === "PERCENT_OFF" && (discountValue < 1 || discountValue > 90)) {
      setFormError("PERCENT_OFF discount value must be between 1 and 90");
      return;
    }
    if (discountType === "FLAT_OFF" && discountValue <= 0) {
      setFormError("FLAT_OFF discount value must be greater than 0");
      return;
    }
    if (maxUsesPerCustomer < 1) {
      setFormError("max_uses_per_customer must be at least 1");
      return;
    }

    const payload = {
      chain_id: chainId,
      store_id: targetScope === "SPECIFIC_STORE" ? specificStoreId : null,
      code: code.trim().toUpperCase(),
      discount_type: discountType,
      discount_value: Number(discountValue),
      min_cart_value_paise: Number(minCartValuePaise),
      max_uses: maxUses ? Number(maxUses) : null,
      max_uses_per_customer: Number(maxUsesPerCustomer),
    };

    try {
      const url = editingCoupon ? `/v1/cart/admin/coupons/${editingCoupon.id}` : "/v1/cart/admin/coupons";
      const method = editingCoupon ? "PUT" : "POST";

      const res = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          "X-User-Role": "OWNER",
          "X-Chain-ID": chainId,
        },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.message || `HTTP ${res.status}`);
      }

      setShowFormModal(false);
      fetchCoupons();
    } catch (err: any) {
      setFormError(err?.message || "Failed to save coupon");
    }
  };

  const handleToggle = async (coupon: CouponItem) => {
    try {
      const res = await fetch(`/v1/cart/admin/coupons/${coupon.id}/toggle`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-User-Role": "OWNER",
          "X-Chain-ID": chainId,
        },
      });
      if (res.ok) {
        fetchCoupons();
      }
    } catch (err) {
      console.error("Failed to toggle coupon:", err);
    }
  };

  const handleDelete = async (coupon: CouponItem) => {
    try {
      const res = await fetch(`/v1/cart/admin/coupons/${coupon.id}`, {
        method: "DELETE",
        headers: {
          "X-User-Role": "OWNER",
          "X-Chain-ID": chainId,
        },
      });
      if (res.ok) {
        fetchCoupons();
      }
    } catch (err) {
      console.error("Failed to delete coupon:", err);
    }
  };

  return (
    <div className="space-y-8 p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">Chain-Wide Promotional Coupons</h1>
          <p className="text-sm text-slate-400">Chain ID: {chainId} — Enterprise Coupon Management & Redis Fan-Out</p>
        </div>

        <button
          onClick={openCreateModal}
          data-testid="new-chain-coupon-btn"
          className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-sm rounded-lg transition-colors shadow"
        >
          + New Chain Coupon
        </button>
      </div>

      {error && (
        <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg text-rose-400 text-sm">
          {error}
        </div>
      )}

      {/* Form Modal */}
      {showFormModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-xl shadow-2xl w-full max-w-lg p-6 space-y-4 text-slate-100">
            <h2 className="text-lg font-bold text-white">
              {editingCoupon ? "Edit Chain Coupon" : "Create New Chain Coupon"}
            </h2>

            {formError && (
              <div className="p-2 bg-rose-500/10 border border-rose-500/30 rounded text-rose-400 text-xs">
                {formError}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-400 uppercase mb-1">Target Scope</label>
                  <select
                    value={targetScope}
                    onChange={(e) => setTargetScope(e.target.value as any)}
                    disabled={!!editingCoupon}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none"
                  >
                    <option value="CHAIN_WIDE">Chain-Wide (All Stores)</option>
                    <option value="SPECIFIC_STORE">Specific Store</option>
                  </select>
                </div>

                {targetScope === "SPECIFIC_STORE" && (
                  <div>
                    <label className="block text-xs font-semibold text-slate-400 uppercase mb-1">Store ID</label>
                    <input
                      type="text"
                      value={specificStoreId}
                      onChange={(e) => setSpecificStoreId(e.target.value)}
                      disabled={!!editingCoupon}
                      placeholder="store-001"
                      className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none"
                    />
                  </div>
                )}
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-400 uppercase mb-1">Coupon Code</label>
                <input
                  type="text"
                  value={code}
                  onChange={(e) => setCode(e.target.value.toUpperCase())}
                  disabled={!!editingCoupon}
                  placeholder="MEGA50"
                  required
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm font-mono text-slate-100 focus:outline-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-400 uppercase mb-1">Discount Type</label>
                  <select
                    value={discountType}
                    onChange={(e) => setDiscountType(e.target.value as any)}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none"
                  >
                    <option value="PERCENT_OFF">PERCENT_OFF (1-90%)</option>
                    <option value="FLAT_OFF">FLAT_OFF (in Paise)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-400 uppercase mb-1">
                    Value ({discountType === "PERCENT_OFF" ? "%" : "Paise"})
                  </label>
                  <input
                    type="number"
                    value={discountValue}
                    onChange={(e) => setDiscountValue(Number(e.target.value))}
                    required
                    min={1}
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-400 uppercase mb-1">Global Max Uses</label>
                  <input
                    type="number"
                    value={maxUses}
                    onChange={(e) => setMaxUses(e.target.value)}
                    placeholder="Unlimited"
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-400 uppercase mb-1">Max Uses Per Customer</label>
                  <input
                    type="number"
                    value={maxUsesPerCustomer}
                    onChange={(e) => setMaxUsesPerCustomer(Number(e.target.value))}
                    min={1}
                    required
                    className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-400 uppercase mb-1">Min Cart Value (Paise)</label>
                <input
                  type="number"
                  value={minCartValuePaise}
                  onChange={(e) => setMinCartValuePaise(Number(e.target.value))}
                  min={0}
                  className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-sm text-slate-100 focus:outline-none"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowFormModal(false)}
                  className="px-4 py-2 border border-slate-700 rounded-lg text-sm text-slate-300 hover:bg-slate-800 font-medium"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-bold shadow"
                >
                  {editingCoupon ? "Save Changes" : "Create & Fan-Out"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Coupons Table */}
      <section className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-sm">
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between">
          <h2 className="font-bold text-slate-200">Active & Available Chain Coupons</h2>
          <span className="text-xs text-slate-400">{coupons.length} coupon(s) configured</span>
        </div>

        {loading ? (
          <div className="p-8 text-center text-slate-400">Loading chain coupons...</div>
        ) : coupons.length === 0 ? (
          <div className="p-8 text-center text-slate-500 text-sm">No coupons found for this chain. Click "+ New Chain Coupon" to create one.</div>
        ) : (
          <table className="w-full text-left border-collapse text-sm">
            <thead>
              <tr className="bg-slate-950/60 border-b border-slate-800 text-slate-400 text-xs uppercase font-semibold">
                <th className="px-6 py-3">Code</th>
                <th className="px-6 py-3">Scope</th>
                <th className="px-6 py-3">Discount</th>
                <th className="px-6 py-3">Usage (Current / Max)</th>
                <th className="px-6 py-3">Per Customer Limit</th>
                <th className="px-6 py-3">Status</th>
                <th className="px-6 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {coupons.map((c) => (
                <tr key={c.id} className="hover:bg-slate-800/50 transition-colors">
                  <td className="px-6 py-4 font-mono font-bold text-indigo-400">{c.code}</td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold ${
                        c.store_id ? "bg-amber-500/10 text-amber-400 border border-amber-500/20" : "bg-indigo-500/10 text-indigo-400 border border-indigo-500/20"
                      }`}
                    >
                      {c.store_id ? `Store: ${c.store_id}` : "Chain-Wide (All Stores)"}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-slate-200 font-medium">
                    {c.discount_type === "PERCENT_OFF" ? `${c.discount_value}% OFF` : `₹${(c.discount_value / 100).toFixed(2)} OFF`}
                  </td>
                  <td className="px-6 py-4 text-slate-300">
                    {c.current_use_count} / {c.max_uses !== undefined && c.max_uses !== null ? c.max_uses : "∞"}
                  </td>
                  <td className="px-6 py-4 text-slate-300">{c.max_uses_per_customer} max/user</td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        c.is_active ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" : "bg-slate-800 text-slate-400"
                      }`}
                    >
                      {c.is_active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right space-x-2">
                    <button
                      onClick={() => openEditModal(c)}
                      className="text-xs text-indigo-400 hover:text-indigo-300 font-medium"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleToggle(c)}
                      className="text-xs text-slate-400 hover:text-slate-200 font-medium"
                    >
                      {c.is_active ? "Deactivate" : "Activate"}
                    </button>
                    <button
                      onClick={() => handleDelete(c)}
                      className="text-xs text-rose-400 hover:text-rose-300 font-medium"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  );
}
