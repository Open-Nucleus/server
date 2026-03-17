import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/clinical_models.dart';
import '../../../shared/models/patient_models.dart';

/// Low-level HTTP client for all clinical FHIR sub-resource endpoints.
///
/// Every method that modifies data accepts an optional [breakGlass] flag.
/// When `true`, the `X-Break-Glass: true` header is added to the request,
/// triggering an automatic 4-hour emergency consent with full audit trail.
class ClinicalApi {
  final Dio _dio;

  ClinicalApi(this._dio);

  // ---------------------------------------------------------------------------
  // Encounters
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{patientId}/encounters`
  Future<ApiEnvelope<ClinicalListResponse>> listEncounters(
    String patientId, {
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.encounters(patientId),
      queryParameters: {'page': page, 'per_page': perPage},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => ClinicalListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/patients/{patientId}/encounters/{encounterId}`
  Future<ApiEnvelope<Map<String, dynamic>>> getEncounter(
    String patientId,
    String encounterId,
  ) async {
    final response = await _dio.get(
      ApiPaths.encounter(patientId, encounterId),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => data as Map<String, dynamic>,
    );
  }

  /// `POST /api/v1/patients/{patientId}/encounters`
  Future<ApiEnvelope<WriteResponse>> createEncounter(
    String patientId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.post(
      ApiPaths.encounters(patientId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `PUT /api/v1/patients/{patientId}/encounters/{encounterId}`
  Future<ApiEnvelope<WriteResponse>> updateEncounter(
    String patientId,
    String encounterId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.put(
      ApiPaths.encounter(patientId, encounterId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Observations
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{patientId}/observations`
  Future<ApiEnvelope<ClinicalListResponse>> listObservations(
    String patientId, {
    ObservationFilters? filters,
    int page = 1,
    int perPage = 25,
  }) async {
    final queryParams = <String, dynamic>{
      'page': page,
      'per_page': perPage,
      ...?filters?.toQueryParameters(),
    };

    final response = await _dio.get(
      ApiPaths.observations(patientId),
      queryParameters: queryParams,
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => ClinicalListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/patients/{patientId}/observations/{observationId}`
  Future<ApiEnvelope<Map<String, dynamic>>> getObservation(
    String patientId,
    String observationId,
  ) async {
    final response = await _dio.get(
      ApiPaths.observation(patientId, observationId),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => data as Map<String, dynamic>,
    );
  }

  /// `POST /api/v1/patients/{patientId}/observations`
  Future<ApiEnvelope<WriteResponse>> createObservation(
    String patientId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.post(
      ApiPaths.observations(patientId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Conditions
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{patientId}/conditions`
  Future<ApiEnvelope<ClinicalListResponse>> listConditions(
    String patientId, {
    ConditionFilters? filters,
    int page = 1,
    int perPage = 25,
  }) async {
    final queryParams = <String, dynamic>{
      'page': page,
      'per_page': perPage,
      ...?filters?.toQueryParameters(),
    };

    final response = await _dio.get(
      ApiPaths.conditions(patientId),
      queryParameters: queryParams,
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => ClinicalListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `POST /api/v1/patients/{patientId}/conditions`
  Future<ApiEnvelope<WriteResponse>> createCondition(
    String patientId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.post(
      ApiPaths.conditions(patientId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `PUT /api/v1/patients/{patientId}/conditions/{conditionId}`
  Future<ApiEnvelope<WriteResponse>> updateCondition(
    String patientId,
    String conditionId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.put(
      ApiPaths.condition(patientId, conditionId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Medication Requests
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{patientId}/medication-requests`
  Future<ApiEnvelope<ClinicalListResponse>> listMedicationRequests(
    String patientId, {
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.medicationRequests(patientId),
      queryParameters: {'page': page, 'per_page': perPage},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => ClinicalListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `POST /api/v1/patients/{patientId}/medication-requests`
  Future<ApiEnvelope<WriteResponse>> createMedicationRequest(
    String patientId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.post(
      ApiPaths.medicationRequests(patientId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `PUT /api/v1/patients/{patientId}/medication-requests/{requestId}`
  Future<ApiEnvelope<WriteResponse>> updateMedicationRequest(
    String patientId,
    String requestId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.put(
      ApiPaths.medicationRequest(patientId, requestId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Allergy Intolerances
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{patientId}/allergy-intolerances`
  Future<ApiEnvelope<ClinicalListResponse>> listAllergyIntolerances(
    String patientId, {
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.allergyIntolerances(patientId),
      queryParameters: {'page': page, 'per_page': perPage},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => ClinicalListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `POST /api/v1/patients/{patientId}/allergy-intolerances`
  Future<ApiEnvelope<WriteResponse>> createAllergyIntolerance(
    String patientId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.post(
      ApiPaths.allergyIntolerances(patientId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `PUT /api/v1/patients/{patientId}/allergy-intolerances/{allergyId}`
  Future<ApiEnvelope<WriteResponse>> updateAllergyIntolerance(
    String patientId,
    String allergyId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.put(
      ApiPaths.allergyIntolerance(patientId, allergyId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Immunizations
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{patientId}/immunizations`
  Future<ApiEnvelope<ClinicalListResponse>> listImmunizations(
    String patientId, {
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.immunizations(patientId),
      queryParameters: {'page': page, 'per_page': perPage},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => ClinicalListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/patients/{patientId}/immunizations/{immunizationId}`
  Future<ApiEnvelope<Map<String, dynamic>>> getImmunization(
    String patientId,
    String immunizationId,
  ) async {
    final response = await _dio.get(
      ApiPaths.immunization(patientId, immunizationId),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => data as Map<String, dynamic>,
    );
  }

  /// `POST /api/v1/patients/{patientId}/immunizations`
  Future<ApiEnvelope<WriteResponse>> createImmunization(
    String patientId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.post(
      ApiPaths.immunizations(patientId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Procedures
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{patientId}/procedures`
  Future<ApiEnvelope<ClinicalListResponse>> listProcedures(
    String patientId, {
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.procedures(patientId),
      queryParameters: {'page': page, 'per_page': perPage},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => ClinicalListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/patients/{patientId}/procedures/{procedureId}`
  Future<ApiEnvelope<Map<String, dynamic>>> getProcedure(
    String patientId,
    String procedureId,
  ) async {
    final response = await _dio.get(
      ApiPaths.procedure(patientId, procedureId),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => data as Map<String, dynamic>,
    );
  }

  /// `POST /api/v1/patients/{patientId}/procedures`
  Future<ApiEnvelope<WriteResponse>> createProcedure(
    String patientId,
    Map<String, dynamic> body, {
    bool breakGlass = false,
  }) async {
    final response = await _dio.post(
      ApiPaths.procedures(patientId),
      data: body,
      options: _breakGlassOptions(breakGlass),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Helpers
  // ---------------------------------------------------------------------------

  /// Returns [Options] with the break-glass header when [breakGlass] is true.
  Options? _breakGlassOptions(bool breakGlass) {
    if (!breakGlass) return null;
    return Options(headers: {'X-Break-Glass': 'true'});
  }
}
