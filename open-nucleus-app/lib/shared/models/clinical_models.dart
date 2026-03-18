/// Clinical DTOs matching Go `service.ClinicalListResponse`,
/// `service.ObservationFilters`, and `service.ConditionFilters`.

/// Paginated list of clinical FHIR resources (encounters, observations,
/// conditions, medication requests, allergy intolerances, immunizations,
/// procedures, etc.).
///
/// Resources are returned as raw `Map<String, dynamic>` — the canonical
/// FHIR JSON straight from Git.
class ClinicalListResponse {
  final List<Map<String, dynamic>> resources;
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const ClinicalListResponse({
    required this.resources,
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory ClinicalListResponse.fromJson(Map<String, dynamic> json) {
    return ClinicalListResponse(
      resources: (json['resources'] as List<dynamic>?)
              ?.map((r) => r as Map<String, dynamic>)
              .toList() ??
          [],
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'resources': resources,
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

/// Query filters for `GET /api/v1/patients/{id}/observations`.
///
/// Matches Go `service.ObservationFilters`. Converted to query parameters
/// by the repository layer.
class ObservationFilters {
  final String? code;
  final String? category;
  final String? dateFrom;
  final String? dateTo;
  final String? encounterId;

  const ObservationFilters({
    this.code,
    this.category,
    this.dateFrom,
    this.dateTo,
    this.encounterId,
  });

  factory ObservationFilters.fromJson(Map<String, dynamic> json) {
    return ObservationFilters(
      code: json['code'] as String?,
      category: json['category'] as String?,
      dateFrom: json['date_from'] as String?,
      dateTo: json['date_to'] as String?,
      encounterId: json['encounter_id'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (code != null) 'code': code,
      if (category != null) 'category': category,
      if (dateFrom != null) 'date_from': dateFrom,
      if (dateTo != null) 'date_to': dateTo,
      if (encounterId != null) 'encounter_id': encounterId,
    };
  }

  /// Converts to query parameter map for HTTP requests.
  Map<String, String> toQueryParameters() {
    return {
      if (code != null) 'code': code!,
      if (category != null) 'category': category!,
      if (dateFrom != null) 'date_from': dateFrom!,
      if (dateTo != null) 'date_to': dateTo!,
      if (encounterId != null) 'encounter_id': encounterId!,
    };
  }
}

/// Query filters for `GET /api/v1/patients/{id}/conditions`.
///
/// Matches Go `service.ConditionFilters`.
class ConditionFilters {
  final String? clinicalStatus;
  final String? category;
  final String? code;

  const ConditionFilters({
    this.clinicalStatus,
    this.category,
    this.code,
  });

  factory ConditionFilters.fromJson(Map<String, dynamic> json) {
    return ConditionFilters(
      clinicalStatus: json['clinical_status'] as String?,
      category: json['category'] as String?,
      code: json['code'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (clinicalStatus != null) 'clinical_status': clinicalStatus,
      if (category != null) 'category': category,
      if (code != null) 'code': code,
    };
  }

  /// Converts to query parameter map for HTTP requests.
  Map<String, String> toQueryParameters() {
    return {
      if (clinicalStatus != null) 'clinical_status': clinicalStatus!,
      if (category != null) 'category': category!,
      if (code != null) 'code': code!,
    };
  }
}
