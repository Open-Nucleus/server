import 'api_envelope.dart';

/// Conflict DTOs matching Go `service.ConflictListResponse`,
/// `service.ConflictDetail`, etc. Used by the merge conflict resolution UI.

/// Response from `GET /api/v1/conflicts`.
class ConflictListResponse {
  final List<ConflictDetail> conflicts;
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const ConflictListResponse({
    required this.conflicts,
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory ConflictListResponse.fromJson(Map<String, dynamic> json) {
    return ConflictListResponse(
      conflicts: (json['conflicts'] as List<dynamic>)
          .map((c) => ConflictDetail.fromJson(c as Map<String, dynamic>))
          .toList(),
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'conflicts': conflicts.map((c) => c.toJson()).toList(),
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

/// A single merge conflict with both local and remote versions.
///
/// Matches Go `service.ConflictDetail`. The `localVersion` and `remoteVersion`
/// fields contain raw FHIR resource JSON.
class ConflictDetail {
  final String id;
  final String resourceType;
  final String resourceId;
  final String status;
  final String detectedAt;
  final Map<String, dynamic>? localVersion;
  final Map<String, dynamic>? remoteVersion;
  final String localNode;
  final String remoteNode;

  const ConflictDetail({
    required this.id,
    required this.resourceType,
    required this.resourceId,
    required this.status,
    required this.detectedAt,
    this.localVersion,
    this.remoteVersion,
    required this.localNode,
    required this.remoteNode,
  });

  factory ConflictDetail.fromJson(Map<String, dynamic> json) {
    return ConflictDetail(
      id: json['id'] as String,
      resourceType: json['resource_type'] as String,
      resourceId: json['resource_id'] as String,
      status: json['status'] as String,
      detectedAt: json['detected_at'] as String,
      localVersion: json['local_version'] as Map<String, dynamic>?,
      remoteVersion: json['remote_version'] as Map<String, dynamic>?,
      localNode: json['local_node'] as String,
      remoteNode: json['remote_node'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'resource_type': resourceType,
      'resource_id': resourceId,
      'status': status,
      'detected_at': detectedAt,
      if (localVersion != null) 'local_version': localVersion,
      if (remoteVersion != null) 'remote_version': remoteVersion,
      'local_node': localNode,
      'remote_node': remoteNode,
    };
  }
}

/// Body of `POST /api/v1/conflicts/{id}/resolve`.
class ResolveConflictRequest {
  final String conflictId;
  final String resolution;
  final Map<String, dynamic>? mergedResource;
  final String author;

  const ResolveConflictRequest({
    required this.conflictId,
    required this.resolution,
    this.mergedResource,
    required this.author,
  });

  factory ResolveConflictRequest.fromJson(Map<String, dynamic> json) {
    return ResolveConflictRequest(
      conflictId: json['conflict_id'] as String,
      resolution: json['resolution'] as String,
      mergedResource: json['merged_resource'] as Map<String, dynamic>?,
      author: json['author'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'conflict_id': conflictId,
      'resolution': resolution,
      if (mergedResource != null) 'merged_resource': mergedResource,
      'author': author,
    };
  }
}

/// Response from `POST /api/v1/conflicts/{id}/resolve`.
class ResolveConflictResponse {
  final GitInfo? git;

  const ResolveConflictResponse({this.git});

  factory ResolveConflictResponse.fromJson(Map<String, dynamic> json) {
    return ResolveConflictResponse(
      git: json['git'] != null
          ? GitInfo.fromJson(json['git'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      if (git != null) 'git': git!.toJson(),
    };
  }
}

/// Body of `POST /api/v1/conflicts/{id}/defer`.
class DeferConflictRequest {
  final String conflictId;
  final String reason;

  const DeferConflictRequest({
    required this.conflictId,
    required this.reason,
  });

  factory DeferConflictRequest.fromJson(Map<String, dynamic> json) {
    return DeferConflictRequest(
      conflictId: json['conflict_id'] as String,
      reason: json['reason'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'conflict_id': conflictId,
      'reason': reason,
    };
  }
}

/// Response from `POST /api/v1/conflicts/{id}/defer`.
class DeferConflictResponse {
  final String status;

  const DeferConflictResponse({required this.status});

  factory DeferConflictResponse.fromJson(Map<String, dynamic> json) {
    return DeferConflictResponse(
      status: json['status'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'status': status,
    };
  }
}
