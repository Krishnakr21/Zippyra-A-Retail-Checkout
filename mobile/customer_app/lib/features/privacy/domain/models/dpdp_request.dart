class DPDPRequest {
  final String id;
  final String requestType;
  final String status;
  final String? detail;
  final String createdAt;

  const DPDPRequest({
    required this.id,
    required this.requestType,
    required this.status,
    this.detail,
    required this.createdAt,
  });

  factory DPDPRequest.fromJson(Map<String, dynamic> json) {
    return DPDPRequest(
      id: json['id'] as String? ?? '',
      requestType: json['request_type'] as String? ?? 'ACCESS',
      status: json['status'] as String? ?? 'RECEIVED',
      detail: json['detail'] as String?,
      createdAt: json['created_at'] as String? ?? '',
    );
  }
}
