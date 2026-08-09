import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:customer_app/features/orders/domain/entities/order_summary.dart';
import 'package:customer_app/features/orders/domain/entities/exit_token.dart';
import 'package:customer_app/features/orders/domain/entities/order_detail.dart';
import 'package:customer_app/features/orders/domain/entities/returnable_item.dart';
import 'package:customer_app/features/orders/domain/repositories/orders_repository.dart';
import 'package:customer_app/features/orders/domain/usecases/get_order_detail_use_case.dart';
import 'package:customer_app/features/orders/domain/usecases/request_return_use_case.dart';
import 'package:customer_app/features/orders/presentation/bloc/order_detail_bloc.dart';

class MockOrdersRepoForDetail implements OrdersRepository {
  bool failReturn = false;

  @override
  Future<List<OrderSummary>> getOrderHistory({int page = 1, int pageSize = 20}) async => [];

  @override
  Future<OrderDetail> getOrderDetail(String orderId) async {
    return OrderDetail(
      id: orderId,
      paymentId: 'pay-1',
      userId: 'user-1',
      storeId: 'store-1',
      items: const [
        ReturnableItem(barcode: '890100', name: 'Item A', qty: 2, returnedQty: 0, isReturnable: true, pricePaise: 5000),
      ],
      subtotalPaise: 10000,
      discountPaise: 0,
      cgstPaise: 900,
      sgstPaise: 900,
      igstPaise: 0,
      totalPaise: 10000,
      loyaltyPointsUsed: 0,
      paymentMethod: 'UPI',
      status: 'COMPLETED',
      createdAt: DateTime.now(),
    );
  }

  @override
  Future<void> requestReturn({required String orderId, required List<String> itemBarcodes, required String reason}) async {
    if (failReturn) {
      throw const ItemNotReturnableFailure('Item is not returnable');
    }
  }

  @override
  Future<ExitToken> getExitToken({required String storeId}) async {
    throw UnimplementedError();
  }
}

void main() {
  late MockOrdersRepoForDetail repo;
  late OrderDetailBloc bloc;

  setUp(() {
    repo = MockOrdersRepoForDetail();
    bloc = OrderDetailBloc(
      getOrderDetailUseCase: GetOrderDetailUseCase(repo),
      requestReturnUseCase: RequestReturnUseCase(repo),
    );
  });

  tearDown(() {
    bloc.close();
  });

  blocTest<OrderDetailBloc, OrderDetailState>(
    'ReturnRequested with valid selection emits ReturnSubmitting then ReturnSubmitted',
    build: () => bloc,
    act: (bloc) => bloc.add(const ReturnRequested(
      orderId: 'ord-100',
      itemBarcodes: ['890100'],
      reason: 'DAMAGED',
    )),
    expect: () => [
      isA<ReturnSubmitting>(),
      isA<ReturnSubmitted>().having((s) => s.orderId, 'orderId', 'ord-100'),
    ],
  );

  blocTest<OrderDetailBloc, OrderDetailState>(
    'Server ITEM_NOT_RETURNABLE error emits ReturnFailed with specific error code',
    build: () {
      repo.failReturn = true;
      return bloc;
    },
    act: (bloc) => bloc.add(const ReturnRequested(
      orderId: 'ord-100',
      itemBarcodes: ['890100'],
      reason: 'DAMAGED',
    )),
    expect: () => [
      isA<ReturnSubmitting>(),
      isA<ReturnFailed>().having((s) => s.errorCode, 'errorCode', ErrorCodes.itemNotReturnable),
    ],
  );
}
