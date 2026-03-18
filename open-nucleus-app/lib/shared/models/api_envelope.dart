/// Standard API response envelope matching the Go backend exactly.
///
/// Every REST response from the Open Nucleus backend is wrapped in this
/// envelope. The generic [T] represents the shape of [data], which varies
/// per endpoint.
class ApiEnvelope<T> {
  final String status;
  final T? data;
  final ErrorBody? error;
  final Pagination? pagination;
  final List<Warning>? warnings;
  final GitInfo? git;
  final Meta? meta;

  const ApiEnvelope({
    required this.status,
    this.data,
    this.error,
    this.pagination,
    this.warnings,
    this.git,
    this.meta,
  });

  /// Deserializes the envelope from JSON.
  ///
  /// [fromJsonT] converts the raw `data` field into a typed [T]. Pass `null`
  /// when the response carries no data (e.g. 204 or error-only).
  factory ApiEnvelope.fromJson(
    Map<String, dynamic> json,
    T Function(dynamic)? fromJsonT,
  ) {
    return ApiEnvelope<T>(
      status: json['status'] as String,
      data: json['data'] != null && fromJsonT != null
          ? fromJsonT(json['data'])
          : null,
      error: json['error'] != null
          ? ErrorBody.fromJson(json['error'] as Map<String, dynamic>)
          : null,
      pagination: json['pagination'] != null
          ? Pagination.fromJson(json['pagination'] as Map<String, dynamic>)
          : null,
      warnings: json['warnings'] != null
          ? (json['warnings'] as List<dynamic>)
              .map((w) => Warning.fromJson(w as Map<String, dynamic>))
              .toList()
          : null,
      git: json['git'] != null
          ? GitInfo.fromJson(json['git'] as Map<String, dynamic>)
          : null,
      meta: json['meta'] != null
          ? Meta.fromJson(json['meta'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson(Map<String, dynamic> Function(T)? toJsonT) {
    return {
      'status': status,
      if (data != null && toJsonT != null) 'data': toJsonT(data as T),
      if (error != null) 'error': error!.toJson(),
      if (pagination != null) 'pagination': pagination!.toJson(),
      if (warnings != null)
        'warnings': warnings!.map((w) => w.toJson()).toList(),
      if (git != null) 'git': git!.toJson(),
      if (meta != null) 'meta': meta!.toJson(),
    };
  }

  bool get isSuccess => status == 'success';
  bool get isError => status == 'error';
}

/// Error body matching Go `model.ErrorBody`.
class ErrorBody {
  final String code;
  final String message;
  final dynamic details;

  const ErrorBody({
    required this.code,
    required this.message,
    this.details,
  });

  factory ErrorBody.fromJson(Map<String, dynamic> json) {
    return ErrorBody(
      code: json['code'] as String,
      message: json['message'] as String,
      details: json['details'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'code': code,
      'message': message,
      if (details != null) 'details': details,
    };
  }
}

/// Warning matching Go `model.Warning`.
class Warning {
  final String severity;
  final String type;
  final String description;
  final String? interactingMedication;
  final String? source;

  const Warning({
    required this.severity,
    required this.type,
    required this.description,
    this.interactingMedication,
    this.source,
  });

  factory Warning.fromJson(Map<String, dynamic> json) {
    return Warning(
      severity: json['severity'] as String,
      type: json['type'] as String,
      description: json['description'] as String,
      interactingMedication: json['interacting_medication'] as String?,
      source: json['source'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'severity': severity,
      'type': type,
      'description': description,
      if (interactingMedication != null)
        'interacting_medication': interactingMedication,
      if (source != null) 'source': source,
    };
  }
}

/// Git commit metadata matching Go `model.GitInfo`.
class GitInfo {
  final String commit;
  final String message;

  const GitInfo({
    required this.commit,
    required this.message,
  });

  factory GitInfo.fromJson(Map<String, dynamic> json) {
    return GitInfo(
      commit: json['commit'] as String,
      message: json['message'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'commit': commit,
      'message': message,
    };
  }
}

/// Request metadata matching Go `model.Meta`.
class Meta {
  final String requestId;
  final int durationMs;
  final String nodeId;

  const Meta({
    required this.requestId,
    required this.durationMs,
    required this.nodeId,
  });

  factory Meta.fromJson(Map<String, dynamic> json) {
    return Meta(
      requestId: json['request_id'] as String,
      durationMs: (json['duration_ms'] as num).toInt(),
      nodeId: json['node_id'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'request_id': requestId,
      'duration_ms': durationMs,
      'node_id': nodeId,
    };
  }
}

/// Pagination metadata matching Go `model.Pagination`.
class Pagination {
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const Pagination({
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory Pagination.fromJson(Map<String, dynamic> json) {
    return Pagination(
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
