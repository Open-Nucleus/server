package server

import (
	"context"

	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
)

func (s *Server) ExportBundle(_ context.Context, req *syncv1.ExportBundleRequest) (*syncv1.ExportBundleResponse, error) {
	data, count, err := s.engine.ExportBundle(req.ResourceTypes, req.Since)
	if err != nil {
		return nil, mapError(err)
	}

	return &syncv1.ExportBundleResponse{
		BundleData:    data,
		Format:        "nucleus-bundle-v1",
		ResourceCount: int32(count),
	}, nil
}

func (s *Server) ImportBundle(_ context.Context, req *syncv1.ImportBundleRequest) (*syncv1.ImportBundleResponse, error) {
	imported, skipped, errors, err := s.engine.ImportBundle(req.BundleData)
	if err != nil {
		return nil, mapError(err)
	}

	return &syncv1.ImportBundleResponse{
		ResourcesImported: int32(imported),
		ResourcesSkipped:  int32(skipped),
		Errors:            errors,
	}, nil
}
