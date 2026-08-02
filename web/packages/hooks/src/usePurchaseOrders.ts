import { api } from '@zippyra/api-client';
import { PurchaseOrder } from '@zippyra/types';

export function usePurchaseOrders() {
  const getPurchaseOrders = async (storeId: string, status?: string, page = 1) => {
    let url = `/api/warehouse/po?store_id=${storeId}&page=${page}`;
    if (status) url += `&status=${status}`;
    return api.get<{ items: PurchaseOrder[]; page: number }>(url);
  };

  const getPODetail = async (id: string) => {
    return api.get<PurchaseOrder>(`/api/warehouse/po/${id}`);
  };

  const createPO = async (poData: Partial<PurchaseOrder>) => {
    return api.post<PurchaseOrder>('/api/warehouse/po', poData);
  };

  const submitPO = async (id: string) => {
    return api.put<{ po_id: string; status: string }>(`/api/warehouse/po/${id}/submit`);
  };

  return {
    getPurchaseOrders,
    getPODetail,
    createPO,
    submitPO,
  };
}
