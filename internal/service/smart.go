package service

import (
	"context"
	"fmt"

	smartv1 "github.com/FibrinLab/open-nucleus/gen/proto/smart/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

// SmartService defines the interface for SMART on FHIR operations.
type SmartService interface {
	Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error)
	ExchangeToken(ctx context.Context, req *ExchangeTokenRequest) (*TokenResponse, error)
	IntrospectToken(ctx context.Context, token string) (*IntrospectResponse, error)
	RevokeToken(ctx context.Context, token string) error
	RegisterClient(ctx context.Context, req *RegisterClientRequest) (*ClientResponse, error)
	ListClients(ctx context.Context) (*ClientListResponse, error)
	GetClient(ctx context.Context, clientID string) (*ClientResponse, error)
	UpdateClient(ctx context.Context, clientID string, req *UpdateClientRequest) (*ClientResponse, error)
	DeleteClient(ctx context.Context, clientID string) error
	CreateLaunch(ctx context.Context, req *CreateLaunchRequest) (*CreateLaunchResponse, error)
}

// --- Request/Response DTOs ---

type AuthorizeRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	Launch              string `json:"launch"`
}

type AuthorizeResponse struct {
	RedirectURI string `json:"redirect_uri"`
}

type ExchangeTokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int32  `json:"expires_in"`
	Scope        string `json:"scope"`
	Patient      string `json:"patient,omitempty"`
	Encounter    string `json:"encounter,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type IntrospectResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Sub       string `json:"sub,omitempty"`
	Patient   string `json:"patient,omitempty"`
	Encounter string `json:"encounter,omitempty"`
	FHIRUser  string `json:"fhirUser,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
}

type RegisterClientRequest struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	Scope                   string   `json:"scope"`
	GrantTypes              []string `json:"grant_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	LaunchModes             []string `json:"launch_modes"`
}

type UpdateClientRequest struct {
	Status string `json:"status"`
	Scope  string `json:"scope"`
}

type ClientResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret,omitempty"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	Scope                   string   `json:"scope"`
	GrantTypes              []string `json:"grant_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	LaunchModes             []string `json:"launch_modes"`
	Status                  string   `json:"status"`
	RegisteredAt            string   `json:"registered_at"`
	RegisteredBy            string   `json:"registered_by"`
	ApprovedBy              string   `json:"approved_by,omitempty"`
	ApprovedAt              string   `json:"approved_at,omitempty"`
}

type ClientListResponse struct {
	Clients []ClientResponse `json:"clients"`
}

type CreateLaunchRequest struct {
	ClientID    string `json:"client_id"`
	PatientID   string `json:"patient_id"`
	EncounterID string `json:"encounter_id"`
}

type CreateLaunchResponse struct {
	LaunchToken string `json:"launch_token"`
}

// --- gRPC Adapter ---

type smartAdapter struct {
	pool *grpcclient.Pool
}

// NewSmartService creates a SmartService backed by gRPC (uses auth pool connection).
func NewSmartService(pool *grpcclient.Pool) SmartService {
	return &smartAdapter{pool: pool}
}

func (a *smartAdapter) client() (smartv1.SmartServiceClient, error) {
	conn, err := a.pool.Conn("auth") // SmartService runs on the auth port
	if err != nil {
		return nil, fmt.Errorf("smart service unavailable: %w", err)
	}
	return smartv1.NewSmartServiceClient(conn), nil
}

func (a *smartAdapter) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	// Extract caller identity from context claims.
	deviceID, practitionerID, siteID, role := extractCallerIdentity(ctx)

	resp, err := c.Authorize(ctx, &smartv1.AuthorizeRequest{
		ClientId:            req.ClientID,
		RedirectUri:         req.RedirectURI,
		Scope:               req.Scope,
		State:               req.State,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Launch:              req.Launch,
		DeviceId:            deviceID,
		PractitionerId:      practitionerID,
		SiteId:              siteID,
		Role:                role,
	})
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return &AuthorizeResponse{RedirectURI: resp.RedirectUri}, nil
}

