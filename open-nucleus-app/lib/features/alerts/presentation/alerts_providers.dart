import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/alert_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/alert_api.dart';

// ---------------------------------------------------------------------------
// Data layer provider
// ---------------------------------------------------------------------------

/// Provides the [AlertApi] HTTP client.
final alertApiProvider = Provider<AlertApi>((ref) {
  final dio = ref.watch(dioProvider);
  return AlertApi(dio);
});

// ---------------------------------------------------------------------------
// Alert list state
// ---------------------------------------------------------------------------

/// Immutable state for the alert list including filters and pagination.
class AlertListState {
  final List<AlertDetail> alerts;
  final int page;
  final int perPage;
  final int totalItems;
  final int totalPages;
  final String? severityFilter;
  final String? statusFilter;
  final bool isLoading;
  final String? error;

  const AlertListState({
    this.alerts = const [],
    this.page = 1,
    this.perPage = 25,
    this.totalItems = 0,
    this.totalPages = 0,
    this.severityFilter,
    this.statusFilter,
    this.isLoading = false,
    this.error,
  });

  AlertListState copyWith({
    List<AlertDetail>? alerts,
    int? page,
    int? perPage,
    int? totalItems,
    int? totalPages,
    String? severityFilter,
    String? statusFilter,
    bool? isLoading,
    String? error,
    bool clearError = false,
    bool clearSeverity = false,
    bool clearStatus = false,
  }) {
    return AlertListState(
      alerts: alerts ?? this.alerts,
      page: page ?? this.page,
      perPage: perPage ?? this.perPage,
      totalItems: totalItems ?? this.totalItems,
      totalPages: totalPages ?? this.totalPages,
      severityFilter:
          clearSeverity ? null : (severityFilter ?? this.severityFilter),
      statusFilter:
          clearStatus ? null : (statusFilter ?? this.statusFilter),
      isLoading: isLoading ?? this.isLoading,
      error: clearError ? null : (error ?? this.error),
    );
  }
}

// ---------------------------------------------------------------------------
// Alert list notifier
// ---------------------------------------------------------------------------

/// StateNotifier managing the alert list: fetching, pagination, and filtering.
class AlertListNotifier extends StateNotifier<AlertListState> {
  final AlertApi _api;

  AlertListNotifier(this._api) : super(const AlertListState()) {
    fetch();
  }

  /// Fetch alerts with current page and filters.
  Future<void> fetch() async {
    state = state.copyWith(isLoading: true, clearError: true);

    try {
      final envelope = await _api.listAlerts(
        page: state.page,
        perPage: state.perPage,
        severity: state.severityFilter,
        status: state.statusFilter,
      );

      final data = envelope.data;
      state = state.copyWith(
        alerts: data?.alerts ?? [],
        totalItems: data?.total ?? 0,
        totalPages: data?.totalPages ?? 0,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Navigate to a specific page.
  Future<void> goToPage(int page) async {
    state = state.copyWith(page: page);
    await fetch();
  }

  /// Change rows per page and reset to page 1.
  Future<void> setPerPage(int perPage) async {
    state = state.copyWith(perPage: perPage, page: 1);
    await fetch();
  }

  /// Set severity filter and reset to page 1.
  Future<void> setSeverityFilter(String? severity) async {
    state = state.copyWith(
      severityFilter: severity,
      page: 1,
      clearSeverity: severity == null,
    );
    await fetch();
  }

  /// Set status filter and reset to page 1.
  Future<void> setStatusFilter(String? status) async {
    state = state.copyWith(
      statusFilter: status,
      page: 1,
      clearStatus: status == null,
    );
    await fetch();
  }

  /// Clear all filters and reset to page 1.
  Future<void> clearFilters() async {
    state = state.copyWith(
      page: 1,
      clearSeverity: true,
      clearStatus: true,
    );
    await fetch();
  }
}

/// Provider for the alert list state.
final alertListProvider =
    StateNotifierProvider.autoDispose<AlertListNotifier, AlertListState>(
  (ref) {
    final api = ref.watch(alertApiProvider);
    return AlertListNotifier(api);
  },
);

// ---------------------------------------------------------------------------
// Alert summary (auto-refresh every 30 seconds)
// ---------------------------------------------------------------------------

/// Fetches aggregate alert counts and auto-refreshes every 30 seconds.
final alertSummaryProvider =
    FutureProvider.autoDispose<AlertSummaryResponse>((ref) async {
  final api = ref.watch(alertApiProvider);

  // Set up a periodic timer that invalidates this provider every 30s.
  final timer = Timer.periodic(const Duration(seconds: 30), (_) {
    ref.invalidateSelf();
  });
  ref.onDispose(timer.cancel);

  final envelope = await api.getSummary();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Selected alert
// ---------------------------------------------------------------------------

/// Holds the ID of the currently selected alert for the detail panel.
final selectedAlertProvider = StateProvider.autoDispose<String?>(
  (ref) => null,
);

// ---------------------------------------------------------------------------
// Alert detail
// ---------------------------------------------------------------------------

/// Fetches the detail of a single alert by its ID.
final alertDetailProvider =
    FutureProvider.autoDispose.family<AlertDetail, String>(
  (ref, alertId) async {
    final api = ref.watch(alertApiProvider);
    final envelope = await api.getAlert(alertId);
    return envelope.data!;
  },
);
