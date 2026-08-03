"use client";

import React, { useEffect, useState } from "react";
import { OfferTable, OfferForm, OfferTableItem, OfferFormValues } from "@zippyra/ui";

export default function ChainHQOffersPage() {
  const chainId = "chain-001";
  const [userRole, setUserRole] = useState<string>("OWNER"); // Default to OWNER, can switch for testing/viewing

  const [offers, setOffers] = useState<OfferTableItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showFormModal, setShowFormModal] = useState(false);
  const [editingOffer, setEditingOffer] = useState<OfferTableItem | null>(null);

  const fetchOffers = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/v1/cart/admin/offers?chain_id=${chainId}`, {
        headers: {
          "X-User-Role": userRole,
          "X-Chain-ID": chainId,
        },
      });
      if (res.ok) {
        const data = await res.json();
        setOffers(data.offers || []);
      } else {
        setError(`Failed to fetch chain offers: HTTP ${res.status}`);
      }
    } catch (err: any) {
      setError(err?.message || "Network error fetching chain offers");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOffers();
  }, [userRole]);

  const handleCreateSubmit = async (values: OfferFormValues) => {
    const payload = {
      ...values,
      chain_id: chainId,
      store_id: null, // Chain-wide offer
    };

    const res = await fetch("/v1/cart/admin/offers", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-User-Role": userRole,
        "X-Chain-ID": chainId,
      },
      body: JSON.stringify(payload),
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.message || `Failed to create chain-wide offer: HTTP ${res.status}`);
    }

    setShowFormModal(false);
    fetchOffers();
  };

  const handleEditSubmit = async (values: OfferFormValues) => {
    if (!editingOffer) return;

    const res = await fetch(`/v1/cart/admin/offers/${editingOffer.id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "X-User-Role": userRole,
        "X-Chain-ID": chainId,
      },
      body: JSON.stringify(values),
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.message || `Failed to update offer: HTTP ${res.status}`);
    }

    setEditingOffer(null);
    fetchOffers();
  };

  const handleToggle = async (offer: OfferTableItem, isActive: boolean) => {
    try {
      const res = await fetch(`/v1/cart/admin/offers/${offer.id}/toggle`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-User-Role": userRole,
          "X-Chain-ID": chainId,
        },
        body: JSON.stringify({ is_active: isActive }),
      });
      if (res.ok) {
        fetchOffers();
      }
    } catch (err) {
      console.error("Failed to toggle offer:", err);
    }
  };

  const handleDelete = async (offer: OfferTableItem) => {
    try {
      const res = await fetch(`/v1/cart/admin/offers/${offer.id}`, {
        method: "DELETE",
        headers: {
          "X-User-Role": userRole,
          "X-Chain-ID": chainId,
        },
      });
      if (res.ok) {
        fetchOffers();
      }
    } catch (err) {
      console.error("Failed to delete offer:", err);
    }
  };

  const chainWideOffers = offers.filter((o) => !o.store_id || o.scope === "CHAIN_WIDE");
  const storeSpecificOffers = offers.filter((o) => o.store_id && o.scope === "STORE_SPECIFIC");

  return (
    <div className="space-y-8 p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">Chain HQ Offer Campaign Authoring</h1>
          <p className="text-sm text-slate-400">Chain ID: {chainId} — Global Chain-Wide Promotions</p>
        </div>

        <div className="flex items-center space-x-4">
          {/* Role simulation selector for testing/verifying gating */}
          <div className="flex items-center space-x-2 text-xs text-slate-400">
            <span>Simulate Role:</span>
            <select
              value={userRole}
              onChange={(e) => setUserRole(e.target.value)}
              className="px-2 py-1 bg-slate-800 border border-slate-700 rounded text-slate-200 text-xs"
            >
              <option value="OWNER">OWNER</option>
              <option value="FINANCE">FINANCE</option>
              <option value="OPERATIONS">OPERATIONS</option>
            </select>
          </div>

          {/* New Chain-Wide Offer button - ONLY visible for OWNER role */}
          {userRole === "OWNER" && (
            <button
              onClick={() => {
                setEditingOffer(null);
                setShowFormModal(true);
              }}
              data-testid="new-chain-offer-btn"
              className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white font-bold text-sm rounded-lg transition-colors shadow"
            >
              + New Chain-Wide Offer
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="p-3 bg-rose-500/10 border border-rose-500/30 rounded-lg text-rose-400 text-sm">
          {error}
        </div>
      )}

      {/* Form Modal for Create / Edit */}
      {(showFormModal || editingOffer) && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 overflow-y-auto">
          <div className="w-full max-w-2xl my-8">
            <OfferForm
              mode={editingOffer ? "edit" : "create"}
              scope="CHAIN_WIDE"
              initialValues={(editingOffer as any) || undefined}
              onSubmit={editingOffer ? handleEditSubmit : handleCreateSubmit}
              onCancel={() => {
                setShowFormModal(false);
                setEditingOffer(null);
              }}
            />
          </div>
        </div>
      )}

      {/* Section 1: Chain-Wide Offers */}
      <section className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-200">Global Chain-Wide Offers</h2>
          <span className="text-xs text-purple-400 font-semibold">{chainWideOffers.length} Active</span>
        </div>

        {loading ? (
          <div className="p-6 text-center text-slate-500">Loading chain-wide offers...</div>
        ) : (
          <OfferTable
            offers={chainWideOffers}
            showScope={false}
            userRole={userRole}
            onEdit={(offer) => setEditingOffer(offer)}
            onToggle={handleToggle}
            onDelete={handleDelete}
          />
        )}
      </section>

      {/* Section 2: Store-Specific Overrides Summary */}
      <section className="space-y-3 pt-6 border-t border-slate-800">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-200">Store-Specific Local Offers Overview</h2>
          <span className="text-xs text-blue-400 font-semibold">{storeSpecificOffers.length} Store Overrides</span>
        </div>

        {loading ? (
          <div className="p-6 text-center text-slate-500">Loading store overrides...</div>
        ) : (
          <OfferTable
            offers={storeSpecificOffers}
            showScope={true}
            userRole="FINANCE" // Read-only for store overrides in HQ
            onEdit={() => {}}
            onToggle={() => {}}
            onDelete={() => {}}
          />
        )}
      </section>
    </div>
  );
}
