package organization

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for organization operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Organization], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Organization, err error)
	Create(ctx context.Context, object *model.OrganizationRequest) (result *model.Organization, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedOrganizationRequest) (result *model.Organization, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Undelete(ctx context.Context, id string) (err error)
	Restore(ctx context.Context, id string) (err error)
	Stats(ctx context.Context, id string) (result *commonmodel.Statistics, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.Organization]
	api.Getter[model.Organization]
	api.Creator[model.OrganizationRequest, model.Organization]
	api.PartialUpdater[model.PatchedOrganizationRequest, model.Organization]
	api.Deleter
	api.ServerSideActionsFetcher
	undelete api.NoIOExecutor
	restore  api.NoIOExecutor
	stats    api.NoInputExecutor[commonmodel.Statistics]
}

func (c *client) Undelete(ctx context.Context, id string) error {
	return c.undelete.Execute(ctx, &id)
}

func (c *client) Restore(ctx context.Context, id string) error {
	return c.restore.Execute(ctx, &id)
}

func (c *client) Stats(ctx context.Context, id string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, &id)
}

// NewClient returns the organization resource client of the organization service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "organizations",
	}

	return &client{
		api.NewLister[model.Organization](c, service, basePathPattern, pathParams),
		api.NewGetter[model.Organization](c, service, basePathPattern, pathParams),
		api.NewCreator[model.OrganizationRequest, model.Organization](c, service, basePathPattern, pathParams),
		api.NewPartialUpdater[model.PatchedOrganizationRequest, model.Organization](c, service, basePathPattern, pathParams),
		api.NewDeleter(c, service, basePathPattern, pathParams),
		api.NewServerSideActionsFetcher(c, service, basePathPattern, pathParams),
		api.NewNoIOExecutor(c, service, basePathPattern, pathParams, "undelete", api.POST, api.IsHTTPStatusNoContent),
		api.NewNoIOExecutor(c, service, basePathPattern, pathParams, "restore", api.POST, api.IsHTTPStatusNoContent),
		api.NewNoInputExecutor[commonmodel.Statistics](c, service, basePathPattern, pathParams, "stats", api.GET, api.IsHTTPStatusOK),
	}
}
