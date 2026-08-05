import '../entities/low_stock_item.dart';
import '../entities/stock_count_entry.dart';
import '../entities/purchase_order_summary.dart';

abstract class InventoryRepository {
  Future<List<LowStockItem>> getLowStockItems(String storeId);
  Future<Map<String, dynamic>> submitStockCount(String storeId, List<StockCountEntry> entries);
  Future<List<PurchaseOrderSummary>> getSubmittedPOs(String storeId);
  Future<Map<String, dynamic>> createGRN({
    required String storeId,
    String? poId,
    String? vendorInvoiceRef,
    required List<Map<String, dynamic>> items,
  });
  Future<void> updateGRNQC({
    required String grnId,
    required List<Map<String, dynamic>> lineItemUpdates,
  });
  Future<Map<String, dynamic>> completeGRN(String grnId);
}
