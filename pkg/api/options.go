package api

import (
	"context"
	"net/http"

	"go.eoscloud.io/go-eoscloud/pkg/auth"
	"golang.org/x/oauth2"
)

// Option configures a client.
type Option func(*client)

// WithBaseURL sets a custom baseURL
func WithBaseURL(baseURL string) Option {
	return func(c *client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *client) {
		c.httpClient = h
	}
}

// sharedTokenSource ensures that resolvePathParams and the HTTP transport share
// the same oauth2.ReuseTokenSource cache. Without this, each caller creates its
// own cache and Token() is called twice per request — once for claim resolution
// and once to attach the Authorization header.
type sharedTokenSource struct {
	inner auth.TokenSourceWithClaims
	reuse oauth2.TokenSource // the ReuseTokenSource created by oauth2.NewClient
}

// Token uses the shared reuse cache — the same instance the HTTP transport uses.
func (s *sharedTokenSource) Token() (*oauth2.Token, error) {
	return s.reuse.Token()
}

// TokenClaims warms the reuse cache first (no-op if already warm), then returns
// the claims cached inside the underlying token source.
func (s *sharedTokenSource) TokenClaims() (*auth.TokenClaims, error) {
	if _, err := s.reuse.Token(); err != nil {
		return nil, err
	}
	return s.inner.TokenClaims()
}

// ClearTokenStore delegates to the underlying token source.
func (s *sharedTokenSource) ClearTokenStore() error {
	return s.inner.ClearTokenStore()
}

// wrapShared lets oauth2.NewClient do its normal setup (including context
// transport), then extracts the ReuseTokenSource it created so we can share it.
func wrapShared(ctx context.Context, ts auth.TokenSourceWithClaims) (*http.Client, *sharedTokenSource) {
	httpClient := oauth2.NewClient(ctx, ts)
	rts := httpClient.Transport.(*oauth2.Transport).Source
	return httpClient, &sharedTokenSource{inner: ts, reuse: rts}
}

// WithTokenSource sets a pre-built TokenSourceWithClaims directly on the client.
// The token source is used as-is, without the shared ReuseTokenSource wrapper, so
// it is intended primarily for testing.
func WithTokenSource(ts auth.TokenSourceWithClaims) Option {
	return func(c *client) {
		c.tokenSource = ts
	}
}

// WithClientCredentialsFlow sets the client to use the client credentials flow.
func WithClientCredentialsFlow(
	ctx context.Context,
	clientID, clientSecret string,
	opts ...auth.ClientCredentialsConfigOption) Option {

	return func(c *client) {
		cc := auth.NewClientCredentialsConfig(clientID, clientSecret)
		cc.With(opts...)
		store := auth.NewTokenStore(cc.KeyRingServiceName(), cc.FallbackTokenStoreDir())
		ts := auth.NewClientCredentialsKeyringTokenSource(ctx, cc, store)
		c.httpClient, c.tokenSource = wrapShared(ctx, ts)
	}
}

// WithDeviceFlow sets the client to use the OAuth2 device authorization flow.
// If no cached token exists, the user will be prompted to authenticate via browser.
// Realm and organization are extracted from the token after login.
func WithDeviceFlow(
	ctx context.Context,
	opts ...auth.DeviceFlowConfigOption) Option {

	return func(c *client) {
		cc := auth.NewDeviceFlowConfig()
		cc.With(opts...)
		store := auth.NewTokenStore(cc.KeyRingServiceName(), cc.FallbackTokenStoreDir())
		ts := auth.NewDeviceKeyringTokenSource(ctx, cc, store)
		c.httpClient, c.tokenSource = wrapShared(ctx, ts)
	}
}
