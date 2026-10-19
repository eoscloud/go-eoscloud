package realm

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for realm operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Realm], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Realm, err error)
	Create(ctx context.Context, object *model.RealmRequest) (result *model.Realm, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedRealmRequest) (result *model.Realm, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Suspend(ctx context.Context, id string) (err error)
	Restore(ctx context.Context, id string) (err error)
	CancelRestore(ctx context.Context, id string) (err error)
	Stats(ctx context.Context, id string) (result *commonmodel.Statistics, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.Realm]
	api.Getter[model.Realm]
	api.Creator[model.RealmRequest, model.Realm]
	api.PartialUpdater[model.PatchedRealmRequest, model.Realm]
	api.Deleter
	api.ServerSideActionsFetcher
	suspend       api.NoIOExecutor
	restore       api.NoIOExecutor
	cancelRestore api.NoIOExecutor
	stats         api.NoInputExecutor[commonmodel.Statistics]
}

func (c *client) Suspend(ctx context.Context, id string) error {
	return c.suspend.Execute(ctx, &id)
}

func (c *client) Restore(ctx context.Context, id string) error {
	return c.restore.Execute(ctx, &id)
}

func (c *client) CancelRestore(ctx context.Context, id string) error {
	return c.cancelRestore.Execute(ctx, &id)
}

func (c *client) Stats(ctx context.Context, id string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, &id)
}

// NewClient returns the realm client of the organization service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "realms",
	}

	return &client{
		api.NewLister[model.Realm](c, service, basePathPattern, pathParams),
		api.NewGetter[model.Realm](c, service, basePathPattern, pathParams),
		api.NewCreator[model.RealmRequest, model.Realm](c, service, basePathPattern, pathParams),
		api.NewPartialUpdater[model.PatchedRealmRequest, model.Realm](c, service, basePathPattern, pathParams),
		api.NewDeleter(c, service, basePathPattern, pathParams),
		api.NewServerSideActionsFetcher(c, service, basePathPattern, pathParams),
		api.NewNoIOExecutor(c, service, basePathPattern, pathParams, "suspend", api.POST, api.IsHTTPStatusNoContent),
		api.NewNoIOExecutor(c, service, basePathPattern, pathParams, "restore", api.POST, api.IsHTTPStatusNoContent),
		api.NewNoIOExecutor(c, service, basePathPattern, pathParams, "cancel-restore", api.POST, api.IsHTTPStatusNoContent),
		api.NewNoInputExecutor[commonmodel.Statistics](c, service, basePathPattern, pathParams, "stats", api.GET, api.IsHTTPStatusOK),
	}
}
