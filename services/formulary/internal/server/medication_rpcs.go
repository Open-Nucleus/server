package server

import (
	"context"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/store"
)

func (s *Server) SearchMedications(_ context.Context, req *formularyv1.SearchMedicationsRequest) (*formularyv1.SearchMedicationsResponse, error) {
	page, perPage := paginationFromProto(req.Pagination)
	result := s.svc.SearchMedications(req.Query, req.Category, page, perPage)

	return &formularyv1.SearchMedicationsResponse{
		Medications: toMedicationProtos(result.Medications),
		Pagination:  paginationToProto(page, perPage, result.Total),
	}, nil
}

func (s *Server) GetMedication(_ context.Context, req *formularyv1.GetMedicationRequest) (*formularyv1.GetMedicationResponse, error) {
	med, err := s.svc.GetMedication(req.Code)
	if err != nil {
		return nil, mapError(err)
	}
	return &formularyv1.GetMedicationResponse{
		Medication: toMedicationProto(med),
	}, nil
}

func (s *Server) ListMedicationsByCategory(_ context.Context, req *formularyv1.ListMedicationsByCategoryRequest) (*formularyv1.ListMedicationsByCategoryResponse, error) {
	page, perPage := paginationFromProto(req.Pagination)
	result := s.svc.ListMedicationsByCategory(req.Category, page, perPage)

	return &formularyv1.ListMedicationsByCategoryResponse{
		Medications: toMedicationProtos(result.Medications),
		Pagination:  paginationToProto(page, perPage, result.Total),
	}, nil
}

// --- Proto converters ---

func toMedicationProto(r *store.MedicationRecord) *formularyv1.Medication {
	if r == nil {
		return nil
	}
	return &formularyv1.Medication{
		Code:              r.Code,
		Display:           r.Display,
		Form:              r.Form,
		Route:             r.Route,
		Category:          r.Category,
		Available:         true,
		WhoEssential:      r.WHOEssential,
		TherapeuticClass:  r.TherapeuticClass,
		CommonFrequencies: r.CommonFrequencies,
		Strength:          r.Strength,
		Unit:              r.Unit,
	}
}

func toMedicationProtos(recs []*store.MedicationRecord) []*formularyv1.Medication {
	out := make([]*formularyv1.Medication, 0, len(recs))
	for _, r := range recs {
		out = append(out, toMedicationProto(r))
	}
	return out
}

func paginationFromProto(pg *commonv1.PaginationRequest) (int, int) {
	page, perPage := 1, 25
	if pg != nil {
		if pg.Page > 0 {
			page = int(pg.Page)
		}
		if pg.PerPage > 0 {
			perPage = int(pg.PerPage)
		}
	}
	return page, perPage
}

func paginationToProto(page, perPage, total int) *commonv1.PaginationResponse {
	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}
	return &commonv1.PaginationResponse{
		Page:       int32(page),
		PerPage:    int32(perPage),
		Total:      int32(total),
		TotalPages: int32(totalPages),
	}
}
