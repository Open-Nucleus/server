import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/patient_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/clinical_api.dart';
import '../data/patient_api.dart';

// ---------------------------------------------------------------------------
// Data layer providers
// ---------------------------------------------------------------------------

/// Provides the [PatientApi] HTTP client.
final patientApiProvider = Provider<PatientApi>((ref) {
  final dio = ref.watch(dioProvider);
  return PatientApi(dio);
});

/// Provides the [ClinicalApi] HTTP client.
final clinicalApiProvider = Provider<ClinicalApi>((ref) {
  final dio = ref.watch(dioProvider);
  return ClinicalApi(dio);
});

// ---------------------------------------------------------------------------
// Patient list state
// ---------------------------------------------------------------------------

/// Immutable state for the patient list screen.
class PatientListState {
  final List<PatientSummary> patients;
  final int page;
  final int perPage;
  final int totalItems;
  final int totalPages;
  final PatientListFilters filters;
  final bool isLoading;
  final String? error;

  const PatientListState({
    this.patients = const [],
    this.page = 1,
    this.perPage = 25,
    this.totalItems = 0,
    this.totalPages = 0,
    this.filters = const PatientListFilters(),
    this.isLoading = false,
    this.error,
  });

  PatientListState copyWith({
    List<PatientSummary>? patients,
    int? page,
    int? perPage,
    int? totalItems,
    int? totalPages,
    PatientListFilters? filters,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) {
    return PatientListState(
      patients: patients ?? this.patients,
      page: page ?? this.page,
      perPage: perPage ?? this.perPage,
      totalItems: totalItems ?? this.totalItems,
      totalPages: totalPages ?? this.totalPages,
      filters: filters ?? this.filters,
      isLoading: isLoading ?? this.isLoading,
      error: clearError ? null : (error ?? this.error),
    );
  }
}

/// Filter parameters for patient list queries.
class PatientListFilters {
  final String? gender;
  final String? birthDateFrom;
  final String? birthDateTo;
  final String? siteId;
  final String? status;
  final bool? hasAlerts;
  final String? sort;

  const PatientListFilters({
    this.gender,
    this.birthDateFrom,
    this.birthDateTo,
    this.siteId,
    this.status,
    this.hasAlerts,
    this.sort,
  });

  PatientListFilters copyWith({
    String? gender,
    String? birthDateFrom,
    String? birthDateTo,
    String? siteId,
    String? status,
    bool? hasAlerts,
    String? sort,
    bool clearGender = false,
    bool clearBirthDateFrom = false,
    bool clearBirthDateTo = false,
    bool clearSiteId = false,
    bool clearStatus = false,
    bool clearHasAlerts = false,
  }) {
    return PatientListFilters(
      gender: clearGender ? null : (gender ?? this.gender),
      birthDateFrom:
          clearBirthDateFrom ? null : (birthDateFrom ?? this.birthDateFrom),
      birthDateTo:
          clearBirthDateTo ? null : (birthDateTo ?? this.birthDateTo),
      siteId: clearSiteId ? null : (siteId ?? this.siteId),
      status: clearStatus ? null : (status ?? this.status),
      hasAlerts: clearHasAlerts ? null : (hasAlerts ?? this.hasAlerts),
      sort: sort ?? this.sort,
    );
  }

  /// Returns true if any filter is active.
  bool get hasActiveFilters =>
      gender != null ||
      birthDateFrom != null ||
      birthDateTo != null ||
      siteId != null ||
      status != null ||
      hasAlerts != null;
}

// ---------------------------------------------------------------------------
// Patient list notifier
// ---------------------------------------------------------------------------

/// StateNotifier managing the patient list: fetching, pagination, filtering.
class PatientListNotifier extends StateNotifier<PatientListState> {
  final PatientApi _api;

  PatientListNotifier(this._api) : super(const PatientListState()) {
    fetch();
  }

  /// Fetch patients with current page and filters.
  Future<void> fetch() async {
    state = state.copyWith(isLoading: true, clearError: true);

    try {
      final envelope = await _api.listPatients(
        page: state.page,
        perPage: state.perPage,
        sort: state.filters.sort,
        gender: state.filters.gender,
        birthDateFrom: state.filters.birthDateFrom,
        birthDateTo: state.filters.birthDateTo,
        siteId: state.filters.siteId,
        status: state.filters.status,
        hasAlerts: state.filters.hasAlerts,
      );

      state = state.copyWith(
        patients: envelope.data ?? [],
        totalItems: envelope.pagination?.total ?? 0,
        totalPages: envelope.pagination?.totalPages ?? 0,
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

  /// Apply new filters and reset to page 1.
  Future<void> applyFilters(PatientListFilters filters) async {
    state = state.copyWith(filters: filters, page: 1);
    await fetch();
  }

  /// Clear all filters and reset to page 1.
  Future<void> clearFilters() async {
    state = state.copyWith(
      filters: const PatientListFilters(),
      page: 1,
    );
    await fetch();
  }
}

/// Provider for the patient list state.
final patientListProvider =
    StateNotifierProvider.autoDispose<PatientListNotifier, PatientListState>(
  (ref) {
    final api = ref.watch(patientApiProvider);
    return PatientListNotifier(api);
  },
);

// ---------------------------------------------------------------------------
// Debounced patient search
// ---------------------------------------------------------------------------

/// Holds the current search query string.
final patientSearchQueryProvider = StateProvider.autoDispose<String>(
  (ref) => '',
);

/// Searches patients with a debounced query string.
///
/// Returns null when the query is empty (caller should show the normal list).
/// Returns the search results when the query is non-empty.
final patientSearchProvider =
    FutureProvider.autoDispose<List<PatientSummary>?>((ref) async {
  final query = ref.watch(patientSearchQueryProvider);

  if (query.trim().isEmpty) return null;

  // Debounce: wait 300ms before actually searching.
  final completer = Completer<void>();
  final timer = Timer(const Duration(milliseconds: 300), completer.complete);
  ref.onDispose(timer.cancel);
  await completer.future;

  final api = ref.watch(patientApiProvider);
  final envelope = await api.searchPatients(query: query.trim());
  return envelope.data ?? [];
});
