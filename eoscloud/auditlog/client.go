// Package auditlog is the go-eoscloud client for the auditlog service's public
// read API (GET /v1/auditlogs). The auditlog service is GLOBAL, not
// realm-scoped: entries are organization-scoped and optionally narrowed with a
// realm_id filter, so the client is wired with the global path pattern
// (/{version}/{kind}/), not the realm pattern.
package auditlog

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// IdentityType identifies the kind of caller that triggered an audit log
// Entry.
// Values: "user" (IdentityTypeUser), "client" (IdentityTypeClient)
type IdentityType string

const (
	IdentityTypeUser   IdentityType = "user"
	IdentityTypeClient IdentityType = "client"
)

// Entry is a single audit-log entry as returned by GET /v1/auditlogs. Only the
// fields consumers currently need are modelled; the service returns more
// (data, etc.) which are ignored on decode.
//
// Note: the service intentionally does NOT serialize the event status (the model
// field is `json:"-"`), so there is no Status field here — a recorded entry with
// a matching type IS the record of the (successful) action.
//
// The Identity* fields are optional: the service omits them when the caller
// lacks permission to see identity information for the entry.
type Entry struct {
	ID             string        `json:"id,omitempty" yaml:"id,omitempty"`
	Timestamp      string        `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	Service        string        `json:"service,omitempty" yaml:"service,omitempty"`
	ResourceID     string        `json:"resource_id,omitempty" yaml:"resource_id,omitempty"`
	ResourceKind   string        `json:"resource_kind,omitempty" yaml:"resource_kind,omitempty"`
	ResourceName   string        `json:"resource_name,omitempty" yaml:"resource_name,omitempty"`
	Type           string        `json:"type,omitempty" yaml:"type,omitempty"`
	RealmID        string        `json:"realm_id,omitempty" yaml:"realm_id,omitempty"`
	OrganizationID string        `json:"organization_id,omitempty" yaml:"organization_id,omitempty"`
	IdentityID     string        `json:"identity_id,omitempty" yaml:"identity_id,omitempty"`
	IdentityName   string        `json:"identity_name,omitempty" yaml:"identity_name,omitempty"`
	IdentityEmail  string        `json:"identity_email,omitempty" yaml:"identity_email,omitempty"`
	IdentityType   *IdentityType `json:"identity_type,omitempty" yaml:"identity_type,omitempty"`
}

// APIClient defines the interface for auditlog read operations.
//
// Supported filters (url.Values) on List include: resource_id, resource_kind,
// type, service, realm_id, timestamp__gte, timestamp__lte, _page, _page_size,
// _ordering. The list envelope uses the standard "count"/"results" shape, so
// both result.Count and result.Results are populated.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[Entry], err error)
	Get(ctx context.Context, id string) (result *Entry, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[Entry]
	h               api.Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

// Get retrieves a single audit log entry by id. It does not use api.Getter
// because that appends a trailing slash after the id, which this service does
// not register (see the auditlogPathPattern comment above).
func (c *client) Get(ctx context.Context, id string) (*Entry, error) {
	var decoded Entry
	v, err := c.h.Call(ctx, &api.CallRequest{
		Method:      api.GET,
		Service:     c.serviceName,
		Path:        strings.TrimRight(c.basePathPattern, "/") + "/" + id,
		PathParams:  c.pathParams,
		Query:       url.Values{},
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
	out, ok := v.(*Entry)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", v)
	}
	return out, nil
}

// auditlogPathPattern is the auditlog list path. Unlike the compute/Django
// services, the auditlog service (eos Go framework) registers its collection
// WITHOUT a trailing slash — GET https://auditlog.<domain>/v1/auditlogs — and a
// trailing slash is served the SPA fallback (HTML), so we must not use the
// shared "/{version}/{kind}/" global pattern here.
const auditlogPathPattern = "/{version}/{kind}"

// newEntryClient returns the auditlog entries client (GET /v1/auditlogs). The
// auditlog service owns its path pattern (no trailing slash), so — unlike the
// compute sub-resource clients — it does not take one from the caller.
func newEntryClient(c api.Caller) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "auditlogs",
	}
	return &client{
		Lister:          api.NewLister[Entry](c, "auditlog", auditlogPathPattern, pathParams),
		h:               c,
		serviceName:     "auditlog",
		basePathPattern: auditlogPathPattern,
		pathParams:      pathParams,
	}
}

// Client represents the auditlog service client.
type Client struct {
	AuditLog     APIClient
	Identity     IdentityAPIClient
	Service      ServiceAPIClient
	ResourceKind ResourceKindAPIClient
	Resource     ResourceAPIClient
	ActionType   ActionTypeAPIClient
}

// NewClient returns an auditlog service client. It reuses the same
// authenticated api client as the rest of the SDK, so no separate auth is
// needed.
func NewClient(c api.Caller) *Client {
	return &Client{
		AuditLog:     newEntryClient(c),
		Identity:     NewIdentityClient(c),
		Service:      NewServiceClient(c),
		ResourceKind: NewResourceKindClient(c),
		Resource:     NewResourceClient(c),
		ActionType:   NewActionTypeClient(c),
	}
}
