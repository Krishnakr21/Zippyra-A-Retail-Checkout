class DeviceSession {
  final String id;
  final String deviceId;
  final String deviceLabel;
  final DateTime createdAt;
  final DateTime? lastUsedAt;
  final bool isCurrent;

  const DeviceSession({
    required this.id,
    required this.deviceId,
    required this.deviceLabel,
    required this.createdAt,
    this.lastUsedAt,
    required this.isCurrent,
  });
}
