import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/main.dart';
import 'package:customer_app/injection_container.dart';

void main() {
  testWidgets('Zippyra Customer App initializes and navigates', (WidgetTester tester) async {
    await initServiceLocator();
    await tester.pumpWidget(const ZippyraCustomerApp());
    expect(find.text('Zippyra'), findsOneWidget);
    await tester.pump(const Duration(seconds: 3));
  });
}
