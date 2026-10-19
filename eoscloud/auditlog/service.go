package auditlog

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// Service is a single service name as returned by GET /v1/services. It is
// used to populate the service filter on the audit log List endpoint.
type Service struct {
	ID   string `json:"id,omitempty" yaml:"id,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

// ServiceAPIClient defines the interface for listing possible services in the
// audit log (GET /v1/services).
type ServiceAPIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[Service], err error)
}

var _ ServiceAPIClient = (*serviceClient)(nil)

type serviceClient struct {
	api.Lister[Service]
}

// NewServiceClient returns the auditlog service filter client (GET /v1/services).
func NewServiceClient(c api.Caller) ServiceAPIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "services",
	}
	return &serviceClient{
		api.NewLister[Service](c, "auditlog", auditlogPathPattern, pathParams),
	}
}
