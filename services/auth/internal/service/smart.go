package service

import (
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"

	"github.com/FibrinLab/open-nucleus/pkg/auth"
	"github.com/FibrinLab/open-nucleus/pkg/smart"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/store"
)

// SmartService provides SMART on FHIR OAuth2 operations.
type SmartService struct {
	authSvc     *AuthService
	clientStore *store.ClientStore
	authCodes   *smart.AuthCodeStore
	launches    *smart.LaunchStore
}

// NewSmartService creates a new SmartService.
func NewSmartService(authSvc *AuthService, clientStore *store.ClientStore) *SmartService {
	return &SmartService{
		authSvc:     authSvc,
		clientStore: clientStore,
		authCodes:   smart.NewAuthCodeStore(60 * time.Second),
		launches:    smart.NewLaunchStore(5 * time.Minute),
	}
}

// Authorize validates the authorization request and returns a redirect URI with auth code.
func (s *SmartService) Authorize(
	clientID, redirectURI, scope, state, codeChallenge, codeChallengeMethod, launchToken string,
	deviceID, practitionerID, siteID, role string,
) (string, error) {
	// Validate client.
	client, err := s.clientStore.Get(clientID)
	if err != nil {
		return "", fmt.Errorf("invalid client_id: %w", err)
	}
	if client.Status != smart.ClientApproved {
		return "", fmt.Errorf("client is not approved (status: %s)", client.Status)
	}

	// Validate redirect URI matches registered URIs.
	validRedirect := false
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		return "", fmt.Errorf("redirect_uri not registered for this client")
	}

	// PKCE required for public clients.
	if client.TokenEndpointAuthMethod == "none" {
		if codeChallenge == "" {
			return "", fmt.Errorf("code_challenge required for public clients (PKCE)")
		}
		if codeChallengeMethod != "S256" {
			return "", fmt.Errorf("only S256 code_challenge_method is supported")
		}
	}

	// Validate requested scope against client's max scope.
	if err := validateScopeSubset(scope, client.Scope); err != nil {
		return "", fmt.Errorf("scope exceeds client registration: %w", err)
	}

	// Handle launch token if present.
	var patientID, encounterID string
	if launchToken != "" {
		lt, err := s.launches.Consume(launchToken)
		if err != nil {
			return "", fmt.Errorf("invalid launch token: %w", err)
		}
		if lt.ClientID != clientID {
			return "", fmt.Errorf("launch token was issued for a different client")
		}
		patientID = lt.PatientID
		encounterID = lt.EncounterID
	}

	// Generate auth code.
	code, err := s.authCodes.Generate(smart.AuthCodeParams{
		ClientID:       clientID,
		RedirectURI:    redirectURI,
		Scope:          scope,
		CodeChallenge:  codeChallenge,
		DeviceID:       deviceID,
		PractitionerID: practitionerID,
		SiteID:         siteID,
		Role:           role,
		PatientID:      patientID,
		EncounterID:    encounterID,
	})
	if err != nil {
		return "", fmt.Errorf("generate auth code: %w", err)
	}

	// Build redirect URI with code + state.
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("parse redirect_uri: %w", err)
	}
	q := u.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// ExchangeToken exchanges an authorization code for an access token.
func (s *SmartService) ExchangeToken(
	grantType, code, redirectURI, codeVerifier, clientID, clientSecret string,
) (accessToken string, expiresIn int32, scope, patient, encounter string, err error) {
	if grantType != "authorization_code" {
		err = fmt.Errorf("unsupported grant_type: %s", grantType)
		return
	}

	// Validate client.
	client, clientErr := s.clientStore.Get(clientID)
	if clientErr != nil {
		err = fmt.Errorf("invalid client_id")
		return
	}

	// Validate client authentication.
	if client.TokenEndpointAuthMethod == "client_secret_basic" {
		if clientSecret != client.ClientSecret {
			err = fmt.Errorf("invalid client_secret")
			return
		}
	}

	// Exchange auth code.
	ac, exchangeErr := s.authCodes.Exchange(code, clientID, codeVerifier, redirectURI)
	if exchangeErr != nil {
		err = fmt.Errorf("token exchange: %w", exchangeErr)
		return
	}

	// Look up role for permissions.
	roleDef, ok := auth.GetRole(ac.Role)
	if !ok {
		err = fmt.Errorf("unknown role: %s", ac.Role)
		return
	}

	// Build SMART scopes → permissions mapping.
	scopes, _ := smart.ParseScopes(ac.Scope)
	smartPerms := smart.ScopesToPermissions(scopes)

	// Merge SMART permissions with role permissions (intersection-like: SMART perms that the role also has).
	perms := intersectPermissions(smartPerms, roleDef.Permissions)

	// Issue JWT with SMART claims.
	jti := uuid.New().String()
	lifetime := s.authSvc.cfg.JWT.AccessLifetime

	fhirUser := ""
	if ac.PractitionerID != "" {
		fhirUser = "Practitioner/" + ac.PractitionerID
	}

	claims := auth.NewSmartAccessClaims(
		ac.PractitionerID, ac.DeviceID, s.authSvc.nodeID, ac.SiteID, ac.Role,
		perms, roleDef.SiteScope,
		ac.Scope, clientID, fhirUser, ac.PatientID, ac.EncounterID,
		jti, s.authSvc.cfg.JWT.Issuer, lifetime,
	)

	accessToken, signErr := auth.SignToken(claims, s.authSvc.nodePrivateKey, s.authSvc.nodeID)
	if signErr != nil {
		err = fmt.Errorf("sign token: %w", signErr)
		return
	}

	expiresIn = int32(lifetime.Seconds())
	scope = ac.Scope
	patient = ac.PatientID
	encounter = ac.EncounterID
	return
}

