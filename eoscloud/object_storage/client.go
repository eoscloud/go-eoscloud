package objectstorage

import (
	"context"
	"strings"

	"go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/bucket"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/district"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/product"
	bucketProduct "go.eoscloud.io/go-eoscloud/eoscloud/object_storage/product/bucket"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/user"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

const serviceName = "object-storage"

// Client represents the object storage service client
type Client struct {
	// Deprecated: use BucketProduct.
	Product       product.APIClient
	BucketProduct bucketProduct.APIClient
	District      district.APIClient
	User          user.APIClient
	Bucket        bucket.APIClient
	statsFetcher  api.NoInputExecutor[model.Statistics]
}

func (c *Client) Stats(ctx context.Context) (result *model.Statistics, err error) {
	return c.statsFetcher.Execute(ctx, nil)
}

// NewClient returns an object storage service client eoscloud
func NewClient(c api.Caller, baseGlobalPathPattern, baseRealmPathPattern string) *Client {
	return &Client{
		Product:       product.NewClient(c, serviceName, baseGlobalPathPattern),
		BucketProduct: bucketProduct.NewClient(c, serviceName, baseGlobalPathPattern),
		District:      district.NewClient(c, serviceName, baseGlobalPathPattern),
		User:          user.NewClient(c, serviceName, baseRealmPathPattern),
		Bucket:        bucket.NewClient(c, serviceName, baseRealmPathPattern),
		statsFetcher: api.NewNoInputExecutor[model.Statistics](
			c,
			serviceName,
			strings.Replace(baseRealmPathPattern, "{kind}", "", 1),
			map[string]string{
				"version": "v1",
			},
			"stats",
			api.GET,
			api.IsHTTPStatusOK,
		),
	}
}
