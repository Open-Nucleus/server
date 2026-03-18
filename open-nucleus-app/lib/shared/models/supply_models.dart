/// Supply chain DTOs matching Go `service.InventoryListResponse`,
/// `service.RecordDeliveryRequest`, `service.PredictionsResponse`, etc.

/// Paginated inventory list from `GET /api/v1/supply/inventory`.
class InventoryListResponse {
  final List<InventoryItemDetail> items;
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const InventoryListResponse({
    required this.items,
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory InventoryListResponse.fromJson(Map<String, dynamic> json) {
    return InventoryListResponse(
      items: (json['items'] as List<dynamic>)
          .map((i) => InventoryItemDetail.fromJson(i as Map<String, dynamic>))
          .toList(),
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'items': items.map((i) => i.toJson()).toList(),
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

/// A single supply inventory item.
class InventoryItemDetail {
  final String itemCode;
  final String display;
  final int quantity;
  final String unit;
  final String siteId;
  final String lastUpdated;
  final int reorderLevel;

  const InventoryItemDetail({
    required this.itemCode,
    required this.display,
    required this.quantity,
    required this.unit,
    required this.siteId,
    required this.lastUpdated,
    required this.reorderLevel,
  });

  factory InventoryItemDetail.fromJson(Map<String, dynamic> json) {
    return InventoryItemDetail(
      itemCode: json['item_code'] as String,
      display: json['display'] as String,
      quantity: (json['quantity'] as num).toInt(),
      unit: json['unit'] as String,
      siteId: json['site_id'] as String,
      lastUpdated: json['last_updated'] as String,
      reorderLevel: (json['reorder_level'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'item_code': itemCode,
      'display': display,
      'quantity': quantity,
      'unit': unit,
      'site_id': siteId,
      'last_updated': lastUpdated,
      'reorder_level': reorderLevel,
    };
  }
}

/// Body of `POST /api/v1/supply/deliveries`.
class RecordDeliveryRequest {
  final String siteId;
  final List<DeliveryItem> items;
  final String receivedBy;
  final String deliveryDate;

  const RecordDeliveryRequest({
    required this.siteId,
    required this.items,
    required this.receivedBy,
    required this.deliveryDate,
  });

  factory RecordDeliveryRequest.fromJson(Map<String, dynamic> json) {
    return RecordDeliveryRequest(
      siteId: json['site_id'] as String,
      items: (json['items'] as List<dynamic>)
          .map((i) => DeliveryItem.fromJson(i as Map<String, dynamic>))
          .toList(),
      receivedBy: json['received_by'] as String,
      deliveryDate: json['delivery_date'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'site_id': siteId,
      'items': items.map((i) => i.toJson()).toList(),
      'received_by': receivedBy,
      'delivery_date': deliveryDate,
    };
  }
}

/// A single item in a delivery.
class DeliveryItem {
  final String itemCode;
  final int quantity;
  final String unit;
  final String batchNumber;
  final String expiryDate;

  const DeliveryItem({
    required this.itemCode,
    required this.quantity,
    required this.unit,
    required this.batchNumber,
    required this.expiryDate,
  });

  factory DeliveryItem.fromJson(Map<String, dynamic> json) {
    return DeliveryItem(
      itemCode: json['item_code'] as String,
      quantity: (json['quantity'] as num).toInt(),
      unit: json['unit'] as String,
      batchNumber: json['batch_number'] as String,
      expiryDate: json['expiry_date'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'item_code': itemCode,
      'quantity': quantity,
      'unit': unit,
      'batch_number': batchNumber,
      'expiry_date': expiryDate,
    };
  }
}

/// Response from `POST /api/v1/supply/deliveries`.
class RecordDeliveryResponse {
  final String deliveryId;
  final int itemsRecorded;

  const RecordDeliveryResponse({
    required this.deliveryId,
    required this.itemsRecorded,
  });

  factory RecordDeliveryResponse.fromJson(Map<String, dynamic> json) {
    return RecordDeliveryResponse(
      deliveryId: json['delivery_id'] as String,
      itemsRecorded: (json['items_recorded'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'delivery_id': deliveryId,
      'items_recorded': itemsRecorded,
    };
  }
}

/// Response from `GET /api/v1/supply/predictions`.
class PredictionsResponse {
  final List<SupplyPrediction> predictions;

  const PredictionsResponse({required this.predictions});

  factory PredictionsResponse.fromJson(Map<String, dynamic> json) {
    return PredictionsResponse(
      predictions: (json['predictions'] as List<dynamic>)
          .map((p) => SupplyPrediction.fromJson(p as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'predictions': predictions.map((p) => p.toJson()).toList(),
    };
  }
}

/// A single supply prediction for an item.
class SupplyPrediction {
  final String itemCode;
  final String display;
  final int currentQuantity;
  final int predictedDaysRemaining;
  final String riskLevel;
  final String recommendedAction;

  const SupplyPrediction({
    required this.itemCode,
    required this.display,
    required this.currentQuantity,
    required this.predictedDaysRemaining,
    required this.riskLevel,
    required this.recommendedAction,
  });

  factory SupplyPrediction.fromJson(Map<String, dynamic> json) {
    return SupplyPrediction(
      itemCode: json['item_code'] as String,
      display: json['display'] as String,
      currentQuantity: (json['current_quantity'] as num).toInt(),
      predictedDaysRemaining:
          (json['predicted_days_remaining'] as num).toInt(),
      riskLevel: json['risk_level'] as String,
      recommendedAction: json['recommended_action'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'item_code': itemCode,
      'display': display,
      'current_quantity': currentQuantity,
      'predicted_days_remaining': predictedDaysRemaining,
      'risk_level': riskLevel,
      'recommended_action': recommendedAction,
    };
  }
}

/// Response from `GET /api/v1/supply/redistribution`.
class RedistributionResponse {
  final List<RedistributionSuggestion> suggestions;

  const RedistributionResponse({required this.suggestions});

  factory RedistributionResponse.fromJson(Map<String, dynamic> json) {
    return RedistributionResponse(
      suggestions: (json['suggestions'] as List<dynamic>)
          .map((s) =>
              RedistributionSuggestion.fromJson(s as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'suggestions': suggestions.map((s) => s.toJson()).toList(),
    };
  }
}

/// A single redistribution suggestion between two sites.
class RedistributionSuggestion {
  final String itemCode;
  final String fromSite;
  final String toSite;
  final int suggestedQuantity;
  final String rationale;

  const RedistributionSuggestion({
    required this.itemCode,
    required this.fromSite,
    required this.toSite,
    required this.suggestedQuantity,
    required this.rationale,
  });

  factory RedistributionSuggestion.fromJson(Map<String, dynamic> json) {
    return RedistributionSuggestion(
      itemCode: json['item_code'] as String,
      fromSite: json['from_site'] as String,
      toSite: json['to_site'] as String,
      suggestedQuantity: (json['suggested_quantity'] as num).toInt(),
      rationale: json['rationale'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'item_code': itemCode,
      'from_site': fromSite,
      'to_site': toSite,
      'suggested_quantity': suggestedQuantity,
      'rationale': rationale,
    };
  }
}
