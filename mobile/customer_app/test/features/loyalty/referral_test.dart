import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:customer_app/features/loyalty/domain/entities/referral_info.dart';
import 'package:customer_app/features/loyalty/domain/repositories/referral_repository.dart';
import 'package:customer_app/features/loyalty/presentation/cubit/referral_cubit.dart';
import 'package:customer_app/features/loyalty/presentation/screens/referral_screen.dart';

class FakeReferralRepository implements ReferralRepository {
  ReferralInfo info = const ReferralInfo(
    referralCode: 'ZIPPY123',
    shareText: 'Join me on Zippyra! Use code ZIPPY123 for 50 points bonus.',
    referrerRewardPoints: 100,
    referredRewardPoints: 50,
  );

  bool applied = false;

  @override
  Future<ReferralInfo> getReferralInfo() async {
    return info;
  }

  @override
  Future<void> applyReferralCode(String referralCode) async {
    applied = true;
  }
}

void main() {
  group('ReferralCubit Tests', () {
    late FakeReferralRepository fakeRepo;
    late ReferralCubit cubit;

    setUp(() {
      fakeRepo = FakeReferralRepository();
      cubit = ReferralCubit(repository: fakeRepo);
    });

    tearDown(() {
      cubit.close();
    });

    test('loadReferralInfo emits ReferralLoading then ReferralLoaded', () async {
      final expectedStates = [
        isA<ReferralLoading>(),
        isA<ReferralLoaded>(),
      ];

      expectLater(cubit.stream, emitsInOrder(expectedStates));
      cubit.loadReferralInfo();
    });

    test('applyReferralCode emits ReferralLoading then ReferralApplySuccess', () async {
      final expectedStates = [
        isA<ReferralLoading>(),
        isA<ReferralApplySuccess>(),
      ];

      expectLater(cubit.stream, emitsInOrder(expectedStates));
      cubit.applyReferralCode('ZIPPY123');
    });
  });

  group('ReferralScreen Widget Tests', () {
    late FakeReferralRepository fakeRepo;

    setUp(() {
      fakeRepo = FakeReferralRepository();
    });

    Widget createWidgetUnderTest() {
      return MaterialApp(
        home: BlocProvider(
          create: (_) => ReferralCubit(repository: fakeRepo),
          child: const ReferralScreen(),
        ),
      );
    }

    testWidgets('renders referral code and share buttons', (tester) async {
      await tester.pumpWidget(createWidgetUnderTest());
      await tester.pumpAndSettle();

      expect(find.text('Refer & Earn'), findsOneWidget);
      expect(find.text('ZIPPY123'), findsOneWidget);
      expect(find.text('YOUR REFERRAL CODE'), findsOneWidget);
      expect(find.text('HOW IT WORKS'), findsOneWidget);
      expect(find.text('Share Code via WhatsApp / Apps'), findsOneWidget);
    });
  });
}
