import 'package:flutter_test/flutter_test.dart';
import 'package:open_nucleus_app/shared/models/api_envelope.dart';

void main() {
  group('ApiEnvelope', () {
    test('fromJson parses a success response with data', () {
      final json = <String, dynamic>{
        'status': 'success',
        'data': {'id': 'patient-001', 'name': 'Test Patient'},
      };

      final envelope = ApiEnvelope<Map<String, dynamic>>.fromJson(
        json,
        (d) => d as Map<String, dynamic>,
      );

      expect(envelope.isSuccess, isTrue);
      expect(envelope.isError, isFalse);
      expect(envelope.status, equals('success'));
      expect(envelope.data, isNotNull);
      expect(envelope.data!['id'], equals('patient-001'));
      expect(envelope.data!['name'], equals('Test Patient'));
      expect(envelope.error, isNull);
    });

    test('fromJson parses an error response', () {
      final json = <String, dynamic>{
        'status': 'error',
        'error': {
          'code': 'NOT_FOUND',
          'message': 'Patient not found',
        },
      };

      final envelope = ApiEnvelope<Map<String, dynamic>>.fromJson(json, null);

      expect(envelope.isError, isTrue);
      expect(envelope.isSuccess, isFalse);
      expect(envelope.data, isNull);
      expect(envelope.error, isNotNull);
      expect(envelope.error!.code, equals('NOT_FOUND'));
      expect(envelope.error!.message, equals('Patient not found'));
    });

    test('fromJson parses pagination metadata', () {
      final json = <String, dynamic>{
        'status': 'success',
        'data': <String, dynamic>{},
        'pagination': {
          'page': 2,
          'per_page': 25,
          'total': 100,
          'total_pages': 4,
        },
      };

      final envelope = ApiEnvelope<Map<String, dynamic>>.fromJson(
        json,
        (d) => d as Map<String, dynamic>,
      );

      expect(envelope.pagination, isNotNull);
      expect(envelope.pagination!.page, equals(2));
      expect(envelope.pagination!.perPage, equals(25));
      expect(envelope.pagination!.total, equals(100));
      expect(envelope.pagination!.totalPages, equals(4));
    });

    test('fromJson parses warnings list', () {
      final json = <String, dynamic>{
        'status': 'success',
        'data': <String, dynamic>{},
        'warnings': [
          {
            'severity': 'high',
            'type': 'drug_interaction',
            'description': 'Potential interaction between Drug A and Drug B',
            'interacting_medication': 'Drug B',
            'source': 'WHO formulary',
          },
          {
            'severity': 'medium',
            'type': 'allergy_alert',
            'description': 'Patient has documented penicillin allergy',
          },
        ],
      };

      final envelope = ApiEnvelope<Map<String, dynamic>>.fromJson(
        json,
        (d) => d as Map<String, dynamic>,
      );

      expect(envelope.warnings, isNotNull);
      expect(envelope.warnings!.length, equals(2));

      final w1 = envelope.warnings![0];
      expect(w1.severity, equals('high'));
      expect(w1.type, equals('drug_interaction'));
      expect(w1.description, contains('Drug A'));
      expect(w1.interactingMedication, equals('Drug B'));
      expect(w1.source, equals('WHO formulary'));

      final w2 = envelope.warnings![1];
      expect(w2.severity, equals('medium'));
      expect(w2.interactingMedication, isNull);
      expect(w2.source, isNull);
    });

    test('fromJson handles null data gracefully', () {
      final json = <String, dynamic>{
        'status': 'success',
        'data': null,
      };

      final envelope = ApiEnvelope<Map<String, dynamic>>.fromJson(
        json,
        (d) => d as Map<String, dynamic>,
      );

      expect(envelope.isSuccess, isTrue);
      expect(envelope.data, isNull);
      expect(envelope.error, isNull);
      expect(envelope.pagination, isNull);
      expect(envelope.warnings, isNull);
      expect(envelope.git, isNull);
      expect(envelope.meta, isNull);
    });

    test('fromJson parses git and meta fields', () {
      final json = <String, dynamic>{
        'status': 'success',
        'data': <String, dynamic>{},
        'git': {
          'commit': 'abc123',
          'message': 'Create patient',
        },
        'meta': {
          'request_id': 'req-001',
          'duration_ms': 42,
          'node_id': 'node-alpha',
        },
      };

      final envelope = ApiEnvelope<Map<String, dynamic>>.fromJson(
        json,
        (d) => d as Map<String, dynamic>,
      );

      expect(envelope.git, isNotNull);
      expect(envelope.git!.commit, equals('abc123'));
      expect(envelope.git!.message, equals('Create patient'));

      expect(envelope.meta, isNotNull);
      expect(envelope.meta!.requestId, equals('req-001'));
      expect(envelope.meta!.durationMs, equals(42));
      expect(envelope.meta!.nodeId, equals('node-alpha'));
    });
  });
}
