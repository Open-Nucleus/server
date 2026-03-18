import 'package:flutter_test/flutter_test.dart';
import 'package:open_nucleus_app/shared/models/auth_models.dart';

void main() {
  group('LoginRequest', () {
    test('toJson serializes all fields correctly', () {
      const request = LoginRequest(
        deviceId: 'dev-001',
        publicKey: 'base64-public-key',
        challengeResponse: ChallengeResponseDTO(
          nonce: 'login:2025-01-01T00:00:00Z',
          signature: 'base64-signature',
          timestamp: '2025-01-01T00:00:00Z',
        ),
        practitionerId: 'practitioner-001',
      );

      final json = request.toJson();

      expect(json['device_id'], equals('dev-001'));
      expect(json['public_key'], equals('base64-public-key'));
      expect(json['practitioner_id'], equals('practitioner-001'));
      expect(json['challenge_response'], isA<Map<String, dynamic>>());
      expect(json['challenge_response']['nonce'],
          equals('login:2025-01-01T00:00:00Z'));
      expect(
          json['challenge_response']['signature'], equals('base64-signature'));
      expect(json['challenge_response']['timestamp'],
          equals('2025-01-01T00:00:00Z'));
    });
  });

  group('LoginResponse', () {
    test('fromJson deserializes all fields correctly', () {
      final json = <String, dynamic>{
        'token': 'jwt-access-token',
        'expires_at': '2025-01-01T01:00:00Z',
        'refresh_token': 'jwt-refresh-token',
        'role': {
          'code': 'MO',
          'display': 'Medical Officer',
          'permissions': ['patient:read', 'patient:write', 'formulary:read'],
        },
        'site_id': 'site-alpha',
        'node_id': 'node-001',
      };

      final response = LoginResponse.fromJson(json);

      expect(response.token, equals('jwt-access-token'));
      expect(response.expiresAt, equals('2025-01-01T01:00:00Z'));
      expect(response.refreshToken, equals('jwt-refresh-token'));
      expect(response.siteId, equals('site-alpha'));
      expect(response.nodeId, equals('node-001'));
      expect(response.role.code, equals('MO'));
      expect(response.role.display, equals('Medical Officer'));
      expect(response.role.permissions.length, equals(3));
    });
  });

  group('RoleDTO', () {
    test('fromJson deserializes correctly', () {
      final json = <String, dynamic>{
        'code': 'CO',
        'display': 'Commanding Officer',
        'permissions': ['admin:all', 'patient:read', 'sync:trigger'],
      };

      final role = RoleDTO.fromJson(json);

      expect(role.code, equals('CO'));
      expect(role.display, equals('Commanding Officer'));
      expect(role.permissions, hasLength(3));
      expect(role.permissions, contains('admin:all'));
      expect(role.permissions, contains('patient:read'));
      expect(role.permissions, contains('sync:trigger'));
    });

    test('toJson roundtrip preserves data', () {
      final original = RoleDTO.fromJson(<String, dynamic>{
        'code': 'NR',
        'display': 'Nurse',
        'permissions': ['patient:read'],
      });

      final json = original.toJson();
      final restored = RoleDTO.fromJson(json);

      expect(restored.code, equals(original.code));
      expect(restored.display, equals(original.display));
      expect(restored.permissions, equals(original.permissions));
    });
  });

  group('WhoamiResponse', () {
    test('fromJson deserializes correctly', () {
      final json = <String, dynamic>{
        'subject': 'practitioner-001',
        'node_id': 'node-alpha',
        'site_id': 'site-bravo',
        'role': {
          'code': 'MO',
          'display': 'Medical Officer',
          'permissions': ['patient:read', 'patient:write'],
        },
      };

      final whoami = WhoamiResponse.fromJson(json);

      expect(whoami.subject, equals('practitioner-001'));
      expect(whoami.nodeId, equals('node-alpha'));
      expect(whoami.siteId, equals('site-bravo'));
      expect(whoami.role, isNotNull);
      expect(whoami.role.code, equals('MO'));
      expect(whoami.role.permissions, hasLength(2));
    });
  });
}
