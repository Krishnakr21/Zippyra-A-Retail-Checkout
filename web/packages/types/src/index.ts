export interface Product {
  id: string;
  store_id: string;
  chain_id: string;
  barcode: string;
  name: string;
  description?: string;
  category_id: string;
  price_paise: number;
  mrp_paise: number;
  hsn_code: string;
  gst_rate_percent: number;
  is_active: boolean;
  is_returnable: boolean;
  image_url?: string;
  thumbnail_url?: string;
  sync_seq?: number;
  created_at: string;
  updated_at: string;
}

export interface StockLevel {
  store_id: string;
  barcode: string;
  on_hand_qty: number;
  reorder_point: number;
  reorder_qty: number;
  low_stock_alerted: boolean;
  updated_at: string;
}

export interface LowStockItem {
  barcode: string;
  product_name?: string;
  on_hand_qty: number;
  reorder_point: number;
  reorder_qty: number;
}

export interface POLineItem {
  id?: string;
  barcode: string;
  qty_ordered: number;
  unit_cost_paise: number;
  qty_received?: number;
}

export interface PurchaseOrder {
  id: string;
  store_id: string;
  chain_id: string;
  vendor_name: string;
  status: 'DRAFT' | 'SUBMITTED' | 'PARTIALLY_RECEIVED' | 'RECEIVED' | 'CANCELLED';
  source: 'MANUAL' | 'AUTO_REORDER';
  created_by?: string;
  expected_delivery_date?: string;
  line_items?: POLineItem[];
  created_at: string;
  submitted_at?: string;
  completed_at?: string;
}

export interface GRNLineItem {
  id: string;
  barcode: string;
  qty_expected?: number;
  qty_received: number;
  unit_cost_paise: number;
  qc_status: 'PENDING' | 'PASSED' | 'REJECTED';
  qc_note?: string;
}

export interface GoodsReceivedNote {
  id: string;
  po_id?: string;
  store_id: string;
  received_by: string;
  vendor_invoice_ref?: string;
  status: 'DRAFT' | 'QC_PENDING' | 'COMPLETED';
  line_items?: GRNLineItem[];
  created_at: string;
  completed_at?: string;
}

export interface TransferLineItem {
  id: string;
  barcode: string;
  qty_requested: number;
  qty_shipped?: number;
  qty_received?: number;
}

export interface TransferOrder {
  id: string;
  source_store_id: string;
  dest_store_id: string;
  chain_id: string;
  status: 'REQUESTED' | 'APPROVED' | 'REJECTED' | 'IN_TRANSIT' | 'RECEIVED' | 'CANCELLED';
  requested_by: string;
  rejection_reason?: string;
  line_items?: TransferLineItem[];
  created_at: string;
  approved_at?: string;
  shipped_at?: string;
  received_at?: string;
}

export interface OrderItem {
  barcode: string;
  name: string;
  qty: number;
  price_paise: number;
  hsn_code: string;
  is_returnable: boolean;
}

export interface Order {
  id: string;
  payment_id: string;
  user_id: string;
  store_id: string;
  items?: OrderItem[];
  subtotal_paise: number;
  discount_paise: number;
  cgst_paise: number;
  sgst_paise: number;
  igst_paise: number;
  total_paise: number;
  loyalty_points_used: number;
  payment_method: string;
  supply_type: string;
  status: 'CREATED' | 'COMPLETED' | 'RETURN_REQUESTED' | 'RETURNED' | 'RETURN_REJECTED';
  invoice_s3_key?: string;
  irn?: string;
  created_at: string;
  completed_at?: string;
}

export interface StaffSessionUser {
  staffId: string;
  role: 'CASHIER' | 'STOCK_ASSOCIATE' | 'SECURITY' | 'MANAGER';
  storeId: string;
  storeName: string;
  token: string;
}
