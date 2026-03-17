import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../../shared/providers/dio_provider.dart';
import '../data/auth_api.dart';
import '../data/auth_repository.dart';
import 'auth_notifier.dart';
import 'device_notifier.dart';

// ---------------------------------------------------------------------------
// Secure storage
// ---------------------------------------------------------------------------

/// Shared [FlutterSecureStorage] instance used across the app.
final secureStorageProvider = Provider<FlutterSecureStorage>(
  (_) => const FlutterSecureStorage(),
);

// ---------------------------------------------------------------------------
// Data layer
// ---------------------------------------------------------------------------

/// Provides the [AuthApi] HTTP client.
final authApiProvider = Provider<AuthApi>((ref) {
  final dio = ref.watch(dioProvider);
  return AuthApi(dio);
});

/// Provides the [AuthRepository] (API + secure storage).
final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final api = ref.watch(authApiProvider);
  final storage = ref.watch(secureStorageProvider);
  return AuthRepository(api, storage);
});

// ---------------------------------------------------------------------------
// Presentation layer
// ---------------------------------------------------------------------------

/// Manages authentication state (login, logout, token refresh).
final authNotifierProvider =
    StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  final repository = ref.watch(authRepositoryProvider);
  return AuthNotifier(repository);
});

/// Manages the device Ed25519 keypair lifecycle.
final deviceNotifierProvider =
    StateNotifierProvider<DeviceNotifier, DeviceState>((ref) {
  final storage = ref.watch(secureStorageProvider);
  return DeviceNotifier(storage);
});
