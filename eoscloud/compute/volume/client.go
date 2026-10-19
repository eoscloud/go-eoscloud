package volume

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for volume operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.ListVolume], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.VolumeDetail, err error)
	Create(ctx context.Context, object *model.VolumeRequest) (result *model.VolumeDetail, err error)
	Update(ctx context.Context, id string, object *model.PatchedVolumeRequest) (result *model.VolumeDetail, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedVolumeRequest) (result *model.VolumeDetail, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Attach(ctx context.Context, id *string, input *model.VolumeAttachRequest) (output *model.VolumeDetail, err error)
	Detach(ctx context.Context, id *string) (output *model.VolumeDetail, err error)
	Grow(ctx context.Context, id *string, object *model.GrowVolumeRequest) (result *model.VolumeDetail, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)

	// Deprecated: Use Attach instead
	AttachVolume(ctx context.Context, id *string, input *model.VolumeAttachRequest) (output *model.VolumeDetail, err error)
	// Deprecated: Use Detach instead
	DetachVolume(ctx context.Context, id *string) (output *model.VolumeDetail, err error)
	// Deprecated: Use Grow instead
	GrowVolume(ctx context.Context, id *string, object *model.GrowVolumeRequest) (result *model.VolumeDetail, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.CRUDClient[
		model.ListVolume,
		model.VolumeRequest,
		model.PatchedVolumeRequest,
		model.PatchedVolumeRequest,
		model.VolumeDetail,
	]
	attach api.Executor[model.VolumeAttachRequest, model.VolumeDetail]
	detach api.NoInputExecutor[model.VolumeDetail]
	grow   api.Executor[model.GrowVolumeRequest, model.VolumeDetail]
	stats  api.NoInputExecutor[commonmodel.Statistics]
}

// Attach attaches a volume to a VM.
func (c *client) Attach(ctx context.Context, id *string, input *model.VolumeAttachRequest) (*model.VolumeDetail, error) {
	return c.attach.Execute(ctx, id, input)
}

// Detach detaches a volume from a VM.
func (c *client) Detach(ctx context.Context, id *string) (*model.VolumeDetail, error) {
	return c.detach.Execute(ctx, id)
}

// Grow grows a volume.
func (c *client) Grow(ctx context.Context, id *string, object *model.GrowVolumeRequest) (result *model.VolumeDetail, err error) {
	return c.grow.Execute(ctx, id, object)
}

// Stats returns the statistics of a volume.
func (c *client) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// AttachVolume attaches a volume to a VM.
// Deprecated: Use Attach instead
func (c *client) AttachVolume(ctx context.Context, id *string, input *model.VolumeAttachRequest) (output *model.VolumeDetail, err error) {
	return c.attach.Execute(ctx, id, input)
}

// DetachVolume detaches a volume from a VM.
// Deprecated: Use Detach instead
func (c *client) DetachVolume(ctx context.Context, id *string) (output *model.VolumeDetail, err error) {
	return c.detach.Execute(ctx, id)
}

// GrowVolume grows a volume.
// Deprecated: Use Grow instead
func (c *client) GrowVolume(ctx context.Context, id *string, object *model.GrowVolumeRequest) (result *model.VolumeDetail, err error) {
	return c.grow.Execute(ctx, id, object)
}

// NewClient returns the VM service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "volumes",
	}

	return &client{
		CRUDClient: api.NewCRUDClient[
			model.ListVolume,
			model.VolumeRequest,
			model.PatchedVolumeRequest,
			model.PatchedVolumeRequest,
			model.VolumeDetail,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		attach: api.NewExecutor[model.VolumeAttachRequest, model.VolumeDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"attach",
			api.POST,
			api.IsHTTPStatusOK,
		),
		detach: api.NewNoInputExecutor[model.VolumeDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"detach",
			api.POST,
			api.IsHTTPStatusOK,
		),
		grow: api.NewExecutor[model.GrowVolumeRequest, model.VolumeDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"grow",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		stats: api.NewNoInputExecutor[commonmodel.Statistics](
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
