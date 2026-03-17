import '../../../shared/models/alert_models.dart';
import '../../../shared/models/anchor_models.dart';
import '../../../shared/models/sync_models.dart';

/// Aggregated dashboard data fetched in parallel from multiple endpoints.
///
/// Each field is nullable because individual requests may fail independently
/// while the dashboard still renders partial data.
class DashboardData {
  final bool healthy;
  final int patientCount;
  final AlertSummaryResponse? alertSummary;
  final SyncStatusResponse? syncStatus;
  final AnchorStatusResponse? anchorStatus;
  final String? nodeId;
  final String? siteId;

  const DashboardData({
    required this.healthy,
    required this.patientCount,
    this.alertSummary,
    this.syncStatus,
    this.anchorStatus,
    this.nodeId,
    this.siteId,
  });

  /// Creates a default empty state for loading/error scenarios.
  factory DashboardData.empty() {
    return const DashboardData(
      healthy: false,
      patientCount: 0,
    );
  }
}
