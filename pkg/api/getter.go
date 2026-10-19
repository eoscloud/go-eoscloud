package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type Getter[D any] interface {
	Get(ctx context.Context, id string, filter url.Values) (result *D, err error)
}

type getter[D any] struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

func (g *getter[D]) buildPathWithId(id string) string {
	return strings.Join([]string{strings.TrimRight(g.basePathPattern, "/"), id}, "/") + "/"
}

func (g *getter[D]) Get(ctx context.Context, id string, filter url.Values) (result *D, err error) {
	var decoded D
	v, err := g.h.Call(ctx, &CallRequest{
		Method:      GET,
		Service:     g.serviceName,
		Path:        g.buildPathWithId(id),
		PathParams:  g.pathParams,
		Query:       filter,
		Encoder:     NoBodyEncoder(),
		StatusCheck: IsHTTPStatusOK,
		Decoder:     NewJSONDecoder(&decoded),
	})
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	out, ok := v.(*D)
	if !ok {
		return result, fmt.Errorf("unexpected type: %T", v)
	}

	return out, nil
}

func NewGetter[D any](h Caller, service, path string, pathParams map[string]string) Getter[D] {
	return &getter[D]{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
	}
}
