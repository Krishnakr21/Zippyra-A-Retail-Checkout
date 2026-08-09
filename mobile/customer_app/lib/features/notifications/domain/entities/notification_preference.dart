import 'package:equatable/equatable.dart';

class NotificationPreference extends Equatable {
  final String notificationType;
  final String channel;
  final bool isMandatory;

  const NotificationPreference({
    required this.notificationType,
    required this.channel,
    this.isMandatory = false,
  });

  bool get isEnabled => channel != 'NONE';

  NotificationPreference copyWith({
    String? notificationType,
    String? channel,
    bool? isMandatory,
  }) {
    return NotificationPreference(
      notificationType: notificationType ?? this.notificationType,
      channel: channel ?? this.channel,
      isMandatory: isMandatory ?? this.isMandatory,
    );
  }

  @override
  List<Object?> get props => [notificationType, channel, isMandatory];
}
