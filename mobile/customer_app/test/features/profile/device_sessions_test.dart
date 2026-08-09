import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:customer_app/features/profile/domain/entities/device_session.dart';
import 'package:customer_app/features/profile/domain/repositories/device_sessions_repository.dart';
import 'package:customer_app/features/profile/presentation/bloc/device_sessions_bloc.dart';
import 'package:customer_app/features/profile/presentation/device_sessions_screen.dart';

class FakeDeviceSessionsRepository implements DeviceSessionsRepository {
  List<DeviceSession> sessions = [
    DeviceSession(
      id: 'sess-1',
      deviceId: 'dev-1',
      deviceLabel: 'iPhone 14 (This Device)',
      createdAt: DateTime.now().subtract(const Duration(days: 2)),
      lastUsedAt: DateTime.now(),
      isCurrent: true,
    ),
    DeviceSession(
      id: 'sess-2',
      deviceId: 'dev-2',
      deviceLabel: 'Pixel 7',
      createdAt: DateTime.now().subtract(const Duration(days: 5)),
      lastUsedAt: DateTime.now().subtract(const Duration(hours: 3)),
      isCurrent: false,
    ),
    DeviceSession(
      id: 'sess-3',
      deviceId: 'dev-3',
      deviceLabel: 'iPad Air',
      createdAt: DateTime.now().subtract(const Duration(days: 10)),
      lastUsedAt: DateTime.now().subtract(const Duration(days: 1)),
      isCurrent: false,
    ),
  ];

  @override
  Future<List<DeviceSession>> getDeviceSessions() async {
    return List.from(sessions);
  }

  @override
  Future<void> revokeSession(String sessionId) async {
    sessions.removeWhere((s) => s.id == sessionId);
  }

  @override
  Future<void> revokeAllOtherSessions() async {
    sessions.removeWhere((s) => !s.isCurrent);
  }
}

void main() {
  group('DeviceSessionsBloc Tests', () {
    late FakeDeviceSessionsRepository fakeRepo;
    late DeviceSessionsBloc bloc;

    setUp(() {
      fakeRepo = FakeDeviceSessionsRepository();
      bloc = DeviceSessionsBloc(repository: fakeRepo);
    });

    tearDown(() {
      bloc.close();
    });

    test('LoadDeviceSessions fetches and emits DeviceSessionsLoaded', () async {
      final expectedStates = [
        isA<DeviceSessionsLoading>(),
        isA<DeviceSessionsLoaded>(),
      ];

      expectLater(bloc.stream, emitsInOrder(expectedStates));

      bloc.add(LoadDeviceSessions());
    });

    test('RevokeDeviceSession removes session without full reload', () async {
      bloc.add(LoadDeviceSessions());
      await bloc.stream.firstWhere((s) => s is DeviceSessionsLoaded);

      bloc.add(RevokeDeviceSession('sess-2'));

      await expectLater(
        bloc.stream,
        emitsThrough(predicate<DeviceSessionsState>((state) {
          if (state is DeviceSessionsLoaded) {
            return state.sessions.length == 2 &&
                !state.sessions.any((s) => s.id == 'sess-2');
          }
          return false;
        })),
      );
    });

    test('RevokeAllOtherDeviceSessions clears down to current session only', () async {
      bloc.add(LoadDeviceSessions());
      await bloc.stream.firstWhere((s) => s is DeviceSessionsLoaded);

      bloc.add(RevokeAllOtherDeviceSessions());

      await expectLater(
        bloc.stream,
        emitsThrough(predicate<DeviceSessionsState>((state) {
          if (state is DeviceSessionsLoaded) {
            return state.sessions.length == 1 && state.sessions.first.isCurrent;
          }
          return false;
        })),
      );
    });
  });

  group('DeviceSessionsScreen Widget Tests', () {
    late FakeDeviceSessionsRepository fakeRepo;

    setUp(() {
      fakeRepo = FakeDeviceSessionsRepository();
    });

    Widget createWidgetUnderTest() {
      return MaterialApp(
        home: BlocProvider(
          create: (_) => DeviceSessionsBloc(repository: fakeRepo),
          child: const DeviceSessionsScreen(),
        ),
      );
    }

    testWidgets('renders active sessions list and shows This Device badge', (tester) async {
      await tester.pumpWidget(createWidgetUnderTest());
      await tester.pumpAndSettle();

      expect(find.text('Logged-in Devices'), findsOneWidget);
      expect(find.text('iPhone 14 (This Device)'), findsOneWidget);
      expect(find.text('This Device'), findsOneWidget);
      expect(find.text('Pixel 7'), findsOneWidget);
      expect(find.text('Sign Out All Other Devices'), findsOneWidget);
    });

    testWidgets('clicking Sign Out All Other Devices shows confirmation dialog', (tester) async {
      await tester.pumpWidget(createWidgetUnderTest());
      await tester.pumpAndSettle();

      final signOutAllBtn = find.text('Sign Out All Other Devices');
      await tester.tap(signOutAllBtn);
      await tester.pumpAndSettle();

      expect(find.text('Sign Out All Other Devices?'), findsOneWidget);
      expect(find.text('This will log you out everywhere except this device.'), findsOneWidget);
    });
  });
}
