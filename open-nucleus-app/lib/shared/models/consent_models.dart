/// Consent DTOs matching Go `service.ConsentAccessDecision`,
/// `service.ConsentGrantResponse`, `service.ConsentListResponse`, etc.

/// Access decision returned by the consent check middleware.
class ConsentAccessDecision {
  final bool allowed;
  final String? consentId;
  final String reason;

  const ConsentAccessDecision({
    required this.allowed,
    this.consentId,
    required this.reason,
  });

  factory ConsentAccessDecision.fromJson(Map<String, dynamic> json) {
    return ConsentAccessDecision(
      allowed: json['allowed'] as bool,
      consentId: json['consent_id'] as String?,
      reason: json['reason'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'allowed': allowed,
      if (consentId != null) 'consent_id': consentId,
      'reason': reason,
    };
  }
}

/// Time period for a consent grant.
class ConsentPeriod {
  final String start;
  final String end;

  const ConsentPeriod({
    required this.start,
    required this.end,
  });

  factory ConsentPeriod.fromJson(Map<String, dynamic> json) {
    return ConsentPeriod(
      start: json['start'] as String,
      end: json['end'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'start': start,
      'end': end,
    };
  }
}

/// Response from `POST /api/v1/patients/{id}/consents`.
class ConsentGrantResponse {
  final String consentId;
  final String commitHash;
  final String status;

  const ConsentGrantResponse({
    required this.consentId,
    required this.commitHash,
    required this.status,
  });

  factory ConsentGrantResponse.fromJson(Map<String, dynamic> json) {
    return ConsentGrantResponse(
      consentId: json['consent_id'] as String,
      commitHash: json['commit_hash'] as String,
      status: json['status'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'consent_id': consentId,
      'commit_hash': commitHash,
      'status': status,
    };
  }
}

/// Paginated consent list from `GET /api/v1/patients/{id}/consents`.
class ConsentListResponse {
  final List<ConsentSummary> consents;
  final ConsentPaginationMeta? pagination;

  const ConsentListResponse({
    required this.consents,
    this.pagination,
  });

  factory ConsentListResponse.fromJson(Map<String, dynamic> json) {
    return ConsentListResponse(
      consents: (json['consents'] as List<dynamic>)
          .map((c) => ConsentSummary.fromJson(c as Map<String, dynamic>))
          .toList(),
      pagination: json['pagination'] != null
          ? ConsentPaginationMeta.fromJson(
              json['pagination'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'consents': consents.map((c) => c.toJson()).toList(),
      if (pagination != null) 'pagination': pagination!.toJson(),
    };
  }
}

/// Summary of a single consent record.
class ConsentSummary {
  final String id;
  final String patientId;
  final String status;
  final String scopeCode;
  final String performerId;
  final String provisionType;
  final String? periodStart;
  final String? periodEnd;
  final String? category;
  final String lastUpdated;

  const ConsentSummary({
    required this.id,
    required this.patientId,
    required this.status,
    required this.scopeCode,
    required this.performerId,
    required this.provisionType,
    this.periodStart,
    this.periodEnd,
    this.category,
    required this.lastUpdated,
  });

  factory ConsentSummary.fromJson(Map<String, dynamic> json) {
    return ConsentSummary(
      id: json['id'] as String,
      patientId: json['patient_id'] as String,
      status: json['status'] as String,
      scopeCode: json['scope_code'] as String,
      performerId: json['performer_id'] as String,
      provisionType: json['provision_type'] as String,
      periodStart: json['period_start'] as String?,
      periodEnd: json['period_end'] as String?,
      category: json['category'] as String?,
      lastUpdated: json['last_updated'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'patient_id': patientId,
      'status': status,
      'scope_code': scopeCode,
      'performer_id': performerId,
      'provision_type': provisionType,
      if (periodStart != null) 'period_start': periodStart,
      if (periodEnd != null) 'period_end': periodEnd,
      if (category != null) 'category': category,
      'last_updated': lastUpdated,
    };
  }
}

/// Pagination metadata for consent lists (matches Go `service.PaginationMeta`).
class ConsentPaginationMeta {
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const ConsentPaginationMeta({
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory ConsentPaginationMeta.fromJson(Map<String, dynamic> json) {
    return ConsentPaginationMeta(
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

/// Response from `POST /api/v1/consents/{consentId}/vc`.
class ConsentVCResponse {
  final Map<String, dynamic> verifiableCredential;

  const ConsentVCResponse({required this.verifiableCredential});

  factory ConsentVCResponse.fromJson(Map<String, dynamic> json) {
    return ConsentVCResponse(
      verifiableCredential:
          json['verifiable_credential'] as Map<String, dynamic>,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'verifiable_credential': verifiableCredential,
    };
  }
}
