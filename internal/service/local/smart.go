package local

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/services/auth/authservice"
)

// smartSvc implements service.SmartService by calling
// authservice.SmartService directly (no gRPC).
type smartSvc struct {
	real *authservice.SmartService
}

// NewSmartService creates a local adapter for SMART on FHIR operations.
func NewSmartService(real *authservice.SmartService) service.SmartService {
	return &smartSvc{real: real}
}

// --- helpers ---

// callerFromCtx extracts device/practitioner/site/role from the JWT
// claims stored in context. Mirrors the gRPC adapter's
// extractCallerIdentity helper.
func callerFromCtx(ctx context.Context) (deviceID, practitionerID, siteID, role string) {
	claims := model.ClaimsFromContext(ctx)
	if claims == nil {
		return
	}
	deviceID = claims.DeviceID
	practitionerID = claims.Subject
	siteID = claims.Site
	role = claims.Role
	return
}

// toClientResponse converts smart.Client fields into the gateway DTO,
// producing the same output as the gRPC adapter's protoToClientResponse.
func toClientResponse(id, secret, name string, redirectURIs []string, scope string, grantTypes []string, authMethod string, launchModes []string, status, registeredAt, registeredBy, approvedBy, approvedAt string) *service.ClientResponse {
	return &service.ClientResponse{
		ClientID:                id,
		ClientSecret:            secret,
		ClientName:              name,
		RedirectURIs:            redirectURIs,
		Scope:                   scope,
		GrantTypes:              grantTypes,
		TokenEndpointAuthMethod: authMethod,
		LaunchModes:             launchModes,
		Status:                  status,
		RegisteredAt:            registeredAt,
		RegisteredBy:            registeredBy,
		ApprovedBy:              approvedBy,
		ApprovedAt:              approvedAt,
	}
}

// --- SmartService interface ---

func (s *smartSvc) Authorize(ctx context.Context, req *service.AuthorizeRequest) (*service.AuthorizeResponse, error) {
	deviceID, practitionerID, siteID, role := callerFromCtx(ctx)

	redirectURI, err := s.real.Authorize(
		req.ClientID, req.RedirectURI, req.Scope, req.State,
		req.CodeChallenge, req.CodeChallengeMethod, req.Launch,
		deviceID, practitionerID, siteID, role,
	)
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return &service.AuthorizeResponse{RedirectURI: redirectURI}, nil
}

func (s *smartSvc) ExchangeToken(_ context.Context, req *service.ExchangeTokenRequest) (*service.TokenResponse, error) {
	accessToken, expiresIn, scope, patient, encounter, err := s.real.ExchangeToken(
		req.GrantType, req.Code, req.RedirectURI, req.CodeVerifier, req.ClientID, req.ClientSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return &service.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		Scope:       scope,
		Patient:     patient,
		Encounter:   encounter,
	}, nil
}

func (s *smartSvc) IntrospectToken(_ context.Context, token string) (*service.IntrospectResponse, error) {
	active, scope, clientID, sub, patient, encounter, fhirUser, exp, iat, err := s.real.IntrospectToken(token)
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return &service.IntrospectResponse{
		Active:    active,
		Scope:     scope,
		ClientID:  clientID,
		Sub:       sub,
		Patient:   patient,
		Encounter: encounter,
		FHIRUser:  fhirUser,
		Exp:       exp,
		Iat:       iat,
	}, nil
}

func (s *smartSvc) RevokeToken(_ context.Context, token string) error {
	if err := s.real.RevokeToken(token); err != nil {
		return fmt.Errorf("smart: %w", err)
	}
	return nil
}

func (s *smartSvc) RegisterClient(ctx context.Context, req *service.RegisterClientRequest) (*service.ClientResponse, error) {
	deviceID, _, _, _ := callerFromCtx(ctx)

	client, err := s.real.RegisterClient(
		req.ClientName, req.RedirectURIs, req.Scope, req.GrantTypes,
		req.TokenEndpointAuthMethod, req.LaunchModes, deviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return toClientResponse(
		client.ClientID, client.ClientSecret, client.ClientName,
		client.RedirectURIs, client.Scope, client.GrantTypes,
		client.TokenEndpointAuthMethod, client.LaunchModes,
		string(client.Status), client.RegisteredAt, client.RegisteredBy,
		client.ApprovedBy, client.ApprovedAt,
	), nil
}

func (s *smartSvc) ListClients(_ context.Context) (*service.ClientListResponse, error) {
	clients, err := s.real.ListClients()
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}

	var out []service.ClientResponse
	for _, c := range clients {
		out = append(out, *toClientResponse(
			c.ClientID, c.ClientSecret, c.ClientName,
			c.RedirectURIs, c.Scope, c.GrantTypes,
			c.TokenEndpointAuthMethod, c.LaunchModes,
			string(c.Status), c.RegisteredAt, c.RegisteredBy,
			c.ApprovedBy, c.ApprovedAt,
		))
	}
	return &service.ClientListResponse{Clients: out}, nil
}

func (s *smartSvc) GetClient(_ context.Context, clientID string) (*service.ClientResponse, error) {
	client, err := s.real.GetClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return toClientResponse(
		client.ClientID, client.ClientSecret, client.ClientName,
		client.RedirectURIs, client.Scope, client.GrantTypes,
		client.TokenEndpointAuthMethod, client.LaunchModes,
		string(client.Status), client.RegisteredAt, client.RegisteredBy,
		client.ApprovedBy, client.ApprovedAt,
	), nil
}

func (s *smartSvc) UpdateClient(ctx context.Context, clientID string, req *service.UpdateClientRequest) (*service.ClientResponse, error) {
	deviceID, _, _, _ := callerFromCtx(ctx)

	client, err := s.real.UpdateClient(clientID, req.Status, deviceID, req.Scope)
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return toClientResponse(
		client.ClientID, client.ClientSecret, client.ClientName,
		client.RedirectURIs, client.Scope, client.GrantTypes,
		client.TokenEndpointAuthMethod, client.LaunchModes,
		string(client.Status), client.RegisteredAt, client.RegisteredBy,
		client.ApprovedBy, client.ApprovedAt,
	), nil
}

func (s *smartSvc) DeleteClient(_ context.Context, clientID string) error {
	if err := s.real.DeleteClient(clientID); err != nil {
		return fmt.Errorf("smart: %w", err)
	}
	return nil
}

func (s *smartSvc) CreateLaunch(ctx context.Context, req *service.CreateLaunchRequest) (*service.CreateLaunchResponse, error) {
	deviceID, _, _, _ := callerFromCtx(ctx)

	token, err := s.real.CreateLaunch(req.ClientID, req.PatientID, req.EncounterID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return &service.CreateLaunchResponse{LaunchToken: token}, nil
}
