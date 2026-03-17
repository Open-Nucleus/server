import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/formulary_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/formulary_api.dart';

// ---------------------------------------------------------------------------
// Data layer
// ---------------------------------------------------------------------------

/// Provides the [FormularyApi] HTTP client.
final formularyApiProvider = Provider<FormularyApi>((ref) {
  final dio = ref.watch(dioProvider);
  return FormularyApi(dio);
});

// ---------------------------------------------------------------------------
// Formulary Info
// ---------------------------------------------------------------------------

/// Fetches formulary metadata (version, categories, total medications, etc.).
final formularyInfoProvider =
    FutureProvider.autoDispose<FormularyInfoResponse?>((ref) async {
  final api = ref.watch(formularyApiProvider);
  try {
    final envelope = await api.getFormularyInfo();
    return envelope.data;
  } catch (_) {
    return null;
  }
});

// ---------------------------------------------------------------------------
// Medication Search
// ---------------------------------------------------------------------------

/// State for the medication search feature.
class MedicationSearchState {
  final String query;
  final String? category;
  final List<MedicationDetail> results;
  final bool loading;
  final int page;
  final int totalPages;
  final String? errorMessage;

  const MedicationSearchState({
    this.query = '',
    this.category,
    this.results = const [],
    this.loading = false,
    this.page = 1,
    this.totalPages = 1,
    this.errorMessage,
  });

  MedicationSearchState copyWith({
    String? query,
    String? Function()? category,
    List<MedicationDetail>? results,
    bool? loading,
    int? page,
    int? totalPages,
    String? Function()? errorMessage,
  }) {
    return MedicationSearchState(
      query: query ?? this.query,
      category: category != null ? category() : this.category,
      results: results ?? this.results,
      loading: loading ?? this.loading,
      page: page ?? this.page,
      totalPages: totalPages ?? this.totalPages,
      errorMessage:
          errorMessage != null ? errorMessage() : this.errorMessage,
    );
  }
}

/// Notifier that manages medication search state.
class MedicationSearchNotifier extends StateNotifier<MedicationSearchState> {
  MedicationSearchNotifier(this._api) : super(const MedicationSearchState());

  final FormularyApi _api;

  /// Performs a search with optional category filter.
  Future<void> search({String? query, String? Function()? category}) async {
    final newQuery = query ?? state.query;
    final newCategory =
        category != null ? category() : state.category;

    state = state.copyWith(
      query: newQuery,
      category: () => newCategory,
      loading: true,
      page: 1,
      errorMessage: () => null,
    );

    try {
      final envelope = await _api.searchMedications(
        query: newQuery.isNotEmpty ? newQuery : null,
        category: newCategory,
        page: 1,
      );

      if (envelope.isSuccess && envelope.data != null) {
        state = state.copyWith(
          results: envelope.data!.medications,
          totalPages: envelope.data!.totalPages,
          loading: false,
        );
      } else {
        state = state.copyWith(
          loading: false,
          errorMessage: () =>
              envelope.error?.message ?? 'Failed to search medications',
        );
      }
    } catch (e) {
      state = state.copyWith(
        loading: false,
        errorMessage: () => 'Search failed: $e',
      );
    }
  }

  /// Loads the next page of results.
  Future<void> loadPage(int page) async {
    if (page < 1 || page > state.totalPages) return;

    state = state.copyWith(loading: true, page: page);

    try {
      final envelope = await _api.searchMedications(
        query: state.query.isNotEmpty ? state.query : null,
        category: state.category,
        page: page,
      );

      if (envelope.isSuccess && envelope.data != null) {
        state = state.copyWith(
          results: envelope.data!.medications,
          totalPages: envelope.data!.totalPages,
          loading: false,
        );
      } else {
        state = state.copyWith(loading: false);
      }
    } catch (_) {
      state = state.copyWith(loading: false);
    }
  }
}

/// Provider for the medication search notifier.
final medicationSearchProvider =
    StateNotifierProvider.autoDispose<MedicationSearchNotifier, MedicationSearchState>(
        (ref) {
  final api = ref.watch(formularyApiProvider);
  return MedicationSearchNotifier(api);
});

