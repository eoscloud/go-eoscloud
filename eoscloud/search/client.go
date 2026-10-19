// Package search is the go-eoscloud client for the search service's public
// API (POST /v1/search). The search service is GLOBAL, not realm-scoped.
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// Hit is a single matching resource as returned by the search service. Its
// shape depends on the indexed resource, so fields are kept as raw JSON.
type Hit map[string]json.RawMessage

// Query captures the supported search parameters.
type Query struct {
	Q string `json:"q"`
	// Filter combines attribute filters with AND/OR, e.g.
	// "(service=compute AND realm_id=my-realm-id) OR kind=realm". Filterable
	// attributes are: realm_id, service, kind, uuid.
	Filter string `json:"filter,omitempty"`
	// Fields lists the fields to display in the returned documents.
	Fields []string `json:"_fields,omitempty"`
	Page   uint64   `json:"_page,omitempty"`
	Size   uint64   `json:"_page_size,omitempty"`
}

// Response is the result of a search query.
type Response struct {
	Hits               []Hit  `json:"hits"`
	EstimatedTotalHits int64  `json:"estimated_total_hits,omitempty"`
	Offset             int64  `json:"offset,omitempty"`
	Limit              int64  `json:"limit,omitempty"`
	Query              string `json:"query"`
}

// APIClient defines the interface for search operations.
type APIClient interface {
	Search(ctx context.Context, query *Query) (result *Response, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	h               api.Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

// searchPathPattern is the search path. Like the auditlog service (eos Go
// framework), the search service registers its collection WITHOUT a trailing
// slash, so we must not use the shared "/{version}/{kind}/" global pattern.
const searchPathPattern = "/{version}/{kind}"

// Search performs a search query. The endpoint answers 200 OK with the
// matching results (it is a query encoded as a POST body, not a resource
// creation), so it cannot use api.Creator, which expects 201 Created.
func (c *client) Search(ctx context.Context, query *Query) (*Response, error) {
	var decoded Response
	v, err := c.h.Call(ctx, &api.CallRequest{
		Method:      api.POST,
		Service:     c.serviceName,
		Path:        c.basePathPattern,
		PathParams:  c.pathParams,
		Query:       url.Values{},
		Encoder:     api.NewJSONEncoder[Query](query),
		StatusCheck: api.IsHTTPStatusOK,
		Decoder:     api.NewJSONDecoder(&decoded),
	})
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	out, ok := v.(*Response)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", v)
	}
	return out, nil
}

// NewClient returns the search service client (POST /v1/search).
func NewClient(c api.Caller) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "search",
	}
	return &client{
		h:               c,
		serviceName:     "search",
		basePathPattern: searchPathPattern,
		pathParams:      pathParams,
	}
}
