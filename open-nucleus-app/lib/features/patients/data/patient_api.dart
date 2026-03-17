import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/patient_models.dart';

/// Low-level HTTP client for all patient CRUD and search endpoints.
///
/// Uses [Dio] directly and deserializes responses into the standard
/// [ApiEnvelope] wrapper. Callers handle errors via [DioException].
class PatientApi {
  final Dio _dio;

  PatientApi(this._dio);

  // ---------------------------------------------------------------------------
  // List / Search
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients`
  ///
  /// Returns a paginated list of patient summaries with optional filters.
  Future<ApiEnvelope<List<PatientSummary>>> listPatients({
    int page = 1,
    int perPage = 25,
    String? sort,
    String? gender,
    String? birthDateFrom,
    String? birthDateTo,
    String? siteId,
    String? status,
    bool? hasAlerts,
  }) async {
    final queryParameters = <String, dynamic>{
      'page': page,
      'per_page': perPage,
      if (sort != null) 'sort': sort,
      if (gender != null) 'gender': gender,
      if (birthDateFrom != null) 'birth_date_from': birthDateFrom,
      if (birthDateTo != null) 'birth_date_to': birthDateTo,
      if (siteId != null) 'site_id': siteId,
      if (status != null) 'status': status,
      if (hasAlerts != null) 'has_alerts': hasAlerts.toString(),
    };

    final response = await _dio.get(
      ApiPaths.patients,
      queryParameters: queryParameters,
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => (data as List<dynamic>)
          .map((item) =>
              PatientSummary.fromFhirMap(item as Map<String, dynamic>))
          .toList(),
    );
  }

  /// `GET /api/v1/patients/search?q=...`
  ///
  /// Full-text / blind-index search for patients.
  Future<ApiEnvelope<List<PatientSummary>>> searchPatients({
    required String query,
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.patientsSearch,
      queryParameters: {
        'q': query,
        'page': page,
        'per_page': perPage,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => (data as List<dynamic>)
          .map((item) =>
              PatientSummary.fromFhirMap(item as Map<String, dynamic>))
          .toList(),
    );
  }

  /// `POST /api/v1/patients/match`
  ///
  /// Probabilistic patient matching using demographic data.
  Future<ApiEnvelope<MatchPatientsResponse>> matchPatients(
    MatchPatientsRequest request,
  ) async {
    final response = await _dio.post(
      ApiPaths.patientsMatch,
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          MatchPatientsResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // CRUD
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{id}`
  ///
  /// Returns the full patient bundle (patient + encounters, observations, etc.).
  Future<ApiEnvelope<PatientBundle>> getPatient(String id) async {
    final response = await _dio.get(ApiPaths.patient(id));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => PatientBundle.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `POST /api/v1/patients`
  ///
  /// Creates a new patient from a FHIR Patient resource map.
  Future<ApiEnvelope<WriteResponse>> createPatient(
    Map<String, dynamic> body,
  ) async {
    final response = await _dio.post(
      ApiPaths.patients,
      data: body,
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `PUT /api/v1/patients/{id}`
  ///
  /// Updates an existing patient.
  Future<ApiEnvelope<WriteResponse>> updatePatient(
    String id,
    Map<String, dynamic> body,
  ) async {
    final response = await _dio.put(
      ApiPaths.patient(id),
      data: body,
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `DELETE /api/v1/patients/{id}`
  ///
  /// Soft-deletes a patient.
  Future<ApiEnvelope<WriteResponse>> deletePatient(String id) async {
    final response = await _dio.delete(ApiPaths.patient(id));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WriteResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `DELETE /api/v1/patients/{id}/erase`
  ///
  /// Crypto-erases a patient: destroys encryption key and purges index data.
  Future<ApiEnvelope<EraseResponse>> erasePatient(String id) async {
    final response = await _dio.delete(ApiPaths.patientErase(id));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => EraseResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // History / Timeline
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{id}/history`
  ///
  /// Returns the Git commit history for a patient's records.
  Future<ApiEnvelope<PatientHistoryResponse>> getHistory(String id) async {
    final response = await _dio.get(ApiPaths.patientHistory(id));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          PatientHistoryResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `GET /api/v1/patients/{id}/timeline`
  ///
  /// Returns a chronological timeline of clinical events for the patient.
  Future<ApiEnvelope<List<Map<String, dynamic>>>> getTimeline(
    String id,
  ) async {
    final response = await _dio.get(ApiPaths.patientTimeline(id));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => (data as List<dynamic>)
          .map((item) => item as Map<String, dynamic>)
          .toList(),
    );
  }
}
