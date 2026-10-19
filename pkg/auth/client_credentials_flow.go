package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// ClientCredentialsConfig holds the parameters needed to run the OAuth2 client credentials flow.
type ClientCredentialsConfig struct {
	SharedFlowConfig
	clientSecret string
}

// ClientCredentialsConfigOption is a functional option for ClientCredentialsConfig.
type ClientCredentialsConfigOption func(*ClientCredentialsConfig)

// WithClientCredentialsAuthBaseURL sets a custom auth base URL.
func WithClientCredentialsAuthBaseURL(authBaseURL string) ClientCredentialsConfigOption {
	return func(c *ClientCredentialsConfig) {
		c.authBaseURL = authBaseURL
	}
}

// WithClientCredentialsKeyringServiceName sets a custom keyring service name.
func WithClientCredentialsKeyringServiceName(keyringServiceName string) ClientCredentialsConfigOption {
	return func(c *ClientCredentialsConfig) {
		c.keyringServiceName = keyringServiceName
	}
}

// WithClientCredentialsFallbackTokenStoreDir overrides the directory used for file-based token storage.
func WithClientCredentialsFallbackTokenStoreDir(dir string) ClientCredentialsConfigOption {
	return func(c *ClientCredentialsConfig) {
		c.fallbackTokenStoreDir = dir
	}
}

// WithAllowOrganizationScope allows usage of client credentials without realm_id (=organization scoped).
func WithAllowOrganizationScope() ClientCredentialsConfigOption {
	return func(c *ClientCredentialsConfig) {
		c.allowOrganizationScope = true
	}
}

// With configures the ClientCredentialsConfig with additional options.
func (c *ClientCredentialsConfig) With(opts ...ClientCredentialsConfigOption) {
	for _, opt := range opts {
		opt(c)
	}
}

func (c *ClientCredentialsConfig) GetClientCredentialsConfig() *clientcredentials.Config {
	return &clientcredentials.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", c.authBaseURL),
	}
}

// NewClientCredentialsConfig creates a ClientCredentialsConfig with the given credentials
// and defaults the auth base URL to DefaultAuthBaseURL.
func NewClientCredentialsConfig(clientID, clientSecret string) *ClientCredentialsConfig {
	cfg := defaultFlowConfig()
	cfg.clientID = clientID
	return &ClientCredentialsConfig{SharedFlowConfig: cfg, clientSecret: clientSecret}
}

type clientCredentialsKeyringTokenSource struct {
	ctx                    context.Context
	tokenKey               string // namespaced primary key
	legacyKey              string // legacy fixed key (read-only fallback)
	authBaseURL            string // used to validate legacy token issuer before adopting
	store                  TokenStore
	conf                   *clientcredentials.Config
	cachedClaims           *TokenClaims
	allowOrganizationScope bool
}

func (s *clientCredentialsKeyringTokenSource) ClearTokenStore() error {
	return s.store.DeleteAll()
}

func (s *clientCredentialsKeyringTokenSource) loadFromKeyring() (*oauth2.Token, error) {
	// Try the primary (namespaced) key first.
	stored, err := s.store.Get(s.tokenKey)
	if err == nil && stored != "" {
		var token oauth2.Token
		if err := json.Unmarshal([]byte(stored), &token); err != nil {
			_ = s.store.Delete(s.tokenKey)
			return nil, err
		}
		return &token, nil
	}

	// Primary key miss — attempt read-only fallback to the legacy fixed key.
	return s.loadLegacyToken()
}

// loadLegacyToken reads the legacy fixed-key entry and adopts it only when its
// issuer matches the current authBaseURL. On adoption the token is migrated to
// the namespaced key and the legacy entry is deleted, so it is evaluated at
// most once. A legacy entry that is absent, unparseable, or issued by a
// different auth server is left untouched.
func (s *clientCredentialsKeyringTokenSource) loadLegacyToken() (*oauth2.Token, error) {
	legacyStored, err := s.store.Get(s.legacyKey)
	if err != nil || legacyStored == "" {
		return nil, nil //nolint:nilnil // nil token signals "not found", distinct from an error
	}

	var legacyToken oauth2.Token
	if err := json.Unmarshal([]byte(legacyStored), &legacyToken); err != nil {
		// Unparseable legacy entry — skip adoption without deleting it.
		return nil, nil //nolint:nilnil
	}

	issuer, err := tokenIssuer(&legacyToken)
	if err != nil || !issuerMatchesAuthBaseURL(issuer, s.authBaseURL) {
		// Issuer absent, unparseable, or belongs to a different auth server.
		return nil, nil //nolint:nilnil
	}

	// Adopt and migrate: persist the token under the namespaced key and delete
	// the legacy entry so it is not re-evaluated on subsequent runs. If the
	// namespaced write fails we keep the legacy entry as a fallback and delete
	// nothing. Once the legacy entry is removed, downgrading to an older tool
	// version requires a one-time re-authentication, which is acceptable.
	if err := s.store.Set(s.tokenKey, legacyStored); err == nil {
		_ = s.store.Delete(s.legacyKey)
	}
	return &legacyToken, nil
}

func (s *clientCredentialsKeyringTokenSource) saveToKeyring(token *oauth2.Token) error {
	data, err := json.Marshal(token) //nolint:gosec // the token is stored in the keyring or token store on purpose
	if err != nil {
		_ = s.store.Delete(s.tokenKey)
		return err
	}

	return s.store.Set(s.tokenKey, string(data))
}

func (s *clientCredentialsKeyringTokenSource) fetchNewToken() (*oauth2.Token, error) {
	return s.conf.TokenSource(s.ctx).Token()
}

func (s *clientCredentialsKeyringTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.loadFromKeyring()
	if err != nil {
		return nil, err
	}

	// Re-grant when there is no cached token, the token is expired, or it is
	// within the proactive refresh buffer (shared with the device flow).
	if token == nil || !token.Valid() || nearExpiry(token) {
		token, err = s.fetchNewToken()
		if err != nil {
			return nil, err
		}
		if err := s.saveToKeyring(token); err != nil {
			return nil, err
		}
	}

	claims, err := ExtractClaims(token)
	if err != nil {
		return nil, fmt.Errorf("token missing required eos claims: %w", err)
	}
	if !s.allowOrganizationScope && claims.RealmID == "" {
		return nil, fmt.Errorf("JWT token missing eos.realm claim")
	}

	s.cachedClaims = claims

	return token, nil
}

func (s *clientCredentialsKeyringTokenSource) TokenClaims() (*TokenClaims, error) {
	if s.cachedClaims != nil {
		return s.cachedClaims, nil
	}
	if _, err := s.Token(); err != nil {
		return nil, err
	}
	return s.cachedClaims, nil
}

// NewClientCredentialsKeyringTokenSource returns a TokenSourceWithClaims that persists tokens using
// the given TokenStore. On the first call or after token expiry, it fetches a new token via client
// credentials grant. EOS claims (realm, org) are extracted and cached inside Token().
func NewClientCredentialsKeyringTokenSource(ctx context.Context, c *ClientCredentialsConfig, store TokenStore) TokenSourceWithClaims {
	return &clientCredentialsKeyringTokenSource{
		ctx:                    ctx,
		tokenKey:               namespacedTokenKey(DefaultClientCredentialsFlowTokenKey, c.authBaseURL, c.clientID),
		legacyKey:              DefaultClientCredentialsFlowTokenKey,
		authBaseURL:            c.authBaseURL,
		store:                  store,
		conf:                   c.GetClientCredentialsConfig(),
		allowOrganizationScope: c.allowOrganizationScope,
	}
}
