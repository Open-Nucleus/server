package server

import (
	"context"

	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
)

func (s *Server) ValidateDosing(_ context.Context, req *formularyv1.ValidateDosingRequest) (*formularyv1.ValidateDosingResponse, error) {
	result, configured := s.svc.ValidateDosing(
		req.MedicationCode, req.DoseValue, req.DoseUnit,
		req.Frequency, req.Route, req.PatientWeightKg,
	)

	resp := &formularyv1.ValidateDosingResponse{
		Configured: configured,
	}
	if result != nil {
		resp.Valid = result.Valid
		resp.Message = result.Message
	} else {
		resp.Message = "Dosing engine not configured"
	}
	return resp, nil
}

func (s *Server) GetDosingOptions(_ context.Context, req *formularyv1.GetDosingOptionsRequest) (*formularyv1.GetDosingOptionsResponse, error) {
	opts, configured := s.svc.GetDosingOptions(req.MedicationCode, req.PatientWeightKg)

	resp := &formularyv1.GetDosingOptionsResponse{
		Configured: configured,
	}
	if !configured {
		resp.Message = "Dosing engine not configured"
	}
	for _, o := range opts {
		resp.Options = append(resp.Options, &formularyv1.DosingOption{
			DoseValue:  o.DoseValue,
			DoseUnit:   o.DoseUnit,
			Frequency:  o.Frequency,
			Route:      o.Route,
			Indication: o.Indication,
		})
	}
	return resp, nil
}

func (s *Server) GenerateSchedule(_ context.Context, req *formularyv1.GenerateScheduleRequest) (*formularyv1.GenerateScheduleResponse, error) {
	entries, configured := s.svc.GenerateSchedule(
		req.MedicationCode, req.DoseValue, req.DoseUnit,
		req.Frequency, req.StartTime, int(req.DurationDays),
	)

	resp := &formularyv1.GenerateScheduleResponse{
		Configured: configured,
	}
	if !configured {
		resp.Message = "Dosing engine not configured"
	}
	for _, e := range entries {
		resp.Entries = append(resp.Entries, &formularyv1.ScheduleEntry{
			Time:      e.Time,
			DoseValue: e.DoseValue,
			DoseUnit:  e.DoseUnit,
			Note:      e.Note,
		})
	}
	return resp, nil
}
