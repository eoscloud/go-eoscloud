package ip

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for public IP product operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.PublicIPProduct], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.PublicIPProduct, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.Lister[model.PublicIPProduct]
	api.Getter[model.PublicIPProduct]
}

// NewClient returns the public IP product client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "public-ip-products",
	}

	return &apiClient{
		api.NewLister[model.PublicIPProduct](c, service, basePathPattern, pathParams),
		api.NewGetter[model.PublicIPProduct](c, service, basePathPattern, pathParams),
	}
}
