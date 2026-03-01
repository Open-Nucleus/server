package dosing

import "errors"

// ErrNotConfigured is returned by the stub engine when dosing is not available.
var ErrNotConfigured = errors.New("dosing engine not configured")

// ValidationResult represents the result of dosing validation.
type ValidationResult struct {
	Valid   bool
	Message string
}

// DosingOption represents a recommended dosing option.
type DosingOption struct {
	DoseValue  float64
	DoseUnit   string
	Frequency  string
	Route      string
	Indication string
}

// ScheduleEntry represents a single dose in a generated schedule.
type ScheduleEntry struct {
	Time      string
	DoseValue float64
	DoseUnit  string
	Note      string
}

// Engine defines the interface for dosing calculation.
// V1 uses StubEngine; future versions will integrate open-pharm-dosing.
type Engine interface {
	ValidateDose(medicationCode string, doseValue float64, doseUnit, frequency, route string, patientWeightKg float64) (*ValidationResult, error)
	GetOptions(medicationCode string, patientWeightKg float64) ([]DosingOption, error)
	GenerateSchedule(medicationCode string, doseValue float64, doseUnit, frequency, startTime string, durationDays int) ([]ScheduleEntry, error)
	Available() bool
}

// StubEngine returns "not configured" for all dosing operations.
type StubEngine struct{}

func NewStubEngine() *StubEngine {
	return &StubEngine{}
}

func (s *StubEngine) ValidateDose(medicationCode string, doseValue float64, doseUnit, frequency, route string, patientWeightKg float64) (*ValidationResult, error) {
	return &ValidationResult{
		Valid:   false,
		Message: "Dosing engine not configured. Install open-pharm-dosing for dosing validation.",
	}, nil
}

func (s *StubEngine) GetOptions(medicationCode string, patientWeightKg float64) ([]DosingOption, error) {
	return nil, nil
}

func (s *StubEngine) GenerateSchedule(medicationCode string, doseValue float64, doseUnit, frequency, startTime string, durationDays int) ([]ScheduleEntry, error) {
	return nil, nil
}

func (s *StubEngine) Available() bool {
	return false
}
