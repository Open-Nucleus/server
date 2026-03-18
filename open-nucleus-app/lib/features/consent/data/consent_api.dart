import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/consent_models.dart';

/// Low-level HTTP client for all consent-related endpoints.
///
/// Uses [Dio] directly and deserializes responses into the standard
/// [ApiEnvelope] wrapper. Callers handle errors via [DioException].
class ConsentApi {
  final Dio _dio;

  ConsentApi(this._dio);

  // ---------------------------------------------------------------------------
  // List Consents
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/patients/{patientId}/consents`
  ///
  /// Returns a paginated list of consent records for the given patient.
  Future<ApiEnvelope<ConsentListResponse>> listConsents(
    String patientId, {
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.patientConsents(patientId),
      queryParameters: {
        'page': page,
        'per_page': perPage,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ConsentListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Grant Consent
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/patients/{patientId}/consents`
  ///
  /// Grants a new consent for the given patient.
  Future<ApiEnvelope<ConsentGrantResponse>> grantConsent({
    required String patientId,
    required String performerId,
    required String scope,
    required String periodStart,
    required String periodEnd,
    String? category,
  }) async {
    final response = await _dio.post(
      ApiPaths.patientConsents(patientId),
      data: {
        'performer_id': performerId,
        'scope': scope,
        'period': {
          'start': periodStart,
          'end': periodEnd,
        },
        if (category != null && category.isNotEmpty) 'category': category,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ConsentGrantResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Revoke Consent
  // ---------------------------------------------------------------------------

  /// `DELETE /api/v1/consents/{consentId}`
  ///
  /// Revokes an existing consent record.
  Future<void> revokeConsent(String consentId) async {
    await _dio.delete(ApiPaths.consent(consentId));
  }

  // ---------------------------------------------------------------------------
  // Issue VC
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/consents/{consentId}/vc`
  ///
  /// Issues a Verifiable Credential for the given consent.
  Future<ApiEnvelope<ConsentVCResponse>> issueVC(String consentId) async {
    final response = await _dio.post(ApiPaths.consentVc(consentId));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ConsentVCResponse.fromJson(data as Map<String, dynamic>),
    );
  }
}
