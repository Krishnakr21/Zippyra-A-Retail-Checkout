import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/entities/notification_preference.dart';
import '../../domain/entities/notification_item.dart';
import '../../domain/usecases/get_preferences_use_case.dart';
import '../../domain/usecases/update_preference_use_case.dart';
import '../../domain/usecases/get_inbox_use_case.dart';
import '../../domain/usecases/mark_read_use_case.dart';
import '../../domain/usecases/get_unread_count_use_case.dart';
import 'notifications_event.dart';
import 'notifications_state.dart';

class NotificationsBloc extends Bloc<NotificationsEvent, NotificationsState> {
  final GetPreferencesUseCase getPreferencesUseCase;
  final UpdatePreferenceUseCase updatePreferenceUseCase;
  final GetInboxUseCase getInboxUseCase;
  final MarkReadUseCase markReadUseCase;
  final GetUnreadCountUseCase getUnreadCountUseCase;

  NotificationsBloc({
    required this.getPreferencesUseCase,
    required this.updatePreferenceUseCase,
    required this.getInboxUseCase,
    required this.markReadUseCase,
    required this.getUnreadCountUseCase,
  }) : super(const NotificationsState()) {
    on<PreferencesRequested>(_onPreferencesRequested);
    on<PreferenceToggled>(_onPreferenceToggled);
    on<InboxRequested>(_onInboxRequested);
    on<InboxNextPageRequested>(_onInboxNextPageRequested);
    on<NotificationMarkedRead>(_onNotificationMarkedRead);
    on<UnreadCountRequested>(_onUnreadCountRequested);
  }

  Future<void> _onPreferencesRequested(
    PreferencesRequested event,
    Emitter<NotificationsState> emit,
  ) async {
    emit(state.copyWith(isPreferencesLoading: true, errorMessage: null));
    try {
      final preferences = await getPreferencesUseCase();
      emit(state.copyWith(
        isPreferencesLoading: false,
        preferences: preferences,
      ));
    } catch (e) {
      emit(state.copyWith(
        isPreferencesLoading: false,
        errorMessage: e.toString(),
      ));
    }
  }

  Future<void> _onPreferenceToggled(
    PreferenceToggled event,
    Emitter<NotificationsState> emit,
  ) async {
    final targetPref = state.preferences.firstWhere(
      (p) => p.notificationType == event.notificationType,
      orElse: () => NotificationPreference(
        notificationType: event.notificationType,
        channel: event.enabled ? 'BOTH' : 'NONE',
      ),
    );

    if (targetPref.isMandatory && !event.enabled) {
      emit(state.copyWith(errorMessage: 'Mandatory notifications cannot be disabled'));
      return;
    }

    final newChannel = event.enabled ? 'BOTH' : 'NONE';
    final previousPreferences = List<NotificationPreference>.from(state.preferences);

    // Optimistic Update
    final updatedPreferences = state.preferences.map((p) {
      if (p.notificationType == event.notificationType) {
        return p.copyWith(channel: newChannel);
      }
      return p;
    }).toList();

    if (!updatedPreferences.any((p) => p.notificationType == event.notificationType)) {
      updatedPreferences.add(NotificationPreference(
        notificationType: event.notificationType,
        channel: newChannel,
      ));
    }

    emit(state.copyWith(preferences: updatedPreferences, errorMessage: null));

    try {
      final updatedPref = await updatePreferenceUseCase(event.notificationType, newChannel);
      final confirmedPrefs = state.preferences.map((p) {
        if (p.notificationType == updatedPref.notificationType) {
          return updatedPref;
        }
        return p;
      }).toList();
      emit(state.copyWith(preferences: confirmedPrefs));
    } catch (e) {
      // Revert on failure
      emit(state.copyWith(
        preferences: previousPreferences,
        errorMessage: e.toString(),
      ));
    }
  }

  Future<void> _onInboxRequested(
    InboxRequested event,
    Emitter<NotificationsState> emit,
  ) async {
    emit(state.copyWith(
      isInboxLoading: true,
      currentPage: 1,
      hasReachedMaxInbox: false,
      errorMessage: null,
    ));
    try {
      final items = await getInboxUseCase(page: 1);
      emit(state.copyWith(
        isInboxLoading: false,
        inbox: items,
        hasReachedMaxInbox: items.length < 20,
        currentPage: 1,
      ));
    } catch (e) {
      emit(state.copyWith(
        isInboxLoading: false,
        errorMessage: e.toString(),
      ));
    }
  }

  Future<void> _onInboxNextPageRequested(
    InboxNextPageRequested event,
    Emitter<NotificationsState> emit,
  ) async {
    if (state.hasReachedMaxInbox || state.isInboxLoading) return;

    final nextPage = state.currentPage + 1;
    try {
      final items = await getInboxUseCase(page: nextPage);
      if (items.isEmpty) {
        emit(state.copyWith(hasReachedMaxInbox: true));
      } else {
        emit(state.copyWith(
          inbox: [...state.inbox, ...items],
          currentPage: nextPage,
          hasReachedMaxInbox: items.length < 20,
        ));
      }
    } catch (e) {
      emit(state.copyWith(errorMessage: e.toString()));
    }
  }

  Future<void> _onNotificationMarkedRead(
    NotificationMarkedRead event,
    Emitter<NotificationsState> emit,
  ) async {
    final targetItem = state.inbox.firstWhere(
      (i) => i.id == event.id,
      orElse: () => NotificationItem(
        id: event.id,
        title: '',
        body: '',
        notificationType: '',
        isRead: false,
        createdAt: DateTime.now(),
      ),
    );

    if (targetItem.isRead) return;

    // Optimistic local update
    final updatedInbox = state.inbox.map((item) {
      if (item.id == event.id) {
        return item.copyWith(isRead: true);
      }
      return item;
    }).toList();

    final newUnreadCount = (state.unreadCount > 0) ? state.unreadCount - 1 : 0;
    emit(state.copyWith(inbox: updatedInbox, unreadCount: newUnreadCount));

    try {
      await markReadUseCase(event.id);
    } catch (_) {}
  }

  Future<void> _onUnreadCountRequested(
    UnreadCountRequested event,
    Emitter<NotificationsState> emit,
  ) async {
    try {
      final count = await getUnreadCountUseCase();
      emit(state.copyWith(unreadCount: count));
    } catch (_) {}
  }
}
