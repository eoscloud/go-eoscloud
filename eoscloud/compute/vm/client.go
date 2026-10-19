package vm

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/coder/websocket"
	commonmodel "go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
	"go.eoscloud.io/go-eoscloud/pkg/ws"
)

// APIClient defines the interface for VM operations.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.ListVM], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.VMDetail, err error)
	Create(ctx context.Context, object *model.VMRequest) (result *model.VMDetail, err error)
	Update(ctx context.Context, id string, object *model.PatchedVMRequest) (result *model.VMDetail, err error)
	UpdatePartial(ctx context.Context, id string, object *model.PatchedVMRequest) (result *model.VMDetail, err error)
	Delete(ctx context.Context, id string) (err error)
	Actions(ctx context.Context, id string) (result *commonmodel.ObjectOptionsResponse, err error)
	PowerOn(ctx context.Context, id *string, input *model.VMPowerActionRequest) (output *model.VMDetail, err error)
	PowerOff(ctx context.Context, id *string, input *model.VMPowerActionRequest) (output *model.VMDetail, err error)
	ForcePowerOff(ctx context.Context, id *string, input *model.VMPowerActionRequest) (output *model.VMDetail, err error)
	Reboot(ctx context.Context, id *string, input *model.VMPowerActionRequest) (output *model.VMDetail, err error)
	Upscale(ctx context.Context, id *string, input *model.VMUpscaleRequest) (output *model.VMDetail, err error)
	ChangeNetwork(ctx context.Context, id *string, input *model.VMChangeNetworkRequest) (output *model.VMDetail, err error)
	Stats(ctx context.Context, id *string) (result *commonmodel.Statistics, err error)
	Console(ctx context.Context, vmID *string, request *model.VMConsoleRequestRequest, rwc io.ReadWriteCloser, httpHeader http.Header) (chan error, chan error, error)
	Exec(ctx context.Context, vmID *string, request *model.VMExecRequest, rwc io.ReadWriteCloser, httpHeader http.Header) (chan error, chan error, error)
}

var _ APIClient = (*client)(nil)

type PowerActionResponse *string

type client struct {
	api.CRUDClient[
		model.ListVM,
		model.VMRequest,
		model.PatchedVMRequest,
		model.PatchedVMRequest,
		model.VMDetail,
	]
	powerOn         api.Executor[model.VMPowerActionRequest, model.VMDetail]
	powerOff        api.Executor[model.VMPowerActionRequest, model.VMDetail]
	forcePowerOff   api.Executor[model.VMPowerActionRequest, model.VMDetail]
	reboot          api.Executor[model.VMPowerActionRequest, model.VMDetail]
	consoleExecutor api.Executor[model.VMConsoleRequestRequest, model.VMConsole]
	upscale         api.Executor[model.VMUpscaleRequest, model.VMDetail]
	changeNetwork   api.Executor[model.VMChangeNetworkRequest, model.VMDetail]
	execExecutor    api.Executor[model.VMExecRequest, model.VMExec]
	stats           api.NoInputExecutor[commonmodel.Statistics]
	userAgent       string
}

func (c *client) PowerOn(ctx context.Context, id *string, input *model.VMPowerActionRequest) (*model.VMDetail, error) {
	return c.powerOn.Execute(ctx, id, input)
}

func (c *client) PowerOff(ctx context.Context, id *string, input *model.VMPowerActionRequest) (*model.VMDetail, error) {
	return c.powerOff.Execute(ctx, id, input)
}

func (c *client) ForcePowerOff(ctx context.Context, id *string, input *model.VMPowerActionRequest) (*model.VMDetail, error) {
	return c.forcePowerOff.Execute(ctx, id, input)
}

func (c *client) Reboot(ctx context.Context, id *string, input *model.VMPowerActionRequest) (*model.VMDetail, error) {
	return c.reboot.Execute(ctx, id, input)
}

func (c *client) Upscale(ctx context.Context, id *string, input *model.VMUpscaleRequest) (*model.VMDetail, error) {
	return c.upscale.Execute(ctx, id, input)
}

func (c *client) ChangeNetwork(ctx context.Context, id *string, input *model.VMChangeNetworkRequest) (*model.VMDetail, error) {
	return c.changeNetwork.Execute(ctx, id, input)
}

