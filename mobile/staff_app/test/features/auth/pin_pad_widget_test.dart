import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:staff_app/features/auth/presentation/widgets/pin_pad.dart';

void _dummyCallback(String pin) {}

void main() {
  testWidgets('PIN pad shows locked text and disables keypad when isLocked is true', (WidgetTester tester) async {
    bool otpTapped = false;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Column(
            children: [
              const PinPadWidget(
                maxLength: 4,
                isLocked: true,
                lockedText: 'PIN locked due to failed attempts. Try again in 900s',
                onCompleted: _dummyCallback,
              ),
              TextButton(
                key: const Key('use_otp_link'),
                onPressed: () {
                  otpTapped = true;
                },
                child: const Text('Use OTP instead'),
              ),
            ],
          ),
        ),
      ),
    );

    // Verify locked text is displayed
    expect(find.text('PIN locked due to failed attempts. Try again in 900s'), findsOneWidget);

    // Verify 'Use OTP instead' link is tappable
    await tester.tap(find.byKey(const Key('use_otp_link')));
    await tester.pump();

    expect(otpTapped, isTrue);
  });
}
