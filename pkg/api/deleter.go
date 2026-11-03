package api

import (
	"context"
	"net/url"
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

func (d *deleter) buildPathWithId(id string) (string, error) {
	return joinIDPath(d.basePathPattern, id)
}

func (d *deleter) Delete(ctx context.Context, id string) error {
	path, pathErr := d.buildPathWithId(id)
	if pathErr != nil {
		return pathErr
	}

	_, callErr := d.h.Call(ctx, &CallRequest{
		Method:      DELETE,
		Service:     d.serviceName,
		Path:        path,
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
