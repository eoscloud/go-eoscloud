package auditlog

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// Identity is a single identity as returned by GET /v1/identities. It is used
// to populate the identity_id filter on the audit log List endpoint.
type Identity struct {
	ID             string       `json:"id,omitempty" yaml:"id,omitempty"`
	Email          string       `json:"email,omitempty" yaml:"email,omitempty"`
	Name           string       `json:"name,omitempty" yaml:"name,omitempty"`
	OrganizationID string       `json:"organization_id,omitempty" yaml:"organization_id,omitempty"`
	Type           IdentityType `json:"type,omitempty" yaml:"type,omitempty"`
}

// IdentityAPIClient defines the interface for listing possible identities in
// the audit log (GET /v1/identities).
type IdentityAPIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[Identity], err error)
}

var _ IdentityAPIClient = (*identityClient)(nil)

type identityClient struct {
	api.Lister[Identity]
}

// NewIdentityClient returns the auditlog identity filter client (GET /v1/identities).
func NewIdentityClient(c api.Caller) IdentityAPIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "identities",
	}
	return &identityClient{
		api.NewLister[Identity](c, "auditlog", auditlogPathPattern, pathParams),
	}
}
