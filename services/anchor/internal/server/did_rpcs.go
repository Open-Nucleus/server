package server

import (
	"context"

	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
	"github.com/FibrinLab/open-nucleus/pkg/openanchor"
)

func (s *Server) GetNodeDID(_ context.Context, _ *anchorv1.GetNodeDIDRequest) (*anchorv1.GetNodeDIDResponse, error) {
	doc, err := s.svc.GetNodeDID()
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.GetNodeDIDResponse{
		Document: didDocToProto(doc),
	}, nil
}

func (s *Server) GetDeviceDID(_ context.Context, req *anchorv1.GetDeviceDIDRequest) (*anchorv1.GetDeviceDIDResponse, error) {
	doc, err := s.svc.GetDeviceDID(req.DeviceId)
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.GetDeviceDIDResponse{
		Document: didDocToProto(doc),
	}, nil
}

func (s *Server) ResolveDID(_ context.Context, req *anchorv1.ResolveDIDRequest) (*anchorv1.ResolveDIDResponse, error) {
	doc, err := s.svc.ResolveDID(req.Did)
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.ResolveDIDResponse{
		Document: didDocToProto(doc),
	}, nil
}

func didDocToProto(doc *openanchor.DIDDocument) *anchorv1.DIDDocument {
	if doc == nil {
		return nil
	}
	vms := make([]*anchorv1.VerificationMethod, 0, len(doc.VerificationMethod))
	for _, vm := range doc.VerificationMethod {
		vms = append(vms, &anchorv1.VerificationMethod{
			Id:                vm.ID,
			Type:              vm.Type,
			Controller:        vm.Controller,
			PublicKeyMultibase: vm.PublicKeyMultibase,
		})
	}
	return &anchorv1.DIDDocument{
		Id:                 doc.ID,
		Context:            doc.Context,
		VerificationMethod: vms,
		Authentication:     doc.Authentication,
		AssertionMethod:    doc.AssertionMethod,
		Created:            doc.Created,
	}
}
