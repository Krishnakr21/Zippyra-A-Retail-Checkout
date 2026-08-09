import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:zippyra_core/zippyra_core.dart';
import 'package:customer_app/features/notifications/domain/entities/notification_preference.dart';
import 'package:customer_app/features/notifications/domain/entities/notification_item.dart';
import 'package:customer_app/features/notifications/domain/usecases/get_preferences_use_case.dart';
import 'package:customer_app/features/notifications/domain/usecases/update_preference_use_case.dart';
import 'package:customer_app/features/notifications/domain/usecases/get_inbox_use_case.dart';
import 'package:customer_app/features/notifications/domain/usecases/mark_read_use_case.dart';
import 'package:customer_app/features/notifications/domain/usecases/get_unread_count_use_case.dart';
import 'package:customer_app/features/notifications/domain/repositories/notifications_repository.dart';
import 'package:customer_app/features/notifications/presentation/bloc/notifications_bloc.dart';
import 'package:customer_app/features/notifications/presentation/bloc/notifications_event.dart';
import 'package:customer_app/features/notifications/presentation/bloc/notifications_state.dart';

class FakeNotificationsRepository implements NotificationsRepository {
  bool shouldFailUpdate = false;
  int markReadCount = 0;
  bool inboxFetched = false;

  @override
  Future<List<NotificationPreference>> getPreferences() async {
    return [
      const NotificationPreference(notificationType: 'MARKETING', channel: 'BOTH'),
    ];
  }

  @override
  Future<NotificationPreference> updatePreference(String type, String channel) async {
    if (shouldFailUpdate) {
      throw const ServerFailure('Network error');
    }
    return NotificationPreference(notificationType: type, channel: channel);
  }

  @override
  Future<List<NotificationItem>> getInbox(int page) async {
    inboxFetched = true;
    return [
      NotificationItem(
        id: 'notif-1',
        title: 'Order Status',
        body: 'Shipped',
        notificationType: 'ORDER_UPDATES',
        isRead: false,
        createdAt: DateTime.now(),
      ),
    ];
  }

  @override
  Future<void> markRead(String id) async {
    markReadCount++;
  }

  @override
  Future<int> getUnreadCount() async {
    return 3;
  }

  @override
  Future<void> registerDeviceToken(String fcmToken, String platform, String deviceId) async {}

  @override
  Future<void> unregisterDeviceToken(String deviceId) async {}
}

void main() {
  late NotificationsBloc bloc;
  late FakeNotificationsRepository fakeRepository;

  setUp(() {
    fakeRepository = FakeNotificationsRepository();
    bloc = NotificationsBloc(
      getPreferencesUseCase: GetPreferencesUseCase(fakeRepository),
      updatePreferenceUseCase: UpdatePreferenceUseCase(fakeRepository),
      getInboxUseCase: GetInboxUseCase(fakeRepository),
      markReadUseCase: MarkReadUseCase(fakeRepository),
      getUnreadCountUseCase: GetUnreadCountUseCase(fakeRepository),
    );
  });

  group('NotificationsBloc Tests', () {
    blocTest<NotificationsBloc, NotificationsState>(
      'PreferenceToggled optimistically updates the UI, reverts on API failure',
      build: () {
        fakeRepository.shouldFailUpdate = true;
        return bloc;
      },
      seed: () => const NotificationsState(
        preferences: [
          NotificationPreference(notificationType: 'MARKETING', channel: 'BOTH'),
        ],
      ),
      act: (b) => b.add(const PreferenceToggled(
        notificationType: 'MARKETING',
        enabled: false,
      )),
      expect: () => [
        // Optimistic update state: channel = NONE
        const NotificationsState(
          preferences: [
            NotificationPreference(notificationType: 'MARKETING', channel: 'NONE'),
          ],
          errorMessage: null,
        ),
        // Reverted state on failure: channel = BOTH
        const NotificationsState(
          preferences: [
            NotificationPreference(notificationType: 'MARKETING', channel: 'BOTH'),
          ],
          errorMessage: 'Network error',
        ),
      ],
    );

    blocTest<NotificationsBloc, NotificationsState>(
      'unread count decreases correctly after NotificationMarkedRead without requiring a full inbox refetch',
      build: () => bloc,
      seed: () => NotificationsState(
        unreadCount: 3,
        inbox: [
          NotificationItem(
            id: 'notif-1',
            title: 'Order Status',
            body: 'Shipped',
            notificationType: 'ORDER_UPDATES',
            isRead: false,
            createdAt: DateTime.now(),
          ),
          NotificationItem(
            id: 'notif-2',
            title: 'Loyalty Bonus',
            body: '100 points added',
            notificationType: 'LOYALTY_UPDATES',
            isRead: false,
            createdAt: DateTime.now(),
          ),
        ],
      ),
      act: (b) => b.add(const NotificationMarkedRead('notif-1')),
      expect: () => [
        predicate<NotificationsState>((state) {
          final item = state.inbox.firstWhere((i) => i.id == 'notif-1');
          return state.unreadCount == 2 && item.isRead == true;
        }),
      ],
      verify: (_) {
        expect(fakeRepository.markReadCount, 1);
        expect(fakeRepository.inboxFetched, false); // No full refetch called
      },
    );
  });
}
