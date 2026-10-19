package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

// debugRoundTripper prints the decoded JWT bearer token claims for every outgoing request.
// It must be placed inside the oauth2.Transport so it sees the Authorization header
// that oauth2 injects before calling the underlying transport.
type debugRoundTripper struct {
	inner http.RoundTripper
}

func (d *debugRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if authHeader := req.Header.Get("Authorization"); authHeader != "" {
		if rawToken, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			printJWTClaims(req.Method, req.URL.String(), rawToken)
		}
	}
	return d.inner.RoundTrip(req)
}

func printJWTClaims(method, url, rawToken string) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "[debug] Bearer token (not a JWT): %s\n", rawToken)
		return
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "[debug] Bearer token payload decode error: %v\n", err)
		return
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		fmt.Fprintf(os.Stderr, "[debug] Bearer token payload unmarshal error: %v\n", err)
		return
	}

	pretty, err := json.MarshalIndent(claims, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[debug] Bearer token claims marshal error: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "[debug] %s %s — bearer token claims:\n%s\n", method, url, pretty)
}

// WithDebugTokenPrinting wraps the HTTP transport to print decoded JWT bearer token
// claims on every request. Apply this option after WithClientCredentialsFlow or
// WithDeviceFlow so the oauth2 transport is already in place.
func WithDebugTokenPrinting() Option {
	return func(c *client) {
		switch t := c.httpClient.Transport.(type) {
		case *oauth2.Transport:
			if t.Base == nil {
				t.Base = http.DefaultTransport
			}
			t.Base = &debugRoundTripper{inner: t.Base}
		default:
			inner := c.httpClient.Transport
			if inner == nil {
				inner = http.DefaultTransport
			}
			c.httpClient.Transport = &debugRoundTripper{inner: inner}
		}
	}
}
