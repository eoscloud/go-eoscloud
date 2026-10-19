package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// tokenRefreshBuffer is the duration before actual expiry at which we proactively
// refresh an access token. This avoids races where a token expires between the
// validity check and the HTTP request.
const tokenRefreshBuffer = 60 * time.Second

// DeviceFlowConfig holds the parameters needed to run the OAuth2 device authorization flow.
type DeviceFlowConfig struct {
	SharedFlowConfig
}

// DeviceFlowConfigOption is a functional option for DeviceFlowConfig.
type DeviceFlowConfigOption func(*DeviceFlowConfig)

// WithDeviceFlowAuthBaseURL sets a custom auth base URL.
func WithDeviceFlowAuthBaseURL(authBaseURL string) DeviceFlowConfigOption {
	return func(d *DeviceFlowConfig) {
		d.authBaseURL = authBaseURL
	}
}

// WithDeviceFlowClientID sets a custom client ID.
func WithDeviceFlowClientID(clientID string) DeviceFlowConfigOption {
	return func(d *DeviceFlowConfig) {
		d.clientID = clientID
	}
}

// WithDeviceFlowKeyringServiceName sets a custom keyring service name.
func WithDeviceFlowKeyringServiceName(keyringServiceName string) DeviceFlowConfigOption {
	return func(d *DeviceFlowConfig) {
		d.keyringServiceName = keyringServiceName
	}
}

// WithDeviceFlowFallbackTokenStoreDir overrides the directory used for file-based token storage.
func WithDeviceFlowFallbackTokenStoreDir(dir string) DeviceFlowConfigOption {
	return func(d *DeviceFlowConfig) {
		d.fallbackTokenStoreDir = dir
	}
}

// With configures the DeviceFlowConfig with additional options.
func (d *DeviceFlowConfig) With(opts ...DeviceFlowConfigOption) {
	for _, opt := range opts {
		opt(d)
	}
}

// NewDeviceFlowConfig creates a DeviceFlowConfig with defaults.
func NewDeviceFlowConfig() *DeviceFlowConfig {
	cfg := defaultFlowConfig()
	cfg.clientID = DefaultClientID
	return &DeviceFlowConfig{SharedFlowConfig: cfg}
}

// PerformDeviceFlow initiates the device authorization flow, prints the verification URL and user
// code to stdout, and blocks until the user completes authentication or the context is cancelled.
func PerformDeviceFlow(ctx context.Context, conf *oauth2.Config) (*oauth2.Token, error) {
	da, err := conf.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("device authorization failed: %w", err)
	}

	if da.VerificationURIComplete != "" {
		fmt.Printf("Trying to open browser to: %s\n", da.VerificationURIComplete)
		openBrowserFunc(da.VerificationURIComplete)
	} else {
		fmt.Printf("Trying to open browser to: %s\n", da.VerificationURI)
		fmt.Printf("Enter code: %s\n", da.UserCode)
		openBrowserFunc(da.VerificationURI)
	}

	token, err := conf.DeviceAccessToken(ctx, da)
	if err != nil {
		return nil, fmt.Errorf("device access token failed: %w", err)
	}

	return token, nil
}

type deviceKeyringTokenSource struct {
	ctx          context.Context
	tokenKey     string // namespaced primary key
	legacyKey    string // legacy fixed key (read-only fallback)
	authBaseURL  string // used to validate legacy token issuer before adopting
	store        TokenStore
	conf         *oauth2.Config
	cachedClaims *TokenClaims
}

func (d *deviceKeyringTokenSource) ClearTokenStore() error {
	return d.store.DeleteAll()
}

func (d *deviceKeyringTokenSource) loadFromKeyring() (*oauth2.Token, error) {
	// Try the primary (namespaced) key first.
	stored, err := d.store.Get(d.tokenKey)
	if err == nil && stored != "" {
		var token oauth2.Token
		if err := json.Unmarshal([]byte(stored), &token); err != nil {
			_ = d.store.Delete(d.tokenKey)
			return nil, err
		}
		return &token, nil
	}

	// Primary key miss — attempt read-only fallback to the legacy fixed key.
	return d.loadLegacyToken()
}

