package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type PartialUpdater[P, D any] interface {
	UpdatePartial(ctx context.Context, id string, object *P) (result *D, err error)
}

type partialUpdater[P, D any] struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

func (p *partialUpdater[P, D]) buildPathWithId(id string) string {
	return strings.Join([]string{strings.TrimRight(p.basePathPattern, "/"), id}, "/") + "/"
}

func (p *partialUpdater[P, D]) UpdatePartial(ctx context.Context, id string, object *P) (result *D, err error) {
	var decoded D
	v, err := p.h.Call(ctx, &CallRequest{
		Method:      PATCH,
		Service:     p.serviceName,
		Path:        p.buildPathWithId(id),
		PathParams:  p.pathParams,
		Query:       url.Values{},
		Encoder:     NewJSONEncoder(object),
		StatusCheck: IsHTTPStatusOKOrAccepted,
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

func NewPartialUpdater[P, D any](h Caller, service, path string, pathParams map[string]string) PartialUpdater[P, D] {
	return &partialUpdater[P, D]{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
	}
}
