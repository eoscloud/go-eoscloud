package snapshot

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for snapshot operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Snapshot], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Snapshot, err error)
	Create(ctx context.Context, object *model.SnapshotRequest) (result *model.Snapshot, err error)
	Update(ctx context.Context, id string, object *model.PatchedSnapshotRequest) (result *model.Snapshot, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedSnapshotRequest) (result *model.Snapshot, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Restore(ctx context.Context, id *string) (output *model.Snapshot, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.CRUDClient[
		model.Snapshot,
		model.SnapshotRequest,
		model.PatchedSnapshotRequest,
		model.PatchedSnapshotRequest,
		model.Snapshot,
	]
	restore api.NoInputExecutor[model.Snapshot]
	stats   api.NoInputExecutor[commonmodel.Statistics]
}

// Restore restores a snapshot.
func (c *apiClient) Restore(ctx context.Context, id *string) (*model.Snapshot, error) {
	return c.restore.Execute(ctx, id)
}

// Stats returns the statistics of a snapshot.
func (c *apiClient) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// NewClient returns the VM snapshot client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "snapshots",
	}

	return &apiClient{
		api.NewCRUDClient[
			model.Snapshot,
			model.SnapshotRequest,
			model.PatchedSnapshotRequest,
			model.PatchedSnapshotRequest,
			model.Snapshot,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		api.NewNoInputExecutor[model.Snapshot](
			c,
			service,
			basePathPattern,
			pathParams,
			"restore",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		api.NewNoInputExecutor[commonmodel.Statistics](
			c,
			service,
			basePathPattern,
			pathParams,
			"stats",
			api.GET,
			api.IsHTTPStatusOK,
		),
	}
}
