package role

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for role operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Role], err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.Role]
}

// NewClient returns the role client of the organization service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "roles",
	}

	return &client{
		api.NewLister[model.Role](c, service, basePathPattern, pathParams),
	}
}
