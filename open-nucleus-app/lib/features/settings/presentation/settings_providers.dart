import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/smart_models.dart';
import '../../../shared/providers/dio_provider.dart';
import '../data/smart_api.dart';

// ---------------------------------------------------------------------------
// Data layer provider
// ---------------------------------------------------------------------------

/// Provides the [SmartApi] HTTP client.
final smartApiProvider = Provider<SmartApi>((ref) {
  final dio = ref.watch(dioProvider);
  return SmartApi(dio);
});

// ---------------------------------------------------------------------------
// SMART client list
// ---------------------------------------------------------------------------

/// Fetches the list of registered SMART clients.
final smartClientListProvider =
    FutureProvider.autoDispose<ClientListResponse>((ref) async {
  final api = ref.watch(smartApiProvider);
  final envelope = await api.listClients();
  return envelope.data!;
});

// ---------------------------------------------------------------------------
// Theme mode
// ---------------------------------------------------------------------------

/// Holds the currently selected [ThemeMode]. Defaults to [ThemeMode.system].
final themeModePr = StateProvider<ThemeMode>((_) => ThemeMode.system);

// ---------------------------------------------------------------------------
// Server URL
// ---------------------------------------------------------------------------

/// Holds the current server URL. Defaults to the standard local backend.
final serverUrlProvider =
    StateProvider<String>((_) => 'https://localhost:8080');