// loadLegacyToken reads the legacy fixed-key entry and adopts it only when its
// issuer matches the current authBaseURL. On adoption the token is migrated to
// the namespaced key and the legacy entry is deleted, so it is evaluated at
// most once. A legacy entry that is absent, unparseable, or issued by a
// different auth server is left untouched.
func (d *deviceKeyringTokenSource) loadLegacyToken() (*oauth2.Token, error) {
	legacyStored, err := d.store.Get(d.legacyKey)
	if err != nil || legacyStored == "" {
		return nil, nil //nolint:nilnil // nil token signals "not found", distinct from an error
	}

	var legacyToken oauth2.Token
	if err := json.Unmarshal([]byte(legacyStored), &legacyToken); err != nil {
		// Unparseable legacy entry — skip adoption without deleting it.
		return nil, nil //nolint:nilnil
	}

	issuer, err := tokenIssuer(&legacyToken)
	if err != nil || !issuerMatchesAuthBaseURL(issuer, d.authBaseURL) {
		// Issuer absent, unparseable, or belongs to a different auth server.
		return nil, nil //nolint:nilnil
	}

	// Adopt and migrate: persist the token under the namespaced key and delete
	// the legacy entry so it is not re-evaluated on subsequent runs. If the
	// namespaced write fails we keep the legacy entry as a fallback and delete
	// nothing. Once the legacy entry is removed, downgrading to an older tool
	// version requires a one-time re-authentication, which is acceptable.
	if err := d.store.Set(d.tokenKey, legacyStored); err == nil {
		_ = d.store.Delete(d.legacyKey)
	}
	return &legacyToken, nil
}

func (d *deviceKeyringTokenSource) saveToKeyring(token *oauth2.Token) error {
	data, err := json.Marshal(token) //nolint:gosec // the token is stored in the keyring or token store on purpose
	if err != nil {
		_ = d.store.Delete(d.tokenKey)
		return err
	}

	return d.store.Set(d.tokenKey, string(data))
}

// nearExpiry reports whether token is expired or will expire within tokenRefreshBuffer.
func nearExpiry(token *oauth2.Token) bool {
	if token.Expiry.IsZero() {
		return false
	}
	return time.Until(token.Expiry) < tokenRefreshBuffer
}

func (d *deviceKeyringTokenSource) refreshOrReauthenticate(token *oauth2.Token) (*oauth2.Token, error) {
	if !nearExpiry(token) {
		// Token is well within its validity window; return it as-is.
		return token, nil
	}

	// Token is expired or within the refresh buffer — force a refresh by
	// passing a copy with expiry in the past so the oauth2 library uses the
	// refresh_token grant instead of returning the (near-)expired access token.
	stale := *token
	stale.Expiry = time.Now().Add(-time.Second)
	refreshed, err := d.conf.TokenSource(d.ctx, &stale).Token()
	if err == nil {
		return refreshed, nil
	}

	// If the refresh_token itself is invalid/expired the server returns
	// "invalid_grant". That is unrecoverable via refresh; fall back to a full
	// interactive device flow. Any other oauth2 error is also treated as
	// unrecoverable (e.g. server misconfiguration) and triggers re-auth.
	if !strings.HasPrefix(err.Error(), "oauth2:") {
		// Non-oauth2 error (network, etc.) — propagate without re-auth.
		return nil, err
	}

	return PerformDeviceFlow(d.ctx, d.conf)
}

func (d *deviceKeyringTokenSource) Token() (*oauth2.Token, error) {
	token, err := d.loadFromKeyring()
	if err != nil {
		return nil, err
	}

	if token == nil {
		token, err = PerformDeviceFlow(d.ctx, d.conf)
		if err != nil {
			return nil, err
		}
	}

	returnToken, err := d.refreshOrReauthenticate(token)
	if err != nil {
		return nil, err
	}

	if err := d.saveToKeyring(returnToken); err != nil {
		return nil, err
	}

	claims, err := ExtractClaims(returnToken)
	if err != nil {
		return nil, fmt.Errorf("token missing required eos claims: %w", err)
	}
	d.cachedClaims = claims

	return returnToken, nil
}

func (d *deviceKeyringTokenSource) TokenClaims() (*TokenClaims, error) {
	if d.cachedClaims != nil {
		return d.cachedClaims, nil
	}
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	return d.cachedClaims, nil
}

// NewDeviceKeyringTokenSource returns a TokenSourceWithClaims that persists tokens using the
// given TokenStore. On the first call or after token expiry, it triggers the device authorization
// flow to get a new token. EOS claims (realm, org) are extracted and cached inside Token().
func NewDeviceKeyringTokenSource(ctx context.Context, c *DeviceFlowConfig, store TokenStore) TokenSourceWithClaims {
	conf := &oauth2.Config{
		ClientID: c.clientID,
		Scopes:   []string{"openid", "offline"},
		Endpoint: oauth2.Endpoint{
			AuthURL:       fmt.Sprintf("%s/oauth2/auth", c.authBaseURL),
			DeviceAuthURL: fmt.Sprintf("%s/oauth2/device/auth", c.authBaseURL),
			TokenURL:      fmt.Sprintf("%s/oauth2/token", c.authBaseURL),
			AuthStyle:     oauth2.AuthStyleInParams,
		},
	}

	return &deviceKeyringTokenSource{
		ctx:         ctx,
		tokenKey:    namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID),
		legacyKey:   DefaultDeviceFlowTokenKey,
		authBaseURL: c.authBaseURL,
		store:       store,
		conf:        conf,
	}
}
