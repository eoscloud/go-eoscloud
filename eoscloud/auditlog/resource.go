package auditlog

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// Resource is a single resource as returned by GET /v1/resources. It is used
// to populate the resource_id filter on the audit log List endpoint.
type Resource struct {
	ID             string `json:"id,omitempty" yaml:"id,omitempty"`
	Name           string `json:"name,omitempty" yaml:"name,omitempty"`
	Service        string `json:"service,omitempty" yaml:"service,omitempty"`
	ResourceKind   string `json:"resource_kind,omitempty" yaml:"resource_kind,omitempty"`
	OrganizationID string `json:"organization_id,omitempty" yaml:"organization_id,omitempty"`
	RealmID        string `json:"realm_id,omitempty" yaml:"realm_id,omitempty"`
}

// ResourceAPIClient defines the interface for listing possible resources in
// the audit log (GET /v1/resources).
//
// Supported filters (url.Values) on List include: service, resource_kind.
type ResourceAPIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[Resource], err error)
}

var _ ResourceAPIClient = (*resourceClient)(nil)

type resourceClient struct {
	api.Lister[Resource]
}

// NewResourceClient returns the auditlog resource filter client (GET /v1/resources).
func NewResourceClient(c api.Caller) ResourceAPIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "resources",
	}
	return &resourceClient{
		api.NewLister[Resource](c, "auditlog", auditlogPathPattern, pathParams),
	}
}
