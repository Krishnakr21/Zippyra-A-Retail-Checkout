import 'package:go_router/go_router.dart';

class DeepLinkService {
  final GoRouter router;

  DeepLinkService({required this.router});

  void handleDeepLink(String? deepLink) {
    if (deepLink == null || deepLink.trim().isEmpty) return;

    final trimmed = deepLink.trim();
    final uri = Uri.parse(trimmed);
    final path = uri.path;

    // Reconcile path shapes from notification-service payloads
    if (path.startsWith('/orders/')) {
      final orderId = path.substring('/orders/'.length);
      if (orderId.isNotEmpty) {
        router.push('/order/$orderId');
        return;
      }
    }

    if (path.startsWith('/order/')) {
      router.push(path);
      return;
    }

    if (path == '/loyalty' || path == '/loyalty/history' || path == '/loyalty/tiers') {
      router.push(path);
      return;
    }

    if (path == '/exit/qr' || path.startsWith('/exit')) {
      router.push(path);
      return;
    }

    if (path == '/notifications' || path == '/notifications/preferences') {
      router.push(path);
      return;
    }

    // Default route push
    router.push(trimmed);
  }
}
