package product

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for product operations.
//
// This endpoint is deprecated in the API; use the bucket product client (product/bucket) instead.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Product], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Product, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.Lister[model.Product]
	api.Getter[model.Product]
}

// NewClient returns the legacy product client eoscloud of the object storage service
//
// This endpoint is deprecated in the API; use bucket.NewClient from product/bucket instead.
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "products",
	}

	return &apiClient{
		api.NewLister[model.Product](c, service, basePathPattern, pathParams),
		api.NewGetter[model.Product](c, service, basePathPattern, pathParams),
	}
}
