import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/orders/domain/entities/order_summary.dart';
import 'package:customer_app/features/orders/domain/entities/exit_token.dart';
import 'package:customer_app/features/orders/domain/entities/order_detail.dart';
import 'package:customer_app/features/orders/domain/repositories/orders_repository.dart';
import 'package:customer_app/features/orders/domain/usecases/get_order_history_use_case.dart';
import 'package:customer_app/features/orders/presentation/bloc/order_history_bloc.dart';

class MockOrdersRepositoryForHistory implements OrdersRepository {
  @override
  Future<List<OrderSummary>> getOrderHistory({int page = 1, int pageSize = 20}) async {
    if (page == 1) {
      return List.generate(
        20,
        (i) => OrderSummary(
          id: 'ord-p1-$i',
          storeId: 'store-1',
          storeName: 'Store 1',
          totalPaise: 10000,
          itemCount: 2,
          status: 'COMPLETED',
          createdAt: DateTime.now(),
        ),
      );
    } else {
      return List.generate(
        5,
        (i) => OrderSummary(
          id: 'ord-p2-$i',
          storeId: 'store-1',
          storeName: 'Store 1',
          totalPaise: 5000,
          itemCount: 1,
          status: 'COMPLETED',
          createdAt: DateTime.now(),
        ),
      );
    }
  }

  @override
  Future<OrderDetail> getOrderDetail(String orderId) async {
    throw UnimplementedError();
  }

  @override
  Future<void> requestReturn({required String orderId, required List<String> itemBarcodes, required String reason}) async {}

  @override
  Future<ExitToken> getExitToken({required String storeId}) async {
    throw UnimplementedError();
  }
}

void main() {
  late MockOrdersRepositoryForHistory repo;
  late OrderHistoryBloc bloc;

  setUp(() {
    repo = MockOrdersRepositoryForHistory();
    bloc = OrderHistoryBloc(getOrderHistoryUseCase: GetOrderHistoryUseCase(repo));
  });

  tearDown(() {
    bloc.close();
  });

  test('initial state is OrderHistoryInitial', () {
    expect(bloc.state, isA<OrderHistoryInitial>());
  });

  blocTest<OrderHistoryBloc, OrderHistoryState>(
    'OrderHistoryRequested fetches page 1 with 20 items and hasMore=true',
    build: () => bloc,
    act: (bloc) => bloc.add(const OrderHistoryRequested(refresh: true)),
    expect: () => [
      isA<OrderHistoryLoading>(),
      isA<OrderHistoryLoaded>()
          .having((s) => s.orders.length, 'orders length', 20)
          .having((s) => s.hasMore, 'hasMore', true),
    ],
  );

  blocTest<OrderHistoryBloc, OrderHistoryState>(
    'OrderHistoryNextPageRequested appends page 2 items to existing list rather than replacing it',
    build: () => bloc,
    act: (bloc) async {
      bloc.add(const OrderHistoryRequested(refresh: true));
      await Future.delayed(const Duration(milliseconds: 10));
      bloc.add(OrderHistoryNextPageRequested());
    },
    expect: () => [
      isA<OrderHistoryLoading>(),
      isA<OrderHistoryLoaded>().having((s) => s.orders.length, 'orders length', 20),
      isA<OrderHistoryLoaded>()
          .having((s) => s.orders.length, 'orders length', 25)
          .having((s) => s.hasMore, 'hasMore', false),
    ],
  );
}
