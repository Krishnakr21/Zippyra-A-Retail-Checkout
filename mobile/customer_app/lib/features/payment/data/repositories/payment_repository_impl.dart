import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../../domain/entities/payment_intent.dart';
import '../../domain/entities/payment_status_entity.dart';
import '../../domain/entities/split_estimate.dart';
import '../../domain/repositories/payment_repository.dart';
import '../datasources/payment_remote_data_source.dart';

class PaymentRepositoryImpl implements PaymentRepository {
  final PaymentRemoteDataSource remoteDataSource;
  final FlutterSecureStorage secureStorage;

  static const String _pendingPaymentIdKey = 'pending_payment_id';

  PaymentRepositoryImpl({
    required this.remoteDataSource,
    FlutterSecureStorage? secureStorage,
  }) : secureStorage = secureStorage ?? const FlutterSecureStorage();

  @override
  Future<SplitEstimate> estimateSplit({
    required String checkoutSessionId,
    required int pointsToRedeem,
  }) {
    return remoteDataSource.estimateSplit(
      checkoutSessionId: checkoutSessionId,
      pointsToRedeem: pointsToRedeem,
    );
  }

  @override
  Future<PaymentIntent> initiatePayment({
    required String checkoutSessionId,
    required String method,
    int pointsToRedeem = 0,
    String? playIntegrityToken,
  }) {
    return remoteDataSource.initiatePayment(
      checkoutSessionId: checkoutSessionId,
      method: method,
      pointsToRedeem: pointsToRedeem,
      playIntegrityToken: playIntegrityToken,
    );
  }

  @override
  Future<PaymentStatusEntity> getPaymentStatus(String paymentId) {
    return remoteDataSource.getPaymentStatus(paymentId);
  }

  @override
  Future<void> savePendingPaymentId(String paymentId) async {
    await secureStorage.write(key: _pendingPaymentIdKey, value: paymentId);
  }

  @override
  Future<String?> getPendingPaymentId() async {
    return await secureStorage.read(key: _pendingPaymentIdKey);
  }

  @override
  Future<void> clearPendingPaymentId() async {
    await secureStorage.delete(key: _pendingPaymentIdKey);
  }
}
