import 'package:flutter_test/flutter_test.dart';
import 'package:open_nucleus_app/shared/models/api_envelope.dart';
import 'package:open_nucleus_app/shared/models/patient_models.dart';

void main() {
  group('PatientSummary', () {
    test('fromFhirMap extracts correct fields from a FHIR Patient resource',
        () {
      final fhirMap = <String, dynamic>{
        'resourceType': 'Patient',
        'id': 'patient-abc',
        'active': true,
        'gender': 'female',
        'birthDate': '1990-05-15',
        'name': [
          {
            'use': 'official',
            'family': 'Doe',
            'given': ['Jane', 'Marie'],
          },
        ],
        'meta': {
          'source': 'site-alpha',
          'lastUpdated': '2025-03-01T12:00:00Z',
        },
      };

      final summary = PatientSummary.fromFhirMap(fhirMap);

      expect(summary.id, equals('patient-abc'));
      expect(summary.familyName, equals('Doe'));
      expect(summary.givenNames, equals(['Jane', 'Marie']));
      expect(summary.gender, equals('female'));
      expect(summary.birthDate, equals('1990-05-15'));
      expect(summary.active, isTrue);
      expect(summary.siteId, equals('site-alpha'));
      expect(summary.lastUpdated, equals('2025-03-01T12:00:00Z'));
      expect(summary.displayName, equals('Jane Marie Doe'));
    });

    test('fromFhirMap handles missing name gracefully', () {
      final fhirMap = <String, dynamic>{
        'id': 'patient-noname',
        'active': false,
      };

      final summary = PatientSummary.fromFhirMap(fhirMap);

      expect(summary.id, equals('patient-noname'));
      expect(summary.familyName, isNull);
      expect(summary.givenNames, isNull);
      expect(summary.active, isFalse);
      expect(summary.displayName, isEmpty);
    });
  });

  group('PatientBundle', () {
    test('fromJson deserializes patient and sub-resources', () {
      final json = <String, dynamic>{
        'patient': {'resourceType': 'Patient', 'id': 'p-001'},
        'encounters': [
          {'resourceType': 'Encounter', 'id': 'enc-001'},
          {'resourceType': 'Encounter', 'id': 'enc-002'},
        ],
        'observations': [
          {'resourceType': 'Observation', 'id': 'obs-001'},
        ],
        'conditions': <Map<String, dynamic>>[],
        'medication_requests': null,
        'allergy_intolerances': null,
        'flags': null,
      };

      final bundle = PatientBundle.fromJson(json);

      expect(bundle.patient['id'], equals('p-001'));
      expect(bundle.encounters.length, equals(2));
      expect(bundle.observations.length, equals(1));
      expect(bundle.conditions, isEmpty);
      expect(bundle.medicationRequests, isEmpty);
      expect(bundle.allergyIntolerances, isEmpty);
      expect(bundle.flags, isEmpty);
    });
  });

  group('WriteResponse', () {
    test('fromJson deserializes with git info', () {
      final json = <String, dynamic>{
        'resource': {
          'resourceType': 'Patient',
          'id': 'p-new',
        },
        'git': {
          'commit': 'deadbeef',
          'message': 'Create Patient p-new',
        },
      };

      final response = WriteResponse.fromJson(json);

      expect(response.resource, isNotNull);
      expect(response.resource!['id'], equals('p-new'));
      expect(response.git, isNotNull);
      expect(response.git!.commit, equals('deadbeef'));
      expect(response.git!.message, equals('Create Patient p-new'));
    });

    test('fromJson handles null resource and git', () {
      final json = <String, dynamic>{};

      final response = WriteResponse.fromJson(json);

      expect(response.resource, isNull);
      expect(response.git, isNull);
    });
  });

  group('HistoryEntry', () {
    test('fromJson deserializes all fields', () {
      final json = <String, dynamic>{
        'commit_hash': 'abc123def456',
        'timestamp': '2025-03-01T10:00:00Z',
        'author': 'practitioner-001',
        'node': 'node-alpha',
        'site': 'site-bravo',
        'operation': 'CREATE',
        'resource_type': 'Patient',
        'resource_id': 'patient-001',
        'message': 'Create Patient patient-001',
      };

      final entry = HistoryEntry.fromJson(json);

      expect(entry.commitHash, equals('abc123def456'));
      expect(entry.timestamp, equals('2025-03-01T10:00:00Z'));
      expect(entry.author, equals('practitioner-001'));
      expect(entry.node, equals('node-alpha'));
      expect(entry.site, equals('site-bravo'));
      expect(entry.operation, equals('CREATE'));
      expect(entry.resourceType, equals('Patient'));
      expect(entry.resourceId, equals('patient-001'));
      expect(entry.message, equals('Create Patient patient-001'));
    });

    test('toJson roundtrip preserves data', () {
      final original = HistoryEntry.fromJson(<String, dynamic>{
        'commit_hash': 'abc123',
        'timestamp': '2025-03-01T10:00:00Z',
        'author': 'doc-1',
        'node': 'n1',
        'site': 's1',
        'operation': 'UPDATE',
        'resource_type': 'Observation',
        'resource_id': 'obs-1',
        'message': 'Update vitals',
      });

      final json = original.toJson();
      final restored = HistoryEntry.fromJson(json);

      expect(restored.commitHash, equals(original.commitHash));
      expect(restored.operation, equals(original.operation));
      expect(restored.resourceType, equals(original.resourceType));
      expect(restored.message, equals(original.message));
    });
  });
}
