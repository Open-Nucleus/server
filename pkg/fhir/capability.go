package fhir

import (
	"encoding/json"
	"sort"
	"time"
)

// CapabilityConfig holds server metadata for the CapabilityStatement.
type CapabilityConfig struct {
	ServerName  string
	ServerURL   string
	Version     string
	PublishedAt time.Time
}

// GenerateCapabilityStatement auto-generates a FHIR R4 CapabilityStatement from the registry.
func GenerateCapabilityStatement(cfg CapabilityConfig) ([]byte, error) {
	published := cfg.PublishedAt
	if published.IsZero() {
		published = time.Now().UTC()
	}

	defs := AllResourceDefs()
	// Sort for deterministic output
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Type < defs[j].Type
	})

	resources := make([]map[string]any, 0, len(defs))
	for _, def := range defs {
		interactions := make([]map[string]any, 0, len(def.Interactions))
		for _, ia := range def.Interactions {
			interactions = append(interactions, map[string]any{"code": ia})
		}

		searchParams := make([]map[string]any, 0, len(def.SearchParams))
		for _, sp := range def.SearchParams {
			searchParams = append(searchParams, map[string]any{
				"name":       sp.Name,
				"type":       sp.Type,
				"definition": "http://hl7.org/fhir/SearchParameter/" + def.Type + "-" + sp.Name,
			})
		}

		res := map[string]any{
			"type":        def.Type,
			"interaction": interactions,
		}
		if len(searchParams) > 0 {
			res["searchParam"] = searchParams
		}
		resources = append(resources, res)
	}

	cs := map[string]any{
		"resourceType":  "CapabilityStatement",
		"status":        "active",
		"date":          published.Format("2006-01-02"),
		"kind":          "instance",
		"fhirVersion":   "4.0.1",
		"format":        []string{"json"},
		"patchFormat":   []string{"application/json-patch+json"},
		"implementation": map[string]any{
			"description": cfg.ServerName,
			"url":         cfg.ServerURL,
		},
		"software": map[string]any{
			"name":    cfg.ServerName,
			"version": cfg.Version,
		},
		"rest": []any{
			map[string]any{
				"mode":     "server",
				"resource": resources,
			},
		},
	}

	return json.Marshal(cs)
}
