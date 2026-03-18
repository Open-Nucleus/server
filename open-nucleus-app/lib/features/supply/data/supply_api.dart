import 'package:dio/dio.dart';

import '../../../core/constants/api_paths.dart';
import '../../../shared/models/api_envelope.dart';
import '../../../shared/models/supply_models.dart';

/// Low-level HTTP client for all supply chain endpoints.
///
/// Uses [Dio] directly and deserializes responses into the standard
/// [ApiEnvelope] wrapper. Callers handle errors via [DioException].
class SupplyApi {
  final Dio _dio;

  SupplyApi(this._dio);

  // ---------------------------------------------------------------------------
  // Inventory — List
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/supply/inventory`
  ///
  /// Returns a paginated list of inventory items.
  Future<ApiEnvelope<InventoryListResponse>> getInventory({
    int page = 1,
    int perPage = 25,
  }) async {
    final response = await _dio.get(
      ApiPaths.supplyInventory,
      queryParameters: {
        'page': page,
        'per_page': perPage,
      },
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          InventoryListResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Inventory — Single Item
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/supply/inventory/{itemCode}`
  ///
  /// Returns the detail of a single inventory item.
  Future<ApiEnvelope<InventoryItemDetail>> getInventoryItem(
    String itemCode,
  ) async {
    final response =
        await _dio.get(ApiPaths.supplyInventoryItem(itemCode));

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          InventoryItemDetail.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Deliveries
  // ---------------------------------------------------------------------------

  /// `POST /api/v1/supply/deliveries`
  ///
  /// Records a new delivery of supply items.
  Future<ApiEnvelope<RecordDeliveryResponse>> recordDelivery(
    RecordDeliveryRequest request,
  ) async {
    final response = await _dio.post(
      ApiPaths.supplyDeliveries,
      data: request.toJson(),
    );

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          RecordDeliveryResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Predictions
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/supply/predictions`
  ///
  /// Returns stock-out predictions for all tracked items.
  Future<ApiEnvelope<PredictionsResponse>> getPredictions() async {
    final response = await _dio.get(ApiPaths.supplyPredictions);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          PredictionsResponse.fromJson(data as Map<String, dynamic>),
    );
  }

  // ---------------------------------------------------------------------------
  // Redistribution
  // ---------------------------------------------------------------------------

  /// `GET /api/v1/supply/redistribution`
  ///
  /// Returns redistribution suggestions across sites.
  Future<ApiEnvelope<RedistributionResponse>> getRedistribution() async {
    final response = await _dio.get(ApiPaths.supplyRedistribution);

    return ApiEnvelope.fromJson(
      response.data as Map<String, dynamic>,
      (data) =>
          RedistributionResponse.fromJson(data as Map<String, dynamic>),
    );
  }
}
