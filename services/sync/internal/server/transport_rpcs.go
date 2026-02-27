package server

import (
	"context"

	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
)

func (s *Server) ListTransports(_ context.Context, _ *syncv1.ListTransportsRequest) (*syncv1.ListTransportsResponse, error) {
	adapters := s.engine.ListTransports()

	protoTransports := make([]*syncv1.TransportInfo, len(adapters))
	for i, a := range adapters {
		caps := a.Capabilities()
		var capList []string
		if caps.Discovery {
			capList = append(capList, "discovery")
		}
		if caps.Streaming {
			capList = append(capList, "streaming")
		}
		if caps.Constrained {
			capList = append(capList, "constrained")
		}

		protoTransports[i] = &syncv1.TransportInfo{
			Name:         a.Name(),
			Enabled:      true,
			Available:    a.Available(),
			Capabilities: capList,
		}
	}

	return &syncv1.ListTransportsResponse{Transports: protoTransports}, nil
}

func (s *Server) EnableTransport(_ context.Context, req *syncv1.EnableTransportRequest) (*syncv1.EnableTransportResponse, error) {
	adapters := s.engine.ListTransports()
	for _, a := range adapters {
		if a.Name() == req.Name {
			return &syncv1.EnableTransportResponse{
				Transport: &syncv1.TransportInfo{
					Name:      a.Name(),
					Enabled:   true,
					Available: a.Available(),
				},
			}, nil
		}
	}
	return &syncv1.EnableTransportResponse{}, nil
}

func (s *Server) DisableTransport(_ context.Context, req *syncv1.DisableTransportRequest) (*syncv1.DisableTransportResponse, error) {
	adapters := s.engine.ListTransports()
	for _, a := range adapters {
		if a.Name() == req.Name {
			return &syncv1.DisableTransportResponse{
				Transport: &syncv1.TransportInfo{
					Name:      a.Name(),
					Enabled:   false,
					Available: a.Available(),
				},
			}, nil
		}
	}
	return &syncv1.DisableTransportResponse{}, nil
}
