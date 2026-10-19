package acl

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for ACL operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.ACL], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.ACL, err error)
	Create(ctx context.Context, object *model.ACLRequest) (result *model.ACL, err error)
	Update(ctx context.Context, id string, object *model.ACLRequest) (result *model.ACL, err error)
	UpdatePartial(ctx context.Context, id string, object *model.ACLRequest) (result *model.ACL, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.CRUDClient[
		model.ACL,
		model.ACLRequest,
		model.ACLRequest,
		model.ACLRequest,
		model.ACL,
	]
	stats api.NoInputExecutor[commonmodel.Statistics]
}

// Stats returns the statistics of an ACL.
func (c *client) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// NewClient returns the ACL service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "acls",
	}

	return &client{
		CRUDClient: api.NewCRUDClient[
			model.ACL,
			model.ACLRequest,
			model.ACLRequest,
			model.ACLRequest,
			model.ACL,
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
