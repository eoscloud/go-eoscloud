package billingaccount

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go.eoscloud.io/go-eoscloud/eoscloud/billing/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// GetCreditInfoOptions are the parameters for GetCreditInfo. Exactly the
// milli-credits or micro-credits amount to price must be provided; the
// service accepts either and combines them (MilliCredits*1000 + MicroCredits).
type GetCreditInfoOptions struct {
	ProductID    string
	MilliCredits uint64
	MicroCredits uint64
}

// APIClient defines the interface for billing account operations.
type APIClient interface {
	Get(ctx context.Context, id string, filter url.Values) (result *model.BillingAccountItem, err error)
	Create(ctx context.Context, object *model.CreateBillingAccountRequest) (result *model.BillingAccountItem, err error)
	UpdatePartial(ctx context.Context, id string, object *model.UpdateBillingAccountRequest) (result *model.BillingAccountItem, err error)
	GetCreditInfo(ctx context.Context, id string, opts GetCreditInfoOptions) (result *model.CreditInfo, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Getter[model.BillingAccountItem]
	api.Creator[model.CreateBillingAccountRequest, model.BillingAccountItem]
	api.PartialUpdater[model.UpdateBillingAccountRequest, model.BillingAccountItem]
	h               api.Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

// GetCreditInfo fetches the price information for a given amount of credits.
// It is not a plain Getter call because it hits the "/:credit" sub-action
// with query parameters.
func (c *client) GetCreditInfo(ctx context.Context, id string, opts GetCreditInfoOptions) (*model.CreditInfo, error) {
	query := url.Values{}
	query.Set("product_id", opts.ProductID)
	if opts.MilliCredits > 0 {
		query.Set("milli_credits", strconv.FormatUint(opts.MilliCredits, 10))
	}
	if opts.MicroCredits > 0 {
		query.Set("micro_credits", strconv.FormatUint(opts.MicroCredits, 10))
	}

	var decoded model.CreditInfo
	v, err := c.h.Call(ctx, &api.CallRequest{
		Method:      api.GET,
		Service:     c.serviceName,
		Path:        strings.TrimRight(c.basePathPattern, "/") + "/" + id + "/:credit",
		PathParams:  c.pathParams,
		Query:       query,
		Encoder:     api.NoBodyEncoder(),
		StatusCheck: api.IsHTTPStatusOK,
		Decoder:     api.NewJSONDecoder(&decoded),
	})
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	out, ok := v.(*model.CreditInfo)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", v)
	}
	return out, nil
}

// NewClient returns the billing account client of the billing service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "billingaccount",
	}

	return &client{
		Getter:          api.NewGetter[model.BillingAccountItem](c, service, basePathPattern, pathParams),
		Creator:         api.NewCreator[model.CreateBillingAccountRequest, model.BillingAccountItem](c, service, basePathPattern, pathParams),
		PartialUpdater:  api.NewPartialUpdater[model.UpdateBillingAccountRequest, model.BillingAccountItem](c, service, basePathPattern, pathParams),
		h:               c,
		serviceName:     service,
		basePathPattern: basePathPattern,
		pathParams:      pathParams,
	}
}
