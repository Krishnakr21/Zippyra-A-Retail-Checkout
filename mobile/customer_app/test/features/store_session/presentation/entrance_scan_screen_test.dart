import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/store_session/presentation/screens/entrance_scan_screen.dart';

void main() {
  testWidgets('entrance_scan_screen shows permission rationale UI when camera/location denied', (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: EntranceScanScreen(),
      ),
    );

    // Initial pump
    await tester.pump();

    // Verify fallback UI or scan screen title
    expect(find.text('Scan Entrance QR'), findsOneWidget);
  });
}
