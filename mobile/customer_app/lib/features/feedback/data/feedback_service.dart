import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:zippyra_core/zippyra_core.dart';

class FeedbackService {
  final ApiClient apiClient;
  final FlutterSecureStorage secureStorage;

  FeedbackService({
    required this.apiClient,
    required this.secureStorage,
  });

  static const String _keyOrderCount = 'feedback_completed_order_count';

  /// Increments local completed order count and checks if feedback modal should show (once per 3 orders).
  Future<bool> incrementOrderAndCheckGating() async {
    final str = await secureStorage.read(key: _keyOrderCount);
    int current = int.tryParse(str ?? '0') ?? 0;
    current += 1;
    await secureStorage.write(key: _keyOrderCount, value: current.toString());
    return (current % 3 == 0);
  }

  /// Submits feedback to support-service POST /v1/support/feedback
  Future<void> submitFeedback({
    int? npsScore,
    String? comment,
    String sourceApp = 'CUSTOMER_APP',
    String context = 'post_checkout',
  }) async {
    try {
      await apiClient.post('/v1/support/feedback', data: {
        if (npsScore != null) 'nps_score': npsScore,
        if (comment != null && comment.isNotEmpty) 'comment': comment,
        'source_app': sourceApp,
        'context': context,
      });
    } catch (_) {
      // Non-blocking failure for feedback
    }
  }
}
