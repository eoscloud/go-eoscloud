package block

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for block operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Block], err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.Block]
}

// NewClient returns the Block client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "blocks",
	}

	return &client{
		api.NewLister[model.Block](c, service, basePathPattern, pathParams),
	}
}
