import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/anchor_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/anchor_api.dart';

// ---------------------------------------------------------------------------
// Data layer provider
// ---------------------------------------------------------------------------

/// Provides the [AnchorApi] HTTP client.
final anchorApiProvider = Provider<AnchorApi>((ref) {
  final dio = ref.watch(dioProvider);
  return AnchorApi(dio);
});

// ---------------------------------------------------------------------------
// Anchor status (auto-refresh every 10 seconds)
// ---------------------------------------------------------------------------

/// Fetches the current anchor status and auto-refreshes every 10 seconds.
final anchorStatusProvider =
    FutureProvider.autoDispose<AnchorStatusResponse>((ref) async {
  final api = ref.watch(anchorApiProvider);

  final timer = Timer.periodic(const Duration(seconds: 10), (_) {
    ref.invalidateSelf();
  });
  ref.onDispose(timer.cancel);

  final envelope = await api.getStatus();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Anchor history
// ---------------------------------------------------------------------------

/// Fetches the paginated anchor history.
final anchorHistoryProvider =
    FutureProvider.autoDispose<AnchorHistoryResponse>((ref) async {
  final api = ref.watch(anchorApiProvider);
  final envelope = await api.getHistory();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Credential list
// ---------------------------------------------------------------------------

/// Fetches the paginated credential list.
final credentialListProvider =
    FutureProvider.autoDispose<CredentialListResponse>((ref) async {
  final api = ref.watch(anchorApiProvider);
  final envelope = await api.listCredentials();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Backend list
// ---------------------------------------------------------------------------

/// Fetches the list of anchor backends.
final backendListProvider =
    FutureProvider.autoDispose<BackendListResponse>((ref) async {
  final api = ref.watch(anchorApiProvider);
  final envelope = await api.listBackends();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Queue status
// ---------------------------------------------------------------------------

/// Fetches the current anchor queue status.
final queueStatusProvider =
    FutureProvider.autoDispose<QueueStatusResponse>((ref) async {
  final api = ref.watch(anchorApiProvider);
  final envelope = await api.getQueueStatus();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Node DID
// ---------------------------------------------------------------------------

/// Fetches the DID document for this node.
final nodeDIDProvider =
    FutureProvider.autoDispose<DIDDocumentResponse>((ref) async {
  final api = ref.watch(anchorApiProvider);
  final envelope = await api.getNodeDID();
  return envelope.data!;
});
