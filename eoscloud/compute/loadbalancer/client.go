package loadbalancer

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for load balancer operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.LoadBalancer], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.LoadBalancer, err error)
	Create(ctx context.Context, object *model.LoadBalancerRequest) (result *model.LoadBalancer, err error)
	Update(ctx context.Context, id string, object *model.LoadBalancerRequest) (result *model.LoadBalancer, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedLoadBalancerRequest) (result *model.LoadBalancer, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
}

var _ APIClient = (*client)(nil)

// client implements the loadbalancer component of the compute API eoscloud
type client struct {
	api.CRUDClient[
		model.LoadBalancer,
		model.LoadBalancerRequest,
		model.LoadBalancerRequest,
		model.PatchedLoadBalancerRequest,
		model.LoadBalancer,
	]
}

// NewClient returns the loadbalancer service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "load-balancers",
	}

	return &client{
		CRUDClient: api.NewCRUDClient[
			model.LoadBalancer,
			model.LoadBalancerRequest,
			model.LoadBalancerRequest,
			model.PatchedLoadBalancerRequest,
			model.LoadBalancer,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
	}
}
