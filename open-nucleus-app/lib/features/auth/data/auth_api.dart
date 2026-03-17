import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/auth_models.dart';

/// Low-level HTTP client for Open Nucleus auth endpoints.
///
/// Uses [Dio] directly and deserializes every response into the standard
/// [ApiEnvelope] wrapper. The caller is responsible for error handling;
/// [DioException]s bubble up as-is.
class AuthApi {
  final Dio _dio;

  AuthApi(this._dio);

  /// `POST /api/v1/auth/login`
  ///
  /// Authenticates a device + practitioner using Ed25519 challenge-response.
  Future<ApiEnvelope<LoginResponse>> login(LoginRequest request) async {
    final response = await _dio.post(
      ApiPaths.authLogin,
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => LoginResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `POST /api/v1/auth/refresh`
  ///
  /// Exchanges a refresh token for a new access + refresh token pair.
  Future<ApiEnvelope<RefreshResponse>> refresh(String refreshToken) async {
    final response = await _dio.post(
      ApiPaths.authRefresh,
      data: {'refresh_token': refreshToken},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => RefreshResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  /// `POST /api/v1/auth/logout`
  ///
  /// Invalidates the given access token on the server.
  Future<void> logout(String token) async {
    await _dio.post(
      ApiPaths.authLogout,
      data: {'token': token},
    );
  }

  /// `GET /api/v1/auth/whoami`
  ///
  /// Returns the identity associated with the current access token.
  Future<ApiEnvelope<WhoamiResponse>> whoami() async {
    final response = await _dio.get(ApiPaths.authWhoami);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => WhoamiResponse.fromJson(data as Map<String, dynamic>),
    );
  }
}
