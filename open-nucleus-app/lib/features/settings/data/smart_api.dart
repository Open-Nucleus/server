import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/smart_models.dart';

/// Low-level HTTP client for SMART on FHIR client management endpoints.
///
/// Uses [Dio] directly and deserializes responses into the standard
/// [ApiEnvelope] wrapper. Callers handle errors via [DioException].
class SmartApi {
  final Dio _dio;

  SmartApi(this._dio);

  // ---------------------------------------------------------------------------
  // List Clients
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/smart/clients`
  ///
  /// Returns the list of registered SMART clients.
  Future<ApiEnvelope<ClientListResponse>> listClients() async {
    final response = await _dio.get(ApiPaths.smartClients);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ClientListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Get Client
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/smart/clients/{clientId}`
  ///
  /// Returns the details of a single SMART client.
  Future<ApiEnvelope<ClientResponse>> getClient(String clientId) async {
    final response = await _dio.get(ApiPaths.smartClient(clientId));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ClientResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Register Client
  // ---------------------------------------------------------------------------

  /// `POST /auth/smart/register`
  ///
  /// Registers a new SMART client.
  Future<ApiEnvelope<ClientResponse>> registerClient(
    RegisterClientRequest request,
  ) async {
    final response = await _dio.post(
      ApiPaths.smartClients,
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ClientResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Update Client
  // ---------------------------------------------------------------------------

  /// `PUT /api/v1/smart/clients/{clientId}`
  ///
  /// Updates a SMART client's status and scope.
  Future<ApiEnvelope<ClientResponse>> updateClient(
    String clientId,
    UpdateClientRequest request,
  ) async {
    final response = await _dio.put(
      ApiPaths.smartClient(clientId),
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ClientResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Delete Client
  // ---------------------------------------------------------------------------

  /// `DELETE /api/v1/smart/clients/{clientId}`
  ///
  /// Deletes a registered SMART client.
  Future<void> deleteClient(String clientId) async {
    await _dio.delete(ApiPaths.smartClient(clientId));
  }
}
