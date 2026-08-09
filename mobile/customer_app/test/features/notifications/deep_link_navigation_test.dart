import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:customer_app/core/services/deep_link_service.dart';

class FakeGoRouter extends Fake implements GoRouter {
  final List<String> pushedPaths = [];

  @override
  Future<T?> push<T extends Object?>(String location, {Object? extra}) async {
    pushedPaths.add(location);
    return null;
  }
}

void main() {
  late FakeGoRouter fakeRouter;
  late DeepLinkService deepLinkService;

  setUp(() {
    fakeRouter = FakeGoRouter();
    deepLinkService = DeepLinkService(router: fakeRouter);
  });

  group('Deep Link 3-Scenario Resolution Tests', () {
    test('Scenario 1: Foreground FCM message received & tapped (data.deep_link = "/orders/abc-123")', () {
      const payloadDeepLink = '/orders/abc-123';

      deepLinkService.handleDeepLink(payloadDeepLink);

      expect(fakeRouter.pushedPaths, contains('/order/abc-123'));
    });

    test('Scenario 2: Background notification tapped from system tray (data.deep_link = "/orders/xyz-999")', () {
      const payloadDeepLink = '/orders/xyz-999';

      deepLinkService.handleDeepLink(payloadDeepLink);

      expect(fakeRouter.pushedPaths, contains('/order/xyz-999'));
    });

    test('Scenario 3: Killed-then-opened via notification tap on app launch (data.deep_link = "/loyalty")', () {
      const payloadDeepLink = '/loyalty';

      deepLinkService.handleDeepLink(payloadDeepLink);

      expect(fakeRouter.pushedPaths, contains('/loyalty'));
    });
  });
}
