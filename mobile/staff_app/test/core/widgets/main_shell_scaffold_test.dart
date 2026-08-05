import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/core/widgets/main_shell_scaffold.dart';
import 'package:staff_app/features/auth/domain/entities/staff_session.dart';
import 'package:staff_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:staff_app/features/inventory/domain/entities/low_stock_item.dart';
import 'package:staff_app/features/inventory/domain/entities/purchase_order_summary.dart';
import 'package:staff_app/features/inventory/domain/entities/stock_count_entry.dart';
import 'package:staff_app/features/inventory/domain/repositories/inventory_repository.dart';
import 'package:staff_app/features/inventory/presentation/bloc/low_stock_bloc.dart';
import 'package:staff_app/features/shift/domain/entities/staff_shift.dart';
import 'package:staff_app/features/shift/domain/repositories/shift_repository.dart';
import 'package:staff_app/features/shift/presentation/bloc/shift_bloc.dart';

import 'role_guard_test.dart';

class StubInventoryRepo implements InventoryRepository {
  @override
  Future<List<LowStockItem>> getLowStockItems(String storeId) async => [];

  @override
  Future<Map<String, dynamic>> completeGRN(String grnId) async => {};

  @override
  Future<Map<String, dynamic>> createGRN({required String storeId, String? poId, String? vendorInvoiceRef, required List<Map<String, dynamic>> items}) async => {};

  @override
  Future<List<PurchaseOrderSummary>> getSubmittedPOs(String storeId) async => [];

  @override
  Future<Map<String, dynamic>> submitStockCount(String storeId, List<StockCountEntry> entries) async => {};

  @override
  Future<void> updateGRNQC({required String grnId, required List<Map<String, dynamic>> lineItemUpdates}) async {}
}

class StubShiftRepo implements ShiftRepository {
  @override
  Future<void> endShift() async {}

  @override
  Future<StaffShiftEntity?> getCurrentShift() async => null;

  @override
  Future<StaffShiftEntity> startShift() async => StaffShiftEntity(
        id: '1',
        staffId: '1',
        storeId: 'store-1',
        startedAt: DateTime.now(),
      );
}

void main() {
  testWidgets('MainShellScaffold hides Inventory tab for SECURITY role', (tester) async {
    final authBloc = AuthBloc(authRepository: DummyAuthRepo());
    authBloc.emit(const AuthAuthenticated(
      staffId: '99',
      role: 'SECURITY',
      storeId: 'store-sec',
      storeName: 'Sec Store',
      session: StaffSession(token: 't', staffId: '99', role: 'SECURITY', storeId: 's', storeName: 'n'),
    ));

    final shiftBloc = ShiftBloc(shiftRepository: StubShiftRepo());
    final lowStockBloc = LowStockBloc(repository: StubInventoryRepo());

    await tester.pumpWidget(
      MaterialApp(
        home: MultiBlocProvider(
          providers: [
            BlocProvider<AuthBloc>.value(value: authBloc),
            BlocProvider<ShiftBloc>.value(value: shiftBloc),
            BlocProvider<LowStockBloc>.value(value: lowStockBloc),
          ],
          child: const MainShellScaffold(initialIndex: 0),
        ),
      ),
    );

    // Verify 4 tabs exist (Inventory hidden for SECURITY)
    expect(find.text('Shift'), findsOneWidget);
    expect(find.text('Devices'), findsOneWidget);
    expect(find.text('POS Assist'), findsOneWidget);
    expect(find.text('Profile'), findsOneWidget);
    expect(find.text('Inventory'), findsNothing);
  });
}
