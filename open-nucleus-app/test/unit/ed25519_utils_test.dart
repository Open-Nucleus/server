import 'dart:convert';

import 'package:cryptography/cryptography.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:open_nucleus_app/shared/utils/ed25519_utils.dart';

void main() {
  group('Ed25519Utils', () {
    test('generateKeypair returns a non-null SimpleKeyPair', () async {
      final keypair = await Ed25519Utils.generateKeypair();

      expect(keypair, isNotNull);
      expect(keypair, isA<SimpleKeyPair>());
    });

    test('sign produces a non-empty base64url string', () async {
      final keypair = await Ed25519Utils.generateKeypair();

      final signature = await Ed25519Utils.sign(keypair, 'test-nonce');

      expect(signature, isNotEmpty);
      // base64url characters only (no padding '=' because we strip them)
      expect(signature, matches(RegExp(r'^[A-Za-z0-9_-]+$')));
    });

    test('getPublicKeyBase64 returns a base64url string', () async {
      final keypair = await Ed25519Utils.generateKeypair();

      final pubKeyB64 = await Ed25519Utils.getPublicKeyBase64(keypair);

      expect(pubKeyB64, isNotEmpty);
      // Should be valid base64url (no padding)
      expect(pubKeyB64, matches(RegExp(r'^[A-Za-z0-9_-]+$')));
    });

    test('getFingerprint returns an 8-character hex string', () async {
      final keypair = await Ed25519Utils.generateKeypair();

      final fingerprint = await Ed25519Utils.getFingerprint(keypair);

      expect(fingerprint, isNotNull);
      expect(fingerprint.length, equals(8));
      // Should be lowercase hex only
      expect(fingerprint, matches(RegExp(r'^[0-9a-f]{8}$')));
    });

    test('serialize/deserialize roundtrip preserves keypair', () async {
      final original = await Ed25519Utils.generateKeypair();

      // Serialize
      final serialized = await Ed25519Utils.serializeKeypair(original);
      expect(serialized, isNotEmpty);

      // Verify it's valid JSON
      final parsed = jsonDecode(serialized) as Map<String, dynamic>;
      expect(parsed, contains('private'));
      expect(parsed, contains('public'));

      // Deserialize
      final restored = await Ed25519Utils.deserializeKeypair(serialized);

      // Verify the public keys match
      final originalPub = await Ed25519Utils.getPublicKeyBase64(original);
      final restoredPub = await Ed25519Utils.getPublicKeyBase64(restored);
      expect(restoredPub, equals(originalPub));

      // Verify the fingerprints match
      final originalFp = await Ed25519Utils.getFingerprint(original);
      final restoredFp = await Ed25519Utils.getFingerprint(restored);
      expect(restoredFp, equals(originalFp));

      // Verify both can sign and produce the same signature
      const nonce = 'roundtrip-test';
      final sig1 = await Ed25519Utils.sign(original, nonce);
      final sig2 = await Ed25519Utils.sign(restored, nonce);
      expect(sig2, equals(sig1));
    });
  });
}
