import { api } from '@zippyra/api-client';
import { Product } from '@zippyra/types';

export function useCatalog() {
  const getProducts = async (storeId: string, page = 1) => {
    return api.get<{ products: Product[]; total: number; page: number }>(
      `/api/catalog?store_id=${storeId}&page=${page}`
    );
  };

  const getProductByBarcode = async (storeId: string, barcode: string) => {
    return api.get<Product>(`/api/catalog/barcode?store_id=${storeId}&barcode=${barcode}`);
  };

  const createProduct = async (productData: Partial<Product>) => {
    return api.post<Product>('/api/catalog', productData);
  };

  const updateProduct = async (id: string, productData: Partial<Product>) => {
    return api.put<Product>(`/api/catalog/${id}`, productData);
  };

  return {
    getProducts,
    getProductByBarcode,
    createProduct,
    updateProduct,
  };
}
