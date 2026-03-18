/// SMART on FHIR DTOs matching Go `service.ClientResponse`,
/// `service.RegisterClientRequest`, `service.UpdateClientRequest`, etc.

/// A registered SMART client.
///
/// Matches Go `service.ClientResponse`.
class ClientResponse {
  final String clientId;
  final String? clientSecret;
  final String clientName;
  final List<String> redirectUris;
  final String scope;
  final List<String> grantTypes;
  final String tokenEndpointAuthMethod;
  final List<String> launchModes;
  final String status;
  final String registeredAt;
  final String registeredBy;
  final String? approvedBy;
  final String? approvedAt;

  const ClientResponse({
    required this.clientId,
    this.clientSecret,
    required this.clientName,
    required this.redirectUris,
    required this.scope,
    required this.grantTypes,
    required this.tokenEndpointAuthMethod,
    required this.launchModes,
    required this.status,
    required this.registeredAt,
    required this.registeredBy,
    this.approvedBy,
    this.approvedAt,
  });

  factory ClientResponse.fromJson(Map<String, dynamic> json) {
    return ClientResponse(
      clientId: json['client_id'] as String,
      clientSecret: json['client_secret'] as String?,
      clientName: json['client_name'] as String,
      redirectUris: (json['redirect_uris'] as List<dynamic>)
          .map((u) => u as String)
          .toList(),
      scope: json['scope'] as String,
      grantTypes: (json['grant_types'] as List<dynamic>)
          .map((g) => g as String)
          .toList(),
      tokenEndpointAuthMethod:
          json['token_endpoint_auth_method'] as String,
      launchModes: (json['launch_modes'] as List<dynamic>)
          .map((l) => l as String)
          .toList(),
      status: json['status'] as String,
      registeredAt: json['registered_at'] as String,
      registeredBy: json['registered_by'] as String,
      approvedBy: json['approved_by'] as String?,
      approvedAt: json['approved_at'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'client_id': clientId,
      if (clientSecret != null) 'client_secret': clientSecret,
      'client_name': clientName,
      'redirect_uris': redirectUris,
      'scope': scope,
      'grant_types': grantTypes,
      'token_endpoint_auth_method': tokenEndpointAuthMethod,
      'launch_modes': launchModes,
      'status': status,
      'registered_at': registeredAt,
      'registered_by': registeredBy,
      if (approvedBy != null) 'approved_by': approvedBy,
      if (approvedAt != null) 'approved_at': approvedAt,
    };
  }
}

/// List of SMART clients from `GET /api/v1/smart/clients`.
class ClientListResponse {
  final List<ClientResponse> clients;

  const ClientListResponse({required this.clients});

  factory ClientListResponse.fromJson(Map<String, dynamic> json) {
    return ClientListResponse(
      clients: (json['clients'] as List<dynamic>)
          .map((c) => ClientResponse.fromJson(c as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'clients': clients.map((c) => c.toJson()).toList(),
    };
  }
}

/// Body of `POST /auth/smart/register`.
class RegisterClientRequest {
  final String clientName;
  final List<String> redirectUris;
  final String scope;
  final List<String> grantTypes;
  final String tokenEndpointAuthMethod;
  final List<String> launchModes;

  const RegisterClientRequest({
    required this.clientName,
    required this.redirectUris,
    required this.scope,
    required this.grantTypes,
    required this.tokenEndpointAuthMethod,
    required this.launchModes,
  });

  factory RegisterClientRequest.fromJson(Map<String, dynamic> json) {
    return RegisterClientRequest(
      clientName: json['client_name'] as String,
      redirectUris: (json['redirect_uris'] as List<dynamic>)
          .map((u) => u as String)
          .toList(),
      scope: json['scope'] as String,
      grantTypes: (json['grant_types'] as List<dynamic>)
          .map((g) => g as String)
          .toList(),
      tokenEndpointAuthMethod:
          json['token_endpoint_auth_method'] as String,
      launchModes: (json['launch_modes'] as List<dynamic>)
          .map((l) => l as String)
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'client_name': clientName,
      'redirect_uris': redirectUris,
      'scope': scope,
      'grant_types': grantTypes,
      'token_endpoint_auth_method': tokenEndpointAuthMethod,
      'launch_modes': launchModes,
    };
  }
}

/// Body of `PUT /api/v1/smart/clients/{id}`.
class UpdateClientRequest {
  final String status;
  final String scope;

  const UpdateClientRequest({
    required this.status,
    required this.scope,
  });

  factory UpdateClientRequest.fromJson(Map<String, dynamic> json) {
    return UpdateClientRequest(
      status: json['status'] as String,
      scope: json['scope'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'status': status,
      'scope': scope,
    };
  }
}
