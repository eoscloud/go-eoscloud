package vm

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for VM product operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.VMProduct], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.VMProduct, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.Lister[model.VMProduct]
	api.Getter[model.VMProduct]
}

// NewClient returns the VM product client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "vm-products",
	}

	return &apiClient{
		api.NewLister[model.VMProduct](c, service, basePathPattern, pathParams),
		api.NewGetter[model.VMProduct](c, service, basePathPattern, pathParams),
	}
}
