package api

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// debugRoundTripper prints a fingerprint and selected JWT claims for every outgoing request.
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

// tokenFingerprint returns a short SHA-256 based fingerprint of the token so
// requests can be correlated without ever printing the token itself.
func tokenFingerprint(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return "sha256:" + hex.EncodeToString(sum[:])[:12]
}

func printJWTClaims(method, url, rawToken string) {
	fp := tokenFingerprint(rawToken)
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		fmt.Fprintf(os.Stderr, "[debug] %s %s - bearer token (not a JWT) %s\n", method, url, fp)
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

	// Only print a minimal, non-sensitive subset of claims.
	shown := map[string]any{}
	for _, k := range []string{"eos", "exp", "iss", "sub"} {
		if v, ok := claims[k]; ok {
			shown[k] = v
		}
	}

	pretty, err := json.MarshalIndent(shown, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[debug] Bearer token claims marshal error: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "[debug] %s %s - bearer token %s claims:\n%s\n", method, url, fp, pretty)
}

// WithDebugTokenPrinting prints a SHA-256 fingerprint of the bearer token and
// the claims eos, exp, iss and sub to stderr on every request. The raw token
// is never printed.
//
// DEBUG ONLY: do not enable this in production; it writes identity
// information to stderr. The option may be placed anywhere in the option list.
func WithDebugTokenPrinting() Option {
	return func(c *client) {
		c.debugTokens = true
	}
}
