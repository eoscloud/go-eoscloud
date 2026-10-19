package eoscloud

import (
	"fmt"

	"go.eoscloud.io/go-eoscloud/eoscloud/auditlog"
	"go.eoscloud.io/go-eoscloud/eoscloud/billing"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute"
	objectstorage "go.eoscloud.io/go-eoscloud/eoscloud/object_storage"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization"
	"go.eoscloud.io/go-eoscloud/eoscloud/search"
	"go.eoscloud.io/go-eoscloud/pkg/api"
	"go.eoscloud.io/go-eoscloud/pkg/auth"
)

const (
	// BaseGlobalPathPattern is used for all services without a realm
	BaseGlobalPathPattern = "/{version}/{kind}/"
	// BaseRealmPathPattern is used for all services with a realm
	BaseRealmPathPattern = "/{version}/{realm_id}/{kind}/"
)

// Client represents EOS cloud API eoscloud
// It consists of multiple sub services
type Client struct {
	Compute       *compute.Client
	ObjectStorage *objectstorage.Client
	// Auditlog is the global (non-realm) auditlog service client. It reuses
	// the same authenticated api client, so no separate auth is needed.
	Auditlog *auditlog.Client
	// Organization is the organization service client (realms, etc.).
	Organization *organization.Client
	// Search is the global search service client.
	Search search.APIClient
	// Billing is the billing service client.
	Billing *billing.Client

	apiClient api.Caller
}

// NewClient returns a client to interact with eoscloud.
// Realm is resolved automatically from the token on the first API call.
func NewClient(options ...api.Option) *Client {
	apiClient := api.NewClient(options...)
	return &Client{
		Compute: compute.NewClient(
			apiClient,
			BaseGlobalPathPattern,
			BaseRealmPathPattern),
		ObjectStorage: objectstorage.NewClient(
			apiClient,
			BaseGlobalPathPattern,
			BaseRealmPathPattern,
		),
		Auditlog: auditlog.NewClient(apiClient),
		Organization: organization.NewClient(
			apiClient,
			BaseGlobalPathPattern,
			BaseRealmPathPattern,
		),
		Search: search.NewClient(apiClient),
		Billing: billing.NewClient(
			apiClient,
			BaseGlobalPathPattern,
		),
		apiClient: apiClient,
	}
}

// TokenClaims returns the realm and organization IDs from the authenticated
// token. It returns an error if the client was not configured with an auth
// flow (WithDeviceFlow, WithClientCredentialsFlow, or WithTokenSource).
func (c *Client) TokenClaims() (*auth.TokenClaims, error) {
	provider, ok := c.apiClient.(api.ClaimsProvider)
	if !ok {
		return nil, fmt.Errorf("token claims not available: client not configured with an auth flow")
	}
	return provider.TokenClaims()
}
