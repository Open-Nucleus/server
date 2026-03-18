import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/sync_models.dart';

/// Low-level HTTP client for all sync-related endpoints.
///
/// Uses [Dio] directly and deserializes responses into the standard
/// [ApiEnvelope] wrapper. Callers handle errors via [DioException].
class SyncApi {
  final Dio _dio;

  SyncApi(this._dio);

  // ---------------------------------------------------------------------------
  // Status
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/sync/status`
  ///
  /// Returns the current sync state, last sync time, pending changes, and
  /// node/site identifiers.
  Future<ApiEnvelope<SyncStatusResponse>> getStatus() async {
    final response = await _dio.get(ApiPaths.syncStatus);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          SyncStatusResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Peers
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/sync/peers`
  ///
  /// Returns the list of discovered peer nodes.
  Future<ApiEnvelope<SyncPeersResponse>> listPeers() async {
    final response = await _dio.get(ApiPaths.syncPeers);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          SyncPeersResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Trigger
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/sync/trigger`
  ///
  /// Triggers a sync operation against the specified target node.
  Future<ApiEnvelope<SyncTriggerResponse>> triggerSync(
    String targetNode,
  ) async {
    final response = await _dio.post(
      ApiPaths.syncTrigger,
      data: {'target_node': targetNode},
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          SyncTriggerResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // History
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/sync/history`
  ///
  /// Returns a paginated list of sync events.
  Future<ApiEnvelope<SyncHistoryResponse>> getHistory({
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.syncHistory,
      queryParameters: {
        'page': page,
        'per_page': perPage,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          SyncHistoryResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Bundle Export
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/sync/bundle/export`
  ///
  /// Exports a FHIR bundle containing the specified resource types since the
  /// given timestamp.
  Future<ApiEnvelope<BundleExportResponse>> exportBundle({
    required List<String> resourceTypes,
    required String since,
  }) async {
    final response = await _dio.post(
      ApiPaths.syncBundleExport,
      data: BundleExportRequest(
        resourceTypes: resourceTypes,
        since: since,
      ).toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          BundleExportResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Bundle Import
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/sync/bundle/import`
  ///
  /// Imports a FHIR bundle into the local node.
  Future<ApiEnvelope<BundleImportResponse>> importBundle({
    required String bundleData,
    required String format,
    required String author,
    required String nodeId,
    required String siteId,
  }) async {
    final response = await _dio.post(
      ApiPaths.syncBundleImport,
      data: BundleImportRequest(
        bundleData: bundleData,
        format: format,
        author: author,
        nodeId: nodeId,
        siteId: siteId,
      ).toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          BundleImportResponse.fromJson(data as Map<String, dynamic>),
    );
  }
}
