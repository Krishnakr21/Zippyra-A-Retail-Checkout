import { api } from '@zippyra/api-client';

export interface ShiftRecord {
  id: string;
  staff_id: string;
  staff_name: string;
  role: string;
  started_at: string;
  ended_at?: string;
  duration_minutes?: number;
}

export function useShiftHistory() {
  const getShiftHistory = async (storeId: string, dateFrom?: string, dateTo?: string) => {
    let url = `/v1/retailer-auth/shift/history?store_id=${storeId}`;
    if (dateFrom) url += `&date_from=${dateFrom}`;
    if (dateTo) url += `&date_to=${dateTo}`;
    return api.get<{ shifts: ShiftRecord[] }>(url);
  };

  return {
    getShiftHistory,
  };
}
