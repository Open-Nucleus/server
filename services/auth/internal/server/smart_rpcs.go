package server

import (
	"context"
	"time"

	smartv1 "github.com/FibrinLab/open-nucleus/gen/proto/smart/v1"
	"github.com/FibrinLab/open-nucleus/pkg/smart"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// SmartServer implements the SmartService gRPC server.
type SmartServer struct {
	smartv1.UnimplementedSmartServiceServer
	svc *service.SmartService
}

// NewSmartServer creates a new SMART gRPC server.
func NewSmartServer(svc *service.SmartService) *SmartServer {
	return &SmartServer{svc: svc}
}

func (s *SmartServer) Authorize(ctx context.Context, req *smartv1.AuthorizeRequest) (*smartv1.AuthorizeResponse, error) {
	redirectURI, err := s.svc.Authorize(
		req.ClientId, req.RedirectUri, req.Scope, req.State,
		req.CodeChallenge, req.CodeChallengeMethod, req.Launch,
		req.DeviceId, req.PractitionerId, req.SiteId, req.Role,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &smartv1.AuthorizeResponse{RedirectUri: redirectURI}, nil
}

func (s *SmartServer) ExchangeToken(ctx context.Context, req *smartv1.ExchangeTokenRequest) (*smartv1.TokenResponse, error) {
	accessToken, expiresIn, scope, patient, encounter, err := s.svc.ExchangeToken(
		req.GrantType, req.Code, req.RedirectUri, req.CodeVerifier, req.ClientId, req.ClientSecret,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &smartv1.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		Scope:       scope,
		Patient:     patient,
		Encounter:   encounter,
	}, nil
}

func (s *SmartServer) IntrospectToken(ctx context.Context, req *smartv1.IntrospectTokenRequest) (*smartv1.IntrospectResponse, error) {
	active, scope, clientID, sub, patient, encounter, fhirUser, exp, iat, err := s.svc.IntrospectToken(req.Token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &smartv1.IntrospectResponse{
		Active:    active,
		Scope:     scope,
		ClientId:  clientID,
		Sub:       sub,
		Patient:   patient,
		Encounter: encounter,
		FhirUser:  fhirUser,
		Exp:       exp,
		Iat:       iat,
	}, nil
}

func (s *SmartServer) RevokeToken(ctx context.Context, req *smartv1.RevokeTokenRequest) (*smartv1.RevokeTokenResponse, error) {
	if err := s.svc.RevokeToken(req.Token); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &smartv1.RevokeTokenResponse{}, nil
}

func (s *SmartServer) RegisterClient(ctx context.Context, req *smartv1.RegisterClientRequest) (*smartv1.ClientInfo, error) {
	client, err := s.svc.RegisterClient(
		req.ClientName, req.RedirectUris, req.Scope, req.GrantTypes,
		req.TokenEndpointAuthMethod, req.LaunchModes, req.RegisteredBy,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toClientInfo(client), nil
}

func (s *SmartServer) ListClients(ctx context.Context, req *smartv1.ListClientsRequest) (*smartv1.ListClientsResponse, error) {
	clients, err := s.svc.ListClients()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var infos []*smartv1.ClientInfo
	for _, c := range clients {
		infos = append(infos, toClientInfo(c))
	}
	return &smartv1.ListClientsResponse{Clients: infos}, nil
}

func (s *SmartServer) GetClient(ctx context.Context, req *smartv1.GetClientRequest) (*smartv1.ClientInfo, error) {
	client, err := s.svc.GetClient(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toClientInfo(client), nil
}

func (s *SmartServer) UpdateClient(ctx context.Context, req *smartv1.UpdateClientRequest) (*smartv1.ClientInfo, error) {
	client, err := s.svc.UpdateClient(req.ClientId, req.Status, req.ApprovedBy, req.Scope)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toClientInfo(client), nil
}

func (s *SmartServer) DeleteClient(ctx context.Context, req *smartv1.DeleteClientRequest) (*smartv1.DeleteClientResponse, error) {
	if err := s.svc.DeleteClient(req.ClientId); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &smartv1.DeleteClientResponse{}, nil
}

func (s *SmartServer) CreateLaunch(ctx context.Context, req *smartv1.CreateLaunchRequest) (*smartv1.CreateLaunchResponse, error) {
	token, err := s.svc.CreateLaunch(req.ClientId, req.PatientId, req.EncounterId, req.CreatedBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &smartv1.CreateLaunchResponse{LaunchToken: token}, nil
}

func (s *SmartServer) Health(ctx context.Context, req *smartv1.HealthRequest) (*smartv1.HealthResponse, error) {
	return &smartv1.HealthResponse{
		Status:    "ok",
		Timestamp: tspb.New(time.Now()),
	}, nil
}

func toClientInfo(c *smart.Client) *smartv1.ClientInfo {
	return &smartv1.ClientInfo{
		ClientId:                c.ClientID,
		ClientSecret:            c.ClientSecret,
		ClientName:              c.ClientName,
		RedirectUris:            c.RedirectURIs,
		Scope:                   c.Scope,
		GrantTypes:              c.GrantTypes,
		TokenEndpointAuthMethod: c.TokenEndpointAuthMethod,
		LaunchModes:             c.LaunchModes,
		Status:                  string(c.Status),
		RegisteredAt:            c.RegisteredAt,
		RegisteredBy:            c.RegisteredBy,
		ApprovedBy:              c.ApprovedBy,
		ApprovedAt:              c.ApprovedAt,
	}
}
