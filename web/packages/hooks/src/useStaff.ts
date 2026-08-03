import { api } from '@zippyra/api-client';

export interface StaffMember {
  id: string;
  name: string;
  phone: string;
  role: 'CASHIER' | 'STOCK_ASSOCIATE' | 'SECURITY' | 'MANAGER';
  is_active: boolean;
  has_pin_set: boolean;
  store_id: string;
  created_at: string;
}

export interface CreateStaffPayload {
  name: string;
  phone: string;
  role: 'CASHIER' | 'STOCK_ASSOCIATE' | 'SECURITY' | 'MANAGER';
  store_id?: string;
}

export interface UpdateStaffPayload {
  name: string;
  role: 'CASHIER' | 'STOCK_ASSOCIATE' | 'SECURITY' | 'MANAGER';
}

export function useStaff() {
  const getStaffList = async (storeId: string, role?: string, activeOnly: boolean = true) => {
    let url = `/api/staff?store_id=${storeId}&active_only=${activeOnly}`;
    if (role && role !== 'ALL') {
      url += `&role=${role}`;
    }
    return api.get<{ staff: StaffMember[] }>(url);
  };

  const createStaff = async (payload: CreateStaffPayload) => {
    return api.post<StaffMember>('/api/staff', payload);
  };

  const updateStaff = async (id: string, payload: UpdateStaffPayload) => {
    return api.put<StaffMember>(`/api/staff/${id}`, payload);
  };

  const deactivateStaff = async (id: string) => {
    return api.delete<{ status: string }>(`/api/staff/${id}`);
  };

  return {
    getStaffList,
    createStaff,
    updateStaff,
    deactivateStaff,
  };
}
