package image

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for VM image operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.VMImage], err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.VMImage]
}

// NewClient returns the VM image client eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "vm-images",
	}

	return &client{
		api.NewLister[model.VMImage](c, service, basePathPattern, pathParams),
	}
}
