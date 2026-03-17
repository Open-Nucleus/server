import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/providers/dio_provider.dart';
import '../data/dashboard_api.dart';
import '../data/dashboard_models.dart';

// ---------------------------------------------------------------------------
// Data layer
// ---------------------------------------------------------------------------

/// Provides the [DashboardApi] HTTP client.
final dashboardApiProvider = Provider<DashboardApi>((ref) {
  final dio = ref.watch(dioProvider);
  return DashboardApi(dio);
});

// ---------------------------------------------------------------------------
// Presentation layer
// ---------------------------------------------------------------------------

/// Fetches all dashboard data from the backend in parallel.
///
/// Returns a [DashboardData] on success or throws on complete failure.
/// Individual sub-requests may fail silently — the model carries nullable
/// fields for partial data display.
final dashboardDataProvider = FutureProvider.autoDispose<DashboardData>((ref) {
  final api = ref.watch(dashboardApiProvider);
  return api.fetchDashboard();
});
