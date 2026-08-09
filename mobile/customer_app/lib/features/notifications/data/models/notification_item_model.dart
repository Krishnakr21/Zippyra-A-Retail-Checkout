import '../../domain/entities/notification_item.dart';

class NotificationItemModel extends NotificationItem {
  const NotificationItemModel({
    required String id,
    required String title,
    required String body,
    String? deepLink,
    required String notificationType,
    required bool isRead,
    required DateTime createdAt,
  }) : super(
          id: id,
          title: title,
          body: body,
          deepLink: deepLink,
          notificationType: notificationType,
          isRead: isRead,
          createdAt: createdAt,
        );

  factory NotificationItemModel.fromJson(Map<String, dynamic> json) {
    return NotificationItemModel(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      body: json['body'] as String? ?? '',
      deepLink: json['deep_link'] as String?,
      notificationType: json['notification_type'] as String? ?? '',
      isRead: json['read_at'] != null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'body': body,
      'deep_link': deepLink,
      'notification_type': notificationType,
      'is_read': isRead,
      'created_at': createdAt.toIso8601String(),
    };
  }
}
