package groups

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for peering-groups operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.PeeringGroup], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.PeeringGroup, err error)
	Create(ctx context.Context, object *model.PeeringGroupRequest) (result *model.PeeringGroup, err error)
	Update(ctx context.Context, id string, object *model.PeeringGroupRequest) (result *model.PeeringGroup, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedPeeringGroupRequest) (result *model.PeeringGroup, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)

	AddMember(ctx context.Context, id *string, input *model.PeeringGroupMemberRequest) (result *model.PeeringGroup, err error)
	RemoveMember(ctx context.Context, id *string, input *model.PeeringGroupMemberRequest) (result *model.PeeringGroup, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.CRUDClient[
		model.PeeringGroup,
		model.PeeringGroupRequest,
		model.PeeringGroupRequest,
		model.PatchedPeeringGroupRequest,
		model.PeeringGroup,
	]
	addMember    api.Executor[model.PeeringGroupMemberRequest, model.PeeringGroup]
	removeMember api.Executor[model.PeeringGroupMemberRequest, model.PeeringGroup]
}

// AddMember adds a Network to a peering group.
func (c *apiClient) AddMember(ctx context.Context, id *string, input *model.PeeringGroupMemberRequest) (*model.PeeringGroup, error) {
	return c.addMember.Execute(ctx, id, input)
}

// RemoveMember removes a Network from a peering group.
func (c *apiClient) RemoveMember(ctx context.Context, id *string, input *model.PeeringGroupMemberRequest) (*model.PeeringGroup, error) {
	return c.removeMember.Execute(ctx, id, input)
}

// NewClient returns the peering-groups client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "peering-groups",
	}
	return &apiClient{
		CRUDClient: api.NewCRUDClient[
			model.PeeringGroup,
			model.PeeringGroupRequest,
			model.PeeringGroupRequest,
			model.PatchedPeeringGroupRequest,
			model.PeeringGroup,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		addMember: api.NewExecutor[model.PeeringGroupMemberRequest, model.PeeringGroup](
			c,
			service,
			basePathPattern,
			pathParams,
			"add-member",
			api.POST,
			api.IsHTTPStatusOK,
		),
		removeMember: api.NewExecutor[model.PeeringGroupMemberRequest, model.PeeringGroup](
			c,
			service,
			basePathPattern,
			pathParams,
			"remove-member",
			api.POST,
			api.IsHTTPStatusOK,
		),
	}
}
