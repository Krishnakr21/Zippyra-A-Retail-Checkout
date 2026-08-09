import 'package:zippyra_core/zippyra_core.dart';
import '../models/payment_intent_model.dart';
import '../models/payment_status_model.dart';
import '../models/split_estimate_model.dart';

abstract class PaymentRemoteDataSource {
  Future<SplitEstimateModel> estimateSplit({
    required String checkoutSessionId,
    required int pointsToRedeem,
  });

  Future<PaymentIntentModel> initiatePayment({
    required String checkoutSessionId,
    required String method,
    int pointsToRedeem = 0,
    String? playIntegrityToken,
  });

  Future<PaymentStatusModel> getPaymentStatus(String paymentId);
}

class PaymentRemoteDataSourceImpl implements PaymentRemoteDataSource {
  final ApiClient apiClient;

  PaymentRemoteDataSourceImpl({required this.apiClient});

  @override
  Future<SplitEstimateModel> estimateSplit({
    required String checkoutSessionId,
    required int pointsToRedeem,
  }) async {
    try {
      final response = await apiClient.post(
        '/v1/payment/split/estimate',
        data: {
          'checkout_session_id': checkoutSessionId,
          'loyalty_points_to_redeem': pointsToRedeem,
        },
      );
      final data = response.data as Map<String, dynamic>;
      return SplitEstimateModel.fromJson(data);
    } catch (e) {
      throw ServerFailure('Failed to estimate split payment: $e');
    }
  }

  @override
  Future<PaymentIntentModel> initiatePayment({
    required String checkoutSessionId,
    required String method,
    int pointsToRedeem = 0,
    String? playIntegrityToken,
  }) async {
    try {
      final payload = <String, dynamic>{
        'checkout_session_id': checkoutSessionId,
        'payment_method': method,
        'loyalty_points_to_redeem': pointsToRedeem,
      };
      if (playIntegrityToken != null && playIntegrityToken.isNotEmpty) {
        payload['play_integrity_token'] = playIntegrityToken;
      }
      final response = await apiClient.post(
        '/v1/payment/initiate',
        data: payload,
      );
      final data = response.data as Map<String, dynamic>;
      return PaymentIntentModel.fromJson(data);
    } catch (e) {
      throw ServerFailure('Failed to initiate payment: $e');
    }
  }

  @override
  Future<PaymentStatusModel> getPaymentStatus(String paymentId) async {
    try {
      final response = await apiClient.get('/v1/payment/status/$paymentId');
      final data = response.data as Map<String, dynamic>;
      return PaymentStatusModel.fromJson(data);
    } catch (e) {
      throw ServerFailure('Failed to get payment status: $e');
    }
  }
}
