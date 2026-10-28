package loadbalancer

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for load balancer product operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.LoadBalancerProduct], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.LoadBalancerProduct, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.Lister[model.LoadBalancerProduct]
	api.Getter[model.LoadBalancerProduct]
}

// NewClient returns the load balancer product client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "load-balancer-products",
	}

	return &apiClient{
		api.NewLister[model.LoadBalancerProduct](c, service, basePathPattern, pathParams),
		api.NewGetter[model.LoadBalancerProduct](c, service, basePathPattern, pathParams),
	}
}
