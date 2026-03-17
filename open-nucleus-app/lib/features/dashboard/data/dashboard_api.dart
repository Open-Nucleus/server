import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/alert_models.dart';
import '../../../shared/models/anchor_models.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/sync_models.dart';
import 'dashboard_models.dart';

/// Low-level HTTP client that fetches all dashboard data in parallel.
///
/// Each endpoint is called concurrently via [Future.wait]. Individual
/// failures are caught so that partial data can still be displayed.
class DashboardApi {
  final Dio _dio;

  DashboardApi(this._dio);

  /// Fetches health, alerts summary, sync status, anchor status, and
  /// patient count in parallel and returns a composite [DashboardData].
  Future<DashboardData> fetchDashboard() async {
    final results = await Future.wait<dynamic>([
      _fetchHealth(), // 0
      _fetchAlertSummary(), // 1
      _fetchSyncStatus(), // 2
      _fetchAnchorStatus(), // 3
      _fetchPatientCount(), // 4
    ], eagerError: false);

    final healthResult = results[0] as _HealthResult?;
    final alertSummary = results[1] as AlertSummaryResponse?;
    final syncStatus = results[2] as SyncStatusResponse?;
    final anchorStatus = results[3] as AnchorStatusResponse?;
    final patientCount = results[4] as int? ?? 0;

    // Node/site ID comes from either health or sync status.
    final nodeId = healthResult?.nodeId ?? syncStatus?.nodeId;
    final siteId = healthResult?.siteId ?? syncStatus?.siteId;

    return DashboardData(
      healthy: healthResult?.healthy ?? false,
      patientCount: patientCount,
      alertSummary: alertSummary,
      syncStatus: syncStatus,
      anchorStatus: anchorStatus,
      nodeId: nodeId,
      siteId: siteId,
    );
  }

  // ---------------------------------------------------------------------------
  // Individual fetchers — each swallows errors and returns null on failure.
  // ---------------------------------------------------------------------------

  Future<_HealthResult?> _fetchHealth() async {
    try {
      final response = await _dio.get(ApiPaths.health);
      final data = response.data as Map<String, dynamic>?;
      if (response.statusCode == 200 && data != null) {
        return _HealthResult(
          healthy: true,
          nodeId: data['node_id'] as String?,
          siteId: data['site_id'] as String?,
        );
      }
      return _HealthResult(healthy: false);
    } catch (_) {
      return null;
    }
  }

  Future<AlertSummaryResponse?> _fetchAlertSummary() async {
    try {
      final response = await _dio.get(ApiPaths.alertsSummary);
      final json = response.data as Map<String, dynamic>;
      final envelope = ApiEnvelope.fromJson(
        json,
        (data) => AlertSummaryResponse.fromJson(data as Map<String, dynamic>),
      );
      return envelope.data;
    } catch (_) {
      return null;
    }
  }

  Future<SyncStatusResponse?> _fetchSyncStatus() async {
    try {
      final response = await _dio.get(ApiPaths.syncStatus);
      final json = response.data as Map<String, dynamic>;
      final envelope = ApiEnvelope.fromJson(
        json,
        (data) => SyncStatusResponse.fromJson(data as Map<String, dynamic>),
      );
      return envelope.data;
    } catch (_) {
      return null;
    }
  }

  Future<AnchorStatusResponse?> _fetchAnchorStatus() async {
    try {
      final response = await _dio.get(ApiPaths.anchorStatus);
      final json = response.data as Map<String, dynamic>;
      final envelope = ApiEnvelope.fromJson(
        json,
        (data) =>
            AnchorStatusResponse.fromJson(data as Map<String, dynamic>),
      );
      return envelope.data;
    } catch (_) {
      return null;
    }
  }

  Future<int?> _fetchPatientCount() async {
    try {
      final response = await _dio.get(
        ApiPaths.patients,
        queryParameters: {'per_page': 1},
      );
      final json = response.data as Map<String, dynamic>;
      final envelope = ApiEnvelope<void>.fromJson(json, null);
      return envelope.pagination?.total ?? 0;
    } catch (_) {
      return null;
    }
  }
}

/// Internal result type for the health check endpoint.
class _HealthResult {
  final bool healthy;
  final String? nodeId;
  final String? siteId;

  const _HealthResult({
    required this.healthy,
    this.nodeId,
    this.siteId,
  });
}
