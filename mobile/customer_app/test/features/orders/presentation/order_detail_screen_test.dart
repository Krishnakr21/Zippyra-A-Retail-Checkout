import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/orders/domain/entities/order_summary.dart';
import 'package:customer_app/features/orders/domain/entities/exit_token.dart';
import 'package:customer_app/features/orders/domain/entities/order_detail.dart';
import 'package:customer_app/features/orders/domain/entities/returnable_item.dart';
import 'package:customer_app/features/orders/domain/repositories/orders_repository.dart';
import 'package:customer_app/features/orders/domain/usecases/get_order_detail_use_case.dart';
import 'package:customer_app/features/orders/domain/usecases/request_return_use_case.dart';
import 'package:customer_app/features/orders/presentation/bloc/order_detail_bloc.dart';
import 'package:customer_app/features/orders/presentation/screens/order_detail_screen.dart';

class MockOrdersRepoForScreen implements OrdersRepository {
  final bool isOld;

  MockOrdersRepoForScreen({this.isOld = false});

  @override
  Future<List<OrderSummary>> getOrderHistory({int page = 1, int pageSize = 20}) async => [];

  @override
  Future<OrderDetail> getOrderDetail(String orderId) async {
    final createdAt = isOld ? DateTime.now().subtract(const Duration(hours: 30)) : DateTime.now();
    return OrderDetail(
      id: orderId,
      paymentId: 'pay-1',
      userId: 'user-1',
      storeId: 'store-1',
      items: const [
        ReturnableItem(barcode: '890100', name: 'Item A', qty: 1, returnedQty: 0, isReturnable: true, pricePaise: 5000),
      ],
      subtotalPaise: 5000,
      discountPaise: 0,
      cgstPaise: 450,
      sgstPaise: 450,
      igstPaise: 0,
      totalPaise: 5000,
      loyaltyPointsUsed: 0,
      paymentMethod: 'UPI',
      status: 'COMPLETED',
      createdAt: createdAt,
    );
  }

  @override
  Future<void> requestReturn({required String orderId, required List<String> itemBarcodes, required String reason}) async {}

  @override
  Future<ExitToken> getExitToken({required String storeId}) async => throw UnimplementedError();
}

void main() {
  testWidgets('Request Return button is HIDDEN when order is older than 24h', (tester) async {
    final repo = MockOrdersRepoForScreen(isOld: true);
    final bloc = OrderDetailBloc(
      getOrderDetailUseCase: GetOrderDetailUseCase(repo),
      requestReturnUseCase: RequestReturnUseCase(repo),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: BlocProvider<OrderDetailBloc>.value(
          value: bloc,
          child: const OrderDetailScreen(orderId: 'ord-old-123'),
        ),
      ),
    );

    await tester.pump(); // trigger initState request
    await tester.pump(const Duration(milliseconds: 50)); // resolve async call

    final btnFinder = find.text('Request Return');
    expect(btnFinder, findsNothing);

    bloc.close();
  });

  testWidgets('Request Return button is VISIBLE when order is within 24h and has returnable items', (tester) async {
    final repo = MockOrdersRepoForScreen(isOld: false);
    final bloc = OrderDetailBloc(
      getOrderDetailUseCase: GetOrderDetailUseCase(repo),
      requestReturnUseCase: RequestReturnUseCase(repo),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: BlocProvider<OrderDetailBloc>.value(
          value: bloc,
          child: const OrderDetailScreen(orderId: 'ord-new-123'),
        ),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    final btnFinder = find.text('Request Return');
    expect(btnFinder, findsOneWidget);

    bloc.close();
  });
}
