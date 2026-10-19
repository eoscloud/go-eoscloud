package user

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for ACL operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.User], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.User, err error)
	Create(ctx context.Context, object *model.UserRequest) (result *model.User, err error)
	Update(ctx context.Context, id string, object *model.PatchedUserRequest) (result *model.User, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedUserRequest) (result *model.User, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Reset(ctx context.Context, id *string) (*model.User, error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.CRUDClient[
		model.User,
		model.UserRequest,
		model.PatchedUserRequest,
		model.PatchedUserRequest,
		model.User,
	]
	reset api.NoInputExecutor[model.User]
}

func (c *client) Reset(ctx context.Context, id *string) (*model.User, error) {
	return c.reset.Execute(ctx, id)
}

// NewClient returns the ACL service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "users",
	}
	return &client{
		CRUDClient: api.NewCRUDClient[
			model.User,
			model.UserRequest,
			model.PatchedUserRequest,
			model.PatchedUserRequest,
			model.User,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		reset: api.NewNoInputExecutor[model.User](
			c,
			service,
			basePathPattern,
			pathParams,
			"reset",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
	}
}
