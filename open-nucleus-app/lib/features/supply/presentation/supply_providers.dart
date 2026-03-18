import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/supply_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/supply_api.dart';

// ---------------------------------------------------------------------------
// Data layer provider
// ---------------------------------------------------------------------------

/// Provides the [SupplyApi] HTTP client.
final supplyApiProvider = Provider<SupplyApi>((ref) {
  final dio = ref.watch(dioProvider);
  return SupplyApi(dio);
});

// ---------------------------------------------------------------------------
// Inventory
// ---------------------------------------------------------------------------

/// Fetches the paginated inventory list.
final inventoryProvider =
    FutureProvider.autoDispose<InventoryListResponse>((ref) async {
  final api = ref.watch(supplyApiProvider);
  final envelope = await api.getInventory();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Predictions
// ---------------------------------------------------------------------------

/// Fetches stock-out predictions.
final predictionsProvider =
    FutureProvider.autoDispose<PredictionsResponse>((ref) async {
  final api = ref.watch(supplyApiProvider);
  final envelope = await api.getPredictions();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Redistribution
// ---------------------------------------------------------------------------

/// Fetches redistribution suggestions.
final redistributionProvider =
    FutureProvider.autoDispose<RedistributionResponse>((ref) async {
  final api = ref.watch(supplyApiProvider);
  final envelope = await api.getRedistribution();
  return envelope.data!;
});
