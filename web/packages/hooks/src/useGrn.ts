import { api } from '@zippyra/api-client';
import { GoodsReceivedNote } from '@zippyra/types';

export function useGrn() {
  const createGRN = async (grnData: Partial<GoodsReceivedNote>) => {
    return api.post<GoodsReceivedNote>('/api/warehouse/grn', grnData);
  };

  const updateQC = async (grnId: string, lineItemUpdates: { grn_line_item_id: string; qc_status: string }[]) => {
    return api.put<{ grn_id: string; updated: number }>(`/api/warehouse/grn/${grnId}/qc`, {
      line_item_updates: lineItemUpdates,
    });
  };

  const completeGRN = async (grnId: string) => {
    return api.post<{ grn_id: string; status: string; items_applied: number }>(`/api/warehouse/grn/${grnId}/complete`);
  };

  return {
    createGRN,
    updateQC,
    completeGRN,
  };
}
