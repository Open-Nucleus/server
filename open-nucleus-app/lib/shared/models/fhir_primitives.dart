/// FHIR R4 primitive/complex types used across the app.
///
/// Plain immutable classes with manual fromJson/toJson — no freezed or
/// code generation. These map to the corresponding FHIR data types as
/// defined in http://hl7.org/fhir/R4/datatypes.html.

/// A concept represented by one or more `Coding` entries plus optional text.
class CodeableConcept {
  final List<Coding>? coding;
  final String? text;

  const CodeableConcept({this.coding, this.text});

  factory CodeableConcept.fromJson(Map<String, dynamic> json) {
    return CodeableConcept(
      coding: json['coding'] != null
          ? (json['coding'] as List<dynamic>)
              .map((c) => Coding.fromJson(c as Map<String, dynamic>))
              .toList()
          : null,
      text: json['text'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (coding != null) 'coding': coding!.map((c) => c.toJson()).toList(),
      if (text != null) 'text': text,
    };
  }
}

/// A single coded value with system, code, and display.
class Coding {
  final String? system;
  final String? code;
  final String? display;

  const Coding({this.system, this.code, this.display});

  factory Coding.fromJson(Map<String, dynamic> json) {
    return Coding(
      system: json['system'] as String?,
      code: json['code'] as String?,
      display: json['display'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (system != null) 'system': system,
      if (code != null) 'code': code,
      if (display != null) 'display': display,
    };
  }
}

/// A reference to another FHIR resource.
class FhirReference {
  final String? reference;
  final String? display;
  final String? type;

  const FhirReference({this.reference, this.display, this.type});

  factory FhirReference.fromJson(Map<String, dynamic> json) {
    return FhirReference(
      reference: json['reference'] as String?,
      display: json['display'] as String?,
      type: json['type'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (reference != null) 'reference': reference,
      if (display != null) 'display': display,
      if (type != null) 'type': type,
    };
  }
}

/// A time period with optional start and end.
class FhirPeriod {
  final String? start;
  final String? end;

  const FhirPeriod({this.start, this.end});

  factory FhirPeriod.fromJson(Map<String, dynamic> json) {
    return FhirPeriod(
      start: json['start'] as String?,
      end: json['end'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (start != null) 'start': start,
      if (end != null) 'end': end,
    };
  }
}

/// A human name with use, family, given names, and optional text.
class HumanName {
  final String? use;
  final String? family;
  final List<String>? given;
  final String? text;

  const HumanName({this.use, this.family, this.given, this.text});

  factory HumanName.fromJson(Map<String, dynamic> json) {
    return HumanName(
      use: json['use'] as String?,
      family: json['family'] as String?,
      given: json['given'] != null
          ? (json['given'] as List<dynamic>).map((g) => g as String).toList()
          : null,
      text: json['text'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (use != null) 'use': use,
      if (family != null) 'family': family,
      if (given != null) 'given': given,
      if (text != null) 'text': text,
    };
  }
}

/// An identifier with system, value, and use.
class FhirIdentifier {
  final String? system;
  final String? value;
  final String? use;

  const FhirIdentifier({this.system, this.value, this.use});

  factory FhirIdentifier.fromJson(Map<String, dynamic> json) {
    return FhirIdentifier(
      system: json['system'] as String?,
      value: json['value'] as String?,
      use: json['use'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (system != null) 'system': system,
      if (value != null) 'value': value,
      if (use != null) 'use': use,
    };
  }
}

/// A contact point (phone, email, etc.).
class ContactPoint {
  final String? system;
  final String? value;
  final String? use;

  const ContactPoint({this.system, this.value, this.use});

  factory ContactPoint.fromJson(Map<String, dynamic> json) {
    return ContactPoint(
      system: json['system'] as String?,
      value: json['value'] as String?,
      use: json['use'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (system != null) 'system': system,
      if (value != null) 'value': value,
      if (use != null) 'use': use,
    };
  }
}

/// A physical/postal address.
class FhirAddress {
  final String? use;
  final List<String>? line;
  final String? city;
  final String? state;
  final String? postalCode;
  final String? country;

  const FhirAddress({
    this.use,
    this.line,
    this.city,
    this.state,
    this.postalCode,
    this.country,
  });

  factory FhirAddress.fromJson(Map<String, dynamic> json) {
    return FhirAddress(
      use: json['use'] as String?,
      line: json['line'] != null
          ? (json['line'] as List<dynamic>).map((l) => l as String).toList()
          : null,
      city: json['city'] as String?,
      state: json['state'] as String?,
      postalCode: json['postalCode'] as String?,
      country: json['country'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (use != null) 'use': use,
      if (line != null) 'line': line,
      if (city != null) 'city': city,
      if (state != null) 'state': state,
      if (postalCode != null) 'postalCode': postalCode,
      if (country != null) 'country': country,
    };
  }
}

/// A measured or measurable amount with unit and optional coding.
class Quantity {
  final num? value;
  final String? unit;
  final String? system;
  final String? code;

  const Quantity({this.value, this.unit, this.system, this.code});

  factory Quantity.fromJson(Map<String, dynamic> json) {
    return Quantity(
      value: json['value'] as num?,
      unit: json['unit'] as String?,
      system: json['system'] as String?,
      code: json['code'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (value != null) 'value': value,
      if (unit != null) 'unit': unit,
      if (system != null) 'system': system,
      if (code != null) 'code': code,
    };
  }
}
