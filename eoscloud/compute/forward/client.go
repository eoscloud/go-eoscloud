package forward

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for port forward operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.PortForward], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.PortForward, err error)
	Create(ctx context.Context, object *model.PortForwardRequest) (result *model.PortForward, err error)
	Update(ctx context.Context, id string, object *model.PortForwardRequest) (result *model.PortForward, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedPortForwardRequest) (result *model.PortForward, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)
}

var _ APIClient = (*client)(nil)

// client implements the forward component of the compute API eoscloud
type client struct {
	api.CRUDClient[
		model.PortForward,
		model.PortForwardRequest,
		model.PortForwardRequest,
		model.PatchedPortForwardRequest,
		model.PortForward,
	]
	stats api.NoInputExecutor[commonmodel.Statistics]
}

// Stats returns the statistics of a port forward.
func (c *client) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// NewClient returns the forward service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "forwards",
	}

	return &client{
		CRUDClient: api.NewCRUDClient[
			model.PortForward,
			model.PortForwardRequest,
			model.PortForwardRequest,
			model.PatchedPortForwardRequest,
			model.PortForward,
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
