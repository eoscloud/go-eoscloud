package groups

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for vm-groups operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.VMGroup], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.VMGroup, err error)
	Create(ctx context.Context, object *model.VMGroupRequest) (result *model.VMGroup, err error)
	Update(ctx context.Context, id string, object *model.PatchedVMGroupRequest) (result *model.VMGroup, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedVMGroupRequest) (result *model.VMGroup, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)

	AddMember(ctx context.Context, id *string, input *model.VMGroupMemberRequest) (result *model.VMGroup, err error)
	RemoveMember(ctx context.Context, id *string, input *model.VMGroupMemberRequest) (result *model.VMGroup, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.CRUDClient[
		model.VMGroup,
		model.VMGroupRequest,
		model.PatchedVMGroupRequest,
		model.PatchedVMGroupRequest,
		model.VMGroup,
	]
	addMember    api.Executor[model.VMGroupMemberRequest, model.VMGroup]
	removeMember api.Executor[model.VMGroupMemberRequest, model.VMGroup]
}

// AddMember adds a VM to a VM group.
func (c *apiClient) AddMember(ctx context.Context, id *string, input *model.VMGroupMemberRequest) (*model.VMGroup, error) {
	return c.addMember.Execute(ctx, id, input)
}

// RemoveMember removes a VM from a VM group.
func (c *apiClient) RemoveMember(ctx context.Context, id *string, input *model.VMGroupMemberRequest) (*model.VMGroup, error) {
	return c.removeMember.Execute(ctx, id, input)
}

// NewClient returns the vm-groups client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "vm-groups",
	}
	return &apiClient{
		CRUDClient: api.NewCRUDClient[
			model.VMGroup,
			model.VMGroupRequest,
			model.PatchedVMGroupRequest,
			model.PatchedVMGroupRequest,
			model.VMGroup,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		addMember: api.NewExecutor[model.VMGroupMemberRequest, model.VMGroup](
			c,
			service,
			basePathPattern,
			pathParams,
			"add-member",
			api.POST,
			api.IsHTTPStatusOK,
		),
		removeMember: api.NewExecutor[model.VMGroupMemberRequest, model.VMGroup](
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
