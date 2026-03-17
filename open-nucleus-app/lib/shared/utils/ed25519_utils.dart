import 'dart:convert';

import 'package:cryptography/cryptography.dart';

/// Utility helpers for Ed25519 key management.
///
/// Uses the `cryptography` package for key generation, signing, and
/// serialization. Keys are persisted via [flutter_secure_storage] using
/// the [serializeKeypair] / [deserializeKeypair] helpers.
class Ed25519Utils {
  Ed25519Utils._();

  static final _algorithm = Ed25519();

  /// Generates a new Ed25519 keypair.
  static Future<SimpleKeyPair> generateKeypair() async {
    final keypair = await _algorithm.newKeyPair();
    return keypair;
  }

  /// Signs [nonce] with the given [keypair] and returns the signature as
  /// a base64url-encoded string (no padding).
  static Future<String> sign(SimpleKeyPair keypair, String nonce) async {
    final data = utf8.encode(nonce);
    final signature = await _algorithm.sign(data, keyPair: keypair);
    return base64Url.encode(signature.bytes).replaceAll('=', '');
  }

  /// Returns the public key as a base64url-encoded string (no padding).
  static Future<String> getPublicKeyBase64(SimpleKeyPair keypair) async {
    final publicKey = await keypair.extractPublicKey();
    return base64Url.encode(publicKey.bytes).replaceAll('=', '');
  }

  /// Returns the first 8 characters of the hex-encoded public key bytes,
  /// used as a short human-readable fingerprint.
  static Future<String> getFingerprint(SimpleKeyPair keypair) async {
    final publicKey = await keypair.extractPublicKey();
    final hex = publicKey.bytes
        .map((b) => b.toRadixString(16).padLeft(2, '0'))
        .join();
    return hex.substring(0, 8);
  }

  /// Serializes a keypair to a JSON string for secure storage.
  ///
  /// Both the private seed and public key are stored as base64url strings.
  static Future<String> serializeKeypair(SimpleKeyPair keypair) async {
    final privateKey = await keypair.extract();
    final publicKey = await keypair.extractPublicKey();
    final map = {
      'private': base64Url.encode(privateKey.bytes),
      'public': base64Url.encode(publicKey.bytes),
    };
    return jsonEncode(map);
  }

  /// Deserializes a keypair from a JSON string produced by [serializeKeypair].
  static Future<SimpleKeyPair> deserializeKeypair(String json) async {
    final map = jsonDecode(json) as Map<String, dynamic>;
    final privateBytes = base64Url.decode(map['private'] as String);
    final publicBytes = base64Url.decode(map['public'] as String);

    return SimpleKeyPairData(
      privateBytes,
      publicKey: SimplePublicKey(publicBytes, type: KeyPairType.ed25519),
      type: KeyPairType.ed25519,
    );
  }
}
