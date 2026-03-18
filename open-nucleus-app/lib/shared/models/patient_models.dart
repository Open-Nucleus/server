import 'api_envelope.dart';

/// Patient DTOs matching the Go backend service layer.
///
/// Raw FHIR payloads are kept as `Map<String, dynamic>` — the full resource
/// lives in Git and is returned as-is by the API.

/// Convenience view-model extracted from a raw FHIR Patient map for display
/// in list views and search results.
class PatientSummary {
  final String id;
  final String? familyName;
  final List<String>? givenNames;
  final String? gender;
  final String? birthDate;
  final bool active;
  final String? siteId;
  final String? lastUpdated;
  final bool hasAlerts;

  const PatientSummary({
    required this.id,
    this.familyName,
    this.givenNames,
    this.gender,
    this.birthDate,
    this.active = true,
    this.siteId,
    this.lastUpdated,
    this.hasAlerts = false,
  });

  /// Extracts a [PatientSummary] from a raw FHIR Patient resource map.
  factory PatientSummary.fromFhirMap(Map<String, dynamic> map) {
    String? familyName;
    List<String>? givenNames;

    final names = map['name'] as List<dynamic>?;
    if (names != null && names.isNotEmpty) {
      final name = names.first as Map<String, dynamic>;
      familyName = name['family'] as String?;
      givenNames = (name['given'] as List<dynamic>?)
          ?.map((g) => g as String)
          .toList();
    }

    return PatientSummary(
      id: map['id'] as String,
      familyName: familyName,
      givenNames: givenNames,
      gender: map['gender'] as String?,
      birthDate: map['birthDate'] as String?,
      active: map['active'] as bool? ?? true,
      siteId: (map['meta'] as Map<String, dynamic>?)?['source'] as String?,
      lastUpdated:
          (map['meta'] as Map<String, dynamic>?)?['lastUpdated'] as String?,
      hasAlerts: false,
    );
  }

  factory PatientSummary.fromJson(Map<String, dynamic> json) {
    return PatientSummary(
      id: json['id'] as String,
      familyName: json['family_name'] as String?,
      givenNames: (json['given_names'] as List<dynamic>?)
          ?.map((g) => g as String)
          .toList(),
      gender: json['gender'] as String?,
      birthDate: json['birth_date'] as String?,
      active: json['active'] as bool? ?? true,
      siteId: json['site_id'] as String?,
      lastUpdated: json['last_updated'] as String?,
      hasAlerts: json['has_alerts'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      if (familyName != null) 'family_name': familyName,
      if (givenNames != null) 'given_names': givenNames,
      if (gender != null) 'gender': gender,
      if (birthDate != null) 'birth_date': birthDate,
      'active': active,
      if (siteId != null) 'site_id': siteId,
      if (lastUpdated != null) 'last_updated': lastUpdated,
      'has_alerts': hasAlerts,
    };
  }

  /// Full display name: "Given Family".
  String get displayName {
    final parts = <String>[];
    if (givenNames != null) parts.addAll(givenNames!);
    if (familyName != null) parts.add(familyName!);
    return parts.join(' ');
  }
}

/// Full patient bundle returned by `GET /api/v1/patients/{id}`.
///
/// Matches Go `service.PatientBundle`. All sub-resources are raw FHIR maps.
class PatientBundle {
  final Map<String, dynamic> patient;
  final List<Map<String, dynamic>> encounters;
  final List<Map<String, dynamic>> observations;
  final List<Map<String, dynamic>> conditions;
  final List<Map<String, dynamic>> medicationRequests;
  final List<Map<String, dynamic>> allergyIntolerances;
  final List<Map<String, dynamic>> flags;

  const PatientBundle({
    required this.patient,
    this.encounters = const [],
    this.observations = const [],
    this.conditions = const [],
    this.medicationRequests = const [],
    this.allergyIntolerances = const [],
    this.flags = const [],
  });

  factory PatientBundle.fromJson(Map<String, dynamic> json) {
    return PatientBundle(
      patient: json['patient'] as Map<String, dynamic>,
      encounters: _mapList(json['encounters']),
      observations: _mapList(json['observations']),
      conditions: _mapList(json['conditions']),
      medicationRequests: _mapList(json['medication_requests']),
      allergyIntolerances: _mapList(json['allergy_intolerances']),
      flags: _mapList(json['flags']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'patient': patient,
      'encounters': encounters,
      'observations': observations,
      'conditions': conditions,
      'medication_requests': medicationRequests,
      'allergy_intolerances': allergyIntolerances,
      'flags': flags,
    };
  }
}

/// Response from write operations (create/update/delete).
///
/// Matches Go `service.WriteResponse`.
class WriteResponse {
  final Map<String, dynamic>? resource;
  final GitInfo? git;

  const WriteResponse({this.resource, this.git});

