package fhir

import "fmt"

// ExtensionDef describes a FHIR extension expected by an Open Nucleus profile.
type ExtensionDef struct {
	URL       string // e.g. "http://opennucleus.dev/fhir/StructureDefinition/national-health-id"
	ValueType string // "valueString", "valueCoding", "valueDecimal", "valueInteger", "valueQuantity", "valueIdentifier"
	Required  bool
	Short     string // human-friendly name for StructureDefinition differential
}

// ExtractExtension returns the value of a named extension from a FHIR resource.
func ExtractExtension(resource map[string]any, url string) (any, bool) {
	exts, ok := getArray(resource, "extension")
	if !ok {
		return nil, false
	}
	for _, ext := range exts {
		extMap, ok := ext.(map[string]any)
		if !ok {
			continue
		}
		if getStr(extMap, "url") == url {
			// Return the value — could be any value[x] type
			for k, v := range extMap {
				if k != "url" && len(k) > 5 && k[:5] == "value" {
					return v, true
				}
			}
			return nil, true // extension present but no value
		}
	}
	return nil, false
}

// HasExtension returns true if the resource contains an extension with the given URL.
func HasExtension(resource map[string]any, url string) bool {
	_, found := ExtractExtension(resource, url)
	return found
}

// ValidateExtensions checks that required extensions are present and value types match.
// Unknown extensions pass through (FHIR open model).
func ValidateExtensions(resource map[string]any, knownExtensions []ExtensionDef) []FieldError {
	var errs []FieldError

	for _, def := range knownExtensions {
		if !def.Required {
			continue
		}
		if !HasExtension(resource, def.URL) {
			errs = append(errs, FieldError{
				Path:    "extension",
				Rule:    "required_extension",
				Message: fmt.Sprintf("Required extension %q is missing", def.URL),
			})
		}
	}

	// Validate value types for present extensions
	exts, ok := getArray(resource, "extension")
	if !ok {
		return errs
	}

	knownByURL := map[string]*ExtensionDef{}
	for i := range knownExtensions {
		knownByURL[knownExtensions[i].URL] = &knownExtensions[i]
	}

	for i, ext := range exts {
		extMap, ok := ext.(map[string]any)
		if !ok {
			continue
		}
		url := getStr(extMap, "url")
		def, known := knownByURL[url]
		if !known {
			continue // unknown extensions pass through
		}

		// Check value type matches
		if _, hasVal := extMap[def.ValueType]; !hasVal {
			// Check if any other value[x] key exists
			hasAnyValue := false
			for k := range extMap {
				if k != "url" && len(k) > 5 && k[:5] == "value" {
					hasAnyValue = true
					break
				}
			}
			if hasAnyValue {
				errs = append(errs, FieldError{
					Path:    fmt.Sprintf("extension[%d]", i),
					Rule:    "extension_type",
					Message: fmt.Sprintf("Extension %q expects %s", url, def.ValueType),
				})
			}
		}
	}

	return errs
}
