package bucket

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for bucket product operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.BucketProduct], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.BucketProduct, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.Lister[model.BucketProduct]
	api.Getter[model.BucketProduct]
}

// NewClient returns the bucket product client eoscloud of the object storage service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "bucket-products",
	}

	return &apiClient{
		api.NewLister[model.BucketProduct](c, service, basePathPattern, pathParams),
		api.NewGetter[model.BucketProduct](c, service, basePathPattern, pathParams),
	}
}
