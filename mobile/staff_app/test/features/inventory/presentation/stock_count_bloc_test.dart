import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/core/services/offline_queue_service.dart';
import 'package:staff_app/features/inventory/domain/entities/low_stock_item.dart';
import 'package:staff_app/features/inventory/domain/entities/purchase_order_summary.dart';
import 'package:staff_app/features/inventory/domain/entities/stock_count_entry.dart';
import 'package:staff_app/features/inventory/domain/repositories/inventory_repository.dart';
import 'package:staff_app/features/inventory/presentation/bloc/stock_count_bloc.dart';

class FailingInventoryRepo implements InventoryRepository {
  @override
  Future<Map<String, dynamic>> submitStockCount(String storeId, List<StockCountEntry> entries) async {
    throw Exception('Network connection timeout');
  }

  @override
  Future<List<LowStockItem>> getLowStockItems(String storeId) async => throw UnimplementedError();

  @override
  Future<List<PurchaseOrderSummary>> getSubmittedPOs(String storeId) async => throw UnimplementedError();

  @override
  Future<Map<String, dynamic>> createGRN({required String storeId, String? poId, String? vendorInvoiceRef, required List<Map<String, dynamic>> items}) async => throw UnimplementedError();

  @override
  Future<void> updateGRNQC({required String grnId, required List<Map<String, dynamic>> lineItemUpdates}) async => throw UnimplementedError();

  @override
  Future<Map<String, dynamic>> completeGRN(String grnId) async => throw UnimplementedError();
}

void main() {
  late FailingInventoryRepo repo;
  late OfflineQueueService queueService;
  late StockCountBloc bloc;

  setUp(() {
    repo = FailingInventoryRepo();
    queueService = OfflineQueueService();
    bloc = StockCountBloc(repository: repo, offlineQueueService: queueService);
  });

  tearDown(() {
    bloc.close();
    queueService.dispose();
  });

  blocTest<StockCountBloc, StockCountState>(
    'CountSubmitted on network failure enqueues via OfflineQueueService and emits StockCountQueuedOffline',
    build: () => bloc,
    act: (b) {
      b.add(const ItemScanned(barcode: '890111', name: 'Item A'));
      b.add(const CountSubmitted('store-001'));
    },
    expect: () => [
      isA<StockCountLoaded>().having((s) => s.entries.length, 'entries count', 1),
      isA<StockCountSubmitting>(),
      isA<StockCountQueuedOffline>(),
    ],
    verify: (_) {
      expect(queueService.getPendingActions().length, equals(1));
      expect(queueService.getPendingActions().first.actionType, equals('STOCK_COUNT'));
    },
  );
}
