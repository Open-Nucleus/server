import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/formulary_models.dart';

/// Low-level HTTP client for all formulary endpoints.
///
/// Provides medication search, detail lookup, interaction checking,
/// allergy conflict detection, stock management, and redistribution.
class FormularyApi {
  final Dio _dio;

  FormularyApi(this._dio);

  // ---------------------------------------------------------------------------
  // Medication Search / Browse
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/formulary/medications?q=&category=&page=&per_page=`
  ///
  /// Searches medications by name or code, optionally filtered by category.
  Future<ApiEnvelope<MedicationListResponse>> searchMedications({
    String? query,
    String? category,
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.formularyMedications,
      queryParameters: {
        if (query != null && query.isNotEmpty) 'q': query,
        if (category != null && category.isNotEmpty) 'category': category,
        'page': page,
        'per_page': perPage,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          MedicationListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/formulary/medications/{code}`
  ///
  /// Returns detailed information about a single medication.
  Future<ApiEnvelope<MedicationDetail>> getMedication(String code) async {
    final response = await _dio.get(ApiPaths.formularyMedication(code));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => MedicationDetail.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/formulary/medications/category/{category}`
  ///
  /// Lists medications filtered by therapeutic category.
  Future<ApiEnvelope<MedicationListResponse>> listByCategory(
    String category, {
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.formularyMedicationsByCategory(category),
      queryParameters: {
        'page': page,
        'per_page': perPage,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          MedicationListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Interaction / Safety Checks
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/formulary/check-interactions`
  ///
  /// Checks for drug-drug interactions, allergy conflicts, and dosing warnings.
  Future<ApiEnvelope<CheckInteractionsResponse>> checkInteractions(
    CheckInteractionsRequest request,
  ) async {
    final response = await _dio.post(
      ApiPaths.formularyCheckInteractions,
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          CheckInteractionsResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `POST /api/v1/formulary/check-allergies`
  ///
  /// Checks for allergy conflicts between medications and known allergies.
  Future<ApiEnvelope<List<AllergyAlertDTO>>> checkAllergyConflicts({
    required List<String> medicationCodes,
    required List<String> allergyCodes,
  }) async {
    final response = await _dio.post(
      ApiPaths.formularyCheckAllergies,
      data: {
        'medication_codes': medicationCodes,
        'allergy_codes': allergyCodes,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => (data as List<dynamic>)
          .map((a) => AllergyAlertDTO.fromJson(a as Map<String, dynamic>))
          .toList(),
    );
  }

  // ---------------------------------------------------------------------------
  // Stock Management
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/formulary/stock/{siteId}/{medicationCode}`
  ///
  /// Returns current stock level for a medication at a given site.
  Future<ApiEnvelope<StockLevelResponse>> getStockLevel(
    String siteId,
    String medicationCode,
  ) async {
    final response =
        await _dio.get(ApiPaths.formularyStock(siteId, medicationCode));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => StockLevelResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/formulary/stock/{siteId}/{medicationCode}/prediction`
  ///
  /// Returns stock prediction (days remaining, risk level, etc.).
  Future<ApiEnvelope<StockPredictionResponse>> getStockPrediction(
    String siteId,
    String medicationCode,
  ) async {
    final response = await _dio
        .get(ApiPaths.formularyStockPrediction(siteId, medicationCode));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          StockPredictionResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Formulary Info / Redistribution
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/formulary/info`
  ///
  /// Returns formulary metadata (version, total medications, categories, etc.).
  Future<ApiEnvelope<FormularyInfoResponse>> getFormularyInfo() async {
    final response = await _dio.get(ApiPaths.formularyInfo);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          FormularyInfoResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/formulary/redistribution/{medicationCode}`
  ///
  /// Returns redistribution suggestions for a medication across sites.
  Future<ApiEnvelope<FormularyRedistributionResponse>>
      getRedistributionSuggestions(String medicationCode) async {
    final response =
        await _dio.get(ApiPaths.formularyRedistribution(medicationCode));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => FormularyRedistributionResponse.fromJson(
          data as Map<String, dynamic>),
    );
  }
}
