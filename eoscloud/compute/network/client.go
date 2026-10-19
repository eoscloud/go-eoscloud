package network

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for network operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Network], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Network, err error)
	Create(ctx context.Context, object *model.NetworkRequest) (result *model.Network, err error)
	Update(ctx context.Context, id string, object *model.PatchedNetworkRequest) (result *model.Network, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedNetworkRequest) (result *model.Network, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)
}

var _ APIClient = (*client)(nil)

// client implements the network component of the compute API eoscloud
type client struct {
	api.CRUDClient[
		model.Network,
		model.NetworkRequest,
		model.PatchedNetworkRequest,
		model.PatchedNetworkRequest,
		model.Network,
	]
	stats api.NoInputExecutor[commonmodel.Statistics]
}

// Stats returns the statistics of a network.
func (c *client) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// NewClient returns the network service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "networks",
	}

	return &client{
		CRUDClient: api.NewCRUDClient[
			model.Network,
			model.NetworkRequest,
			model.PatchedNetworkRequest,
			model.PatchedNetworkRequest,
			model.Network,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		stats: api.NewNoInputExecutor[commonmodel.Statistics](
			c,
			service,
			basePathPattern,
			pathParams,
			"stats",
			api.GET,
			api.IsHTTPStatusOK,
		),
	}
}
