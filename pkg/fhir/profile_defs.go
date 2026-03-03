package fhir

const extBaseURL = "http://opennucleus.dev/fhir/StructureDefinition/"

func registerAllProfiles() {
	registerProfile(buildPatientProfile())
	registerProfile(buildImmunizationProfile())
	registerProfile(buildGrowthObservationProfile())
	registerProfile(buildDetectedIssueProfile())
	registerProfile(buildMeasureReportProfile())
}

func buildPatientProfile() *ProfileDef {
	return &ProfileDef{
		URL:          ProfilePatient,
		Name:         "OpenNucleus-Patient",
		Title:        "Open Nucleus Patient Profile",
		BaseResource: ResourcePatient,
		Version:      "0.1.0",
		Description:  "Patient profile for African healthcare deployment with national ID and ethnic group extensions.",
		Extensions: []ExtensionDef{
			{
				URL:       extBaseURL + "national-health-id",
				ValueType: "valueIdentifier",
				Short:     "National health identifier (Nigeria NIN, Kenya NHIF, South Africa ID)",
			},
			{
				URL:       extBaseURL + "ethnic-group",
				ValueType: "valueCoding",
				Short:     "Ethnic group classification",
			},
		},
	}
}

func buildImmunizationProfile() *ProfileDef {
	return &ProfileDef{
		URL:          ProfileImmunization,
		Name:         "OpenNucleus-Immunization",
		Title:        "Open Nucleus Immunization Profile",
		BaseResource: ResourceImmunization,
		Version:      "0.1.0",
		Description:  "Immunization profile with WHO vaccine schedule extensions and CVX/ATC code guidance.",
		Extensions: []ExtensionDef{
			{
				URL:       extBaseURL + "dose-schedule-name",
				ValueType: "valueString",
				Short:     "Vaccine schedule name (e.g. WHO EPI Schedule)",
			},
			{
				URL:       extBaseURL + "dose-expected-age",
				ValueType: "valueString",
				Short:     "Expected age for this dose (e.g. 6 weeks)",
			},
		},
		ValidateFunc: validateImmunizationProfile,
	}
}

func validateImmunizationProfile(r map[string]any) []FieldError {
	// vaccineCode.coding should include CVX or WHO ATC system (warning only)
	vc, ok := r["vaccineCode"].(map[string]any)
	if !ok {
		return nil
	}
	codings, ok := getArray(vc, "coding")
	if !ok {
		return nil
	}
	for _, c := range codings {
		cMap, ok := c.(map[string]any)
		if !ok {
			continue
		}
		sys := getStr(cMap, "system")
		if sys == "http://hl7.org/fhir/sid/cvx" || sys == "urn:oid:1.3.6.1.4.1.58785" {
			return nil // found a recognized system
		}
	}
	return []FieldError{{
		Path:    "vaccineCode.coding",
		Rule:    "profile_warning",
		Message: "vaccineCode should include a coding with system http://hl7.org/fhir/sid/cvx or urn:oid:1.3.6.1.4.1.58785 (WHO ATC)",
	}}
}

func buildGrowthObservationProfile() *ProfileDef {
	return &ProfileDef{
		URL:          ProfileGrowthObservation,
		Name:         "OpenNucleus-GrowthObservation",
		Title:        "Open Nucleus Growth Observation Profile",
		BaseResource: ResourceObservation,
		Version:      "0.1.0",
		Description:  "Observation profile for paediatric growth monitoring with WHO z-score and nutritional classification.",
		Extensions: []ExtensionDef{
			{
				URL:       extBaseURL + "who-zscore",
				ValueType: "valueDecimal",
				Short:     "WHO z-score value",
			},
			{
				URL:       extBaseURL + "nutritional-classification",
				ValueType: "valueCoding",
				Short:     "Nutritional classification (severely-underweight, underweight, normal, overweight, obese)",
			},
		},
		ValidateFunc: validateGrowthObservationProfile,
	}
}

