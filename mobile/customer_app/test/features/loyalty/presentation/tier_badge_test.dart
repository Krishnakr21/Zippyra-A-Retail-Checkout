import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/loyalty/presentation/widgets/tier_badge.dart';

void main() {
  testWidgets('TierBadge renders Bronze tier correctly', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: TierBadge(tier: 'BRONZE', displayName: 'Bronze Tier'),
        ),
      ),
    );

    expect(find.text('Bronze Tier'), findsOneWidget);
    expect(find.byIcon(Icons.shield_outlined), findsOneWidget);
  });

  testWidgets('TierBadge renders Silver tier correctly', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: TierBadge(tier: 'SILVER', displayName: 'Silver Tier'),
        ),
      ),
    );

    expect(find.text('Silver Tier'), findsOneWidget);
    expect(find.byIcon(Icons.stars), findsOneWidget);
  });

  testWidgets('TierBadge renders Gold tier correctly', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: TierBadge(tier: 'GOLD', displayName: 'Gold Tier'),
        ),
      ),
    );

    expect(find.text('Gold Tier'), findsOneWidget);
    expect(find.byIcon(Icons.military_tech), findsOneWidget);
  });

  testWidgets('TierBadge renders Platinum tier correctly', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: TierBadge(tier: 'PLATINUM', displayName: 'Platinum Tier'),
        ),
      ),
    );

    expect(find.text('Platinum Tier'), findsOneWidget);
    expect(find.byIcon(Icons.workspace_premium), findsOneWidget);
  });
}
