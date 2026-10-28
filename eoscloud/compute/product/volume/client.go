package volume

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for volume product operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.VolumeProduct], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.VolumeProduct, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.Lister[model.VolumeProduct]
	api.Getter[model.VolumeProduct]
}

// NewClient returns the volume product client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "volume-products",
	}

	return &apiClient{
		api.NewLister[model.VolumeProduct](c, service, basePathPattern, pathParams),
		api.NewGetter[model.VolumeProduct](c, service, basePathPattern, pathParams),
	}
}
