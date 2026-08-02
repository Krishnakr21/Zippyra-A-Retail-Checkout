import { api } from '@zippyra/api-client';
import { LowStockItem } from '@zippyra/types';

export function useInventory() {
  const getLowStockItems = async (storeId: string) => {
    return api.get<{ items: LowStockItem[] }>(`/api/inventory/low-stock?store_id=${storeId}`);
  };

  return {
    getLowStockItems,
  };
}
