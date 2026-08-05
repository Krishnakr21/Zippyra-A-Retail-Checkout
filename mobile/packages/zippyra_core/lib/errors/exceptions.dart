class ApiException implements Exception {
  final String message;
  final String? code;
  final int? statusCode;

  ApiException(this.message, {this.code, this.statusCode});

  @override
  String toString() => 'ApiException: $message (code: $code, statusCode: $statusCode)';
}
