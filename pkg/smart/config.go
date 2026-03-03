package smart

// SmartConfiguration represents the SMART on FHIR well-known configuration document.
// See: https://build.fhir.org/ig/HL7/smart-app-launch/conformance.html
type SmartConfiguration struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint,omitempty"`
	IntrospectionEndpoint             string   `json:"introspection_endpoint,omitempty"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	Capabilities                      []string `json:"capabilities"`
}

// GenerateSmartConfiguration builds the SMART well-known configuration for the given base URL.
func GenerateSmartConfiguration(baseURL string) *SmartConfiguration {
	return &SmartConfiguration{
		Issuer:                baseURL,
		AuthorizationEndpoint: baseURL + "/auth/smart/authorize",
		TokenEndpoint:         baseURL + "/auth/smart/token",
		RevocationEndpoint:    baseURL + "/auth/smart/revoke",
		IntrospectionEndpoint: baseURL + "/auth/smart/introspect",
		RegistrationEndpoint:  baseURL + "/auth/smart/register",
		ScopesSupported:       AllSupportedScopes(),
		ResponseTypesSupported: []string{"code"},
		GrantTypesSupported:    []string{"authorization_code"},
		TokenEndpointAuthMethodsSupported: []string{"none", "client_secret_basic"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		Capabilities: []string{
			"launch-ehr",
			"launch-standalone",
			"client-public",
			"client-confidential-symmetric",
			"permission-v2",
			"context-ehr-patient",
			"context-ehr-encounter",
			"context-standalone-patient",
		},
	}
}
