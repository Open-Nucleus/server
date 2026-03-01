package store

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InteractionRule represents a drug-drug interaction or cross-reactivity rule.
type InteractionRule struct {
	MedicationA          string `json:"medication_a"`
	MedicationB          string `json:"medication_b"`
	Severity             string `json:"severity"`
	Type                 string `json:"type"`
	Description          string `json:"description"`
	ClinicalEffect       string `json:"clinical_effect"`
	Recommendation       string `json:"recommendation"`
	Source               string `json:"source"`
	ClassA               string `json:"class_a,omitempty"`
	ClassB               string `json:"class_b,omitempty"`
	CrossReactivityClass string `json:"cross_reactivity_class,omitempty"`
}

// AllergyRule represents a cross-reactivity allergy rule.
type AllergyRule struct {
	AllergyCode          string `json:"allergy_code"`
	MedicationClass      string `json:"medication_class"`
	Severity             string `json:"severity"`
	Description          string `json:"description"`
	CrossReactivityClass string `json:"cross_reactivity_class"`
}

// InteractionIndex provides O(1) lookup for drug-drug interactions.
type InteractionIndex struct {
	byPair   map[string]*InteractionRule // canonical pair key → rule
	byClass  map[string][]*InteractionRule // ATC prefix → rules
	allergies []AllergyRule
	all      []*InteractionRule
}

// NewInteractionIndex creates an empty index.
func NewInteractionIndex() *InteractionIndex {
	return &InteractionIndex{
		byPair:  make(map[string]*InteractionRule),
		byClass: make(map[string][]*InteractionRule),
	}
}

// LoadFromJSON loads interaction rules from JSON (array of InteractionRule).
func (idx *InteractionIndex) LoadFromJSON(data []byte) error {
	var parsed struct {
		Interactions []InteractionRule `json:"interactions"`
		Allergies    []AllergyRule     `json:"allergies"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("unmarshal interactions: %w", err)
	}

	for i := range parsed.Interactions {
		r := &parsed.Interactions[i]
		key := canonicalKey(r.MedicationA, r.MedicationB)
		idx.byPair[key] = r
		idx.all = append(idx.all, r)

		// Index by class prefix if specified
		if r.ClassA != "" {
			idx.byClass[r.ClassA] = append(idx.byClass[r.ClassA], r)
		}
		if r.ClassB != "" {
			idx.byClass[r.ClassB] = append(idx.byClass[r.ClassB], r)
		}
	}

	idx.allergies = parsed.Allergies
	return nil
}

// CheckPair checks for a direct interaction between two medication codes.
func (idx *InteractionIndex) CheckPair(codeA, codeB string) *InteractionRule {
	key := canonicalKey(codeA, codeB)
	return idx.byPair[key]
}

// CheckClass checks for class-level interactions using ATC prefix matching.
func (idx *InteractionIndex) CheckClass(code string) []*InteractionRule {
	var results []*InteractionRule
	// Try full code and progressively shorter prefixes
	for prefixLen := len(code); prefixLen >= 3; prefixLen-- {
		prefix := strings.ToUpper(code[:prefixLen])
		if rules, ok := idx.byClass[prefix]; ok {
			results = append(results, rules...)
		}
	}
	return results
}

// CheckAllergies checks if any medication codes conflict with allergy codes.
func (idx *InteractionIndex) CheckAllergies(medicationCodes, allergyCodes []string) []AllergyMatch {
	var matches []AllergyMatch

	for _, allergyCode := range allergyCodes {
		allergyLower := strings.ToLower(allergyCode)
		for _, medCode := range medicationCodes {
			// Check direct allergy rules
			for _, rule := range idx.allergies {
				if strings.ToLower(rule.AllergyCode) == allergyLower {
					// Check if medication belongs to the class
					if matchesMedicationClass(medCode, rule.MedicationClass) {
						matches = append(matches, AllergyMatch{
							AllergyCode:          allergyCode,
							MedicationCode:       medCode,
							Severity:             rule.Severity,
							Description:          rule.Description,
							CrossReactivityClass: rule.CrossReactivityClass,
						})
					}
				}
			}
		}
	}
	return matches
}

// Count returns the number of interaction rules.
func (idx *InteractionIndex) Count() int {
	return len(idx.all)
}

// AllergyMatch represents a matched allergy conflict.
type AllergyMatch struct {
	AllergyCode          string
	MedicationCode       string
	Severity             string
	Description          string
	CrossReactivityClass string
}

// canonicalKey creates a canonical key for a pair of medication codes.
// Always uses the lexicographically smaller code first.
func canonicalKey(a, b string) string {
	a = strings.ToUpper(a)
	b = strings.ToUpper(b)
	if a > b {
		a, b = b, a
	}
	return a + ":" + b
}

// matchesMedicationClass checks if a medication code matches a class pattern.
// Class patterns can be an ATC prefix (e.g., "J01C" matches "J01CA04")
// or a direct code match.
func matchesMedicationClass(medCode, classPattern string) bool {
	medUpper := strings.ToUpper(medCode)
	patternUpper := strings.ToUpper(classPattern)
	return strings.HasPrefix(medUpper, patternUpper) || medUpper == patternUpper
}
