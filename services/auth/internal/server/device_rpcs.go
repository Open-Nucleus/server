package server

import (
	"context"
	"time"

	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/service"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListDevices(_ context.Context, _ *authv1.ListDevicesRequest) (*authv1.ListDevicesResponse, error) {
	devices, err := s.svc.ListDevices()
	if err != nil {
		return nil, mapError(err)
	}
	protoDevices := make([]*authv1.DeviceInfo, len(devices))
	for i, d := range devices {
		protoDevices[i] = deviceToProto(d)
	}
	return &authv1.ListDevicesResponse{Devices: protoDevices}, nil
}

func (s *Server) RevokeDevice(_ context.Context, req *authv1.RevokeDeviceRequest) (*authv1.RevokeDeviceResponse, error) {
	device, err := s.svc.RevokeDevice(req.DeviceId, req.RevokedBy, req.Reason)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.RevokeDeviceResponse{Device: deviceToProto(device)}, nil
}

func (s *Server) CheckRevocation(_ context.Context, req *authv1.CheckRevocationRequest) (*authv1.CheckRevocationResponse, error) {
	// Check Git-based device status
	devices, err := s.svc.ListDevices()
	if err != nil {
		return nil, mapError(err)
	}
	for _, d := range devices {
		if d.DeviceID == req.DeviceId {
			return &authv1.CheckRevocationResponse{
				Revoked:   d.Status == "revoked",
				RevokedAt: d.RevokedAt,
				Reason:    d.RevocationReason,
			}, nil
		}
	}
	return &authv1.CheckRevocationResponse{Revoked: false}, nil
}

func deviceToProto(d *service.DeviceRecord) *authv1.DeviceInfo {
	return &authv1.DeviceInfo{
		DeviceId:         d.DeviceID,
		PublicKey:        d.PublicKey,
		PractitionerId:   d.PractitionerID,
		SiteId:           d.SiteID,
		DeviceName:       d.DeviceName,
		Role:             d.Role,
		Status:           d.Status,
		RegisteredAt:     d.RegisteredAt,
		RevokedAt:        d.RevokedAt,
		RevokedBy:        d.RevokedBy,
		RevocationReason: d.RevocationReason,
	}
}

func timestamppb(t time.Time) *tspb.Timestamp {
	return tspb.New(t)
}
