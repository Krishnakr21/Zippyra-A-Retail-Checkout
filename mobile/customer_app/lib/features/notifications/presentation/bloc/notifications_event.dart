import 'package:equatable/equatable.dart';

abstract class NotificationsEvent extends Equatable {
  const NotificationsEvent();

  @override
  List<Object?> get props => [];
}

class PreferencesRequested extends NotificationsEvent {}

class PreferenceToggled extends NotificationsEvent {
  final String notificationType;
  final bool enabled;

  const PreferenceToggled({
    required this.notificationType,
    required this.enabled,
  });

  @override
  List<Object?> get props => [notificationType, enabled];
}

class InboxRequested extends NotificationsEvent {
  final bool refresh;

  const InboxRequested({this.refresh = false});

  @override
  List<Object?> get props => [refresh];
}

class InboxNextPageRequested extends NotificationsEvent {}

class NotificationMarkedRead extends NotificationsEvent {
  final String id;

  const NotificationMarkedRead(this.id);

  @override
  List<Object?> get props => [id];
}

class UnreadCountRequested extends NotificationsEvent {}
