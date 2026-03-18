import 'api_envelope.dart';

/// Anchor / IOTA Tangle DTOs matching Go `service.AnchorStatusResponse`,
/// `service.DIDDocumentResponse`, `service.CredentialResponse`, etc.

// ---------------------------------------------------------------------------
// Anchoring
// ---------------------------------------------------------------------------

/// Response from `GET /api/v1/anchor/status`.
class AnchorStatusResponse {
  final String state;
  final String lastAnchorId;
  final String lastAnchorTime;
  final String merkleRoot;
  final String nodeDid;
  final int queueDepth;
  final String backend;
  final int pendingCommits;

  const AnchorStatusResponse({
    required this.state,
    required this.lastAnchorId,
    required this.lastAnchorTime,
    required this.merkleRoot,
    required this.nodeDid,
    required this.queueDepth,
    required this.backend,
    required this.pendingCommits,
  });

  factory AnchorStatusResponse.fromJson(Map<String, dynamic> json) {
    return AnchorStatusResponse(
      state: json['state'] as String,
      lastAnchorId: json['last_anchor_id'] as String,
      lastAnchorTime: json['last_anchor_time'] as String,
      merkleRoot: json['merkle_root'] as String,
      nodeDid: json['node_did'] as String,
      queueDepth: (json['queue_depth'] as num).toInt(),
      backend: json['backend'] as String,
      pendingCommits: (json['pending_commits'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'state': state,
      'last_anchor_id': lastAnchorId,
      'last_anchor_time': lastAnchorTime,
      'merkle_root': merkleRoot,
      'node_did': nodeDid,
      'queue_depth': queueDepth,
      'backend': backend,
      'pending_commits': pendingCommits,
    };
  }
}

/// Response from `POST /api/v1/anchor/verify`.
class AnchorVerifyResponse {
  final bool verified;
  final String anchorId;
  final String merkleRoot;
  final String anchoredAt;
  final String commitHash;
  final String state;

  const AnchorVerifyResponse({
    required this.verified,
    required this.anchorId,
    required this.merkleRoot,
    required this.anchoredAt,
    required this.commitHash,
    required this.state,
  });

  factory AnchorVerifyResponse.fromJson(Map<String, dynamic> json) {
    return AnchorVerifyResponse(
      verified: json['verified'] as bool,
      anchorId: json['anchor_id'] as String,
      merkleRoot: json['merkle_root'] as String,
      anchoredAt: json['anchored_at'] as String,
      commitHash: json['commit_hash'] as String,
      state: json['state'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'verified': verified,
      'anchor_id': anchorId,
      'merkle_root': merkleRoot,
      'anchored_at': anchoredAt,
      'commit_hash': commitHash,
      'state': state,
    };
  }
}

/// Paginated anchor history from `GET /api/v1/anchor/history`.
class AnchorHistoryResponse {
  final List<AnchorRecord> records;
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const AnchorHistoryResponse({
    required this.records,
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory AnchorHistoryResponse.fromJson(Map<String, dynamic> json) {
    return AnchorHistoryResponse(
      records: (json['records'] as List<dynamic>)
          .map((r) => AnchorRecord.fromJson(r as Map<String, dynamic>))
          .toList(),
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'records': records.map((r) => r.toJson()).toList(),
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

/// A single anchor record in the history.
class AnchorRecord {
  final String anchorId;
  final String merkleRoot;
  final String gitHead;
  final String state;
  final String timestamp;
  final String backend;
  final String txId;
  final String nodeDid;

  const AnchorRecord({
    required this.anchorId,
    required this.merkleRoot,
    required this.gitHead,
    required this.state,
    required this.timestamp,
    required this.backend,
    required this.txId,
    required this.nodeDid,
  });

  factory AnchorRecord.fromJson(Map<String, dynamic> json) {
    return AnchorRecord(
      anchorId: json['anchor_id'] as String,
      merkleRoot: json['merkle_root'] as String,
      gitHead: json['git_head'] as String,
      state: json['state'] as String,
      timestamp: json['timestamp'] as String,
      backend: json['backend'] as String,
      txId: json['tx_id'] as String,
      nodeDid: json['node_did'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'anchor_id': anchorId,
      'merkle_root': merkleRoot,
      'git_head': gitHead,
      'state': state,
      'timestamp': timestamp,
      'backend': backend,
      'tx_id': txId,
      'node_did': nodeDid,
    };
  }
}

/// Response from `POST /api/v1/anchor/trigger`.
class AnchorTriggerResponse {
  final String anchorId;
  final String state;
  final String merkleRoot;
  final String gitHead;
  final bool skipped;
  final String message;
  final GitInfo? git;

  const AnchorTriggerResponse({
    required this.anchorId,
    required this.state,
    required this.merkleRoot,
    required this.gitHead,
    required this.skipped,
    required this.message,
    this.git,
  });

  factory AnchorTriggerResponse.fromJson(Map<String, dynamic> json) {
    return AnchorTriggerResponse(
      anchorId: json['anchor_id'] as String,
      state: json['state'] as String,
      merkleRoot: json['merkle_root'] as String,
      gitHead: json['git_head'] as String,
      skipped: json['skipped'] as bool,
      message: json['message'] as String,
      git: json['git'] != null
          ? GitInfo.fromJson(json['git'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'anchor_id': anchorId,
      'state': state,
      'merkle_root': merkleRoot,
      'git_head': gitHead,
      'skipped': skipped,
      'message': message,
      if (git != null) 'git': git!.toJson(),
    };
  }
}

// ---------------------------------------------------------------------------
// DID
// ---------------------------------------------------------------------------

/// DID Document response from `GET /api/v1/anchor/did/*`.
class DIDDocumentResponse {
  final String id;
  final List<String> context;
  final List<VerificationMethodDTO> verificationMethod;
  final List<String> authentication;
  final List<String> assertionMethod;
  final String? created;

  const DIDDocumentResponse({
    required this.id,
    required this.context,
    required this.verificationMethod,
    required this.authentication,
    required this.assertionMethod,
    this.created,
  });

  factory DIDDocumentResponse.fromJson(Map<String, dynamic> json) {
    return DIDDocumentResponse(
      id: json['id'] as String,
      context: (json['@context'] as List<dynamic>)
          .map((c) => c as String)
          .toList(),
      verificationMethod: (json['verificationMethod'] as List<dynamic>)
          .map((v) =>
              VerificationMethodDTO.fromJson(v as Map<String, dynamic>))
          .toList(),
      authentication: (json['authentication'] as List<dynamic>)
          .map((a) => a as String)
          .toList(),
      assertionMethod: (json['assertionMethod'] as List<dynamic>)
          .map((a) => a as String)
          .toList(),
      created: json['created'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      '@context': context,
      'verificationMethod':
          verificationMethod.map((v) => v.toJson()).toList(),
      'authentication': authentication,
      'assertionMethod': assertionMethod,
      if (created != null) 'created': created,
    };
  }
}

/// A DID verification method entry.
class VerificationMethodDTO {
  final String id;
  final String type;
  final String controller;
  final String publicKeyMultibase;

  const VerificationMethodDTO({
    required this.id,
    required this.type,
    required this.controller,
    required this.publicKeyMultibase,
  });

  factory VerificationMethodDTO.fromJson(Map<String, dynamic> json) {
    return VerificationMethodDTO(
      id: json['id'] as String,
      type: json['type'] as String,
      controller: json['controller'] as String,
      publicKeyMultibase: json['publicKeyMultibase'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type,
      'controller': controller,
      'publicKeyMultibase': publicKeyMultibase,
    };
  }
}

// ---------------------------------------------------------------------------
// Credentials
// ---------------------------------------------------------------------------

/// Body of `POST /api/v1/anchor/credentials/issue`.
class IssueCredentialRequest {
  final String anchorId;
  final List<String>? types;
  final Map<String, String>? additionalClaims;

  const IssueCredentialRequest({
    required this.anchorId,
    this.types,
    this.additionalClaims,
  });

  factory IssueCredentialRequest.fromJson(Map<String, dynamic> json) {
    return IssueCredentialRequest(
      anchorId: json['anchor_id'] as String,
      types: (json['types'] as List<dynamic>?)
          ?.map((t) => t as String)
          .toList(),
      additionalClaims: (json['additional_claims'] as Map<String, dynamic>?)
          ?.map((k, v) => MapEntry(k, v as String)),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'anchor_id': anchorId,
      if (types != null) 'types': types,
      if (additionalClaims != null) 'additional_claims': additionalClaims,
    };
  }
}

/// A Verifiable Credential response.
class CredentialResponse {
  final String id;
  final List<String> context;
  final List<String> type;
  final String issuer;
  final String issuanceDate;
  final String? expirationDate;
  final String credentialSubjectJson;
  final CredentialProofDTO? proof;

  const CredentialResponse({
    required this.id,
    required this.context,
    required this.type,
    required this.issuer,
    required this.issuanceDate,
    this.expirationDate,
    required this.credentialSubjectJson,
    this.proof,
  });

  factory CredentialResponse.fromJson(Map<String, dynamic> json) {
    return CredentialResponse(
      id: json['id'] as String,
      context: (json['@context'] as List<dynamic>)
          .map((c) => c as String)
          .toList(),
      type:
          (json['type'] as List<dynamic>).map((t) => t as String).toList(),
      issuer: json['issuer'] as String,
      issuanceDate: json['issuanceDate'] as String,
      expirationDate: json['expirationDate'] as String?,
      credentialSubjectJson: json['credentialSubjectJson'] as String,
      proof: json['proof'] != null
          ? CredentialProofDTO.fromJson(
              json['proof'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      '@context': context,
      'type': type,
      'issuer': issuer,
      'issuanceDate': issuanceDate,
      if (expirationDate != null) 'expirationDate': expirationDate,
      'credentialSubjectJson': credentialSubjectJson,
      if (proof != null) 'proof': proof!.toJson(),
    };
  }
}

/// Cryptographic proof on a Verifiable Credential.
class CredentialProofDTO {
  final String type;
  final String created;
  final String verificationMethod;
  final String proofPurpose;
  final String proofValue;

  const CredentialProofDTO({
    required this.type,
    required this.created,
    required this.verificationMethod,
    required this.proofPurpose,
    required this.proofValue,
  });

  factory CredentialProofDTO.fromJson(Map<String, dynamic> json) {
    return CredentialProofDTO(
      type: json['type'] as String,
      created: json['created'] as String,
      verificationMethod: json['verificationMethod'] as String,
      proofPurpose: json['proofPurpose'] as String,
      proofValue: json['proofValue'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'created': created,
      'verificationMethod': verificationMethod,
      'proofPurpose': proofPurpose,
      'proofValue': proofValue,
    };
  }
}

/// Response from `POST /api/v1/anchor/credentials/verify`.
class CredentialVerificationResponse {
  final bool valid;
  final String issuer;
  final String message;

  const CredentialVerificationResponse({
    required this.valid,
    required this.issuer,
    required this.message,
  });

  factory CredentialVerificationResponse.fromJson(Map<String, dynamic> json) {
    return CredentialVerificationResponse(
      valid: json['valid'] as bool,
      issuer: json['issuer'] as String,
      message: json['message'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'valid': valid,
      'issuer': issuer,
      'message': message,
    };
  }
}

/// Paginated credential list from `GET /api/v1/anchor/credentials`.
class CredentialListResponse {
  final List<CredentialResponse> credentials;
  final int page;
  final int perPage;
  final int total;
  final int totalPages;

  const CredentialListResponse({
    required this.credentials,
    required this.page,
    required this.perPage,
    required this.total,
    required this.totalPages,
  });

  factory CredentialListResponse.fromJson(Map<String, dynamic> json) {
    return CredentialListResponse(
      credentials: (json['credentials'] as List<dynamic>)
          .map((c) => CredentialResponse.fromJson(c as Map<String, dynamic>))
          .toList(),
      page: (json['page'] as num).toInt(),
      perPage: (json['per_page'] as num).toInt(),
      total: (json['total'] as num).toInt(),
      totalPages: (json['total_pages'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'credentials': credentials.map((c) => c.toJson()).toList(),
      'page': page,
      'per_page': perPage,
      'total': total,
      'total_pages': totalPages,
    };
  }
}

// ---------------------------------------------------------------------------
// Backend
// ---------------------------------------------------------------------------

/// List of anchor backends from `GET /api/v1/anchor/backends`.
class BackendListResponse {
  final List<BackendInfoDTO> backends;

  const BackendListResponse({required this.backends});

  factory BackendListResponse.fromJson(Map<String, dynamic> json) {
    return BackendListResponse(
      backends: (json['backends'] as List<dynamic>)
          .map((b) => BackendInfoDTO.fromJson(b as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'backends': backends.map((b) => b.toJson()).toList(),
    };
  }
}

/// Info about a single anchor backend.
class BackendInfoDTO {
  final String name;
  final bool available;
  final String description;

  const BackendInfoDTO({
    required this.name,
    required this.available,
    required this.description,
  });

  factory BackendInfoDTO.fromJson(Map<String, dynamic> json) {
    return BackendInfoDTO(
      name: json['name'] as String,
      available: json['available'] as bool,
      description: json['description'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'available': available,
      'description': description,
    };
  }
}

/// Status of a specific anchor backend from `GET /api/v1/anchor/backends/{name}`.
class BackendStatusResponse {
  final String name;
  final bool available;
  final String description;
  final int anchoredCount;
  final String lastAnchorTime;

  const BackendStatusResponse({
    required this.name,
    required this.available,
    required this.description,
    required this.anchoredCount,
    required this.lastAnchorTime,
  });

  factory BackendStatusResponse.fromJson(Map<String, dynamic> json) {
    return BackendStatusResponse(
      name: json['name'] as String,
      available: json['available'] as bool,
      description: json['description'] as String,
      anchoredCount: (json['anchored_count'] as num).toInt(),
      lastAnchorTime: json['last_anchor_time'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'available': available,
      'description': description,
      'anchored_count': anchoredCount,
      'last_anchor_time': lastAnchorTime,
    };
  }
}

/// Anchor queue status from `GET /api/v1/anchor/queue`.
class QueueStatusResponse {
  final int pending;
  final int totalProcessed;
  final List<QueueEntryDTO> entries;

  const QueueStatusResponse({
    required this.pending,
    required this.totalProcessed,
    required this.entries,
  });

  factory QueueStatusResponse.fromJson(Map<String, dynamic> json) {
    return QueueStatusResponse(
      pending: (json['pending'] as num).toInt(),
      totalProcessed: (json['total_processed'] as num).toInt(),
      entries: (json['entries'] as List<dynamic>)
          .map((e) => QueueEntryDTO.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'pending': pending,
      'total_processed': totalProcessed,
      'entries': entries.map((e) => e.toJson()).toList(),
    };
  }
}

/// A single entry in the anchor queue.
class QueueEntryDTO {
  final String anchorId;
  final String merkleRoot;
  final String gitHead;
  final String enqueuedAt;
  final String state;

  const QueueEntryDTO({
    required this.anchorId,
    required this.merkleRoot,
    required this.gitHead,
    required this.enqueuedAt,
    required this.state,
  });

  factory QueueEntryDTO.fromJson(Map<String, dynamic> json) {
    return QueueEntryDTO(
      anchorId: json['anchor_id'] as String,
      merkleRoot: json['merkle_root'] as String,
      gitHead: json['git_head'] as String,
      enqueuedAt: json['enqueued_at'] as String,
      state: json['state'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'anchor_id': anchorId,
      'merkle_root': merkleRoot,
      'git_head': gitHead,
      'enqueued_at': enqueuedAt,
      'state': state,
    };
  }
}

/// Anchor health check response.
class AnchorHealthResponse {
  final String status;
  final String nodeDid;
  final String backend;
  final int anchorCount;
  final int queueDepth;

  const AnchorHealthResponse({
    required this.status,
    required this.nodeDid,
    required this.backend,
    required this.anchorCount,
    required this.queueDepth,
  });

  factory AnchorHealthResponse.fromJson(Map<String, dynamic> json) {
    return AnchorHealthResponse(
      status: json['status'] as String,
      nodeDid: json['node_did'] as String,
      backend: json['backend'] as String,
      anchorCount: (json['anchor_count'] as num).toInt(),
      queueDepth: (json['queue_depth'] as num).toInt(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'status': status,
      'node_did': nodeDid,
      'backend': backend,
      'anchor_count': anchorCount,
      'queue_depth': queueDepth,
    };
  }
}
