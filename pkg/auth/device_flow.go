package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
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
	promptWriter     io.Writer
	browserOpener    func(url string) error
	browserOpenerSet bool
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

// WithDeviceFlowPromptWriter sets the writer that receives the verification
// URL and user code prompts. The default is os.Stderr; the library never
// writes to stdout.
func WithDeviceFlowPromptWriter(w io.Writer) DeviceFlowConfigOption {
	return func(d *DeviceFlowConfig) {
		d.promptWriter = w
	}
}

// WithDeviceFlowBrowserOpener sets the function used to open the verification
// URL. It is only called for URLs that passed validation (https, or http for
// loopback, and the same host as the auth server). Passing nil disables
// opening a browser; the URL and code are still printed to the prompt writer.
// The default is the built-in platform opener.
func WithDeviceFlowBrowserOpener(open func(url string) error) DeviceFlowConfigOption {
	return func(d *DeviceFlowConfig) {
		d.browserOpener = open
		d.browserOpenerSet = true
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
// code to stderr, and blocks until the user completes authentication or the context is cancelled.
// The browser is only opened for a verification URL on the same host as the device
// authorization endpoint.
func PerformDeviceFlow(ctx context.Context, conf *oauth2.Config) (*oauth2.Token, error) {
	return performDeviceFlow(ctx, conf, os.Stderr, openBrowserFunc, conf.Endpoint.DeviceAuthURL)
}

// performDeviceFlow runs the device flow. A nil open disables opening a
// browser; authURL is any URL on the auth server and defines the only host a
// browser may be pointed at.
func performDeviceFlow(ctx context.Context, conf *oauth2.Config, out io.Writer, open func(string) error, authURL string) (*oauth2.Token, error) {
	if out == nil {
		out = os.Stderr
	}

	da, err := conf.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("device authorization failed: %w", err)
	}

	target := da.VerificationURIComplete
	if target == "" {
		target = da.VerificationURI
		_, _ = fmt.Fprintf(out, "Open: %s\n", sanitizeForTerminal(target))
		_, _ = fmt.Fprintf(out, "Enter code: %s\n", sanitizeForTerminal(da.UserCode))
	} else {
		_, _ = fmt.Fprintf(out, "Open: %s\n", sanitizeForTerminal(target))
	}

	if open != nil {
		if verr := validateVerificationURL(target, authURL); verr != nil {
			_, _ = fmt.Fprintf(out, "Not opening a browser automatically: %v\n", verr)
		} else if oerr := open(target); oerr != nil {
			_, _ = fmt.Fprintf(out, "Failed to open browser. Please open the URL above manually and enter the verification code.\n")
		}
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
	promptWriter io.Writer
	openBrowser  func(string) error // nil disables opening a browser
	initErr      error              // configuration error surfaced by Token()

	mu           sync.Mutex // serializes Token() so only one load/refresh/flow runs at a time
	cachedClaims atomic.Pointer[TokenClaims]
}

// ClearTokenStore removes only this flow's own (namespaced) token entry.
func (d *deviceKeyringTokenSource) ClearTokenStore() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cachedClaims.Store(nil)
	if err := d.store.Delete(d.tokenKey); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return nil
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

// LoadDeviceFlowToken returns the token that a device flow token source for
// the given auth server and client ID stored in store, without running the
// interactive flow or refreshing it. It looks the token up exactly as the
// token source does, including the adoption of a legacy entry. It returns
// nil, nil when no token is cached.
func LoadDeviceFlowToken(store TokenStore, authBaseURL, clientID string) (*oauth2.Token, error) {
	d := &deviceKeyringTokenSource{
		tokenKey:    DeviceFlowTokenKey(authBaseURL, clientID),
		legacyKey:   DefaultDeviceFlowTokenKey,
		authBaseURL: authBaseURL,
		store:       store,
	}
	return d.loadFromKeyring()
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

func (d *deviceKeyringTokenSource) performDeviceFlow() (*oauth2.Token, error) {
	return performDeviceFlow(d.ctx, d.conf, d.promptWriter, d.openBrowser, d.authBaseURL)
}

func (d *deviceKeyringTokenSource) refreshOrReauthenticate(token *oauth2.Token) (*oauth2.Token, error) {
	if !nearExpiry(token) {
		// Token is well within its validity window; return it as-is.
		return token, nil
	}

	// Without a refresh token there is nothing to refresh with.
	if token.RefreshToken == "" {
		return d.performDeviceFlow()
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

	// Only an explicit invalid_grant (refresh token expired/revoked) is
	// unrecoverable via refresh and justifies a new interactive device flow.
	// Transient failures (network, 5xx, 429, ...) are returned to the caller.
	var rerr *oauth2.RetrieveError
	if errors.As(err, &rerr) && rerr.ErrorCode == "invalid_grant" {
		return d.performDeviceFlow()
	}

	return nil, err
}

func (d *deviceKeyringTokenSource) Token() (*oauth2.Token, error) {
	if d.initErr != nil {
		return nil, d.initErr
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	token, err := d.loadFromKeyring()
	if err != nil {
		return nil, err
	}

	if token == nil {
		token, err = d.performDeviceFlow()
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
	d.cachedClaims.Store(claims)

	return returnToken, nil
}

func (d *deviceKeyringTokenSource) TokenClaims() (*TokenClaims, error) {
	if c := d.cachedClaims.Load(); c != nil {
		cp := *c
		return &cp, nil
	}
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	c := d.cachedClaims.Load()
	if c == nil {
		return nil, errors.New("token claims unavailable")
	}
	cp := *c
	return &cp, nil
}

// NewDeviceKeyringTokenSource returns a TokenSourceWithClaims that persists tokens using the
// given TokenStore. On the first call or after token expiry, it triggers the device authorization
// flow to get a new token. EOS claims (realm, org) are extracted and cached inside Token().
//
// The auth base URL must be https (http only for loopback hosts); an invalid
// URL is reported as an error from Token() and TokenClaims(). The context is
// kept for the lifetime of the source but detached from cancellation (its
// values, such as oauth2.HTTPClient, are preserved). Token() is safe for
// concurrent use and serializes refreshes and interactive flows.
func NewDeviceKeyringTokenSource(ctx context.Context, c *DeviceFlowConfig, store TokenStore) TokenSourceWithClaims {
	if ctx == nil {
		ctx = context.Background()
	}

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

	open := openBrowserFunc
	if c.browserOpenerSet {
		open = c.browserOpener
	}

	return &deviceKeyringTokenSource{
		ctx:          context.WithoutCancel(ctx),
		tokenKey:     DeviceFlowTokenKey(c.authBaseURL, c.clientID),
		legacyKey:    DefaultDeviceFlowTokenKey,
		authBaseURL:  c.authBaseURL,
		store:        store,
		conf:         conf,
		promptWriter: c.promptWriter,
		openBrowser:  open,
		initErr:      validateAuthBaseURL(c.authBaseURL),
	}
}
