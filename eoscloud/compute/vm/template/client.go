package template

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for VM template operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Template], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Template, err error)
	Create(ctx context.Context, object *model.TemplateRequest) (result *model.Template, err error)
	Update(ctx context.Context, id string, object *model.PatchedTemplateRequest) (result *model.Template, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedTemplateRequest) (result *model.Template, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)
}

var _ APIClient = (*apiClient)(nil)

type apiClient struct {
	api.CRUDClient[
		model.Template,
		model.TemplateRequest,
		model.PatchedTemplateRequest,
		model.PatchedTemplateRequest,
		model.Template,
	]
	stats api.NoInputExecutor[commonmodel.Statistics]
}

// Stats returns the statistics of a VM template.
func (c *apiClient) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// NewClient returns the VM templates client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "templates",
	}

	return &apiClient{
		CRUDClient: api.NewCRUDClient[
			model.Template,
			model.TemplateRequest,
			model.PatchedTemplateRequest,
			model.PatchedTemplateRequest,
			model.Template,
		](
			c,
			service,
			basePathPattern,
			pathParams,
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
