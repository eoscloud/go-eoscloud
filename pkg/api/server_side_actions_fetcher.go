package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.eoscloud.io/go-eoscloud/eoscloud/common/model"
)

type ServerSideActionsFetcher interface {
	Actions(ctx context.Context, id string) (result *model.ObjectOptionsResponse, err error)
}

type serverSideActionsFetcher struct {
	h               Caller
	serviceName     string
	basePathPattern string
	pathParams      map[string]string
}

func (a *serverSideActionsFetcher) buildPathWithId(id string) string {
	return strings.Join([]string{strings.TrimRight(a.basePathPattern, "/"), id}, "/") + "/"
}

func (a *serverSideActionsFetcher) Actions(ctx context.Context, id string) (result *model.ObjectOptionsResponse, err error) {
	var decoded model.ObjectOptionsResponse
	v, err := a.h.Call(ctx, &CallRequest{
		Method:      OPTIONS,
		Service:     a.serviceName,
		Path:        a.buildPathWithId(id),
		PathParams:  a.pathParams,
		Query:       url.Values{},
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

	out, ok := v.(*model.ObjectOptionsResponse)
	if !ok {
		return result, fmt.Errorf("unexpected type: %T", v)
	}

	return out, nil
}

func NewServerSideActionsFetcher(h Caller, service, path string, pathParams map[string]string) ServerSideActionsFetcher {
	return &serverSideActionsFetcher{
		h:               h,
		serviceName:     service,
		basePathPattern: path,
		pathParams:      pathParams,
	}
}
