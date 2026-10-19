package ip

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for public IP operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.PublicIP], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.PublicIP, err error)
	Create(ctx context.Context, object *model.PublicIPRequest) (result *model.PublicIP, err error)
	Update(ctx context.Context, id string, object *model.PatchedPublicIPRequest) (result *model.PublicIP, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedPublicIPRequest) (result *model.PublicIP, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)

	Attach(ctx context.Context, id *string, input *model.PublicIPAttachRequest) (output *model.PublicIP, err error)
	Detach(ctx context.Context, id *string) (output *model.PublicIP, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)

	// Deprecated: Use Attach instead
	AttachVM(ctx context.Context, id *string, input *model.PublicIPAttachRequest) (output *model.PublicIP, err error)
	// Deprecated: Use Detach instead
	DetachVM(ctx context.Context, id *string) (output *model.PublicIP, err error)
}

var _ APIClient = (*apiClient)(nil)

// apiClient implements the public IP component of the compute API eoscloud
type apiClient struct {
	api.Lister[model.PublicIP]
	api.Creator[model.PublicIPRequest, model.PublicIP]
	api.Getter[model.PublicIP]
	api.Updater[model.PatchedPublicIPRequest, model.PublicIP]
	api.PartialUpdater[model.PatchedPublicIPRequest, model.PublicIP]
	api.Deleter
	api.ServerSideActionsFetcher
	attach api.Executor[model.PublicIPAttachRequest, model.PublicIP]
	detach api.NoInputExecutor[model.PublicIP]
	stats  api.NoInputExecutor[commonmodel.Statistics]

	// Deprecated: Use attach instead
	attachVM api.Executor[model.PublicIPAttachRequest, model.PublicIP]
	// Deprecated: Use detach instead
	detachVM api.NoInputExecutor[model.PublicIP]
}

// Attach attaches a public IP to a VM.
func (c *apiClient) Attach(ctx context.Context, id *string, input *model.PublicIPAttachRequest) (*model.PublicIP, error) {
	return c.attach.Execute(ctx, id, input)
}

// Detach detaches a public IP from a VM.
func (c *apiClient) Detach(ctx context.Context, id *string) (*model.PublicIP, error) {
	return c.detach.Execute(ctx, id)
}

// Stats returns the statistics of a public IP.
func (c *apiClient) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// AttachVM attaches a public IP to a VM.
// Deprecated: Use Attach instead
func (c *apiClient) AttachVM(ctx context.Context, id *string, input *model.PublicIPAttachRequest) (*model.PublicIP, error) {
	return c.attachVM.Execute(ctx, id, input)
}

// DetachVM detaches a public IP from a VM.
// Deprecated: Use Detach instead
func (c *apiClient) DetachVM(ctx context.Context, id *string) (*model.PublicIP, error) {
	return c.detachVM.Execute(ctx, id)
}

// NewClient returns the public IP service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "public-ips",
	}

	return &apiClient{
		Lister: api.NewLister[model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		Creator: api.NewCreator[model.PublicIPRequest, model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		Getter: api.NewGetter[model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		Updater: api.NewUpdater[model.PatchedPublicIPRequest, model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		PartialUpdater: api.NewPartialUpdater[model.PatchedPublicIPRequest, model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		Deleter: api.NewDeleter(
			c,
			service,
			basePathPattern,
			pathParams,
		),
		ServerSideActionsFetcher: api.NewServerSideActionsFetcher(
			c,
			service,
			basePathPattern,
			pathParams,
		),
		attach: api.NewExecutor[model.PublicIPAttachRequest, model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
			"attach",
			api.POST,
			api.IsHTTPStatusOK,
		),
		detach: api.NewNoInputExecutor[model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
			"detach",
			api.POST,
			api.IsHTTPStatusOK,
		),
		attachVM: api.NewExecutor[model.PublicIPAttachRequest, model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
			"attach-vm",
			api.POST,
			api.IsHTTPStatusOK,
		),
		detachVM: api.NewNoInputExecutor[model.PublicIP](
			c,
			service,
			basePathPattern,
			pathParams,
			"detach-vm",
			api.POST,
			api.IsHTTPStatusOK,
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
