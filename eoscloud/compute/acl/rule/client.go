package rule

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for ACL rule operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.ACLRuleList], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.ACLRule, err error)
	Create(ctx context.Context, object *model.ACLRuleRequest) (result *model.ACLRule, err error)
	Update(ctx context.Context, id string, object *model.ACLRuleRequest) (result *model.ACLRule, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedACLRuleRequest) (result *model.ACLRule, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.CRUDClient[
		model.ACLRuleList,
		model.ACLRuleRequest,
		model.ACLRuleRequest,
		model.PatchedACLRuleRequest,
		model.ACLRule,
	]
	stats api.NoInputExecutor[commonmodel.Statistics]
}

// Stats returns the statistics of an ACL rule.
func (c *client) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// NewClient returns the ACL rules client from eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "acl-rules",
	}

	return &client{
		CRUDClient: api.NewCRUDClient[
			model.ACLRuleList,
			model.ACLRuleRequest,
			model.ACLRuleRequest,
			model.PatchedACLRuleRequest,
			model.ACLRule,
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
