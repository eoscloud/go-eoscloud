package api

import (
	"context"
	"net/url"
	"strings"
)

type Deleter interface {
	Delete(ctx context.Context, id string) (err error)
}

type deleter struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

func (d *deleter) buildPathWithId(id string) string {
	return strings.Join([]string{strings.TrimRight(d.basePathPattern, "/"), id}, "/") + "/"
}

func (d *deleter) Delete(ctx context.Context, id string) error {

	_, callErr := d.h.Call(ctx, &CallRequest{
		Method:      DELETE,
		Service:     d.serviceName,
		Path:        d.buildPathWithId(id),
		PathParams:  d.pathParams,
		Query:       url.Values{},
		Encoder:     NoBodyEncoder(),
		StatusCheck: IsHTTPStatusNoContent,
		Decoder:     NoResultDecoder(),
	})
	return callErr
}

func NewDeleter(h Caller, service, path string, pathParams map[string]string) Deleter {
	return &deleter{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
	}
}
