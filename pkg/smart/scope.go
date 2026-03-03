package smart

import (
	"fmt"
	"sort"
	"strings"

	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// Scope represents a parsed SMART v2 scope.
// Format: context/ResourceType.interactions (e.g. "patient/Observation.rs")
type Scope struct {
	Context      string // "patient", "user", "system"
	Resource     string // "Patient", "Observation", "*"
	Interactions string // "cruds", "r", "rs", "*"
	Raw          string // original string
}

// Special (non-resource) scopes that are stored as-is.
var specialScopes = map[string]bool{
	"launch":           true,
	"launch/patient":   true,
	"launch/encounter": true,
	"fhirUser":         true,
	"offline_access":   true,
	"openid":           true,
}

// validContexts are the allowed scope context prefixes.
var validContexts = map[string]bool{
	"patient": true,
	"user":    true,
	"system":  true,
}

// validInteractions are the characters allowed in the interactions segment.
const validInteractionChars = "cruds"

// IsSpecialScope returns true if s is a non-resource SMART scope.
func IsSpecialScope(s string) bool {
	return specialScopes[s]
}

// ParseScope parses a single SMART v2 scope string.
func ParseScope(s string) (Scope, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Scope{}, fmt.Errorf("empty scope")
	}

	// Special scopes are returned as-is with no resource/interaction.
	if specialScopes[s] {
		return Scope{Raw: s}, nil
	}

	// Format: context/Resource.interactions
	slashIdx := strings.Index(s, "/")
	if slashIdx < 0 {
		return Scope{}, fmt.Errorf("invalid scope %q: missing '/' separator", s)
	}

	ctx := s[:slashIdx]
	rest := s[slashIdx+1:]

	if !validContexts[ctx] {
		return Scope{}, fmt.Errorf("invalid scope context %q in %q: must be patient, user, or system", ctx, s)
	}

	dotIdx := strings.Index(rest, ".")
	if dotIdx < 0 {
		return Scope{}, fmt.Errorf("invalid scope %q: missing '.' separator between resource and interactions", s)
	}

	resource := rest[:dotIdx]
	interactions := rest[dotIdx+1:]

	if resource == "" {
		return Scope{}, fmt.Errorf("invalid scope %q: empty resource type", s)
	}
	if interactions == "" {
		return Scope{}, fmt.Errorf("invalid scope %q: empty interactions", s)
	}

	// Validate resource type: must be known or wildcard.
	if resource != "*" && !fhir.IsKnownResource(resource) {
		return Scope{}, fmt.Errorf("invalid scope %q: unknown resource type %q", s, resource)
	}

	// Validate interactions: must be subset of "cruds" or wildcard.
	if interactions != "*" {
		for _, ch := range interactions {
			if !strings.ContainsRune(validInteractionChars, ch) {
				return Scope{}, fmt.Errorf("invalid scope %q: unknown interaction %q (must be subset of 'cruds')", s, string(ch))
			}
		}
	}

	return Scope{
		Context:      ctx,
		Resource:     resource,
		Interactions: interactions,
		Raw:          s,
	}, nil
}

// ParseScopes parses a space-delimited scope string into individual scopes.
func ParseScopes(spaceDelimited string) ([]Scope, error) {
	spaceDelimited = strings.TrimSpace(spaceDelimited)
	if spaceDelimited == "" {
		return nil, nil
	}

	parts := strings.Fields(spaceDelimited)
	scopes := make([]Scope, 0, len(parts))
	for _, p := range parts {
		sc, err := ParseScope(p)
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, sc)
	}
	return scopes, nil
}

// String returns the canonical scope string.
func (s Scope) String() string {
	if s.Raw != "" {
		return s.Raw
	}
	if s.Context == "" {
		return ""
	}
	return s.Context + "/" + s.Resource + "." + s.Interactions
}

// IsSpecial returns true if this is a non-resource scope.
func (s Scope) IsSpecial() bool {
	return s.Context == "" && s.Raw != ""
}

// Allows checks if this scope permits a given interaction on a resource type.
// interaction is one of "c", "r", "u", "d", "s" (create, read, update, delete, search).
func (s Scope) Allows(interaction, resourceType string) bool {
	if s.IsSpecial() {
		return false
	}

	// Check resource match: wildcard or exact.
	if s.Resource != "*" && s.Resource != resourceType {
		return false
	}

	// Check interaction match: wildcard or contained.
	if s.Interactions == "*" {
		return true
	}
	return strings.Contains(s.Interactions, interaction)
}

// FilterByResource returns scopes that apply to the given resource type.
func FilterByResource(scopes []Scope, resourceType string) []Scope {
	var result []Scope
	for _, sc := range scopes {
		if sc.IsSpecial() {
			continue
		}
		if sc.Resource == "*" || sc.Resource == resourceType {
			result = append(result, sc)
		}
	}
	return result
}

// interactionToPermSuffix maps SMART interaction letters to RBAC permission suffixes.
var interactionToPermSuffix = map[byte]string{
	'c': "write",
	'r': "read",
	'u': "write",
	'd': "write",
	's': "read",
}

// ScopesToPermissions converts SMART scopes into RBAC permission strings.
// For example, "patient/Observation.rs" → ["observation:read"].
func ScopesToPermissions(scopes []Scope) []string {
	seen := map[string]bool{}
	for _, sc := range scopes {
		if sc.IsSpecial() {
			continue
		}

		resources := []string{sc.Resource}
		if sc.Resource == "*" {
			for _, def := range fhir.AllResourceDefs() {
				resources = append(resources, def.Type)
			}
		}

		interactions := sc.Interactions
		if interactions == "*" {
			interactions = validInteractionChars
		}

		for _, res := range resources {
			prefix := strings.ToLower(res)
			for i := 0; i < len(interactions); i++ {
				suffix, ok := interactionToPermSuffix[interactions[i]]
				if !ok {
					continue
				}
				perm := prefix + ":" + suffix
				seen[perm] = true
			}
		}
	}

	perms := make([]string, 0, len(seen))
	for p := range seen {
		perms = append(perms, p)
	}
	sort.Strings(perms)
	return perms
}

// AllSupportedScopes returns all SMART scopes supported by this server,
// built dynamically from the FHIR resource registry.
func AllSupportedScopes() []string {
	defs := fhir.AllResourceDefs()
	contexts := []string{"patient", "user"}
	interactionSets := []string{"cruds", "r", "rs"}

	scopes := make([]string, 0, len(defs)*len(contexts)*len(interactionSets)+len(specialScopes))

	// Resource scopes.
	for _, def := range defs {
		for _, ctx := range contexts {
			for _, inter := range interactionSets {
				scopes = append(scopes, ctx+"/"+def.Type+"."+inter)
			}
		}
	}

	// Special scopes.
	for s := range specialScopes {
		scopes = append(scopes, s)
	}

	sort.Strings(scopes)
	return scopes
}
