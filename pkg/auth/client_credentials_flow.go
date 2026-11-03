package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

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

// clientCredentialsKeyringTokenSource keeps the client-credentials token in
// memory only. The grant has no refresh token and a new token can be fetched
// at any time with the client secret, so persisting it would only widen the
// exposure of a bearer credential.
type clientCredentialsKeyringTokenSource struct {
	ctx                    context.Context
	tokenKey               string // namespaced key, only used to clean up entries persisted by older versions
	store                  TokenStore
	conf                   *clientcredentials.Config
	allowOrganizationScope bool
	initErr                error // configuration error surfaced by Token()

	mu           sync.Mutex // guards token and serializes fetches
	token        *oauth2.Token
	cachedClaims atomic.Pointer[TokenClaims]
}

// ClearTokenStore drops the in-memory token and, best effort, removes a token
// entry persisted under this flow's namespaced key by older versions.
func (s *clientCredentialsKeyringTokenSource) ClearTokenStore() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = nil
	s.cachedClaims.Store(nil)

	if s.store == nil {
		return nil
	}
	if err := s.store.Delete(s.tokenKey); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return nil
}

func (s *clientCredentialsKeyringTokenSource) fetchNewToken() (*oauth2.Token, error) {
	return s.conf.TokenSource(s.ctx).Token()
}

func (s *clientCredentialsKeyringTokenSource) Token() (*oauth2.Token, error) {
	if s.initErr != nil {
		return nil, s.initErr
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-grant when there is no cached token, the token is expired, or it is
	// within the proactive refresh buffer (shared with the device flow).
	token := s.token
	if token == nil || !token.Valid() || nearExpiry(token) {
		var err error
		token, err = s.fetchNewToken()
		if err != nil {
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

	s.token = token
	s.cachedClaims.Store(claims)

	return token, nil
}

func (s *clientCredentialsKeyringTokenSource) TokenClaims() (*TokenClaims, error) {
	if c := s.cachedClaims.Load(); c != nil {
		cp := *c
		return &cp, nil
	}
	if _, err := s.Token(); err != nil {
		return nil, err
	}
	c := s.cachedClaims.Load()
	if c == nil {
		return nil, errors.New("token claims unavailable")
	}
	cp := *c
	return &cp, nil
}

// NewClientCredentialsKeyringTokenSource returns a TokenSourceWithClaims that fetches tokens via the
// client credentials grant and caches them in memory only. EOS claims (realm, org) are extracted and
// cached inside Token().
//
// Deprecated: the store parameter is unused for new tokens. Client credentials
// tokens are no longer persisted to the keyring or token file (there is no
// refresh token, and a new token can always be requested). The store, which
// may be nil, is only used by ClearTokenStore to remove entries written by
// older versions. The constructor name is kept for compatibility.
//
// The auth base URL must be https (http only for loopback hosts); an invalid
// URL is reported as an error from Token() and TokenClaims(). The context is
// kept for the lifetime of the source but detached from cancellation (its
// values, such as oauth2.HTTPClient, are preserved).
func NewClientCredentialsKeyringTokenSource(ctx context.Context, c *ClientCredentialsConfig, store TokenStore) TokenSourceWithClaims {
	if ctx == nil {
		ctx = context.Background()
	}

	return &clientCredentialsKeyringTokenSource{
		ctx:                    context.WithoutCancel(ctx),
		tokenKey:               namespacedTokenKey(DefaultClientCredentialsFlowTokenKey, c.authBaseURL, c.clientID),
		store:                  store,
		conf:                   c.GetClientCredentialsConfig(),
		allowOrganizationScope: c.allowOrganizationScope,
		initErr:                validateAuthBaseURL(c.authBaseURL),
	}
}
