package fhir

// ProfileBaseURL is the base URL for Open Nucleus FHIR profiles.
const ProfileBaseURL = "http://opennucleus.dev/fhir/StructureDefinition/"

// Profile URL constants.
const (
	ProfilePatient           = ProfileBaseURL + "OpenNucleus-Patient"
	ProfileImmunization      = ProfileBaseURL + "OpenNucleus-Immunization"
	ProfileGrowthObservation = ProfileBaseURL + "OpenNucleus-GrowthObservation"
	ProfileDetectedIssue     = ProfileBaseURL + "OpenNucleus-DetectedIssue"
	ProfileMeasureReport     = ProfileBaseURL + "OpenNucleus-MeasureReport"
)

// ProfileDef describes a FHIR profile supported by Open Nucleus.
type ProfileDef struct {
	URL          string         // full canonical URL
	Name         string         // "OpenNucleus-Patient"
	Title        string         // "Open Nucleus Patient Profile"
	BaseResource string         // "Patient"
	Version      string         // "0.1.0"
	Description  string
	Extensions   []ExtensionDef
	ValidateFunc func(resource map[string]any) []FieldError // profile-specific constraints
}

// profileRegistry holds all registered profile definitions keyed by profile URL.
var profileRegistry = map[string]*ProfileDef{}

func init() {
	registerAllProfiles()
}

func registerProfile(def *ProfileDef) {
	profileRegistry[def.URL] = def
}

// GetProfileDef returns the definition for a profile URL, or nil if unknown.
func GetProfileDef(profileURL string) *ProfileDef {
	return profileRegistry[profileURL]
}

// AllProfileDefs returns all registered profile definitions.
func AllProfileDefs() []*ProfileDef {
	defs := make([]*ProfileDef, 0, len(profileRegistry))
	for _, def := range profileRegistry {
		defs = append(defs, def)
	}
	return defs
}

// ProfilesForResource returns all profiles targeting a specific resource type.
func ProfilesForResource(resourceType string) []*ProfileDef {
	var defs []*ProfileDef
	for _, def := range profileRegistry {
		if def.BaseResource == resourceType {
			defs = append(defs, def)
		}
	}
	return defs
}

// GetMetaProfiles extracts the meta.profile array from a FHIR resource.
func GetMetaProfiles(resource map[string]any) []string {
	meta, ok := resource["meta"].(map[string]any)
	if !ok {
		return nil
	}
	profiles, ok := getArray(meta, "profile")
	if !ok {
		return nil
	}
	var result []string
	for _, p := range profiles {
		if s, ok := p.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
