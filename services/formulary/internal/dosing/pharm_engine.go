package dosing

import (
	"fmt"
	"time"

	pharm "github.com/Open-Nucleus/open-pharm-dosing"
)

// PharmEngine wraps the open-pharm-dosing library to implement the Engine interface.
type PharmEngine struct{}

// NewPharmEngine creates a dosing engine backed by open-pharm-dosing.
func NewPharmEngine() *PharmEngine {
	return &PharmEngine{}
}

func (e *PharmEngine) Available() bool {
	return true
}

func (e *PharmEngine) ValidateDose(medicationCode string, doseValue float64, doseUnit, frequency, route string, patientWeightKg float64) (*ValidationResult, error) {
	// Validate frequency code
	if err := pharm.Validate(frequency); err != nil {
		return &ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("Invalid frequency %q: %v", frequency, err),
		}, nil
	}

	// Build a DosingInstruction for clinical validation
	fc, err := pharm.Parse(frequency)
	if err != nil {
		return &ValidationResult{Valid: false, Message: err.Error()}, nil
	}

	instruction := &pharm.DosingInstruction{
		Frequency: fc,
		Dose:      &pharm.Dose{Value: doseValue, Unit: doseUnit},
		Route:     route,
	}

	warnings := pharm.ValidateInstruction(instruction)
	for _, w := range warnings {
		if w.Level == "error" {
			return &ValidationResult{
				Valid:   false,
				Message: fmt.Sprintf("%s: %s", w.Field, w.Message),
			}, nil
		}
	}

	msg := "Dosing validated"
	if len(warnings) > 0 {
		msg = fmt.Sprintf("Valid with %d warning(s): %s", len(warnings), warnings[0].Message)
	}

	return &ValidationResult{Valid: true, Message: msg}, nil
}

func (e *PharmEngine) GetOptions(medicationCode string, patientWeightKg float64) ([]DosingOption, error) {
	// Return all available frequency codes as dosing options
	codes := pharm.List()
	options := make([]DosingOption, 0, len(codes))
	for _, fc := range codes {
		display, _ := pharm.ToText(fc.Code, pharm.LocaleEnGB)
		options = append(options, DosingOption{
			Frequency:  fc.Code,
			Indication: display,
		})
	}
	return options, nil
}

func (e *PharmEngine) GenerateSchedule(medicationCode string, doseValue float64, doseUnit, frequency, startTime string, durationDays int) ([]ScheduleEntry, error) {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}

	times, err := pharm.Schedule(frequency, start, durationDays)
	if err != nil {
		return nil, fmt.Errorf("schedule generation: %w", err)
	}

	display, _ := pharm.ToText(frequency, pharm.LocaleEnGB)

	entries := make([]ScheduleEntry, len(times))
	for i, t := range times {
		entries[i] = ScheduleEntry{
			Time:      t.Format(time.RFC3339),
			DoseValue: doseValue,
			DoseUnit:  doseUnit,
			Note:      display,
		}
	}
	return entries, nil
}
