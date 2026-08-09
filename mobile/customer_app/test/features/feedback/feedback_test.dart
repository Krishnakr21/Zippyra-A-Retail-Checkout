import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:dio/dio.dart';
import 'package:customer_app/features/feedback/data/feedback_service.dart';
import 'package:customer_app/features/feedback/presentation/feedback_modal.dart';

class FakeApiClient implements ApiClient {
  @override
  Dio get dio => Dio();

  Map<String, dynamic>? lastBody;

  @override
  Future<Response> get(String path, {Map<String, dynamic>? queryParameters}) async {
    return Response(requestOptions: RequestOptions(path: path), data: {});
  }

  @override
  Future<Response> post(String path, {dynamic data}) async {
    lastBody = data as Map<String, dynamic>?;
    return Response(requestOptions: RequestOptions(path: path), data: {'id': 'fb-1'});
  }

  @override
  Future<Response> put(String path, {dynamic data, Map<String, dynamic>? queryParameters}) async {
    return Response(requestOptions: RequestOptions(path: path), data: {});
  }

  @override
  Future<Response> delete(String path, {dynamic data, Map<String, dynamic>? queryParameters}) async {
    return Response(requestOptions: RequestOptions(path: path), data: {});
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('FeedbackService Gating Tests', () {
    late FlutterSecureStorage storage;
    late FeedbackService service;
    late FakeApiClient apiClient;

    setUp(() {
      FlutterSecureStorage.setMockInitialValues({});
      storage = const FlutterSecureStorage();
      apiClient = FakeApiClient();
      service = FeedbackService(
        apiClient: apiClient,
        secureStorage: storage,
      );
    });

    test('incrementOrderAndCheckGating triggers strictly once per 3 orders', () async {
      expect(await service.incrementOrderAndCheckGating(), isFalse); // Order 1
      expect(await service.incrementOrderAndCheckGating(), isFalse); // Order 2
      expect(await service.incrementOrderAndCheckGating(), isTrue);  // Order 3!

      expect(await service.incrementOrderAndCheckGating(), isFalse); // Order 4
      expect(await service.incrementOrderAndCheckGating(), isFalse); // Order 5
      expect(await service.incrementOrderAndCheckGating(), isTrue);  // Order 6!
    });
  });

  group('FeedbackModal Widget Tests', () {
    late FlutterSecureStorage storage;
    late FeedbackService service;
    late FakeApiClient apiClient;

    setUp(() {
      FlutterSecureStorage.setMockInitialValues({});
      storage = const FlutterSecureStorage();
      apiClient = FakeApiClient();
      service = FeedbackService(
        apiClient: apiClient,
        secureStorage: storage,
      );
    });

    Widget createWidgetUnderTest() {
      return MaterialApp(
        home: Scaffold(
          body: FeedbackModal(
            feedbackService: service,
            sourceApp: 'CUSTOMER_APP',
            contextStr: 'post_checkout',
          ),
        ),
      );
    }

    testWidgets('renders rating stars, comment input, and submit button', (tester) async {
      await tester.pumpWidget(createWidgetUnderTest());
      await tester.pumpAndSettle();

      expect(find.text('How was your experience?'), findsOneWidget);
      expect(find.text('Submit'), findsOneWidget);
      expect(find.text('Not Now'), findsOneWidget);
      expect(find.byType(TextField), findsOneWidget);
    });
  });
}
