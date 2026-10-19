package api

import (
	"context"
	"fmt"
	"net/url"
)

type Creator[C, D any] interface {
	Create(ctx context.Context, object *C) (result *D, err error)
}

type creator[C, D any] struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

func (c *creator[C, D]) Create(ctx context.Context, object *C) (result *D, err error) {
	var decoded D
	v, err := c.h.Call(ctx, &CallRequest{
		Method:      POST,
		Service:     c.serviceName,
		Path:        c.basePathPattern,
		PathParams:  c.pathParams,
		Query:       url.Values{},
		Encoder:     NewJSONEncoder[C](object),
		StatusCheck: IsHTTPStatusCreated,
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
		return result, fmt.Errorf("unexpected create decode type: %T", v)
	}

	return out, nil
}

func NewCreator[C, D any](h Caller, service, path string, pathParams map[string]string) Creator[C, D] {
	return &creator[C, D]{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
	}
}
