package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"go.eoscloud.io/go-eoscloud/pkg/auth"
	"golang.org/x/oauth2"
)

// Option configures a client. Options only record configuration; the final
// http.Client is assembled by NewClient after all options ran, so the order of
// options does not matter.
type Option func(*client)

// WithBaseURL sets a custom baseURL. The URL must use https; http is only
// accepted for loopback hosts (localhost, 127.0.0.0/8, ::1). The pattern may
// contain the "{service}" placeholder. An invalid URL does not panic: the
// error is returned by every Call.
func WithBaseURL(baseURL string) Option {
	return func(c *client) {
		if err := validateBaseURL(baseURL); err != nil {
			c.optErr = err
			return
		}
		c.baseURL = baseURL
	}
}

// validateBaseURL checks scheme and host of a base URL pattern.
func validateBaseURL(baseURL string) error {
	u, err := url.Parse(strings.ReplaceAll(baseURL, "{service}", "service"))
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("invalid base URL %q: missing host", baseURL)
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if isLoopbackHost(u.Hostname()) {
			return nil
		}
		return fmt.Errorf("invalid base URL %q: http is only allowed for loopback hosts", baseURL)
	default:
		return fmt.Errorf("invalid base URL %q: scheme must be https", baseURL)
	}
}

// isLoopbackHost reports whether host is localhost or a loopback IP.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// WithWebSocketSubdomains allows additional subdomains of the API base domain
// for server-provided WebSocket URLs; "compute-console" (the VM console) is
// always allowed. Each entry must be a
// single DNS label; it is resolved against the base domain and port of the
// configured base URL ("console" with https://compute.eoscloud.io allows
// console.eoscloud.io). The API host itself is always allowed. An invalid entry
// does not panic: the error is returned by every Call.
func WithWebSocketSubdomains(subdomains ...string) Option {
	return func(c *client) {
		for _, sub := range subdomains {
			if !dnsLabel.MatchString(sub) {
				c.optErr = fmt.Errorf("invalid websocket subdomain %q: want a single DNS label", sub)
				return
			}
			c.wsSubdomains = append(c.wsSubdomains, sub)
		}
	}
}

// dnsLabel matches a single DNS label (RFC 1123).
var dnsLabel = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

// WithHTTPClient sets a custom base http.Client. Its Transport and Timeout are
// kept; when an auth flow is configured, the oauth2 transport is layered on
// top of the client's transport (in any option order).
func WithHTTPClient(h *http.Client) Option {
	return func(c *client) {
		c.httpClient = h
	}
}

// sharedTokenSource ensures that resolvePathParams and the HTTP transport share
// the same oauth2.ReuseTokenSource cache and that all access to the underlying
// token source (including its cached claims) is serialized by one mutex.
type sharedTokenSource struct {
	mu    sync.Mutex
	inner auth.TokenSourceWithClaims
	reuse oauth2.TokenSource // ReuseTokenSource wrapping inner
}

// Token uses the shared reuse cache. The HTTP transport uses this method too.
func (s *sharedTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reuse.Token()
}

// TokenClaims warms the reuse cache first (no-op if already warm), then returns
// the claims cached inside the underlying token source.
func (s *sharedTokenSource) TokenClaims() (*auth.TokenClaims, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.reuse.Token(); err != nil {
		return nil, err
	}
	return s.inner.TokenClaims()
}

// ClearTokenStore delegates to the underlying token source.
func (s *sharedTokenSource) ClearTokenStore() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inner.ClearTokenStore()
}

// WithTokenSource sets a pre-built TokenSourceWithClaims directly on the client.
// The token source is used as-is, without the shared ReuseTokenSource wrapper
// and without attaching an Authorization header, so it is intended primarily
// for testing.
func WithTokenSource(ts auth.TokenSourceWithClaims) Option {
	return func(c *client) {
		c.tokenSource = ts
		c.authFactory = nil
	}
}

// WithClientCredentialsFlow sets the client to use the client credentials flow.
func WithClientCredentialsFlow(
	ctx context.Context,
	clientID, clientSecret string,
	opts ...auth.ClientCredentialsConfigOption) Option {

	return func(c *client) {
		c.authCtx = ctx
		c.authFactory = func(ctx context.Context) auth.TokenSourceWithClaims {
			cc := auth.NewClientCredentialsConfig(clientID, clientSecret)
			cc.With(opts...)
			store := auth.NewTokenStore(cc.KeyRingServiceName(), cc.FallbackTokenStoreDir())
			// store is only used to clear legacy entries; no replacement constructor exists.
			return auth.NewClientCredentialsKeyringTokenSource(ctx, cc, store) //nolint:staticcheck
		}
		c.tokenSource = nil
	}
}

// WithDeviceFlow sets the client to use the OAuth2 device authorization flow.
// If no cached token exists, the user will be prompted to authenticate via browser.
// Realm and organization are extracted from the token after login.
func WithDeviceFlow(
	ctx context.Context,
	opts ...auth.DeviceFlowConfigOption) Option {

	return func(c *client) {
		c.authCtx = ctx
		c.authFactory = func(ctx context.Context) auth.TokenSourceWithClaims {
			cc := auth.NewDeviceFlowConfig()
			cc.With(opts...)
			store := auth.NewTokenStore(cc.KeyRingServiceName(), cc.FallbackTokenStoreDir())
			return auth.NewDeviceKeyringTokenSource(ctx, cc, store)
		}
		c.tokenSource = nil
	}
}

// buildHTTPClient assembles the final http.Client from the base client and the
// recorded auth/debug configuration.
func (c *client) buildHTTPClient() {
	base := c.httpClient
	transport := base.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if c.debugTokens {
		transport = &debugRoundTripper{inner: transport}
	}

	if c.authFactory != nil {
		// Token endpoint requests use the base client (timeout, transport)
		// but never the oauth2 transport itself.
		tokenClient := &http.Client{
			Transport: base.Transport,
			Timeout:   base.Timeout,
			Jar:       base.Jar,
		}
		ctx := c.authCtx
		if ctx == nil {
			ctx = context.Background()
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, tokenClient)
		ts := c.authFactory(ctx)
		shared := &sharedTokenSource{inner: ts, reuse: oauth2.ReuseTokenSource(nil, ts)}
		c.tokenSource = shared
		transport = &oauth2.Transport{Source: shared, Base: transport}
	}

	c.httpClient = &http.Client{
		Transport:     transport,
		Timeout:       base.Timeout,
		Jar:           base.Jar,
		CheckRedirect: chainCheckRedirect(base.CheckRedirect),
	}
}

// chainCheckRedirect enforces safeCheckRedirect first and then applies a
// CheckRedirect configured on the base client, if any.
func chainCheckRedirect(next func(*http.Request, []*http.Request) error) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if err := safeCheckRedirect(req, via); err != nil {
			return err
		}
		if next != nil {
			return next(req, via)
		}
		return nil
	}
}

// safeCheckRedirect only follows redirects to the same scheme and host as the
// original request, and only over https (http only for loopback). Anything
// else returns the redirect response unfollowed, so credentials never leave
// the original origin.
func safeCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("stopped after %d redirects", len(via))
	}
	orig := via[0].URL
	if req.URL.Scheme != orig.Scheme || !strings.EqualFold(req.URL.Host, orig.Host) {
		return http.ErrUseLastResponse
	}
	switch req.URL.Scheme {
	case "https":
		return nil
	case "http":
		if isLoopbackHost(orig.Hostname()) {
			return nil
		}
	}
	return http.ErrUseLastResponse
}
