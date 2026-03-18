import 'package:flutter_test/flutter_test.dart';
import 'package:open_nucleus_app/shared/models/fhir_primitives.dart';

void main() {
  group('CodeableConcept', () {
    test('fromJson/toJson roundtrip preserves data', () {
      final json = <String, dynamic>{
        'coding': [
          {
            'system': 'http://snomed.info/sct',
            'code': '386661006',
            'display': 'Fever',
          },
          {
            'system': 'http://hl7.org/fhir/sid/icd-10',
            'code': 'R50.9',
            'display': 'Fever, unspecified',
          },
        ],
        'text': 'Fever',
      };

      final concept = CodeableConcept.fromJson(json);

      expect(concept.text, equals('Fever'));
      expect(concept.coding, isNotNull);
      expect(concept.coding!.length, equals(2));
      expect(concept.coding![0].system, equals('http://snomed.info/sct'));
      expect(concept.coding![0].code, equals('386661006'));
      expect(concept.coding![0].display, equals('Fever'));
      expect(concept.coding![1].code, equals('R50.9'));

      // Roundtrip
      final restored = CodeableConcept.fromJson(concept.toJson());
      expect(restored.text, equals(concept.text));
      expect(restored.coding!.length, equals(concept.coding!.length));
      expect(restored.coding![0].code, equals(concept.coding![0].code));
    });

    test('fromJson handles null coding', () {
      final json = <String, dynamic>{
        'text': 'Unknown condition',
      };

      final concept = CodeableConcept.fromJson(json);

      expect(concept.coding, isNull);
      expect(concept.text, equals('Unknown condition'));
    });
  });

  group('HumanName', () {
    test('fromJson with given names', () {
      final json = <String, dynamic>{
        'use': 'official',
        'family': 'Smith',
        'given': ['John', 'James'],
        'text': 'John James Smith',
      };

      final name = HumanName.fromJson(json);

      expect(name.use, equals('official'));
      expect(name.family, equals('Smith'));
      expect(name.given, isNotNull);
      expect(name.given!.length, equals(2));
      expect(name.given![0], equals('John'));
      expect(name.given![1], equals('James'));
      expect(name.text, equals('John James Smith'));
    });

    test('fromJson with minimal fields', () {
      final json = <String, dynamic>{
        'family': 'Doe',
      };

      final name = HumanName.fromJson(json);

      expect(name.family, equals('Doe'));
      expect(name.given, isNull);
      expect(name.use, isNull);
      expect(name.text, isNull);
    });

    test('toJson omits null fields', () {
      const name = HumanName(family: 'Doe');

      final json = name.toJson();

      expect(json.containsKey('family'), isTrue);
      expect(json.containsKey('given'), isFalse);
      expect(json.containsKey('use'), isFalse);
      expect(json.containsKey('text'), isFalse);
    });
  });

  group('Quantity', () {
    test('fromJson/toJson roundtrip', () {
      final json = <String, dynamic>{
        'value': 37.5,
        'unit': 'Cel',
        'system': 'http://unitsofmeasure.org',
        'code': 'Cel',
      };

      final quantity = Quantity.fromJson(json);

      expect(quantity.value, equals(37.5));
      expect(quantity.unit, equals('Cel'));
      expect(quantity.system, equals('http://unitsofmeasure.org'));
      expect(quantity.code, equals('Cel'));

      // Roundtrip
      final restored = Quantity.fromJson(quantity.toJson());
      expect(restored.value, equals(quantity.value));
      expect(restored.unit, equals(quantity.unit));
      expect(restored.system, equals(quantity.system));
      expect(restored.code, equals(quantity.code));
    });

    test('fromJson handles integer values', () {
      final json = <String, dynamic>{
        'value': 120,
        'unit': 'mmHg',
      };

      final quantity = Quantity.fromJson(json);

      expect(quantity.value, equals(120));
      expect(quantity.unit, equals('mmHg'));
      expect(quantity.system, isNull);
      expect(quantity.code, isNull);
    });

    test('toJson omits null fields', () {
      const quantity = Quantity(value: 98.6, unit: 'degF');

      final json = quantity.toJson();

      expect(json.containsKey('value'), isTrue);
      expect(json.containsKey('unit'), isTrue);
      expect(json.containsKey('system'), isFalse);
      expect(json.containsKey('code'), isFalse);
    });
  });

  group('FhirAddress', () {
    test('fromJson/toJson roundtrip', () {
      final json = <String, dynamic>{
        'use': 'home',
        'line': ['123 Main Street', 'Apt 4B'],
        'city': 'Nairobi',
        'state': 'Nairobi County',
        'postalCode': '00100',
        'country': 'KE',
      };

      final address = FhirAddress.fromJson(json);

      expect(address.use, equals('home'));
      expect(address.line, isNotNull);
      expect(address.line!.length, equals(2));
      expect(address.line![0], equals('123 Main Street'));
      expect(address.city, equals('Nairobi'));
      expect(address.state, equals('Nairobi County'));
      expect(address.postalCode, equals('00100'));
      expect(address.country, equals('KE'));

      // Roundtrip
      final restored = FhirAddress.fromJson(address.toJson());
      expect(restored.use, equals(address.use));
      expect(restored.city, equals(address.city));
      expect(restored.line, equals(address.line));
      expect(restored.postalCode, equals(address.postalCode));
      expect(restored.country, equals(address.country));
    });

    test('fromJson handles minimal address', () {
      final json = <String, dynamic>{
        'city': 'Lagos',
        'country': 'NG',
      };

      final address = FhirAddress.fromJson(json);

      expect(address.city, equals('Lagos'));
      expect(address.country, equals('NG'));
      expect(address.use, isNull);
      expect(address.line, isNull);
      expect(address.state, isNull);
      expect(address.postalCode, isNull);
    });

    test('toJson omits null fields', () {
      const address = FhirAddress(city: 'Kampala', country: 'UG');

      final json = address.toJson();

      expect(json.containsKey('city'), isTrue);
      expect(json.containsKey('country'), isTrue);
      expect(json.containsKey('use'), isFalse);
      expect(json.containsKey('line'), isFalse);
      expect(json.containsKey('state'), isFalse);
      expect(json.containsKey('postalCode'), isFalse);
    });
  });

  group('Coding', () {
    test('fromJson/toJson roundtrip', () {
      final json = <String, dynamic>{
        'system': 'http://loinc.org',
        'code': '8310-5',
        'display': 'Body temperature',
      };

      final coding = Coding.fromJson(json);

      expect(coding.system, equals('http://loinc.org'));
      expect(coding.code, equals('8310-5'));
      expect(coding.display, equals('Body temperature'));

      final restored = Coding.fromJson(coding.toJson());
      expect(restored.system, equals(coding.system));
      expect(restored.code, equals(coding.code));
      expect(restored.display, equals(coding.display));
    });
  });

  group('FhirReference', () {
    test('fromJson parses reference fields', () {
      final json = <String, dynamic>{
        'reference': 'Patient/p-001',
        'display': 'Jane Doe',
        'type': 'Patient',
      };

      final ref = FhirReference.fromJson(json);

      expect(ref.reference, equals('Patient/p-001'));
      expect(ref.display, equals('Jane Doe'));
      expect(ref.type, equals('Patient'));
    });
  });

  group('FhirPeriod', () {
    test('fromJson/toJson roundtrip', () {
      final json = <String, dynamic>{
        'start': '2025-01-01T08:00:00Z',
        'end': '2025-01-01T16:00:00Z',
      };

      final period = FhirPeriod.fromJson(json);

      expect(period.start, equals('2025-01-01T08:00:00Z'));
      expect(period.end, equals('2025-01-01T16:00:00Z'));

      final restored = FhirPeriod.fromJson(period.toJson());
      expect(restored.start, equals(period.start));
      expect(restored.end, equals(period.end));
    });
  });
}
