import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/main.dart';
import 'package:staff_app/injection_container.dart';

void main() {
  testWidgets('Zippyra Staff App initializes auth screen', (WidgetTester tester) async {
    await initServiceLocator();
    await tester.pumpWidget(const StaffApp());
    expect(find.text('Staff Sign In'), findsOneWidget);
  });
}
