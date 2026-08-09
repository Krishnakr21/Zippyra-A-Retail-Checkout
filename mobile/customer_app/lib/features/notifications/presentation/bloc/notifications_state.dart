import 'package:equatable/equatable.dart';
import '../../domain/entities/notification_preference.dart';
import '../../domain/entities/notification_item.dart';

class NotificationsState extends Equatable {
  final List<NotificationPreference> preferences;
  final List<NotificationItem> inbox;
  final bool isPreferencesLoading;
  final bool isInboxLoading;
  final bool hasReachedMaxInbox;
  final int currentPage;
  final int unreadCount;
  final String? errorMessage;

  const NotificationsState({
    this.preferences = const [],
    this.inbox = const [],
    this.isPreferencesLoading = false,
    this.isInboxLoading = false,
    this.hasReachedMaxInbox = false,
    this.currentPage = 1,
    this.unreadCount = 0,
    this.errorMessage,
  });

  NotificationsState copyWith({
    List<NotificationPreference>? preferences,
    List<NotificationItem>? inbox,
    bool? isPreferencesLoading,
    bool? isInboxLoading,
    bool? hasReachedMaxInbox,
    int? currentPage,
    int? unreadCount,
    String? errorMessage,
  }) {
    return NotificationsState(
      preferences: preferences ?? this.preferences,
      inbox: inbox ?? this.inbox,
      isPreferencesLoading: isPreferencesLoading ?? this.isPreferencesLoading,
      isInboxLoading: isInboxLoading ?? this.isInboxLoading,
      hasReachedMaxInbox: hasReachedMaxInbox ?? this.hasReachedMaxInbox,
      currentPage: currentPage ?? this.currentPage,
      unreadCount: unreadCount ?? this.unreadCount,
      errorMessage: errorMessage,
    );
  }

  @override
  List<Object?> get props => [
        preferences,
        inbox,
        isPreferencesLoading,
        isInboxLoading,
        hasReachedMaxInbox,
        currentPage,
        unreadCount,
        errorMessage,
      ];
}
