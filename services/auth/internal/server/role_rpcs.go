package server

import (
	"context"

	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListRoles(_ context.Context, _ *authv1.ListRolesRequest) (*authv1.ListRolesResponse, error) {
	roles := s.svc.ListRoles()
	protoRoles := make([]*authv1.RoleInfo, len(roles))
	for i, r := range roles {
		protoRoles[i] = roleToProto(r)
	}
	return &authv1.ListRolesResponse{Roles: protoRoles}, nil
}

func (s *Server) GetRole(_ context.Context, req *authv1.GetRoleRequest) (*authv1.GetRoleResponse, error) {
	role, ok := s.svc.GetRole(req.RoleCode)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "role not found: %s", req.RoleCode)
	}
	return &authv1.GetRoleResponse{Role: roleToProto(role)}, nil
}

func (s *Server) AssignRole(_ context.Context, req *authv1.AssignRoleRequest) (*authv1.AssignRoleResponse, error) {
	device, err := s.svc.AssignRole(req.DeviceId, req.Role, req.AssignedBy)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.AssignRoleResponse{Device: deviceToProto(device)}, nil
}
