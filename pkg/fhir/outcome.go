package fhir

import "encoding/json"

// OutcomeIssue represents a single issue within a FHIR OperationOutcome.
type OutcomeIssue struct {
	Severity    string // fatal | error | warning | information
	Code        string // FHIR issue-type code (e.g., "invalid", "not-found", "processing")
	Diagnostics string // human-readable description
	Location    string // FHIRPath or element path (optional)
}

// NewOperationOutcome builds a FHIR R4 OperationOutcome JSON from a list of issues.
func NewOperationOutcome(issues []OutcomeIssue) ([]byte, error) {
	fhirIssues := make([]map[string]any, 0, len(issues))
	for _, iss := range issues {
		entry := map[string]any{
			"severity":    iss.Severity,
			"code":        iss.Code,
			"diagnostics": iss.Diagnostics,
		}
		if iss.Location != "" {
			entry["expression"] = []string{iss.Location}
		}
		fhirIssues = append(fhirIssues, entry)
	}

	outcome := map[string]any{
		"resourceType": "OperationOutcome",
		"issue":        fhirIssues,
	}
	return json.Marshal(outcome)
}

// FromFieldErrors converts a slice of FieldError (from validation) into OperationOutcome JSON.
func FromFieldErrors(errs []FieldError) ([]byte, error) {
	issues := make([]OutcomeIssue, 0, len(errs))
	for _, fe := range errs {
		issues = append(issues, OutcomeIssue{
			Severity:    "error",
			Code:        fieldRuleToIssueCode(fe.Rule),
			Diagnostics: fe.Message,
			Location:    fe.Path,
		})
	}
	return NewOperationOutcome(issues)
}

// FromError creates a single-issue OperationOutcome from an error code and message.
func FromError(code, message string) ([]byte, error) {
	return NewOperationOutcome([]OutcomeIssue{
		{
			Severity:    "error",
			Code:        code,
			Diagnostics: message,
		},
	})
}

// fieldRuleToIssueCode maps validation rule names to FHIR issue-type codes.
func fieldRuleToIssueCode(rule string) string {
	switch rule {
	case "required":
		return "required"
	case "value_set", "value":
		return "code-invalid"
	case "json", "type":
		return "structure"
	default:
		return "invalid"
	}
}