// IntrospectToken checks if a token is active and returns its metadata.
func (s *SmartService) IntrospectToken(tokenStr string) (active bool, scope, clientID, sub, patient, encounter, fhirUser string, exp, iat int64, err error) {
	claims, errCode, verifyErr := s.authSvc.ValidateToken(tokenStr)
	if verifyErr != nil || errCode != "" {
		// Token is not active.
		return
	}

	active = true
	scope = claims.Scope
	clientID = claims.ClientID
	sub = claims.Subject
	patient = claims.LaunchPatient
	encounter = claims.LaunchEncounter
	fhirUser = claims.FHIRUser

	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Unix()
	}
	if claims.IssuedAt != nil {
		iat = claims.IssuedAt.Unix()
	}
	return
}

// RevokeToken revokes a token by adding its JTI to the deny list.
func (s *SmartService) RevokeToken(tokenStr string) error {
	return s.authSvc.Logout(tokenStr)
}

// RegisterClient registers a new SMART client application.
func (s *SmartService) RegisterClient(name string, redirectURIs []string, scope string, grantTypes []string, authMethod string, launchModes []string, registeredBy string) (*smart.Client, error) {
	client := &smart.Client{
		ClientID:                smart.GenerateClientID(),
		ClientName:              name,
		RedirectURIs:            redirectURIs,
		Scope:                   scope,
		GrantTypes:              grantTypes,
		TokenEndpointAuthMethod: authMethod,
		LaunchModes:             launchModes,
		Status:                  smart.ClientPending,
		RegisteredAt:            time.Now().UTC().Format(time.RFC3339),
		RegisteredBy:            registeredBy,
	}

	// Generate secret for confidential clients.
	if authMethod == "client_secret_basic" {
		client.ClientSecret = smart.GenerateClientSecret()
	}

	if err := smart.ValidateClient(client); err != nil {
		return nil, fmt.Errorf("invalid client: %w", err)
	}

	s.authSvc.mu.Lock()
	defer s.authSvc.mu.Unlock()

	if _, err := s.clientStore.Save(client); err != nil {
		return nil, err
	}

	return client, nil
}

// ListClients returns all registered SMART clients.
func (s *SmartService) ListClients() ([]*smart.Client, error) {
	return s.clientStore.List()
}

// GetClient returns a single SMART client.
func (s *SmartService) GetClient(clientID string) (*smart.Client, error) {
	return s.clientStore.Get(clientID)
}

// UpdateClient updates a SMART client's status or scope.
func (s *SmartService) UpdateClient(clientID, status, approvedBy, scope string) (*smart.Client, error) {
	s.authSvc.mu.Lock()
	defer s.authSvc.mu.Unlock()
	return s.clientStore.Update(clientID, status, approvedBy, scope)
}

// DeleteClient removes a SMART client registration.
func (s *SmartService) DeleteClient(clientID string) error {
	s.authSvc.mu.Lock()
	defer s.authSvc.mu.Unlock()
	return s.clientStore.Delete(clientID)
}

// CreateLaunch generates an EHR launch token.
func (s *SmartService) CreateLaunch(clientID, patientID, encounterID, createdBy string) (string, error) {
	// Validate client exists.
	if _, err := s.clientStore.Get(clientID); err != nil {
		return "", fmt.Errorf("invalid client_id: %w", err)
	}
	return s.launches.Create(clientID, patientID, encounterID, createdBy)
}

// validateScopeSubset checks that all requested scopes are within the client's max scope.
func validateScopeSubset(requested, maxAllowed string) error {
	reqScopes, err := smart.ParseScopes(requested)
	if err != nil {
		return err
	}
	maxScopes, err := smart.ParseScopes(maxAllowed)
	if err != nil {
		return err
	}

	for _, req := range reqScopes {
		if req.IsSpecial() {
			// Special scopes: check exact match in max.
			found := false
			for _, m := range maxScopes {
				if m.Raw == req.Raw {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("scope %q not in client registration", req.Raw)
			}
			continue
		}

		// Resource scopes: check that a matching max scope covers it.
		found := false
		for _, m := range maxScopes {
			if m.IsSpecial() {
				continue
			}
			if (m.Resource == "*" || m.Resource == req.Resource) &&
				(m.Interactions == "*" || containsAllChars(m.Interactions, req.Interactions)) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("scope %q exceeds client registration", req.String())
		}
	}
	return nil
}

func containsAllChars(haystack, needles string) bool {
	for _, ch := range needles {
		found := false
		for _, h := range haystack {
			if h == ch {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func intersectPermissions(smartPerms, rolePerms []string) []string {
	roleSet := make(map[string]bool, len(rolePerms))
	for _, p := range rolePerms {
		roleSet[p] = true
	}
	var result []string
	for _, p := range smartPerms {
		if roleSet[p] {
			result = append(result, p)
		}
	}
	return result
}
