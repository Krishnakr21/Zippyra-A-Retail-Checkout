import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/cart/presentation/widgets/sticky_checkout_bar.dart';

void main() {
  testWidgets('StickyCheckoutBar button is disabled when isEnabled is false', (tester) async {
    bool pressed = false;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: StickyCheckoutBar(
            totalPaise: 5000,
            isEnabled: false,
            isLoading: false,
            onCheckout: () => pressed = true,
          ),
        ),
      ),
    );

    final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
    expect(button.onPressed, isNull);

    await tester.tap(find.byType(ElevatedButton));
    expect(pressed, isFalse);
  });

  testWidgets('StickyCheckoutBar button is enabled when isEnabled is true', (tester) async {
    bool pressed = false;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: StickyCheckoutBar(
            totalPaise: 5000,
            isEnabled: true,
            isLoading: false,
            onCheckout: () => pressed = true,
          ),
        ),
      ),
    );

    final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
    expect(button.onPressed, isNotNull);

    await tester.tap(find.byType(ElevatedButton));
    expect(pressed, isTrue);
  });
}
