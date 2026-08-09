import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:customer_app/features/membership/domain/entities/subscription_plan.dart';
import 'package:customer_app/features/membership/domain/entities/member_subscription.dart';
import 'package:customer_app/features/membership/domain/repositories/membership_repository.dart';
import 'package:customer_app/features/membership/presentation/cubit/membership_cubit.dart';
import 'package:customer_app/features/membership/presentation/screens/membership_screen.dart';

class FakeMembershipRepository implements MembershipRepository {
  List<SubscriptionPlan> mockPlans = const [
    SubscriptionPlan(
      id: 'plan-1',
      chainId: 'chain-1',
      name: 'Smart Saver Monthly',
      pricePaise: 19900,
      billingInterval: 'MONTHLY',
      loyaltyMultiplierBonus: 0.5,
      freeDelivery: true,
      isActive: true,
    ),
    SubscriptionPlan(
      id: 'plan-2',
      chainId: 'chain-1',
      name: 'Smart Saver Annual',
      pricePaise: 149900,
      billingInterval: 'ANNUAL',
      loyaltyMultiplierBonus: 0.5,
      freeDelivery: true,
      isActive: true,
    ),
  ];

  MemberSubscription? mockSub;

  @override
  Future<List<SubscriptionPlan>> getPlans() async {
    return mockPlans;
  }

  @override
  Future<MemberSubscription?> getMySubscription() async {
    return mockSub;
  }

  @override
  Future<MemberSubscription> subscribe(String planId) async {
    mockSub = MemberSubscription(
      id: 'sub-new-1',
      userId: 'usr-1',
      planId: planId,
      status: 'ACTIVE',
      createdAt: DateTime.now(),
      plan: mockPlans.firstWhere((p) => p.id == planId, orElse: () => mockPlans[0]),
    );
    return mockSub!;
  }

  @override
  Future<void> cancelSubscription() async {
    mockSub = null;
  }
}

void main() {
  group('MembershipCubit Tests', () {
    late FakeMembershipRepository fakeRepo;
    late MembershipCubit cubit;

    setUp(() {
      fakeRepo = FakeMembershipRepository();
      cubit = MembershipCubit(repository: fakeRepo);
    });

    tearDown(() {
      cubit.close();
    });

    test('loadMembershipData emits MembershipLoading then MembershipLoaded', () async {
      final expectedStates = [
        isA<MembershipLoading>(),
        isA<MembershipLoaded>(),
      ];

      expectLater(cubit.stream, emitsInOrder(expectedStates));
      cubit.loadMembershipData();
    });

    test('subscribe creates active subscription and reloads', () async {
      cubit.subscribe('plan-1');
      await untilCalled(() => fakeRepo.getMySubscription());
      expect(fakeRepo.mockSub?.planId, equals('plan-1'));
    });
  });

  group('MembershipScreen Widget Tests', () {
    late FakeMembershipRepository fakeRepo;

    setUp(() {
      fakeRepo = FakeMembershipRepository();
    });

    Widget createWidgetUnderTest() {
      return MaterialApp(
        home: BlocProvider(
          create: (_) => MembershipCubit(repository: fakeRepo),
          child: const MembershipScreen(),
        ),
      );
    }

    testWidgets('renders Smart Saver plans and member benefits', (tester) async {
      await tester.pumpWidget(createWidgetUnderTest());
      await tester.pumpAndSettle();

      expect(find.text('Smart Saver Membership'), findsOneWidget);
      expect(find.text('SMART SAVER'), findsOneWidget);
      expect(find.text('Smart Saver Monthly'), findsOneWidget);
      expect(find.text('Smart Saver Annual'), findsOneWidget);
      expect(find.text('+0.5x Extra Loyalty Multiplier'), findsOneWidget);
      expect(find.text('Free Delivery Equivalent Perks'), findsOneWidget);
    });
  });
}

Future<void> untilCalled(Function() fn) async {
  await Future.delayed(const Duration(milliseconds: 50));
}
