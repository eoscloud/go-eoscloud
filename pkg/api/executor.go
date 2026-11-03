package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type Executor[I, O any] interface {
	Execute(ctx context.Context, id *string, input *I) (output *O, err error)
}

type executor[I, O any] struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
	method          Method
	action          string
	statusCheck     HTTPStatusCheck
}

func (e *executor[I, O]) buildPathWithId(id *string) (string, error) {
	if id == nil {
		return strings.TrimRight(e.basePathPattern, "/") + "/:" + e.action, nil
	}
	p, err := joinIDPath(e.basePathPattern, *id)
	if err != nil {
		return "", err
	}
	return p + ":" + e.action, nil
}

func (e *executor[I, O]) Execute(ctx context.Context, id *string, input *I) (output *O, err error) {
	path, pathErr := e.buildPathWithId(id)
	if pathErr != nil {
		return nil, pathErr
	}
	var decoded O
	v, err := e.h.Call(ctx, &CallRequest{
		Method:      e.method,
		Service:     e.serviceName,
		Path:        path,
		PathParams:  e.pathParams,
		Query:       url.Values{},
		Encoder:     NewJSONEncoder[I](input),
		StatusCheck: e.statusCheck,
		Decoder:     NewJSONDecoder(&decoded),
	})
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	out, ok := v.(*O)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", v)
	}

	return out, nil
}

func NewExecutor[I, O any](
	h Caller,
	service, path string,
	pathParams map[string]string,
	action string,
	method Method,
	statusCheck HTTPStatusCheck,
) Executor[I, O] {

	return &executor[I, O]{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
		method:          method,
		action:          action,
		statusCheck:     statusCheck,
	}
}

type NoResultExecutor[I any] interface {
	Execute(ctx context.Context, id *string, input *I) (err error)
}

type noResultExecutor[I any] struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
	method          Method
	action          string
	statusCheck     HTTPStatusCheck
}

func (e *noResultExecutor[I]) buildPathWithId(id *string) (string, error) {
	if id == nil {
		return strings.TrimRight(e.basePathPattern, "/") + "/:" + e.action, nil
	}
	p, err := joinIDPath(e.basePathPattern, *id)
	if err != nil {
		return "", err
	}
	return p + ":" + e.action, nil
}

func (e *noResultExecutor[I]) Execute(ctx context.Context, id *string, input *I) (err error) {
	path, pathErr := e.buildPathWithId(id)
	if pathErr != nil {
		return pathErr
	}
	v, err := e.h.Call(ctx, &CallRequest{
		Method:      e.method,
		Service:     e.serviceName,
		Path:        path,
		PathParams:  e.pathParams,
		Query:       url.Values{},
		Encoder:     NewJSONEncoder[I](input),
		StatusCheck: e.statusCheck,
		Decoder:     nil,
	})
	if err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	return fmt.Errorf("unexpected response: %v", v)
}

func NewNoResultExecutor[I any](
	h Caller,
	service, path string,
	pathParams map[string]string,
	action string,
	method Method,
	statusCheck HTTPStatusCheck,
) NoResultExecutor[I] {

	return &noResultExecutor[I]{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
		method:          method,
		action:          action,
		statusCheck:     statusCheck,
	}
}

type NoInputExecutor[O any] interface {
	Execute(ctx context.Context, id *string) (o *O, err error)
}

type noInputExecutor[O any] struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
	method          Method
	action          string
	statusCheck     HTTPStatusCheck
}

func (e *noInputExecutor[O]) buildPathWithId(id *string) (string, error) {
	if id == nil {
		return strings.TrimRight(e.basePathPattern, "/") + "/:" + e.action, nil
	}
	p, err := joinIDPath(e.basePathPattern, *id)
	if err != nil {
		return "", err
	}
	return p + ":" + e.action, nil
}

func (e *noInputExecutor[O]) Execute(ctx context.Context, id *string) (o *O, err error) {
	path, pathErr := e.buildPathWithId(id)
	if pathErr != nil {
		return nil, pathErr
	}
	var decoded O
	v, err := e.h.Call(ctx, &CallRequest{
		Method:      e.method,
		Service:     e.serviceName,
		Path:        path,
		PathParams:  e.pathParams,
		Query:       url.Values{},
		Encoder:     NoBodyEncoder(),
		StatusCheck: e.statusCheck,
		Decoder:     NewJSONDecoder(&decoded),
	})
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	out, ok := v.(*O)
	if !ok {
		return nil, fmt.Errorf("unexpected type: %T", v)
	}

	return out, nil
}

func NewNoInputExecutor[O any](
	h Caller,
	service, path string,
	pathParams map[string]string,
	action string,
	method Method,
	statusCheck HTTPStatusCheck,
) NoInputExecutor[O] {

	return &noInputExecutor[O]{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
		method:          method,
		action:          action,
		statusCheck:     statusCheck,
	}
}

type NoIOExecutor interface {
	Execute(ctx context.Context, id *string) (err error)
}
type noIOExecutor struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
	method          Method
	action          string
	statusCheck     HTTPStatusCheck
}

func (e *noIOExecutor) buildPathWithId(id *string) (string, error) {
	if id == nil {
		return strings.TrimRight(e.basePathPattern, "/") + "/:" + e.action, nil
	}
	p, err := joinIDPath(e.basePathPattern, *id)
	if err != nil {
		return "", err
	}
	return p + ":" + e.action, nil
}

func (e *noIOExecutor) Execute(ctx context.Context, id *string) (err error) {
	path, pathErr := e.buildPathWithId(id)
	if pathErr != nil {
		return pathErr
	}
	_, callErr := e.h.Call(ctx, &CallRequest{
		Method:      e.method,
		Service:     e.serviceName,
		Path:        path,
		PathParams:  e.pathParams,
		Query:       url.Values{},
		Encoder:     NoBodyEncoder(),
		StatusCheck: e.statusCheck,
		Decoder:     NoResultDecoder(),
	})
	if callErr != nil {
		return callErr
	}
	return nil
}

func NewNoIOExecutor(
	h Caller,
	service, path string,
	pathParams map[string]string,
	action string,
	method Method,
	statusCheck HTTPStatusCheck,
) NoIOExecutor {

	return &noIOExecutor{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
		method:          method,
		action:          action,
		statusCheck:     statusCheck,
	}
}
