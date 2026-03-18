/// Alert / Sentinel DTOs matching Go `service.AlertListResponse`,
/// `service.AlertSummaryResponse`, and `service.AlertDetail`.

/// Paginated list of alerts from `GET /api/v1/alerts`.
class AlertListResponse {
  final List<AlertDetail> alerts;
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const AlertListResponse({
    required this.alerts,
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory AlertListResponse.fromJson(Map<String, dynamic> json) {
    return AlertListResponse(
      alerts: (json['alerts'] as List<dynamic>)
          .map((a) => AlertDetail.fromJson(a as Map<String, dynamic>))
          .toList(),
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'alerts': alerts.map((a) => a.toJson()).toList(),
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

/// Aggregate alert counts from `GET /api/v1/alerts/summary`.
class AlertSummaryResponse {
  final int total;
  final int critical;
  final int warning;
  final int info;
  final int unacknowledged;

  const AlertSummaryResponse({
    required this.total,
    required this.critical,
    required this.warning,
    required this.info,
    required this.unacknowledged,
  });

  factory AlertSummaryResponse.fromJson(Map<String, dynamic> json) {
    return AlertSummaryResponse(
      total: (json['total'] as num).toInt(),
      critical: (json['critical'] as num).toInt(),
      warning: (json['warning'] as num).toInt(),
      info: (json['info'] as num).toInt(),
      unacknowledged: (json['unacknowledged'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'total': total,
      'critical': critical,
      'warning': warning,
      'info': info,
      'unacknowledged': unacknowledged,
    };
  }
}

/// A single Sentinel alert.
///
/// Matches Go `service.AlertDetail`.
class AlertDetail {
  final String id;
  final String type;
  final String severity;
  final String status;
  final String title;
  final String description;
  final String patientId;
  final String createdAt;
  final String? acknowledgedAt;
  final String? acknowledgedBy;

  const AlertDetail({
    required this.id,
    required this.type,
    required this.severity,
    required this.status,
    required this.title,
    required this.description,
    required this.patientId,
    required this.createdAt,
    this.acknowledgedAt,
    this.acknowledgedBy,
  });

  factory AlertDetail.fromJson(Map<String, dynamic> json) {
    return AlertDetail(
      id: json['id'] as String,
      type: json['type'] as String,
      severity: json['severity'] as String,
      status: json['status'] as String,
      title: json['title'] as String,
      description: json['description'] as String,
      patientId: json['patient_id'] as String,
      createdAt: json['created_at'] as String,
      acknowledgedAt: json['acknowledged_at'] as String?,
      acknowledgedBy: json['acknowledged_by'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type,
      'severity': severity,
      'status': status,
      'title': title,
      'description': description,
      'patient_id': patientId,
      'created_at': createdAt,
      if (acknowledgedAt != null) 'acknowledged_at': acknowledgedAt,
      if (acknowledgedBy != null) 'acknowledged_by': acknowledgedBy,
    };
  }
}