  factory WriteResponse.fromJson(Map<String, dynamic> json) {
    return WriteResponse(
      resource: json['resource'] as Map<String, dynamic>?,
      git: json['git'] != null
          ? GitInfo.fromJson(json['git'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (resource != null) 'resource': resource,
      if (git != null) 'git': git!.toJson(),
    };
  }
}

/// Response from `DELETE /api/v1/patients/{id}/erase`.
///
/// Matches Go `service.EraseResponse`.
class EraseResponse {
  final bool erased;
  final String patientId;

  const EraseResponse({
    required this.erased,
    required this.patientId,
  });

  factory EraseResponse.fromJson(Map<String, dynamic> json) {
    return EraseResponse(
      erased: json['erased'] as bool,
      patientId: json['patient_id'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'erased': erased,
      'patient_id': patientId,
    };
  }
}

/// Body of `POST /api/v1/patients/match`.
///
/// Matches Go `service.MatchPatientsRequest`.
class MatchPatientsRequest {
  final String familyName;
  final List<String> givenNames;
  final String gender;
  final String birthDateApprox;
  final String district;
  final double threshold;

  const MatchPatientsRequest({
    required this.familyName,
    required this.givenNames,
    required this.gender,
    required this.birthDateApprox,
    required this.district,
    required this.threshold,
  });

  factory MatchPatientsRequest.fromJson(Map<String, dynamic> json) {
    return MatchPatientsRequest(
      familyName: json['family_name'] as String,
      givenNames: (json['given_names'] as List<dynamic>)
          .map((g) => g as String)
          .toList(),
      gender: json['gender'] as String,
      birthDateApprox: json['birth_date_approx'] as String,
      district: json['district'] as String,
      threshold: (json['threshold'] as num).toDouble(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'family_name': familyName,
      'given_names': givenNames,
      'gender': gender,
      'birth_date_approx': birthDateApprox,
      'district': district,
      'threshold': threshold,
    };
  }
}

/// Response from `POST /api/v1/patients/match`.
///
/// Matches Go `service.MatchPatientsResponse`.
class MatchPatientsResponse {
  final List<PatientMatch> matches;

  const MatchPatientsResponse({required this.matches});

  factory MatchPatientsResponse.fromJson(Map<String, dynamic> json) {
    return MatchPatientsResponse(
      matches: (json['matches'] as List<dynamic>)
          .map((m) => PatientMatch.fromJson(m as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'matches': matches.map((m) => m.toJson()).toList(),
    };
  }
}

/// A single patient match result with confidence score.
class PatientMatch {
  final String patientId;
  final double confidence;
  final List<String> matchFactors;

  const PatientMatch({
    required this.patientId,
    required this.confidence,
    required this.matchFactors,
  });

  factory PatientMatch.fromJson(Map<String, dynamic> json) {
    return PatientMatch(
      patientId: json['patient_id'] as String,
      confidence: (json['confidence'] as num).toDouble(),
      matchFactors: (json['match_factors'] as List<dynamic>)
          .map((f) => f as String)
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'patient_id': patientId,
      'confidence': confidence,
      'match_factors': matchFactors,
    };
  }
}

/// Response from `GET /api/v1/patients/{id}/history`.
///
/// Matches Go `service.PatientHistoryResponse`.
class PatientHistoryResponse {
  final List<HistoryEntry> entries;

  const PatientHistoryResponse({required this.entries});

  factory PatientHistoryResponse.fromJson(Map<String, dynamic> json) {
    return PatientHistoryResponse(
      entries: (json['entries'] as List<dynamic>)
          .map((e) => HistoryEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'entries': entries.map((e) => e.toJson()).toList(),
    };
  }
}

/// A single Git history entry for a patient.
class HistoryEntry {
  final String commitHash;
  final String timestamp;
  final String author;
  final String node;
  final String site;
  final String operation;
  final String resourceType;
  final String resourceId;
  final String message;

  const HistoryEntry({
    required this.commitHash,
    required this.timestamp,
    required this.author,
    required this.node,
    required this.site,
    required this.operation,
    required this.resourceType,
    required this.resourceId,
    required this.message,
  });

  factory HistoryEntry.fromJson(Map<String, dynamic> json) {
    return HistoryEntry(
      commitHash: json['commit_hash'] as String,
      timestamp: json['timestamp'] as String,
      author: json['author'] as String,
      node: json['node'] as String,
      site: json['site'] as String,
      operation: json['operation'] as String,
      resourceType: json['resource_type'] as String,
      resourceId: json['resource_id'] as String,
      message: json['message'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'commit_hash': commitHash,
      'timestamp': timestamp,
      'author': author,
      'node': node,
      'site': site,
      'operation': operation,
      'resource_type': resourceType,
      'resource_id': resourceId,
      'message': message,
    };
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

List<Map<String, dynamic>> _mapList(dynamic value) {
  if (value == null) return [];
  return (value as List<dynamic>)
      .map((item) => item as Map<String, dynamic>)
      .toList();
}
