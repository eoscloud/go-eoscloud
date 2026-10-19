package auditlog

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// ActionType is a single action type as returned by GET /v1/action-types. It
// is used to populate the type filter on the audit log List endpoint.
type ActionType struct {
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	Service      string `json:"service,omitempty" yaml:"service,omitempty"`
	ResourceKind string `json:"resource_kind,omitempty" yaml:"resource_kind,omitempty"`
}

// ActionTypeAPIClient defines the interface for listing possible action types
// in the audit log (GET /v1/action-types).
//
// Supported filters (url.Values) on List include: service, resource_kind.
type ActionTypeAPIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[ActionType], err error)
}

var _ ActionTypeAPIClient = (*actionTypeClient)(nil)

type actionTypeClient struct {
	api.Lister[ActionType]
}

// NewActionTypeClient returns the auditlog action-type filter client (GET /v1/action-types).
func NewActionTypeClient(c api.Caller) ActionTypeAPIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "action-types",
	}
	return &actionTypeClient{
		api.NewLister[ActionType](c, "auditlog", auditlogPathPattern, pathParams),
	}
}
