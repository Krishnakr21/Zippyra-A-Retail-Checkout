import 'package:equatable/equatable.dart';

class NotificationItem extends Equatable {
  final String id;
  final String title;
  final String body;
  final String? deepLink;
  final String notificationType;
  final bool isRead;
  final DateTime createdAt;

  const NotificationItem({
    required this.id,
    required this.title,
    required this.body,
    this.deepLink,
    required this.notificationType,
    required this.isRead,
    required this.createdAt,
  });

  NotificationItem copyWith({
    String? id,
    String? title,
    String? body,
    String? deepLink,
    String? notificationType,
    bool? isRead,
    DateTime? createdAt,
  }) {
    return NotificationItem(
      id: id ?? this.id,
      title: title ?? this.title,
      body: body ?? this.body,
      deepLink: deepLink ?? this.deepLink,
      notificationType: notificationType ?? this.notificationType,
      isRead: isRead ?? this.isRead,
      createdAt: createdAt ?? this.createdAt,
    );
  }

  @override
  List<Object?> get props => [
        id,
        title,
        body,
        deepLink,
        notificationType,
        isRead,
        createdAt,
      ];
}
