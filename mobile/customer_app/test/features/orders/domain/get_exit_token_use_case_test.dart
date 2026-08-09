import 'package:flutter_test/flutter_test.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:customer_app/features/orders/domain/entities/order_summary.dart';
import 'package:customer_app/features/orders/domain/entities/exit_token.dart';
import 'package:customer_app/features/orders/domain/entities/order_detail.dart';
import 'package:customer_app/features/orders/domain/repositories/orders_repository.dart';
import 'package:customer_app/features/orders/domain/usecases/get_exit_token_use_case.dart';

class MockOrdersRepoForExitToken implements OrdersRepository {
  int attempts = 0;

  @override
  Future<List<OrderSummary>> getOrderHistory({int page = 1, int pageSize = 20}) async => [];

  @override
  Future<OrderDetail> getOrderDetail(String orderId) async => throw UnimplementedError();

  @override
  Future<void> requestReturn({required String orderId, required List<String> itemBarcodes, required String reason}) async {}

  @override
  Future<ExitToken> getExitToken({required String storeId}) async {
    attempts++;
    throw const NoPendingExitFailure('No pending exit pass found');
  }
}

void main() {
  test('GetExitTokenUseCase retry-on-not-ready logic attempts exactly 3 times before surfacing failure', () async {
    final repo = MockOrdersRepoForExitToken();
    final useCase = GetExitTokenUseCase(
      repo,
      maxRetries: 3,
      retryDelay: const Duration(milliseconds: 1),
    );

    expect(
      () => useCase(storeId: 'store-1'),
      throwsA(isA<NoPendingExitFailure>()),
    );

    await Future.delayed(const Duration(milliseconds: 10));
    expect(repo.attempts, equals(3));
  });
}
