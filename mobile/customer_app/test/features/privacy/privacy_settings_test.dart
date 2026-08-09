import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/privacy/data/datasources/privacy_remote_data_source.dart';
import 'package:customer_app/features/privacy/data/repositories/privacy_repository_impl.dart';
import 'package:customer_app/features/privacy/presentation/screens/privacy_settings_screen.dart';

void main() {
  late MockPrivacyRemoteDataSource dataSource;
  late PrivacyRepositoryImpl repository;

  setUp(() {
    dataSource = MockPrivacyRemoteDataSource();
    repository = PrivacyRepositoryImpl(remoteDataSource: dataSource);
  });

  testWidgets('PrivacySettingsScreen renders consent options and handles optimistic toggle', (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: PrivacySettingsScreen(repository: repository),
      ),
    );

    // Initial loading
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    await tester.pumpAndSettle();

    // Verify consents loaded
    expect(find.text('Marketing Communications'), findsOneWidget);
    expect(find.text('Location Personalization'), findsOneWidget);
    expect(find.text('Analytics Sharing'), findsOneWidget);

    // Verify needs_reconfirmation badge exists for LOCATION_TRACKING
    expect(find.byKey(const Key('reconfirmation_badge_LOCATION_TRACKING')), findsOneWidget);

    // Toggle LOCATION_TRACKING switch
    final locationTile = find.byKey(const Key('consent_card_LOCATION_TRACKING'));
    final switchFinder = find.descendant(of: locationTile, matching: find.byType(Switch));
    expect(switchFinder, findsOneWidget);

    await tester.tap(switchFinder);
    await tester.pumpAndSettle();

    // Verify needs_reconfirmation badge DISAPPEARS after re-toggling
    expect(find.byKey(const Key('reconfirmation_badge_LOCATION_TRACKING')), findsNothing);

    // Verify Grievance Officer section renders
    expect(find.text('Grievance Officer'), findsOneWidget);
    expect(find.text('Nisha Sharma'), findsOneWidget);
    expect(find.textContaining('72 hours'), findsOneWidget);
  });
}
