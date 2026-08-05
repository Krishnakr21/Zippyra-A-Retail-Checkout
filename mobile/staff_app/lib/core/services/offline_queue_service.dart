import 'dart:async';

class PendingAction {
  final String id;
  final String actionType;
  final Map<String, dynamic> payload;
  final DateTime createdAt;
  int retryCount;
  bool needsAttention;

  PendingAction({
    required this.id,
    required this.actionType,
    required this.payload,
    required this.createdAt,
    this.retryCount = 0,
    this.needsAttention = false,
  });
}

typedef ActionHandler = Future<bool> Function(Map<String, dynamic> payload);

class OfflineQueueService {
  final List<PendingAction> _queue = [];
  final Map<String, ActionHandler> _handlers = {};
  Timer? _retryTimer;
  static const int maxRetryAttempts = 3;

  OfflineQueueService() {
    _startRetryLoop();
  }

  void registerHandler(String actionType, ActionHandler handler) {
    _handlers[actionType] = handler;
  }

  void enqueue(String id, String actionType, Map<String, dynamic> payload) {
    _queue.add(PendingAction(
      id: id,
      actionType: actionType,
      payload: payload,
      createdAt: DateTime.now(),
    ));
  }

  List<PendingAction> getPendingActions() => List.unmodifiable(_queue);

  Future<void> triggerFlush() async {
    for (final action in List<PendingAction>.from(_queue)) {
      if (action.needsAttention) continue;

      final handler = _handlers[action.actionType];
      if (handler != null) {
        final success = await handler(action.payload);
        if (success) {
          _queue.removeWhere((item) => item.id == action.id);
        } else {
          action.retryCount++;
          if (action.retryCount >= maxRetryAttempts) {
            action.needsAttention = true;
          }
        }
      }
    }
  }

  void _startRetryLoop() {
    _retryTimer = Timer.periodic(const Duration(seconds: 15), (_) {
      triggerFlush();
    });
  }

  void dispose() {
    _retryTimer?.cancel();
  }
}
