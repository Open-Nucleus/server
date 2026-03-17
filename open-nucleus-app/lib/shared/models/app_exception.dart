/// Application-level exception that wraps HTTP errors, Dio failures, and
/// business logic errors into a single, consistent type.
///
/// Mirrors the backend's `ErrorBody` structure but adds an optional
/// [statusCode] for transport-level context.
class AppException implements Exception {
  /// Machine-readable error code (e.g. "VALIDATION_ERROR", "UNAUTHORIZED").
  final String code;

  /// Human-readable description of what went wrong.
  final String message;

  /// HTTP status code, when the error originated from an HTTP response.
  final int? statusCode;

  /// Optional structured details (validation errors, debug info, etc.).
  final dynamic details;

  const AppException({
    required this.code,
    required this.message,
    this.statusCode,
    this.details,
  });

  @override
  String toString() =>
      'AppException($code, $message${statusCode != null ? ', status=$statusCode' : ''})';
}
