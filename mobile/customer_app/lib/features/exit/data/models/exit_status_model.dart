class ExitStatusModel {
  final String result;
  final String? gateId;
  final String? reason;

  const ExitStatusModel({
    required this.result,
    this.gateId,
    this.reason,
  });

  factory ExitStatusModel.fromJson(Map<String, dynamic> json) {
    return ExitStatusModel(
      result: json['result'] as String? ?? 'PENDING',
      gateId: json['gate_id'] as String?,
      reason: json['reason'] as String?,
    );
  }
}
