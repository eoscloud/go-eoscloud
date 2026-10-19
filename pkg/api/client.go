package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/pkg/auth"
)

const (
	// BaseURLPattern represents the default eoscloud baseURL
	BaseURLPattern = "https://{service}.eoscloud.io"
)

var (
	AllowedErrorStatusCodes = []int{
		400,
		401,
		403,
		404,
		405,
		409,
		422,
		429,
		500,
	}
)

type RequestModifier func(req *http.Request)

// Caller is the interface used by all API operation types (Lister, Getter, Creator, etc.)
// to perform HTTP requests. Implement this interface to provide a custom HTTP client for testing.
type Caller interface {
	Call(ctx context.Context, r *CallRequest) (any, error)
}

// ClaimsProvider is optionally implemented by a Caller to expose the realm and
// organization IDs from the authenticated token. The Caller returned by
// NewClient implements this interface when configured with an auth flow
// (WithDeviceFlow, WithClientCredentialsFlow, or WithTokenSource).
type ClaimsProvider interface {
	TokenClaims() (*auth.TokenClaims, error)
}

var _ ClaimsProvider = (*client)(nil)

// TokenClaims returns the realm and organization IDs extracted from the
// authenticated token. It returns an error if no auth flow was configured.
func (c *client) TokenClaims() (*auth.TokenClaims, error) {
	if c.tokenSource == nil {
		return nil, fmt.Errorf("token claims not available: client not configured with an auth flow")
	}
	return c.tokenSource.TokenClaims()
}

// client represents an HTTP eoscloud for interacting with an API.
// baseURL defines the base URL for the API endpoint in form scheme://host:port
// httpClient is the underlying HTTP client used for making requests.
type client struct {
	baseURL     string
	httpClient  *http.Client
	tokenSource auth.TokenSourceWithClaims
	// userAgentApp is the optional "name/version" set via WithUserAgent.
	userAgentApp string
	userAgent    string
}

// resolvePathParams merges the caller-supplied pathParams with realm_id sourced
// from the token claims (if an auth flow is configured). Returns the original
// map unchanged when no token source is set or the realm is empty.
func (c *client) resolvePathParams(base map[string]string) (map[string]string, error) {
	if c.tokenSource == nil {
		return base, nil
	}

	claims, err := c.tokenSource.TokenClaims()
	if err != nil {
		return nil, fmt.Errorf("resolving realm from token: %w", err)
	}

	if claims.RealmID == "" {
		return base, nil
	}

	// Copy so we never mutate the shared PathParams map stored in sub-clients.
	merged := make(map[string]string, len(base)+1)
	for k, v := range base {
		merged[k] = v
	}
	merged["realm_id"] = claims.RealmID
	return merged, nil
}

func (c *client) buildServiceBaseURL(service string) string {
	return strings.ReplaceAll(c.baseURL, "{service}", url.PathEscape(service))
}

// NewClient creates a new API client.
func NewClient(opts ...Option) Caller {
	c := &client{
		baseURL: BaseURLPattern,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.userAgent = buildUserAgent(c.userAgentApp)
	return c
}

func (c *client) buildRequestURL(service, path string, pathParams map[string]string, query url.Values) (string, error) {
	base := strings.TrimRight(c.buildServiceBaseURL(service), "/")
	for k, v := range pathParams {
		path = strings.ReplaceAll(path, "{"+k+"}", url.PathEscape(v))
	}
	if strings.Contains(path, "{") {
		return "", fmt.Errorf("missing basePathPattern parameters for %q", path)
	}

	requestURL := base + path
	if len(query) > 0 {
		if qs := query.Encode(); qs != "" {
			requestURL += "?" + qs
		}
	}

	return requestURL, nil
}

func (c *client) clearKeyringOnAuthError(err error, _ int) {
	// A bare HTTP 401 must NOT clear the token store: the stored refresh token
	// is still valid and the token source (device flow) will use it to obtain a
	// new access token on the next call. Wiping the store on every 401 would
	// destroy the refresh token and force unnecessary interactive re-logins.
	//
	// We only clear the store when the oauth2 transport itself surfaces an
	// oauth2 error during token acquisition (e.g. a network error talking to
	// the token endpoint), which signals a problem with the stored credentials
	// rather than a transient API response. Recovery from invalid_grant (dead
	// refresh token) is handled inside the token source itself.
	if c.tokenSource == nil {
		return
	}

	if err != nil && strings.Contains(err.Error(), "oauth2:") {
		_ = c.tokenSource.ClearTokenStore()
	}
}

func handleResponseStatusError(resp *http.Response, method Method, requestURL string) error {
	if slices.Contains(AllowedErrorStatusCodes, resp.StatusCode) {
		var r model.APIError
		if decErr := json.NewDecoder(resp.Body).Decode(&r); decErr != nil {
			return decErr
		}
		r.StatusCode = resp.StatusCode
		return &r
	}

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	return &UnexpectedHttpStatusError{
		StatusCode: resp.StatusCode,
		Method:     method,
		URL:        requestURL,
		Body:       string(bodyBytes),
	}
}

// CallRequest holds all parameters for a Client.Call invocation.
type CallRequest struct {
	Method      Method
	Service     string
	Path        string
	PathParams  map[string]string
	Query       url.Values
	Encoder     Encoder
	StatusCheck HTTPStatusCheck
	Decoder     Decoder[any]
}

// Call performs the HTTP request described by the given CallRequest.
func (c *client) Call(ctx context.Context, r *CallRequest) (any, error) {
	pathParams, err := c.resolvePathParams(r.PathParams)
	if err != nil {
		return nil, err
	}

	requestURL, err := c.buildRequestURL(r.Service, r.Path, pathParams, r.Query)
	if err != nil {
		return nil, err
	}

	contentType, requestBody, encodeErr := r.Encoder()
	if encodeErr != nil {
		return nil, encodeErr
	}

	req, err := http.NewRequestWithContext(ctx, string(r.Method), requestURL, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", appendUserAgent(req.Header.Get("User-Agent"), c.userAgent))

	if contentType != "" {
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.clearKeyringOnAuthError(err, 0)
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	c.clearKeyringOnAuthError(nil, resp.StatusCode)

	if !r.StatusCheck(resp.StatusCode) {
		return nil, handleResponseStatusError(resp, r.Method, requestURL)
	}

	if r.Decoder != nil {
		return r.Decoder(resp.Body)
	}

	return nil, nil
}
