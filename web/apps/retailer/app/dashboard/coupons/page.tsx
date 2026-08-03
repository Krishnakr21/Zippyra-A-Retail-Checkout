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

export default function RetailerCouponsPage() {
  const storeId = "store-001";
  const chainId = "chain-default-001";
  const [coupons, setCoupons] = useState<CouponItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showFormModal, setShowFormModal] = useState(false);
  const [editingCoupon, setEditingCoupon] = useState<CouponItem | null>(null);

  // Form State
  const [code, setCode] = useState("");
  const [discountType, setDiscountType] = useState<"PERCENT_OFF" | "FLAT_OFF">("PERCENT_OFF");
  const [discountValue, setDiscountValue] = useState(10);
  const [minCartValuePaise, setMinCartValuePaise] = useState(0);
  const [maxUses, setMaxUses] = useState<string>("");
  const [maxUsesPerCustomer, setMaxUsesPerCustomer] = useState(1);
  const [formError, setFormError] = useState<string | null>(null);

  const fetchCoupons = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/v1/cart/admin/coupons?chain_id=${chainId}&store_id=${storeId}&include_inactive=true`);
      if (res.ok) {
        const data = await res.json();
        setCoupons(data.coupons || []);
      } else {
        setError(`Failed to fetch store coupons: HTTP ${res.status}`);
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
    setDiscountValue(10);
    setMinCartValuePaise(0);
    setMaxUses("");
    setMaxUsesPerCustomer(1);
    setFormError(null);
    setShowFormModal(true);
  };

  const openEditModal = (coupon: CouponItem) => {
    setEditingCoupon(coupon);
    setCode(coupon.code);
    setDiscountType(coupon.discount_type);
    setDiscountValue(coupon.discount_value);
    setMinCartValuePaise(coupon.min_cart_value_paise);
    setMaxUses(coupon.max_uses !== undefined ? String(coupon.max_uses) : "");
    setMaxUsesPerCustomer(coupon.max_uses_per_customer);
    setFormError(null);
    setShowFormModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    // Validation
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
      store_id: storeId,
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
          "X-User-Role": "MANAGER",
          "X-Store-ID": storeId,
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
          "X-User-Role": "MANAGER",
          "X-Store-ID": storeId,
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
          "X-User-Role": "MANAGER",
          "X-Store-ID": storeId,
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
          <h1 className="text-2xl font-bold text-slate-900">Store Promotional Coupons</h1>
          <p className="text-sm text-slate-500">Store ID: {storeId} — Manage Store-Specific Promotional Codes</p>
        </div>

        <button
          onClick={openCreateModal}
          data-testid="new-store-coupon-btn"
          className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-sm rounded-lg transition-colors shadow"
        >
          + New Store Coupon
        </button>
      </div>

      {error && (
        <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg text-rose-600 text-sm">
          {error}
        </div>
      )}

      {/* Form Modal */}
      {showFormModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-lg p-6 space-y-4">
            <h2 className="text-lg font-bold text-slate-900">
              {editingCoupon ? "Edit Store Coupon" : "Create New Store Coupon"}
            </h2>

            {formError && (
              <div className="p-2 bg-rose-50 border border-rose-200 rounded text-rose-600 text-xs">
                {formError}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">Coupon Code</label>
                <input
                  type="text"
                  value={code}
                  onChange={(e) => setCode(e.target.value.toUpperCase())}
                  disabled={!!editingCoupon}
                  placeholder="SAVE50"
                  required
                  className="w-full px-3 py-2 border rounded-lg text-sm font-mono focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">Discount Type</label>
                  <select
                    value={discountType}
                    onChange={(e) => setDiscountType(e.target.value as any)}
                    className="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                  >
                    <option value="PERCENT_OFF">PERCENT_OFF (1-90%)</option>
                    <option value="FLAT_OFF">FLAT_OFF (in Paise)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">
                    Value ({discountType === "PERCENT_OFF" ? "%" : "Paise"})
                  </label>
                  <input
                    type="number"
                    value={discountValue}
                    onChange={(e) => setDiscountValue(Number(e.target.value))}
                    required
                    min={1}
                    className="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                  />
                  {discountType === "FLAT_OFF" && discountValue > 500000 && (
                    <p className="text-amber-600 text-xs mt-1">⚠ High discount value (&gt; ₹5,000)</p>
                  )}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">Global Max Uses</label>
                  <input
                    type="number"
                    value={maxUses}
                    onChange={(e) => setMaxUses(e.target.value)}
                    placeholder="Unlimited"
                    className="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">Max Uses Per Customer</label>
                  <input
                    type="number"
                    value={maxUsesPerCustomer}
                    onChange={(e) => setMaxUsesPerCustomer(Number(e.target.value))}
                    min={1}
                    required
                    className="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase mb-1">Min Cart Value (Paise)</label>
                <input
                  type="number"
                  value={minCartValuePaise}
                  onChange={(e) => setMinCartValuePaise(Number(e.target.value))}
                  min={0}
                  className="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setShowFormModal(false)}
                  className="px-4 py-2 border rounded-lg text-sm text-slate-600 hover:bg-slate-50 font-medium"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-bold shadow"
                >
                  {editingCoupon ? "Save Changes" : "Create Coupon"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Coupons Table */}
      <section className="bg-white border border-slate-200 rounded-xl overflow-hidden shadow-sm">
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between">
          <h2 className="font-bold text-slate-800">Store Coupons List</h2>
          <span className="text-xs text-slate-500">{coupons.length} coupon(s) configured</span>
        </div>

        {loading ? (
          <div className="p-8 text-center text-slate-500">Loading coupons...</div>
        ) : coupons.length === 0 ? (
          <div className="p-8 text-center text-slate-400 text-sm">No coupons found for this store. Click "+ New Store Coupon" to create one.</div>
        ) : (
          <table className="w-full text-left border-collapse text-sm">
            <thead>
              <tr className="bg-slate-50 border-b border-slate-200 text-slate-600 text-xs uppercase font-semibold">
                <th className="px-6 py-3">Code</th>
                <th className="px-6 py-3">Discount</th>
                <th className="px-6 py-3">Min Cart</th>
                <th className="px-6 py-3">Usage (Current / Max)</th>
                <th className="px-6 py-3">Per Customer Limit</th>
                <th className="px-6 py-3">Status</th>
                <th className="px-6 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {coupons.map((c) => (
                <tr key={c.id} className="hover:bg-slate-50/80 transition-colors">
                  <td className="px-6 py-4 font-mono font-bold text-indigo-600">{c.code}</td>
                  <td className="px-6 py-4 text-slate-800 font-medium">
                    {c.discount_type === "PERCENT_OFF" ? `${c.discount_value}% OFF` : `₹${(c.discount_value / 100).toFixed(2)} OFF`}
                  </td>
                  <td className="px-6 py-4 text-slate-600">
                    {c.min_cart_value_paise > 0 ? `₹${(c.min_cart_value_paise / 100).toFixed(2)}` : "None"}
                  </td>
                  <td className="px-6 py-4 text-slate-600">
                    {c.current_use_count} / {c.max_uses !== undefined && c.max_uses !== null ? c.max_uses : "∞"}
                  </td>
                  <td className="px-6 py-4 text-slate-600">{c.max_uses_per_customer} max/user</td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        c.is_active ? "bg-emerald-100 text-emerald-800" : "bg-slate-100 text-slate-600"
                      }`}
                    >
                      {c.is_active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right space-x-2">
                    <button
                      onClick={() => openEditModal(c)}
                      className="text-xs text-indigo-600 hover:text-indigo-800 font-medium"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleToggle(c)}
                      className="text-xs text-slate-600 hover:text-slate-800 font-medium"
                    >
                      {c.is_active ? "Deactivate" : "Activate"}
                    </button>
                    <button
                      onClick={() => handleDelete(c)}
                      className="text-xs text-rose-600 hover:text-rose-800 font-medium"
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
