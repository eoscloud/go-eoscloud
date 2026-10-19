package auditlog

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// ResourceKind is a single resource kind as returned by GET
// /v1/resource-kinds. It is used to populate the resource_kind filter on the
// audit log List endpoint.
type ResourceKind struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty"`
	Service string `json:"service,omitempty" yaml:"service,omitempty"`
}

// ResourceKindAPIClient defines the interface for listing possible resource
// kinds in the audit log (GET /v1/resource-kinds).
//
// Supported filters (url.Values) on List include: service.
type ResourceKindAPIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[ResourceKind], err error)
}

var _ ResourceKindAPIClient = (*resourceKindClient)(nil)

type resourceKindClient struct {
	api.Lister[ResourceKind]
}

// NewResourceKindClient returns the auditlog resource-kind filter client (GET /v1/resource-kinds).
func NewResourceKindClient(c api.Caller) ResourceKindAPIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "resource-kinds",
	}
	return &resourceKindClient{
		api.NewLister[ResourceKind](c, "auditlog", auditlogPathPattern, pathParams),
	}
}
