import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/low_stock_item.dart';
import '../../domain/entities/purchase_order_summary.dart';

abstract class InventoryRemoteDataSource {
  Future<List<LowStockItem>> getLowStockItems(String storeId);
  Future<Map<String, dynamic>> submitStockCount(String storeId, List<Map<String, dynamic>> entries);
  Future<List<PurchaseOrderSummary>> getSubmittedPOs(String storeId);
  Future<Map<String, dynamic>> createGRN(Map<String, dynamic> body);
  Future<void> updateGRNQC(String grnId, List<Map<String, dynamic>> lineItemUpdates);
  Future<Map<String, dynamic>> completeGRN(String grnId);
}

class InventoryRemoteDataSourceImpl implements InventoryRemoteDataSource {
  final ApiClient apiClient;

  InventoryRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<List<LowStockItem>> getLowStockItems(String storeId) async {
    final response = await apiClient.get('/v1/inventory/low-stock', queryParameters: {'store_id': storeId});
    final data = response.data as Map<String, dynamic>;
    final items = data['items'] as List<dynamic>? ?? [];

    return items.map((item) {
      final m = item as Map<String, dynamic>;
      return LowStockItem(
        barcode: m['barcode'] as String? ?? '',
        productName: m['product_name'] as String? ?? m['barcode'] as String? ?? '',
        onHandQty: (m['on_hand_qty'] as num?)?.toInt() ?? 0,
        reorderPoint: (m['reorder_point'] as num?)?.toInt() ?? 10,
        reorderQty: (m['reorder_qty'] as num?)?.toInt() ?? 50,
      );
    }).toList();
  }

  @override
  Future<Map<String, dynamic>> submitStockCount(String storeId, List<Map<String, dynamic>> entries) async {
    final response = await apiClient.post('/v1/inventory/stock-count', data: {
      'store_id': storeId,
      'entries': entries,
    });
    return response.data as Map<String, dynamic>;
  }

  @override
  Future<List<PurchaseOrderSummary>> getSubmittedPOs(String storeId) async {
    final response = await apiClient.get('/v1/warehouse/po', queryParameters: {
      'store_id': storeId,
      'status': 'SUBMITTED',
    });
    final data = response.data as Map<String, dynamic>;
    final items = data['items'] as List<dynamic>? ?? [];

    return items.map((item) {
      final m = item as Map<String, dynamic>;
      return PurchaseOrderSummary(
        id: m['id'] as String? ?? '',
        storeId: m['store_id'] as String? ?? '',
        vendorName: m['vendor_name'] as String? ?? '',
        status: m['status'] as String? ?? '',
        createdAt: m['created_at'] as String? ?? '',
      );
    }).toList();
  }

  @override
  Future<Map<String, dynamic>> createGRN(Map<String, dynamic> body) async {
    final response = await apiClient.post('/v1/warehouse/grn', data: body);
    return response.data as Map<String, dynamic>;
  }

  @override
  Future<void> updateGRNQC(String grnId, List<Map<String, dynamic>> lineItemUpdates) async {
    await apiClient.put('/v1/warehouse/grn/$grnId/qc', data: {
      'line_item_updates': lineItemUpdates,
    });
  }

  @override
  Future<Map<String, dynamic>> completeGRN(String grnId) async {
    final response = await apiClient.post('/v1/warehouse/grn/$grnId/complete');
    return response.data as Map<String, dynamic>;
  }
}
