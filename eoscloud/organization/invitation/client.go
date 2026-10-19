package invitation

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for invitation operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Invitation], err error)
	Create(ctx context.Context, object *model.InvitationRequest) (result *model.Invitation, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.Invitation]
	api.Creator[model.InvitationRequest, model.Invitation]
	api.Deleter
	api.ServerSideActionsFetcher
}

// NewClient returns the invitation client of the organization service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "invitations",
	}

	return &client{
		api.NewLister[model.Invitation](c, service, basePathPattern, pathParams),
		api.NewCreator[model.InvitationRequest, model.Invitation](c, service, basePathPattern, pathParams),
		api.NewDeleter(c, service, basePathPattern, pathParams),
		api.NewServerSideActionsFetcher(c, service, basePathPattern, pathParams),
	}
}
