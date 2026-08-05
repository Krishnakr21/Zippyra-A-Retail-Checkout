import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/core/services/offline_queue_service.dart';

void main() {
  late OfflineQueueService queueService;

  setUp(() {
    queueService = OfflineQueueService();
  });

  tearDown(() {
    queueService.dispose();
  });

  test('OfflineQueueService retries failing handler up to maxRetryAttempts then marks needsAttention', () async {
    int attempts = 0;
    queueService.registerHandler('TEST_ACTION', (payload) async {
      attempts++;
      return false; // Always fails
    });

    queueService.enqueue('action-1', 'TEST_ACTION', {'key': 'val'});

    // Flush 1
    await queueService.triggerFlush();
    expect(attempts, equals(1));
    expect(queueService.getPendingActions().first.retryCount, equals(1));
    expect(queueService.getPendingActions().first.needsAttention, isFalse);

    // Flush 2
    await queueService.triggerFlush();
    expect(attempts, equals(2));
    expect(queueService.getPendingActions().first.retryCount, equals(2));
    expect(queueService.getPendingActions().first.needsAttention, isFalse);

    // Flush 3 (Max retry reached!)
    await queueService.triggerFlush();
    expect(attempts, equals(3));
    expect(queueService.getPendingActions().first.retryCount, equals(3));
    expect(queueService.getPendingActions().first.needsAttention, isTrue);

    // Flush 4 (Skipped because needsAttention = true)
    await queueService.triggerFlush();
    expect(attempts, equals(3)); // No 4th attempt made!
  });
}
