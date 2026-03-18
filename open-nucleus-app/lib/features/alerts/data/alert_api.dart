import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/alert_models.dart';
import '../../../shared/models/api_envelope.dart';

/// Low-level HTTP client for all Sentinel alert endpoints.
///
/// Uses [Dio] directly and deserializes responses into the standard
/// [ApiEnvelope] wrapper. Callers handle errors via [DioException].
class AlertApi {
  final Dio _dio;

  AlertApi(this._dio);

  // ---------------------------------------------------------------------------
  // List
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/alerts`
  ///
  /// Returns a paginated list of alerts with optional severity and status
  /// filters.
  Future<ApiEnvelope<AlertListResponse>> listAlerts({
    int page = 1,
    int perPage = 25,
    String? severity,
    String? status,
  }) async {
    final response = await _dio.get(
      ApiPaths.alerts,
      queryParameters: {
        'page': page,
        'per_page': perPage,
        if (severity != null) 'severity': severity,
        if (status != null) 'status': status,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          AlertListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Summary
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/alerts/summary`
  ///
  /// Returns aggregate alert counts by severity and acknowledgement status.
  Future<ApiEnvelope<AlertSummaryResponse>> getSummary() async {
    final response = await _dio.get(ApiPaths.alertsSummary);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          AlertSummaryResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Detail
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/alerts/{id}`
  ///
  /// Returns the full detail of a single alert.
  Future<ApiEnvelope<AlertDetail>> getAlert(String alertId) async {
    final response = await _dio.get(ApiPaths.alert(alertId));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => AlertDetail.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Acknowledge
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/alerts/{id}/acknowledge`
  ///
  /// Acknowledges an alert. Returns the updated alert detail.
  Future<ApiEnvelope<AlertDetail>> acknowledgeAlert(
    String alertId,
  ) async {
    final response = await _dio.post(ApiPaths.alertAcknowledge(alertId));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => AlertDetail.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Dismiss
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/alerts/{id}/dismiss`
  ///
  /// Dismisses an alert with a reason. Returns the updated alert detail.
  Future<ApiEnvelope<AlertDetail>> dismissAlert(
    String alertId,
    String reason,
  ) async {
    final response = await _dio.post(
      ApiPaths.alertDismiss(alertId),
      data: {'reason': reason},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => AlertDetail.fromJson(data as Map<String, dynamic>),
    );
  }
}
