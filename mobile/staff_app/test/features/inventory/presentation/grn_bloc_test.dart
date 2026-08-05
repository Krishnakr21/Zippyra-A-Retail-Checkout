import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/features/inventory/domain/entities/low_stock_item.dart';
import 'package:staff_app/features/inventory/domain/entities/purchase_order_summary.dart';
import 'package:staff_app/features/inventory/domain/entities/stock_count_entry.dart';
import 'package:staff_app/features/inventory/domain/repositories/inventory_repository.dart';
import 'package:staff_app/features/inventory/presentation/bloc/grn_bloc.dart';

class MockGrnRepo implements InventoryRepository {
  @override
  Future<Map<String, dynamic>> completeGRN(String grnId) async {
    throw Exception('QC_INCOMPLETE: barcodes 890111, 890222 pending');
  }

  @override
  Future<List<LowStockItem>> getLowStockItems(String storeId) async => throw UnimplementedError();

  @override
  Future<Map<String, dynamic>> submitStockCount(String storeId, List<StockCountEntry> entries) async => throw UnimplementedError();

  @override
  Future<List<PurchaseOrderSummary>> getSubmittedPOs(String storeId) async => throw UnimplementedError();

  @override
  Future<Map<String, dynamic>> createGRN({required String storeId, String? poId, String? vendorInvoiceRef, required List<Map<String, dynamic>> items}) async => throw UnimplementedError();

  @override
  Future<void> updateGRNQC({required String grnId, required List<Map<String, dynamic>> lineItemUpdates}) async => throw UnimplementedError();
}

void main() {
  late MockGrnRepo repo;
  late GrnBloc bloc;

  setUp(() {
    repo = MockGrnRepo();
    bloc = GrnBloc(repository: repo);
  });

  tearDown(() {
    bloc.close();
  });

  blocTest<GrnBloc, GrnState>(
    'GrnCompleteRequested emits QcIncomplete state when backend returns QC_INCOMPLETE',
    build: () => bloc,
    act: (b) => b.add(const GrnCompleteRequested('grn-100')),
    expect: () => [
      isA<GrnLoading>(),
      isA<QcIncomplete>(),
    ],
  );
}
