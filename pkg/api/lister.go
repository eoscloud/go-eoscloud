package api

import (
	"context"
	"fmt"
	"net/url"
)

type Lister[L any] interface {
	List(ctx context.Context, filter url.Values) (result *Page[L], err error)
}

type lister[L any] struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

// List retrieves a paged list of the object
func (l *lister[L]) List(ctx context.Context, filter url.Values) (result *Page[L], err error) {
	var decoded Page[L]
	v, callErr := l.h.Call(ctx, &CallRequest{
		Method:      GET,
		Service:     l.serviceName,
		Path:        l.basePathPattern,
		PathParams:  l.pathParams,
		Query:       filter,
		Encoder:     NoBodyEncoder(),
		StatusCheck: IsHTTPStatusOK,
		Decoder:     NewJSONDecoder(&decoded),
	})
	if callErr != nil {
		return nil, callErr
	}

	if v == nil {
		return nil, nil
	}

	out, ok := v.(*Page[L])
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", v)
	}
	return out, nil
}

func NewLister[L any](h Caller, service, path string, pathParams map[string]string) Lister[L] {
	return &lister[L]{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
	}
}
