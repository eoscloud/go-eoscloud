package api

import (
	"context"
	"fmt"
	"net/url"
)

type Updater[U, D any] interface {
	Update(ctx context.Context, id string, object *U) (result *D, err error)
}

type updater[U, D any] struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

func (u *updater[U, D]) buildPathWithId(id string) (string, error) {
	return joinIDPath(u.basePathPattern, id)
}

func (u *updater[U, D]) Update(ctx context.Context, id string, object *U) (result *D, err error) {
	path, pathErr := u.buildPathWithId(id)
	if pathErr != nil {
		return nil, pathErr
	}
	var decoded D
	v, err := u.h.Call(ctx, &CallRequest{
		Method:      PUT,
		Service:     u.serviceName,
		Path:        path,
		PathParams:  u.pathParams,
		Query:       url.Values{},
		Encoder:     NewJSONEncoder[U](object),
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

func NewUpdater[U, D any](h Caller, service, path string, pathParams map[string]string) Updater[U, D] {
	return &updater[U, D]{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
	}
}