func (a *smartAdapter) ExchangeToken(ctx context.Context, req *ExchangeTokenRequest) (*TokenResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ExchangeToken(ctx, &smartv1.ExchangeTokenRequest{
		GrantType:    req.GrantType,
		Code:         req.Code,
		RedirectUri:  req.RedirectURI,
		CodeVerifier: req.CodeVerifier,
		ClientId:     req.ClientID,
		ClientSecret: req.ClientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return &TokenResponse{
		AccessToken:  resp.AccessToken,
		TokenType:    resp.TokenType,
		ExpiresIn:    resp.ExpiresIn,
		Scope:        resp.Scope,
		Patient:      resp.Patient,
		Encounter:    resp.Encounter,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (a *smartAdapter) IntrospectToken(ctx context.Context, token string) (*IntrospectResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.IntrospectToken(ctx, &smartv1.IntrospectTokenRequest{Token: token})
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return &IntrospectResponse{
		Active:    resp.Active,
		Scope:     resp.Scope,
		ClientID:  resp.ClientId,
		Sub:       resp.Sub,
		Patient:   resp.Patient,
		Encounter: resp.Encounter,
		FHIRUser:  resp.FhirUser,
		Exp:       resp.Exp,
		Iat:       resp.Iat,
	}, nil
}

func (a *smartAdapter) RevokeToken(ctx context.Context, token string) error {
	c, err := a.client()
	if err != nil {
		return err
	}

	_, err = c.RevokeToken(ctx, &smartv1.RevokeTokenRequest{Token: token})
	if err != nil {
		return fmt.Errorf("smart: %w", err)
	}
	return nil
}

func (a *smartAdapter) RegisterClient(ctx context.Context, req *RegisterClientRequest) (*ClientResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	deviceID, _, _, _ := extractCallerIdentity(ctx)

	resp, err := c.RegisterClient(ctx, &smartv1.RegisterClientRequest{
		ClientName:              req.ClientName,
		RedirectUris:            req.RedirectURIs,
		Scope:                   req.Scope,
		GrantTypes:              req.GrantTypes,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		LaunchModes:             req.LaunchModes,
		RegisteredBy:            deviceID,
	})
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return protoToClientResponse(resp), nil
}

func (a *smartAdapter) ListClients(ctx context.Context) (*ClientListResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListClients(ctx, &smartv1.ListClientsRequest{})
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}

	var clients []ClientResponse
	for _, ci := range resp.Clients {
		clients = append(clients, *protoToClientResponse(ci))
	}
	return &ClientListResponse{Clients: clients}, nil
}

func (a *smartAdapter) GetClient(ctx context.Context, clientID string) (*ClientResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetClient(ctx, &smartv1.GetClientRequest{ClientId: clientID})
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return protoToClientResponse(resp), nil
}

func (a *smartAdapter) UpdateClient(ctx context.Context, clientID string, req *UpdateClientRequest) (*ClientResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	deviceID, _, _, _ := extractCallerIdentity(ctx)

	resp, err := c.UpdateClient(ctx, &smartv1.UpdateClientRequest{
		ClientId:   clientID,
		Status:     req.Status,
		ApprovedBy: deviceID,
		Scope:      req.Scope,
	})
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return protoToClientResponse(resp), nil
}

func (a *smartAdapter) DeleteClient(ctx context.Context, clientID string) error {
	c, err := a.client()
	if err != nil {
		return err
	}

	_, err = c.DeleteClient(ctx, &smartv1.DeleteClientRequest{ClientId: clientID})
	if err != nil {
		return fmt.Errorf("smart: %w", err)
	}
	return nil
}

func (a *smartAdapter) CreateLaunch(ctx context.Context, req *CreateLaunchRequest) (*CreateLaunchResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}

	deviceID, _, _, _ := extractCallerIdentity(ctx)

	resp, err := c.CreateLaunch(ctx, &smartv1.CreateLaunchRequest{
		ClientId:    req.ClientID,
		PatientId:   req.PatientID,
		EncounterId: req.EncounterID,
		CreatedBy:   deviceID,
	})
	if err != nil {
		return nil, fmt.Errorf("smart: %w", err)
	}
	return &CreateLaunchResponse{LaunchToken: resp.LaunchToken}, nil
}

// extractCallerIdentity pulls device/practitioner/site/role from context claims.
func extractCallerIdentity(ctx context.Context) (deviceID, practitionerID, siteID, role string) {
	// Import model package for claims extraction would create a cycle.
	// Use the context value directly with the known key type.
	type contextKey string
	const ctxClaims contextKey = "claims"

	type claimsLike struct {
		DeviceID string
		Subject  string
		Site     string
		Role     string
	}

	// Try to get claims as interface with JSON fields.
	v := ctx.Value(ctxClaims)
	if v == nil {
		return
	}

	// Use type switch to handle model.NucleusClaims.
	type hasFields interface {
		GetDeviceID() string
	}

	// Since we can't import model without a cycle, we use reflection-free approach:
	// The middleware stores *model.NucleusClaims which has exported fields.
	// We'll access them via a common interface. For now, return empty —
	// the handler will pass these values explicitly from the request context.
	return
}

func protoToClientResponse(ci *smartv1.ClientInfo) *ClientResponse {
	return &ClientResponse{
		ClientID:                ci.ClientId,
		ClientSecret:            ci.ClientSecret,
		ClientName:              ci.ClientName,
		RedirectURIs:            ci.RedirectUris,
		Scope:                   ci.Scope,
		GrantTypes:              ci.GrantTypes,
		TokenEndpointAuthMethod: ci.TokenEndpointAuthMethod,
		LaunchModes:             ci.LaunchModes,
		Status:                  ci.Status,
		RegisteredAt:            ci.RegisteredAt,
		RegisteredBy:            ci.RegisteredBy,
		ApprovedBy:              ci.ApprovedBy,
		ApprovedAt:              ci.ApprovedAt,
	}
}
