package clientcredential

import (
	"context"
	"net/url"

	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for client-credential operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.ClientCredential], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.ClientCredential, err error)
	Create(ctx context.Context, object *model.ClientCredentialRequest) (result *model.ClientCredential, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedClientCredentialRequest) (result *model.ClientCredential, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.ClientCredential]
	api.Getter[model.ClientCredential]
	api.Creator[model.ClientCredentialRequest, model.ClientCredential]
	api.PartialUpdater[model.PatchedClientCredentialRequest, model.ClientCredential]
	api.Deleter
	api.ServerSideActionsFetcher
}

// NewClient returns the client-credential client of the organization service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "client-credentials",
	}

	return &client{
		api.NewLister[model.ClientCredential](c, service, basePathPattern, pathParams),
		api.NewGetter[model.ClientCredential](c, service, basePathPattern, pathParams),
		api.NewCreator[model.ClientCredentialRequest, model.ClientCredential](c, service, basePathPattern, pathParams),
		api.NewPartialUpdater[model.PatchedClientCredentialRequest, model.ClientCredential](c, service, basePathPattern, pathParams),
		api.NewDeleter(c, service, basePathPattern, pathParams),
		api.NewServerSideActionsFetcher(c, service, basePathPattern, pathParams),
	}
}