var validGrowthCodes = map[string]bool{
	"29463-7": true, // weight-for-age
	"8302-2":  true, // height-for-age
	"77606-2": true, // weight-for-height
	"59576-9": true, // BMI-for-age
	"9843-4":  true, // head-circumference-for-age
}

func validateGrowthObservationProfile(r map[string]any) []FieldError {
	var errs []FieldError

	// code.coding[].code must be one of the growth codes
	code, ok := r["code"].(map[string]any)
	if ok {
		codings, ok := getArray(code, "coding")
		if ok {
			found := false
			for _, c := range codings {
				cMap, ok := c.(map[string]any)
				if !ok {
					continue
				}
				if validGrowthCodes[getStr(cMap, "code")] {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, FieldError{
					Path:    "code.coding",
					Rule:    "profile_constraint",
					Message: "Growth observation code must be one of: 29463-7, 8302-2, 77606-2, 59576-9, 9843-4",
				})
			}
		}
	}

	// category must include vital-signs
	cats, ok := getArray(r, "category")
	if ok {
		found := false
		for _, cat := range cats {
			catMap, ok := cat.(map[string]any)
			if !ok {
				continue
			}
			catCodings, ok := getArray(catMap, "coding")
			if !ok {
				continue
			}
			for _, cc := range catCodings {
				ccMap, ok := cc.(map[string]any)
				if !ok {
					continue
				}
				if getStr(ccMap, "code") == "vital-signs" {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			errs = append(errs, FieldError{
				Path:    "category",
				Rule:    "profile_constraint",
				Message: "Growth observation must include category with code vital-signs",
			})
		}
	} else {
		errs = append(errs, FieldError{
			Path:    "category",
			Rule:    "profile_constraint",
			Message: "Growth observation must include category with code vital-signs",
		})
	}

	return errs
}

func buildDetectedIssueProfile() *ProfileDef {
	return &ProfileDef{
		URL:          ProfileDetectedIssue,
		Name:         "OpenNucleus-DetectedIssue",
		Title:        "Open Nucleus Detected Issue Profile",
		BaseResource: ResourceDetectedIssue,
		Version:      "0.1.0",
		Description:  "DetectedIssue profile for AI-generated alerts with model provenance extensions.",
		Extensions: []ExtensionDef{
			{
				URL:       extBaseURL + "ai-model-name",
				ValueType: "valueString",
				Short:     "AI model name that generated this issue",
			},
			{
				URL:       extBaseURL + "ai-confidence-score",
				ValueType: "valueDecimal",
				Short:     "AI confidence score (0.0-1.0)",
			},
			{
				URL:       extBaseURL + "ai-reflection-count",
				ValueType: "valueInteger",
				Short:     "Number of AI reflection/reasoning steps",
			},
			{
				URL:       extBaseURL + "ai-reasoning-chain",
				ValueType: "valueString",
				Short:     "AI reasoning chain description",
			},
		},
	}
}

func buildMeasureReportProfile() *ProfileDef {
	return &ProfileDef{
		URL:          ProfileMeasureReport,
		Name:         "OpenNucleus-MeasureReport",
		Title:        "Open Nucleus MeasureReport Profile",
		BaseResource: ResourceMeasureReport,
		Version:      "0.1.0",
		Description:  "MeasureReport profile for DHIS2 reporting integration extensions.",
		Extensions: []ExtensionDef{
			{
				URL:       extBaseURL + "dhis2-data-element",
				ValueType: "valueString",
				Short:     "DHIS2 data element ID",
			},
			{
				URL:       extBaseURL + "dhis2-org-unit",
				ValueType: "valueString",
				Short:     "DHIS2 organisation unit ID",
			},
			{
				URL:       extBaseURL + "dhis2-period",
				ValueType: "valueString",
				Short:     "DHIS2 reporting period (e.g. 202603)",
			},
		},
	}
}
