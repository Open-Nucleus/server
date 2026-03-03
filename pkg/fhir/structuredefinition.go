package fhir

import "encoding/json"

// GenerateStructureDefinition generates a FHIR R4 StructureDefinition JSON
// from a ProfileDef. Used for the /fhir/StructureDefinition endpoint.
func GenerateStructureDefinition(def *ProfileDef) ([]byte, error) {
	elements := make([]map[string]any, 0, len(def.Extensions)+1)

	// Root element
	elements = append(elements, map[string]any{
		"id":   def.BaseResource,
		"path": def.BaseResource,
	})

	// One element per extension
	for _, ext := range def.Extensions {
		elem := map[string]any{
			"id":    def.BaseResource + ".extension:" + ext.Short,
			"path":  def.BaseResource + ".extension",
			"short": ext.Short,
			"type": []map[string]any{
				{"code": "Extension", "profile": []string{ext.URL}},
			},
		}
		if ext.Required {
			elem["min"] = 1
		} else {
			elem["min"] = 0
		}
		elem["max"] = "1"
		elements = append(elements, elem)
	}

	sd := map[string]any{
		"resourceType":   "StructureDefinition",
		"id":             def.Name,
		"url":            def.URL,
		"name":           def.Name,
		"title":          def.Title,
		"status":         "active",
		"kind":           "resource",
		"abstract":       false,
		"type":           def.BaseResource,
		"baseDefinition": "http://hl7.org/fhir/StructureDefinition/" + def.BaseResource,
		"derivation":     "constraint",
		"fhirVersion":    "4.0.1",
		"version":        def.Version,
		"description":    def.Description,
		"differential": map[string]any{
			"element": elements,
		},
	}

	return json.Marshal(sd)
}

// GenerateAllStructureDefinitions generates StructureDefinition JSON for every
// registered profile, keyed by profile name.
func GenerateAllStructureDefinitions() (map[string][]byte, error) {
	defs := AllProfileDefs()
	result := make(map[string][]byte, len(defs))
	for _, def := range defs {
		data, err := GenerateStructureDefinition(def)
		if err != nil {
			return nil, err
		}
		result[def.Name] = data
	}
	return result, nil
}
