import 'package:zippyra_core/zippyra_core.dart';
import '../../domain/entities/low_stock_item.dart';
import '../../domain/entities/stock_count_entry.dart';
import '../../domain/entities/purchase_order_summary.dart';
import '../../domain/repositories/inventory_repository.dart';
import '../datasources/inventory_remote_data_source.dart';

class InventoryRepositoryImpl implements InventoryRepository {
  final InventoryRemoteDataSource remoteDataSource;

  InventoryRepositoryImpl({required this.remoteDataSource});

  @override
  Future<List<LowStockItem>> getLowStockItems(String storeId) async {
    try {
      return await remoteDataSource.getLowStockItems(storeId);
    } catch (e) {
      throw ServerFailure('Failed to load low stock items: $e');
    }
  }

  @override
  Future<Map<String, dynamic>> submitStockCount(String storeId, List<StockCountEntry> entries) async {
    try {
      final payloadEntries = entries
          .map((e) => {
                'barcode': e.barcode,
                'counted_qty': e.countedQty,
              })
          .toList();
      return await remoteDataSource.submitStockCount(storeId, payloadEntries);
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<List<PurchaseOrderSummary>> getSubmittedPOs(String storeId) async {
    try {
      return await remoteDataSource.getSubmittedPOs(storeId);
    } catch (e) {
      throw ServerFailure('Failed to load submitted POs: $e');
    }
  }

  @override
  Future<Map<String, dynamic>> createGRN({
    required String storeId,
    String? poId,
    String? vendorInvoiceRef,
    required List<Map<String, dynamic>> items,
  }) async {
    try {
      final body = <String, dynamic>{
        'store_id': storeId,
        'items': items,
      };
      if (poId != null) body['po_id'] = poId;
      if (vendorInvoiceRef != null) body['vendor_invoice_ref'] = vendorInvoiceRef;

      return await remoteDataSource.createGRN(body);
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<void> updateGRNQC({
    required String grnId,
    required List<Map<String, dynamic>> lineItemUpdates,
  }) async {
    try {
      await remoteDataSource.updateGRNQC(grnId, lineItemUpdates);
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }

  @override
  Future<Map<String, dynamic>> completeGRN(String grnId) async {
    try {
      return await remoteDataSource.completeGRN(grnId);
    } catch (e) {
      throw ServerFailure(e.toString());
    }
  }
}
