/// Auth DTOs matching Go `service.LoginRequest`, `service.LoginResponse`, etc.
///
/// JSON keys use snake_case to match the Go backend exactly.

/// Body of `POST /api/v1/auth/login`.
class LoginRequest {
  final String deviceId;
  final String publicKey;
  final ChallengeResponseDTO challengeResponse;
  final String practitionerId;

  const LoginRequest({
    required this.deviceId,
    required this.publicKey,
    required this.challengeResponse,
    required this.practitionerId,
  });

  factory LoginRequest.fromJson(Map<String, dynamic> json) {
    return LoginRequest(
      deviceId: json['device_id'] as String,
      publicKey: json['public_key'] as String,
      challengeResponse: ChallengeResponseDTO.fromJson(
        json['challenge_response'] as Map<String, dynamic>,
      ),
      practitionerId: json['practitioner_id'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'device_id': deviceId,
      'public_key': publicKey,
      'challenge_response': challengeResponse.toJson(),
      'practitioner_id': practitionerId,
    };
  }
}

/// Ed25519 challenge-response for login.
class ChallengeResponseDTO {
  final String nonce;
  final String signature;
  final String timestamp;

  const ChallengeResponseDTO({
    required this.nonce,
    required this.signature,
    required this.timestamp,
  });

  factory ChallengeResponseDTO.fromJson(Map<String, dynamic> json) {
    return ChallengeResponseDTO(
      nonce: json['nonce'] as String,
      signature: json['signature'] as String,
      timestamp: json['timestamp'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'nonce': nonce,
      'signature': signature,
      'timestamp': timestamp,
    };
  }
}

/// Response from `POST /api/v1/auth/login`.
class LoginResponse {
  final String token;
  final String expiresAt;
  final String refreshToken;
  final RoleDTO role;
  final String siteId;
  final String nodeId;

  const LoginResponse({
    required this.token,
    required this.expiresAt,
    required this.refreshToken,
    required this.role,
    required this.siteId,
    required this.nodeId,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) {
    return LoginResponse(
      token: json['token'] as String,
      expiresAt: json['expires_at'] as String,
      refreshToken: json['refresh_token'] as String,
      role: RoleDTO.fromJson(json['role'] as Map<String, dynamic>),
      siteId: json['site_id'] as String,
      nodeId: json['node_id'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'token': token,
      'expires_at': expiresAt,
      'refresh_token': refreshToken,
      'role': role.toJson(),
      'site_id': siteId,
      'node_id': nodeId,
    };
  }
}

/// RBAC role with permissions list.
class RoleDTO {
  final String code;
  final String display;
  final List<String> permissions;

  const RoleDTO({
    required this.code,
    required this.display,
    required this.permissions,
  });

  factory RoleDTO.fromJson(Map<String, dynamic> json) {
    return RoleDTO(
      code: json['code'] as String,
      display: json['display'] as String,
      permissions: (json['permissions'] as List<dynamic>)
          .map((p) => p as String)
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'code': code,
      'display': display,
      'permissions': permissions,
    };
  }
}

/// Response from `POST /api/v1/auth/refresh`.
class RefreshResponse {
  final String token;
  final String expiresAt;
  final String refreshToken;

  const RefreshResponse({
    required this.token,
    required this.expiresAt,
    required this.refreshToken,
  });

  factory RefreshResponse.fromJson(Map<String, dynamic> json) {
    return RefreshResponse(
      token: json['token'] as String,
      expiresAt: json['expires_at'] as String,
      refreshToken: json['refresh_token'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'token': token,
      'expires_at': expiresAt,
      'refresh_token': refreshToken,
    };
  }
}

/// Response from `GET /api/v1/auth/whoami`.
class WhoamiResponse {
  final String subject;
  final String nodeId;
  final String siteId;
  final RoleDTO role;

  const WhoamiResponse({
    required this.subject,
    required this.nodeId,
    required this.siteId,
    required this.role,
  });

  factory WhoamiResponse.fromJson(Map<String, dynamic> json) {
    return WhoamiResponse(
      subject: json['subject'] as String,
      nodeId: json['node_id'] as String,
      siteId: json['site_id'] as String,
      role: RoleDTO.fromJson(json['role'] as Map<String, dynamic>),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'subject': subject,
      'node_id': nodeId,
      'site_id': siteId,
      'role': role.toJson(),
    };
  }
}
