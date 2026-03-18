import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/conflict_models.dart';

/// Low-level HTTP client for all conflict resolution endpoints.
///
/// Uses [Dio] directly and deserializes responses into the standard
/// [ApiEnvelope] wrapper. Callers handle errors via [DioException].
class ConflictApi {
  final Dio _dio;

  ConflictApi(this._dio);

  // ---------------------------------------------------------------------------
  // List
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/conflicts`
  ///
  /// Returns a paginated list of merge conflicts.
  Future<ApiEnvelope<ConflictListResponse>> listConflicts({
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.conflicts,
      queryParameters: {
        'page': page,
        'per_page': perPage,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ConflictListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Detail
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/conflicts/{id}`
  ///
  /// Returns the full conflict detail including local and remote resource
  /// versions.
  Future<ApiEnvelope<ConflictDetail>> getConflict(String conflictId) async {
    final response = await _dio.get(ApiPaths.conflict(conflictId));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) => ConflictDetail.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Resolve
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/conflicts/{id}/resolve`
  ///
  /// Resolves a conflict by accepting local, remote, or a merged resource.
  Future<ApiEnvelope<ResolveConflictResponse>> resolveConflict(
    ResolveConflictRequest request,
  ) async {
    final response = await _dio.post(
      ApiPaths.conflictResolve(request.conflictId),
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          ResolveConflictResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Defer
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/conflicts/{id}/defer`
  ///
  /// Defers a conflict resolution with a reason.
  Future<ApiEnvelope<DeferConflictResponse>> deferConflict(
    DeferConflictRequest request,
  ) async {
    final response = await _dio.post(
      ApiPaths.conflictDefer(request.conflictId),
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          DeferConflictResponse.fromJson(data as Map<String, dynamic>),
    );
  }
}
