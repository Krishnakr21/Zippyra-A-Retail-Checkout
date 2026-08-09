import '../../domain/entities/notification_preference.dart';

class NotificationPreferenceModel extends NotificationPreference {
  const NotificationPreferenceModel({
    required String notificationType,
    required String channel,
    bool isMandatory = false,
  }) : super(
          notificationType: notificationType,
          channel: channel,
          isMandatory: isMandatory,
        );

  factory NotificationPreferenceModel.fromJson(Map<String, dynamic> json) {
    return NotificationPreferenceModel(
      notificationType: json['notification_type'] as String? ?? '',
      channel: json['channel'] as String? ?? 'BOTH',
      isMandatory: json['is_mandatory'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'notification_type': notificationType,
      'channel': channel,
      'is_mandatory': isMandatory,
    };
  }
}