func (c *client) Stats(ctx context.Context, id *string) (*commonmodel.Statistics, error) {
	return c.stats.Execute(ctx, id)
}

// withUserAgent returns a copy of h that carries the library's User-Agent
// appended to any User-Agent the caller already set.
func (c *client) withUserAgent(h http.Header) http.Header {
	if c.userAgent == "" {
		return h
	}
	out := h.Clone()
	if out == nil {
		out = http.Header{}
	}
	if existing := out.Get("User-Agent"); existing != "" {
		out.Set("User-Agent", existing+" "+c.userAgent)
	} else {
		out.Set("User-Agent", c.userAgent)
	}
	return out
}

func (c *client) Console(
	ctx context.Context,
	vmID *string,
	request *model.VMConsoleRequestRequest,
	rwc io.ReadWriteCloser,
	httpHeader http.Header,
) (chan error, chan error, error) {
	response, err := c.consoleExecutor.Execute(ctx, vmID, request)
	if err != nil {
		return nil, nil, err
	}

	wsc, _, err := websocket.Dial(ctx, response.ConsoleURL, &websocket.DialOptions{HTTPHeader: c.withUserAgent(httpHeader)})
	if err != nil {
		return nil, nil, err
	}
	readChan, WriteChan := ws.Mirror(ctx, wsc, rwc)
	return readChan, WriteChan, nil
}

func (c *client) Exec(
	ctx context.Context,
	vmID *string,
	request *model.VMExecRequest,
	rwc io.ReadWriteCloser,
	httpHeader http.Header,
) (chan error, chan error, error) {
	response, err := c.execExecutor.Execute(ctx, vmID, request)
	if err != nil {
		return nil, nil, err
	}

	wsc, _, err := websocket.Dial(ctx, response.ConsoleURL, &websocket.DialOptions{HTTPHeader: c.withUserAgent(httpHeader)})
	if err != nil {
		return nil, nil, err
	}
	readChan, WriteChan := ws.Mirror(ctx, wsc, rwc)
	return readChan, WriteChan, nil
}

// NewClient returns the VM service eoscloud of the compute service
func NewClient(
	c api.Caller,
	service, basePathPattern string,
) APIClient {

	pathParams := map[string]string{
		"version": "v1",
		"kind":    "vms",
	}

	return &client{
		userAgent: api.UserAgent(c),
		CRUDClient: api.NewCRUDClient[
			model.ListVM,
			model.VMRequest,
			model.PatchedVMRequest,
			model.PatchedVMRequest,
			model.VMDetail,
		](
			c,
			service,
			basePathPattern,
			pathParams,
		),
		powerOn: api.NewExecutor[model.VMPowerActionRequest, model.VMDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"power-on",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		powerOff: api.NewExecutor[model.VMPowerActionRequest, model.VMDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"power-off",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		forcePowerOff: api.NewExecutor[model.VMPowerActionRequest, model.VMDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"force-power-off",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		reboot: api.NewExecutor[model.VMPowerActionRequest, model.VMDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"reboot",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		consoleExecutor: api.NewExecutor[model.VMConsoleRequestRequest, model.VMConsole](
			c,
			service,
			basePathPattern,
			pathParams,
			"console",
			api.POST,
			api.IsHTTPStatusOK,
		),
		upscale: api.NewExecutor[model.VMUpscaleRequest, model.VMDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"upscale",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		changeNetwork: api.NewExecutor[model.VMChangeNetworkRequest, model.VMDetail](
			c,
			service,
			basePathPattern,
			pathParams,
			"change-network",
			api.POST,
			api.IsHTTPStatusAccepted,
		),
		execExecutor: api.NewExecutor[model.VMExecRequest, model.VMExec](
			c,
			service,
			basePathPattern,
			pathParams,
			"exec",
			api.POST,
			api.IsHTTPStatusOK,
		),
		stats: api.NewNoInputExecutor[commonmodel.Statistics](
			c,
			service,
			basePathPattern,
			pathParams,
			"stats",
			api.GET,
			api.IsHTTPStatusOK,
		),
	}
}
