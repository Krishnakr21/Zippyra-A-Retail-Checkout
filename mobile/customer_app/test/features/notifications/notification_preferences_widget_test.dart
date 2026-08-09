import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:customer_app/features/notifications/domain/entities/notification_preference.dart';
import 'package:customer_app/features/notifications/domain/entities/notification_item.dart';
import 'package:customer_app/features/notifications/domain/usecases/get_preferences_use_case.dart';
import 'package:customer_app/features/notifications/domain/usecases/update_preference_use_case.dart';
import 'package:customer_app/features/notifications/domain/usecases/get_inbox_use_case.dart';
import 'package:customer_app/features/notifications/domain/usecases/mark_read_use_case.dart';
import 'package:customer_app/features/notifications/domain/usecases/get_unread_count_use_case.dart';
import 'package:customer_app/features/notifications/domain/repositories/notifications_repository.dart';
import 'package:customer_app/features/notifications/presentation/bloc/notifications_bloc.dart';
import 'package:customer_app/features/notifications/presentation/screens/notification_preferences_screen.dart';

class FakeNotificationsRepositoryWidgetTest implements NotificationsRepository {
  @override
  Future<List<NotificationPreference>> getPreferences() async {
    return const [
      NotificationPreference(
        notificationType: 'ORDER_UPDATES',
        channel: 'BOTH',
        isMandatory: true,
      ),
      NotificationPreference(
        notificationType: 'MARKETING',
        channel: 'BOTH',
        isMandatory: false,
      ),
    ];
  }

  @override
  Future<NotificationPreference> updatePreference(String type, String channel) async {
    return NotificationPreference(notificationType: type, channel: channel);
  }

  @override
  Future<List<NotificationItem>> getInbox(int page) async => [];
  @override
  Future<void> markRead(String id) async {}
  @override
  Future<int> getUnreadCount() async => 0;
  @override
  Future<void> registerDeviceToken(String fcmToken, String platform, String deviceId) async {}
  @override
  Future<void> unregisterDeviceToken(String deviceId) async {}
}

void main() {
  late NotificationsBloc bloc;
  late FakeNotificationsRepositoryWidgetTest fakeRepository;

  setUp(() {
    fakeRepository = FakeNotificationsRepositoryWidgetTest();
    bloc = NotificationsBloc(
      getPreferencesUseCase: GetPreferencesUseCase(fakeRepository),
      updatePreferenceUseCase: UpdatePreferenceUseCase(fakeRepository),
      getInboxUseCase: GetInboxUseCase(fakeRepository),
      markReadUseCase: MarkReadUseCase(fakeRepository),
      getUnreadCountUseCase: GetUnreadCountUseCase(fakeRepository),
    );
  });

  testWidgets('mandatory notification type renders without an interactive toggle and displays mandatory note', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: BlocProvider<NotificationsBloc>.value(
          value: bloc,
          child: const NotificationPreferencesScreen(),
        ),
      ),
    );

    await tester.pumpAndSettle();

    // Verify mandatory note text is displayed
    expect(find.byKey(const Key('mandatory_note_text')), findsOneWidget);

    // Verify switch for ORDER_UPDATES is disabled (onChanged == null)
    final orderSwitchFinder = find.byKey(const Key('switch_ORDER_UPDATES'));
    expect(orderSwitchFinder, findsOneWidget);
    final orderSwitch = tester.widget<Switch>(orderSwitchFinder);
    expect(orderSwitch.onChanged, isNull);

    // Verify switch for MARKETING is interactive (onChanged != null)
    final marketingSwitchFinder = find.byKey(const Key('switch_MARKETING'));
    expect(marketingSwitchFinder, findsOneWidget);
    final marketingSwitch = tester.widget<Switch>(marketingSwitchFinder);
    expect(marketingSwitch.onChanged, isNotNull);
  });
}
