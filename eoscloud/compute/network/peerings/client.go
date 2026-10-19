package peerings

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for network-peerings operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.NetworkPeering], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.NetworkPeering, err error)
	Create(ctx context.Context, object *model.NetworkPeeringRequest) (result *model.NetworkPeering, err error)
	Update(ctx context.Context, id string, object *model.PatchedNetworkPeeringRequest) (result *model.NetworkPeering, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedNetworkPeeringRequest) (result *model.NetworkPeering, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)

	AddMember(ctx context.Context, id *string, input *model.NetworkPeeringMemberRequest) (result *model.NetworkPeering, err error)
	RemoveMember(ctx context.Context, id *string, input *model.NetworkPeeringMemberRequest) (result *model.NetworkPeering, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.CRUDClient[
		model.NetworkPeering,
		model.NetworkPeeringRequest,
		model.PatchedNetworkPeeringRequest,
		model.PatchedNetworkPeeringRequest,
		model.NetworkPeering,
	]
	addMember    api.Executor[model.NetworkPeeringMemberRequest, model.NetworkPeering]
	removeMember api.Executor[model.NetworkPeeringMemberRequest, model.NetworkPeering]
}

// AddMember adds a Network to a Network peering.
func (c *apiClient) AddMember(ctx context.Context, id *string, input *model.NetworkPeeringMemberRequest) (*model.NetworkPeering, error) {
	return c.addMember.Execute(ctx, id, input)
}

// RemoveMember removes a Network from a Network peering.
func (c *apiClient) RemoveMember(ctx context.Context, id *string, input *model.NetworkPeeringMemberRequest) (*model.NetworkPeering, error) {
	return c.removeMember.Execute(ctx, id, input)
}

// NewClient returns the network-peerings client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "network-peerings",
	}
	return &apiClient{
		CRUDClient: api.NewCRUDClient[
			model.NetworkPeering,
			model.NetworkPeeringRequest,
			model.PatchedNetworkPeeringRequest,
			model.PatchedNetworkPeeringRequest,
			model.NetworkPeering,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		addMember: api.NewExecutor[model.NetworkPeeringMemberRequest, model.NetworkPeering](
			c,
			service,
			basePathPattern,
			pathParams,
			"add-member",
			api.POST,
			api.IsHTTPStatusOK,
		),
		removeMember: api.NewExecutor[model.NetworkPeeringMemberRequest, model.NetworkPeering](
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
