import { api } from '@zippyra/api-client';
import { TransferOrder } from '@zippyra/types';

export function useTransfers() {
  const createTransfer = async (transferData: Partial<TransferOrder>) => {
    return api.post<TransferOrder>('/api/warehouse/transfer', transferData);
  };

  const approveTransfer = async (id: string) => {
    return api.put<{ transfer_id: string; status: string }>(`/api/warehouse/transfer/${id}/approve`);
  };

  const rejectTransfer = async (id: string, reason: string) => {
    return api.put<{ transfer_id: string; status: string }>(`/api/warehouse/transfer/${id}/reject`, { reason });
  };

  const shipTransfer = async (id: string, items: { barcode: string; qty_shipped: number }[]) => {
    return api.put<{ transfer_id: string; status: string }>(`/api/warehouse/transfer/${id}/ship`, { items });
  };

  const receiveTransfer = async (id: string, items: { barcode: string; qty_received: number }[]) => {
    return api.put<{ transfer_id: string; status: string }>(`/api/warehouse/transfer/${id}/receive`, { items });
  };

  return {
    createTransfer,
    approveTransfer,
    rejectTransfer,
    shipTransfer,
    receiveTransfer,
  };
}
