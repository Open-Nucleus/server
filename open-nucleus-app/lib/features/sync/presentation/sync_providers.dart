import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/conflict_models.dart';
import '../../../shared/models/sync_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/conflict_api.dart';
import '../data/sync_api.dart';

// ---------------------------------------------------------------------------
// Data layer providers
// ---------------------------------------------------------------------------

/// Provides the [SyncApi] HTTP client.
final syncApiProvider = Provider<SyncApi>((ref) {
  final dio = ref.watch(dioProvider);
  return SyncApi(dio);
});

/// Provides the [ConflictApi] HTTP client.
final conflictApiProvider = Provider<ConflictApi>((ref) {
  final dio = ref.watch(dioProvider);
  return ConflictApi(dio);
});

// ---------------------------------------------------------------------------
// Sync status (auto-refresh every 5 seconds)
// ---------------------------------------------------------------------------

/// Fetches the current sync status and auto-refreshes every 5 seconds.
final syncStatusProvider =
    FutureProvider.autoDispose<SyncStatusResponse>((ref) async {
  final api = ref.watch(syncApiProvider);

  // Set up a periodic timer that invalidates this provider every 5s.
  final timer = Timer.periodic(const Duration(seconds: 5), (_) {
    ref.invalidateSelf();
  });
  ref.onDispose(timer.cancel);

  final envelope = await api.getStatus();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Peers
// ---------------------------------------------------------------------------

/// Fetches the list of discovered peer nodes.
final syncPeersProvider =
    FutureProvider.autoDispose<SyncPeersResponse>((ref) async {
  final api = ref.watch(syncApiProvider);
  final envelope = await api.listPeers();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Sync history
// ---------------------------------------------------------------------------

/// Fetches the sync event history.
final syncHistoryProvider =
    FutureProvider.autoDispose<SyncHistoryResponse>((ref) async {
  final api = ref.watch(syncApiProvider);
  final envelope = await api.getHistory();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Conflict list
// ---------------------------------------------------------------------------

/// Fetches the current list of merge conflicts.
final conflictListProvider =
    FutureProvider.autoDispose<ConflictListResponse>((ref) async {
  final api = ref.watch(conflictApiProvider);
  final envelope = await api.listConflicts();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Selected conflict
// ---------------------------------------------------------------------------

/// Holds the ID of the currently selected conflict for the detail pane.
final selectedConflictProvider = StateProvider.autoDispose<String?>(
  (ref) => null,
);

// ---------------------------------------------------------------------------
// Conflict detail
// ---------------------------------------------------------------------------

/// Fetches the detail of a single conflict by its ID.
final conflictDetailProvider =
    FutureProvider.autoDispose.family<ConflictDetail, String>(
  (ref, conflictId) async {
    final api = ref.watch(conflictApiProvider);
    final envelope = await api.getConflict(conflictId);
    return envelope.data!;
  },
);
