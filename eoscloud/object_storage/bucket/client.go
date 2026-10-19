package bucket

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for ACL operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Bucket], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Bucket, err error)
	Create(ctx context.Context, object *model.BucketRequest) (result *model.Bucket, err error)
	Update(ctx context.Context, id string, object *model.PatchedBucketRequest) (result *model.Bucket, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedBucketRequest) (result *model.Bucket, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)

	Grow(ctx context.Context, id *string, object *model.BucketGrowRequest) (*model.Bucket, error)
	RecoverPolicy(ctx context.Context, id *string) (*model.Bucket, error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.CRUDClient[
		model.Bucket,
		model.BucketRequest,
		model.PatchedBucketRequest,
		model.PatchedBucketRequest,
		model.Bucket,
	]
	grow          api.Executor[model.BucketGrowRequest, model.Bucket]
	recoverPolicy api.NoInputExecutor[model.Bucket]
}

func (c *client) Grow(ctx context.Context, id *string, object *model.BucketGrowRequest) (*model.Bucket, error) {
	return c.grow.Execute(ctx, id, object)
}

func (c *client) RecoverPolicy(ctx context.Context, id *string) (*model.Bucket, error) {
	return c.recoverPolicy.Execute(ctx, id)
}

// NewClient returns the ACL service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "buckets",
	}
	return &client{
		CRUDClient: api.NewCRUDClient[
			model.Bucket,
			model.BucketRequest,
			model.PatchedBucketRequest,
			model.PatchedBucketRequest,
			model.Bucket,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		grow: api.NewExecutor[model.BucketGrowRequest, model.Bucket](
			c,
			service,
			basePathPattern,
			pathParams,
			"grow",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		recoverPolicy: api.NewNoInputExecutor[model.Bucket](
			c,
			service,
			basePathPattern,
			pathParams,
			"recover-policy",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
	}
}
