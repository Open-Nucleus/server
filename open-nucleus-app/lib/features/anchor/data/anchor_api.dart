import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/anchor_models.dart';
import '../../../shared/models/api_envelope.dart';

/// Low-level HTTP client for all anchor / integrity endpoints.
///
/// Uses [Dio] directly and deserializes responses into the standard
/// [ApiEnvelope] wrapper. Callers handle errors via [DioException].
class AnchorApi {
  final Dio _dio;

  AnchorApi(this._dio);

  // ---------------------------------------------------------------------------
  // Status
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/anchor/status`
  ///
  /// Returns the current anchoring state, merkle root, queue depth, etc.
  Future<ApiEnvelope<AnchorStatusResponse>> getStatus() async {
    final response = await _dio.get(ApiPaths.anchorStatus);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          AnchorStatusResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Verify
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/anchor/verify`
  ///
  /// Verifies whether a specific git commit has been anchored.
  Future<ApiEnvelope<AnchorVerifyResponse>> verify(String commitHash) async {
    final response = await _dio.post(
      ApiPaths.anchorVerify,
      data: {'commit_hash': commitHash},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          AnchorVerifyResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // History
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/anchor/history`
  ///
  /// Returns a paginated list of anchor records.
  Future<ApiEnvelope<AnchorHistoryResponse>> getHistory({
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.anchorHistory,
      queryParameters: {
        'page': page,
        'per_page': perPage,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          AnchorHistoryResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Trigger
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/anchor/trigger`
  ///
  /// Triggers a manual anchor operation.
  Future<ApiEnvelope<AnchorTriggerResponse>> triggerAnchor() async {
    final response = await _dio.post(ApiPaths.anchorTrigger);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          AnchorTriggerResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // DID — Node
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/anchor/did/node`
  ///
  /// Returns the DID document for this node.
  Future<ApiEnvelope<DIDDocumentResponse>> getNodeDID() async {
    final response = await _dio.get(ApiPaths.anchorDidNode);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          DIDDocumentResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // DID — Device
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/anchor/did/device/{deviceId}`
  ///
  /// Returns the DID document for a specific device.
  Future<ApiEnvelope<DIDDocumentResponse>> getDeviceDID(
    String deviceId,
  ) async {
    final response = await _dio.get(ApiPaths.anchorDidDevice(deviceId));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          DIDDocumentResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // DID — Resolve
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/anchor/did/resolve`
  ///
  /// Resolves an arbitrary DID string to its DID document.
  Future<ApiEnvelope<DIDDocumentResponse>> resolveDID(String did) async {
    final response = await _dio.post(
      ApiPaths.anchorDidResolve,
      data: {'did': did},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          DIDDocumentResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Credentials — Issue
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/anchor/credentials/issue`
  ///
  /// Issues a Verifiable Credential for the given anchor.
  Future<ApiEnvelope<CredentialResponse>> issueCredential(
    IssueCredentialRequest request,
  ) async {
    final response = await _dio.post(
      ApiPaths.anchorCredentialsIssue,
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          CredentialResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Credentials — Verify
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/anchor/credentials/verify`
  ///
  /// Verifies a Verifiable Credential JSON payload.
  Future<ApiEnvelope<CredentialVerificationResponse>> verifyCredential(
    Map<String, dynamic> credentialJson,
  ) async {
    final response = await _dio.post(
      ApiPaths.anchorCredentialsVerify,
      data: {'credential': credentialJson},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => CredentialVerificationResponse.fromJson(
          data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Credentials — List
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/anchor/credentials`
  ///
  /// Returns a paginated list of issued credentials, optionally filtered by
  /// credential type.
  Future<ApiEnvelope<CredentialListResponse>> listCredentials({
    String? credType,
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.anchorCredentials,
      queryParameters: {
        'page': page,
        'per_page': perPage,
        if (credType != null) 'type': credType,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          CredentialListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Backends — List
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/anchor/backends`
  ///
  /// Returns the list of available anchor backends.
  Future<ApiEnvelope<BackendListResponse>> listBackends() async {
    final response = await _dio.get(ApiPaths.anchorBackends);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          BackendListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Backends — Status
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/anchor/backends/{name}`
  ///
  /// Returns the status of a specific anchor backend.
  Future<ApiEnvelope<BackendStatusResponse>> getBackendStatus(
    String name,
  ) async {
    final response = await _dio.get(ApiPaths.anchorBackend(name));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          BackendStatusResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Queue
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/anchor/queue`
  ///
  /// Returns the current anchor queue status.
  Future<ApiEnvelope<QueueStatusResponse>> getQueueStatus() async {
    final response = await _dio.get(ApiPaths.anchorQueue);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          QueueStatusResponse.fromJson(data as Map<String, dynamic>),
    );
  }
}
