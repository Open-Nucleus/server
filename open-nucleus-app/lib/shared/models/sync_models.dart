import 'api_envelope.dart';

/// Sync DTOs matching Go `service.SyncStatusResponse`, `service.SyncPeersResponse`,
/// etc. Used by the sync status panel, peer discovery, and bundle import/export.

/// Response from `GET /api/v1/sync/status`.
class SyncStatusResponse {
  final String state;
  final String lastSync;
  final int pendingChanges;
  final String nodeId;
  final String siteId;

  const SyncStatusResponse({
    required this.state,
    required this.lastSync,
    required this.pendingChanges,
    required this.nodeId,
    required this.siteId,
  });

  factory SyncStatusResponse.fromJson(Map<String, dynamic> json) {
    return SyncStatusResponse(
      state: json['state'] as String,
      lastSync: json['last_sync'] as String,
      pendingChanges: (json['pending_changes'] as num).toInt(),
      nodeId: json['node_id'] as String,
      siteId: json['site_id'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'state': state,
      'last_sync': lastSync,
      'pending_changes': pendingChanges,
      'node_id': nodeId,
      'site_id': siteId,
    };
  }
}

/// Response from `GET /api/v1/sync/peers`.
class SyncPeersResponse {
  final List<PeerInfo> peers;

  const SyncPeersResponse({required this.peers});

  factory SyncPeersResponse.fromJson(Map<String, dynamic> json) {
    return SyncPeersResponse(
      peers: (json['peers'] as List<dynamic>)
          .map((p) => PeerInfo.fromJson(p as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'peers': peers.map((p) => p.toJson()).toList(),
    };
  }
}

/// A discovered sync peer.
class PeerInfo {
  final String nodeId;
  final String siteId;
  final String lastSeen;
  final String state;
  final int latencyMs;

  const PeerInfo({
    required this.nodeId,
    required this.siteId,
    required this.lastSeen,
    required this.state,
    required this.latencyMs,
  });

  factory PeerInfo.fromJson(Map<String, dynamic> json) {
    return PeerInfo(
      nodeId: json['node_id'] as String,
      siteId: json['site_id'] as String,
      lastSeen: json['last_seen'] as String,
      state: json['state'] as String,
      latencyMs: (json['latency_ms'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'node_id': nodeId,
      'site_id': siteId,
      'last_seen': lastSeen,
      'state': state,
      'latency_ms': latencyMs,
    };
  }
}

/// Response from `POST /api/v1/sync/trigger`.
class SyncTriggerResponse {
  final String syncId;
  final String state;
  final GitInfo? git;

  const SyncTriggerResponse({
    required this.syncId,
    required this.state,
    this.git,
  });

  factory SyncTriggerResponse.fromJson(Map<String, dynamic> json) {
    return SyncTriggerResponse(
      syncId: json['sync_id'] as String,
      state: json['state'] as String,
      git: json['git'] != null
          ? GitInfo.fromJson(json['git'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'sync_id': syncId,
      'state': state,
      if (git != null) 'git': git!.toJson(),
    };
  }
}

/// Response from `GET /api/v1/sync/history`.
class SyncHistoryResponse {
  final List<SyncEvent> events;
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const SyncHistoryResponse({
    required this.events,
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory SyncHistoryResponse.fromJson(Map<String, dynamic> json) {
    return SyncHistoryResponse(
      events: (json['events'] as List<dynamic>)
          .map((e) => SyncEvent.fromJson(e as Map<String, dynamic>))
          .toList(),
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'events': events.map((e) => e.toJson()).toList(),
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

/// A single sync event in the history log.
class SyncEvent {
  final String syncId;
  final String timestamp;
  final String direction;
  final String peerNode;
  final String state;
  final int resourcesTransferred;

  const SyncEvent({
    required this.syncId,
    required this.timestamp,
    required this.direction,
    required this.peerNode,
    required this.state,
    required this.resourcesTransferred,
  });

  factory SyncEvent.fromJson(Map<String, dynamic> json) {
    return SyncEvent(
      syncId: json['sync_id'] as String,
      timestamp: json['timestamp'] as String,
      direction: json['direction'] as String,
      peerNode: json['peer_node'] as String,
      state: json['state'] as String,
      resourcesTransferred: (json['resources_transferred'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'sync_id': syncId,
      'timestamp': timestamp,
      'direction': direction,
      'peer_node': peerNode,
      'state': state,
      'resources_transferred': resourcesTransferred,
    };
  }
}

/// Body of `POST /api/v1/sync/bundle/export`.
class BundleExportRequest {
  final List<String> resourceTypes;
  final String since;

  const BundleExportRequest({
    required this.resourceTypes,
    required this.since,
  });

  factory BundleExportRequest.fromJson(Map<String, dynamic> json) {
    return BundleExportRequest(
      resourceTypes: (json['resource_types'] as List<dynamic>)
          .map((r) => r as String)
          .toList(),
      since: json['since'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'resource_types': resourceTypes,
      'since': since,
    };
  }
}

/// Response from `POST /api/v1/sync/bundle/export`.
class BundleExportResponse {
  final String bundleData;
  final String format;
  final int resourceCount;
  final GitInfo? git;

  const BundleExportResponse({
    required this.bundleData,
    required this.format,
    required this.resourceCount,
    this.git,
  });

  factory BundleExportResponse.fromJson(Map<String, dynamic> json) {
    return BundleExportResponse(
      bundleData: json['bundle_data'] as String,
      format: json['format'] as String,
      resourceCount: (json['resource_count'] as num).toInt(),
      git: json['git'] != null
          ? GitInfo.fromJson(json['git'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'bundle_data': bundleData,
      'format': format,
      'resource_count': resourceCount,
      if (git != null) 'git': git!.toJson(),
    };
  }
}

/// Body of `POST /api/v1/sync/bundle/import`.
class BundleImportRequest {
  final String bundleData;
  final String format;
  final String author;
  final String nodeId;
  final String siteId;

  const BundleImportRequest({
    required this.bundleData,
    required this.format,
    required this.author,
    required this.nodeId,
    required this.siteId,
  });

  factory BundleImportRequest.fromJson(Map<String, dynamic> json) {
    return BundleImportRequest(
      bundleData: json['bundle_data'] as String,
      format: json['format'] as String,
      author: json['author'] as String,
      nodeId: json['node_id'] as String,
      siteId: json['site_id'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'bundle_data': bundleData,
      'format': format,
      'author': author,
      'node_id': nodeId,
      'site_id': siteId,
    };
  }
}

/// Response from `POST /api/v1/sync/bundle/import`.
class BundleImportResponse {
  final int resourcesImported;
  final int resourcesSkipped;
  final List<String> errors;
  final GitInfo? git;

  const BundleImportResponse({
    required this.resourcesImported,
    required this.resourcesSkipped,
    required this.errors,
    this.git,
  });

  factory BundleImportResponse.fromJson(Map<String, dynamic> json) {
    return BundleImportResponse(
      resourcesImported: (json['resources_imported'] as num).toInt(),
      resourcesSkipped: (json['resources_skipped'] as num).toInt(),
      errors: (json['errors'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      git: json['git'] != null
          ? GitInfo.fromJson(json['git'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'resources_imported': resourcesImported,
      'resources_skipped': resourcesSkipped,
      'errors': errors,
      if (git != null) 'git': git!.toJson(),
    };
  }
}
