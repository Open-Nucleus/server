package fhir

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ProvenanceContext holds the metadata needed to auto-generate a Provenance resource.
type ProvenanceContext struct {
	TargetResourceType string // e.g. "Patient"
	TargetResourceID   string // the resource's ID
	Activity           string // "CREATE", "UPDATE", "DELETE" — mapped to HL7 v3-DataOperation
	PractitionerID     string // who performed the action
	DeviceID           string // which node/device acted as custodian
	SiteID             string // source site
	Recorded           time.Time
}

// GenerateProvenance creates a FHIR R4 Provenance resource with target reference,
// activity coding (HL7 v3-DataOperation), agent, and recorded timestamp.
func GenerateProvenance(ctx ProvenanceContext) (jsonBytes []byte, id string, err error) {
	id = uuid.New().String()
	recorded := ctx.Recorded
	if recorded.IsZero() {
		recorded = time.Now().UTC()
	}

	activityCode := mapActivityCode(ctx.Activity)

	agents := []map[string]any{
		{
			"type": map[string]any{
				"coding": []any{
					map[string]any{
						"system":  "http://terminology.hl7.org/CodeSystem/provenance-participant-type",
						"code":    "author",
						"display": "Author",
					},
				},
			},
			"who": map[string]any{
				"reference": "Practitioner/" + ctx.PractitionerID,
			},
		},
	}

	if ctx.DeviceID != "" {
		agents = append(agents, map[string]any{
			"type": map[string]any{
				"coding": []any{
					map[string]any{
						"system":  "http://terminology.hl7.org/CodeSystem/provenance-participant-type",
						"code":    "custodian",
						"display": "Custodian",
					},
				},
			},
			"who": map[string]any{
				"reference": "Device/" + ctx.DeviceID,
			},
		})
	}

	prov := map[string]any{
		"resourceType": "Provenance",
		"id":           id,
		"target": []any{
			map[string]any{
				"reference": ctx.TargetResourceType + "/" + ctx.TargetResourceID,
			},
		},
		"recorded": recorded.Format(time.RFC3339),
		"activity": map[string]any{
			"coding": []any{
				map[string]any{
					"system":  "http://terminology.hl7.org/CodeSystem/v3-DataOperation",
					"code":    activityCode,
					"display": activityDisplay(activityCode),
				},
			},
		},
		"agent": agents,
		"meta": map[string]any{
			"source": ctx.SiteID,
		},
	}

	jsonBytes, err = json.Marshal(prov)
	return
}

func mapActivityCode(op string) string {
	switch op {
	case OpCreate:
		return "CREATE"
	case OpUpdate:
		return "UPDATE"
	case OpDelete:
		return "DELETE"
	default:
		return "UPDATE"
	}
}

func activityDisplay(code string) string {
	switch code {
	case "CREATE":
		return "create"
	case "UPDATE":
		return "revise"
	case "DELETE":
		return "delete"
	default:
		return "revise"
	}
}
