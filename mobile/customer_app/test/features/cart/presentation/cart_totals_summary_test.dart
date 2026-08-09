import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/cart/domain/entities/cart_summary.dart';
import 'package:customer_app/features/cart/presentation/widgets/cart_totals_summary.dart';

void main() {
  testWidgets('CartTotalsSummary shows CGST/SGST and hides IGST when igstPaise == 0', (tester) async {
    const summary = CartSummary(
      items: [],
      subtotalPaise: 50000,
      discountPaise: 0,
      cgstPaise: 4500,
      sgstPaise: 4500,
      igstPaise: 0,
      totalPaise: 59000,
      appliedOffers: [],
      itemCount: 1,
    );

    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: CartTotalsSummary(summary: summary),
        ),
      ),
    );

    expect(find.text('CGST'), findsOneWidget);
    expect(find.text('SGST'), findsOneWidget);
    expect(find.text('IGST'), findsNothing);
  });

  testWidgets('CartTotalsSummary shows IGST and hides CGST/SGST when igstPaise > 0', (tester) async {
    const summary = CartSummary(
      items: [],
      subtotalPaise: 50000,
      discountPaise: 0,
      cgstPaise: 0,
      sgstPaise: 0,
      igstPaise: 9000,
      totalPaise: 59000,
      appliedOffers: [],
      itemCount: 1,
    );

    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: CartTotalsSummary(summary: summary),
        ),
      ),
    );

    expect(find.text('IGST'), findsOneWidget);
    expect(find.text('CGST'), findsNothing);
    expect(find.text('SGST'), findsNothing);
  });
}
