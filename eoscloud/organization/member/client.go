package member

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for member operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Member], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Member, err error)
	Create(ctx context.Context, object *model.MemberRequest) (result *model.Member, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedMemberRequest) (result *model.Member, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.Member]
	api.Getter[model.Member]
	api.Creator[model.MemberRequest, model.Member]
	api.PartialUpdater[model.PatchedMemberRequest, model.Member]
	api.Deleter
	api.ServerSideActionsFetcher
}

// NewClient returns the member client of the organization service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "members",
	}

	return &client{
		api.NewLister[model.Member](c, service, basePathPattern, pathParams),
		api.NewGetter[model.Member](c, service, basePathPattern, pathParams),
		api.NewCreator[model.MemberRequest, model.Member](c, service, basePathPattern, pathParams),
		api.NewPartialUpdater[model.PatchedMemberRequest, model.Member](c, service, basePathPattern, pathParams),
		api.NewDeleter(c, service, basePathPattern, pathParams),
		api.NewServerSideActionsFetcher(c, service, basePathPattern, pathParams),
	}
}
