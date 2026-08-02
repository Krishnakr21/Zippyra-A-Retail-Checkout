import { api } from '@zippyra/api-client';
import { Order } from '@zippyra/types';

export function useOrders() {
  const getStoreOrders = async (storeId: string, page = 1) => {
    return api.get<{ orders: Order[]; page: number }>(`/api/orders?store_id=${storeId}&page=${page}`);
  };

  const getOrderDetail = async (orderId: string) => {
    return api.get<{ order: Order; signed_invoice_url?: string }>(`/api/orders/${orderId}`);
  };

  const acceptReturn = async (orderId: string) => {
    return api.post<{ order_id: string; status: string }>(`/api/orders/${orderId}/return/accept`);
  };

  const rejectReturn = async (orderId: string, reason: string) => {
    return api.post<{ order_id: string; status: string; reason: string }>(`/api/orders/${orderId}/return/reject`, { reason });
  };

  return {
    getStoreOrders,
    getOrderDetail,
    acceptReturn,
    rejectReturn,
  };
}