// ---------------------------------------------------------------------------
// Selected Medication
// ---------------------------------------------------------------------------

/// Currently selected medication for the detail pane.
final selectedMedicationProvider =
    StateProvider.autoDispose<MedicationDetail?>((ref) => null);

// ---------------------------------------------------------------------------
// Interaction Checker
// ---------------------------------------------------------------------------

/// State for the interaction checker feature.
class InteractionCheckerState {
  final List<MedicationDetail> selectedMedications;
  final CheckInteractionsResponse? result;
  final bool loading;
  final String? errorMessage;

  const InteractionCheckerState({
    this.selectedMedications = const [],
    this.result,
    this.loading = false,
    this.errorMessage,
  });

  InteractionCheckerState copyWith({
    List<MedicationDetail>? selectedMedications,
    CheckInteractionsResponse? Function()? result,
    bool? loading,
    String? Function()? errorMessage,
  }) {
    return InteractionCheckerState(
      selectedMedications: selectedMedications ?? this.selectedMedications,
      result: result != null ? result() : this.result,
      loading: loading ?? this.loading,
      errorMessage:
          errorMessage != null ? errorMessage() : this.errorMessage,
    );
  }
}

/// Notifier that manages the interaction checker.
class InteractionCheckerNotifier
    extends StateNotifier<InteractionCheckerState> {
  InteractionCheckerNotifier(this._api)
      : super(const InteractionCheckerState());

  final FormularyApi _api;

  void addMedication(MedicationDetail med) {
    if (state.selectedMedications.any((m) => m.code == med.code)) return;
    state = state.copyWith(
      selectedMedications: [...state.selectedMedications, med],
    );
  }

  void removeMedication(String code) {
    state = state.copyWith(
      selectedMedications:
          state.selectedMedications.where((m) => m.code != code).toList(),
    );
  }

  void clear() {
    state = const InteractionCheckerState();
  }

  Future<void> checkInteractions({String patientId = ''}) async {
    if (state.selectedMedications.length < 2) return;

    state = state.copyWith(loading: true, errorMessage: () => null);

    try {
      final request = CheckInteractionsRequest(
        medicationCodes:
            state.selectedMedications.map((m) => m.code).toList(),
        patientId: patientId,
      );

      final envelope = await _api.checkInteractions(request);

      if (envelope.isSuccess && envelope.data != null) {
        state = state.copyWith(
          result: () => envelope.data,
          loading: false,
        );
      } else {
        state = state.copyWith(
          loading: false,
          errorMessage: () =>
              envelope.error?.message ?? 'Interaction check failed',
        );
      }
    } catch (e) {
      state = state.copyWith(
        loading: false,
        errorMessage: () => 'Error: $e',
      );
    }
  }
}

/// Provider for the interaction checker notifier.
final interactionCheckerProvider = StateNotifierProvider.autoDispose<
    InteractionCheckerNotifier, InteractionCheckerState>((ref) {
  final api = ref.watch(formularyApiProvider);
  return InteractionCheckerNotifier(api);
});

// ---------------------------------------------------------------------------
// Stock Info
// ---------------------------------------------------------------------------

/// Fetches stock level and prediction for a given site + medication.
///
/// Returns a tuple of (StockLevelResponse?, StockPredictionResponse?).
final stockInfoProvider = FutureProvider.autoDispose
    .family<({StockLevelResponse? level, StockPredictionResponse? prediction}),
        ({String siteId, String medicationCode})>((ref, params) async {
  final api = ref.watch(formularyApiProvider);

  StockLevelResponse? level;
  StockPredictionResponse? prediction;

  // Fetch both in parallel; each is independent so partial failures are OK.
  await Future.wait<void>([
    () async {
      try {
        final envelope =
            await api.getStockLevel(params.siteId, params.medicationCode);
        level = envelope.data;
      } catch (_) {}
    }(),
    () async {
      try {
        final envelope =
            await api.getStockPrediction(params.siteId, params.medicationCode);
        prediction = envelope.data;
      } catch (_) {}
    }(),
  ]);

  return (level: level, prediction: prediction);
});
