import 'package:cryptography/cryptography.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../../shared/utils/ed25519_utils.dart';

// ---------------------------------------------------------------------------
// Storage key
// ---------------------------------------------------------------------------

const _keypairStorageKey = 'device_ed25519_keypair';

// ---------------------------------------------------------------------------
// DeviceState
// ---------------------------------------------------------------------------

/// State for the device keypair lifecycle.
///
/// [DeviceLoading] — loading or generating keypair.
/// [DeviceReady] — keypair available for signing.
/// [DeviceError] — something went wrong.
abstract class DeviceState {
  const DeviceState();
}

class DeviceLoading extends DeviceState {
  const DeviceLoading();
}

class DeviceReady extends DeviceState {
  final SimpleKeyPair keypair;
  final String fingerprint;
  final String publicKeyBase64;

  const DeviceReady({
    required this.keypair,
    required this.fingerprint,
    required this.publicKeyBase64,
  });
}

class DeviceError extends DeviceState {
  final String message;

  const DeviceError(this.message);
}

// ---------------------------------------------------------------------------
// DeviceNotifier
// ---------------------------------------------------------------------------

/// Manages the device Ed25519 keypair.
///
/// On construction, tries to load an existing keypair from
/// [FlutterSecureStorage]. If none is found, generates a new one and
/// persists it.
class DeviceNotifier extends StateNotifier<DeviceState> {
  final FlutterSecureStorage _storage;

  DeviceNotifier(this._storage) : super(const DeviceLoading()) {
    _init();
  }

  Future<void> _init() async {
    try {
      final saved = await _storage.read(key: _keypairStorageKey);

      SimpleKeyPair keypair;
      if (saved != null) {
        keypair = await Ed25519Utils.deserializeKeypair(saved);
      } else {
        keypair = await Ed25519Utils.generateKeypair();
        final serialized = await Ed25519Utils.serializeKeypair(keypair);
        await _storage.write(key: _keypairStorageKey, value: serialized);
      }

      final fingerprint = await Ed25519Utils.getFingerprint(keypair);
      final publicKeyBase64 = await Ed25519Utils.getPublicKeyBase64(keypair);

      state = DeviceReady(
        keypair: keypair,
        fingerprint: fingerprint,
        publicKeyBase64: publicKeyBase64,
      );
    } catch (e) {
      state = DeviceError(e.toString());
    }
  }

  /// Generates a new keypair, replacing the existing one.
  ///
  /// Use with caution: the backend ties device identity to the public key.
  /// Generating a new keypair means the device will need to re-register.
  Future<void> generateNewKeypair() async {
    state = const DeviceLoading();

    try {
      final keypair = await Ed25519Utils.generateKeypair();
      final serialized = await Ed25519Utils.serializeKeypair(keypair);
      await _storage.write(key: _keypairStorageKey, value: serialized);

      final fingerprint = await Ed25519Utils.getFingerprint(keypair);
      final publicKeyBase64 = await Ed25519Utils.getPublicKeyBase64(keypair);

      state = DeviceReady(
        keypair: keypair,
        fingerprint: fingerprint,
        publicKeyBase64: publicKeyBase64,
      );
    } catch (e) {
      state = DeviceError(e.toString());
    }
  }
}
