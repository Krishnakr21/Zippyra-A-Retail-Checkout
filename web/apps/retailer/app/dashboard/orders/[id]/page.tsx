'use client';

import React, { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Badge } from '@zippyra/ui';
import { Order } from '@zippyra/types';
import { useOrders } from '@zippyra/hooks';

export default function OrderDetailPage() {
  const params = useParams();
  const router = useRouter();
  const orderId = params.id as string;
  const { getOrderDetail, acceptReturn, rejectReturn } = useOrders();

  const [order, setOrder] = useState<Order | null>(null);
  const [signedInvoiceUrl, setSignedInvoiceUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [msg, setMsg] = useState('');

  const loadOrder = () => {
    setLoading(true);
    getOrderDetail(orderId)
      .then((res) => {
        setOrder(res.order);
        setSignedInvoiceUrl(res.signed_invoice_url || null);
      })
      .catch((err) => setMsg(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (orderId) loadOrder();
  }, [orderId]);

  const handleAcceptReturn = async () => {
    setActionLoading(true);
    try {
      await acceptReturn(orderId);
      setMsg('Return request ACCEPTED. Stock reversal and loyalty reversal events emitted.');
      loadOrder();
    } catch (err: any) {
      setMsg(err.message || 'Failed to accept return');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRejectReturn = async () => {
    setActionLoading(true);
    try {
      await rejectReturn(orderId, 'Item does not meet store return policy condition');
      setMsg('Return request REJECTED.');
      loadOrder();
    } catch (err: any) {
      setMsg(err.message || 'Failed to reject return');
    } finally {
      setActionLoading(false);
    }
  };

  if (loading) return <div className="p-8 text-center text-gray-500">Loading order details...</div>;
  if (!order) return <div className="p-8 text-center text-red-500">{msg || 'Order not found'}</div>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <button onClick={() => router.back()} className="text-xs text-blue-600 hover:underline mb-1">
            ← Back to Orders
          </button>
          <h1 className="text-2xl font-bold text-gray-900">Order #{order.id.substring(0, 8)}</h1>
          <p className="text-sm text-gray-500 mt-1">Payment ID: {order.payment_id}</p>
        </div>
        <div className="flex items-center gap-4">
          <Badge status={order.status} />
          {signedInvoiceUrl && (
            <a
              href={signedInvoiceUrl}
              target="_blank"
              rel="noreferrer"
              className="px-3 py-1.5 bg-blue-50 text-blue-700 hover:bg-blue-100 text-xs font-semibold rounded-md border border-blue-200"
            >
              📄 Download Tax Invoice
            </a>
          )}
        </div>
      </div>

      {msg && (
        <div className="p-4 bg-blue-50 text-blue-800 rounded-lg text-sm border border-blue-200">
          {msg}
        </div>
      )}

      {/* Return Request Action Card */}
      {order.status === 'RETURN_REQUESTED' && (
        <div className="p-6 bg-amber-50 border border-amber-200 rounded-xl space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-base font-bold text-amber-900">Review Customer Return Request</h3>
              <p className="text-xs text-amber-700 mt-1">Customer requested a return for items in this order.</p>
            </div>
            <div className="flex gap-3">
              <button
                onClick={handleRejectReturn}
                disabled={actionLoading}
                className="px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white text-xs font-bold rounded-lg disabled:opacity-50"
              >
                Reject Return
              </button>
              <button
                onClick={handleAcceptReturn}
                disabled={actionLoading}
                className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold rounded-lg disabled:opacity-50"
              >
                Accept Return
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Summary Card */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="p-6 bg-white rounded-xl border border-gray-200 shadow-sm space-y-2">
          <p className="text-xs font-semibold text-gray-500">TOTAL AMOUNT</p>
          <p className="text-2xl font-bold text-gray-900">₹{(order.total_paise / 100.0).toFixed(2)}</p>
          <p className="text-xs text-gray-400">Payment Method: {order.payment_method}</p>
        </div>
        <div className="p-6 bg-white rounded-xl border border-gray-200 shadow-sm space-y-2">
          <p className="text-xs font-semibold text-gray-500">GST BREAKDOWN</p>
          <p className="text-sm font-semibold text-gray-800">
            CGST + SGST: ₹{((order.cgst_paise + order.sgst_paise) / 100.0).toFixed(2)}
          </p>
          <p className="text-xs text-gray-400">Discount: ₹{(order.discount_paise / 100.0).toFixed(2)}</p>
        </div>
        <div className="p-6 bg-white rounded-xl border border-gray-200 shadow-sm space-y-2">
          <p className="text-xs font-semibold text-gray-500">ORDER DATE</p>
          <p className="text-base font-bold text-gray-900">{new Date(order.created_at).toLocaleString()}</p>
        </div>
      </div>
    </div>
  );
}
