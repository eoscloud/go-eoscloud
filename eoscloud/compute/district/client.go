package district

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for block operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.District], err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.District]
}

// NewClient returns the Block client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "districts",
	}

	return &client{
		api.NewLister[model.District](c, service, basePathPattern, pathParams),
	}
}
