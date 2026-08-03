"use client";

import React, { useEffect, useState } from "react";
import { OfferTable, OfferForm, OfferPreviewPanel, OfferTableItem, OfferFormValues } from "@zippyra/ui";

export default function RetailerOffersPage() {
  const storeId = "store-001";
  const [offers, setOffers] = useState<OfferTableItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showFormModal, setShowFormModal] = useState(false);
  const [editingOffer, setEditingOffer] = useState<OfferTableItem | null>(null);

  const fetchOffers = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/v1/cart/admin/offers?store_id=${storeId}`);
      if (res.ok) {
        const data = await res.json();
        setOffers(data.offers || []);
      } else {
        setError(`Failed to fetch store offers: HTTP ${res.status}`);
      }
    } catch (err: any) {
      setError(err?.message || "Network error fetching offers");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOffers();
  }, []);

  const handleCreateSubmit = async (values: OfferFormValues) => {
    const payload = {
      ...values,
      store_id: storeId, // Strictly store-scoped for store manager
    };

    const res = await fetch("/v1/cart/admin/offers", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-User-Role": "MANAGER",
        "X-Store-ID": storeId,
      },
      body: JSON.stringify(payload),
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.message || `Failed to create offer: HTTP ${res.status}`);
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
        "X-User-Role": "MANAGER",
        "X-Store-ID": storeId,
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
          "X-User-Role": "MANAGER",
          "X-Store-ID": storeId,
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
          "X-User-Role": "MANAGER",
          "X-Store-ID": storeId,
        },
      });
      if (res.ok) {
        fetchOffers();
      }
    } catch (err) {
      console.error("Failed to delete offer:", err);
    }
  };

  return (
    <div className="space-y-8 p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">Store Promotional Offers & Campaigns</h1>
          <p className="text-sm text-slate-400">Store ID: {storeId} — Active Effective Offers</p>
        </div>

        <button
          onClick={() => {
            setEditingOffer(null);
            setShowFormModal(true);
          }}
          data-testid="new-store-offer-btn"
          className="px-4 py-2 bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold text-sm rounded-lg transition-colors shadow"
        >
          + New Store Offer
        </button>
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
              scope="STORE"
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

      {/* Main Effective Offer Table */}
      <section className="space-y-3">
        <h2 className="text-lg font-semibold text-slate-200">Effective Offer Rules (Chain-Wide & Store-Specific)</h2>
        {loading ? (
          <div className="p-6 text-center text-slate-500">Loading store offers...</div>
        ) : (
          <OfferTable
            offers={offers}
            showScope={true}
            userRole="MANAGER"
            onEdit={(offer) => setEditingOffer(offer)}
            onToggle={handleToggle}
            onDelete={handleDelete}
          />
        )}
      </section>

      {/* Live Preview Panel */}
      <section>
        <OfferPreviewPanel storeId={storeId} />
      </section>
    </div>
  );
}
