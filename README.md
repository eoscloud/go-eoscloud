# go-eoscloud

[![CI](https://github.com/eoscloud/go-eoscloud/actions/workflows/ci.yaml/badge.svg)](https://github.com/eoscloud/go-eoscloud/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/go.eoscloud.io/go-eoscloud.svg)](https://pkg.go.dev/go.eoscloud.io/go-eoscloud)
[![Release](https://img.shields.io/github/v/release/eoscloud/go-eoscloud)](https://github.com/eoscloud/go-eoscloud/releases/latest)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

Go client library for the [eos cloud](https://eoscloud.io) API.

## Installation

```bash
go get go.eoscloud.io/go-eoscloud
```

Requires Go 1.27 or later.

## Usage

```go
import (
    "go.eoscloud.io/go-eoscloud/eoscloud"
    "go.eoscloud.io/go-eoscloud/pkg/api"
    "go.eoscloud.io/go-eoscloud/pkg/auth"
)
```

### Creating a Client

`NewClient` takes a realm ID and one or more authentication options:

```go
client := eoscloud.NewClient(realmID, /* auth option */)
```

#### Identifying your application

Please tell the library the name and version of your application. They are reported in
the `User-Agent` sent to the API, e.g. `go-eoscloud/0.4.0 my-app/1.2.3 (go1.27; linux/amd64)`,
so usage can be attributed to clients and versions. Neither may be empty, the name must not
contain spaces or `/`, the version must not contain spaces.

```go
userAgent, err := api.WithUserAgent("my-app", "1.2.3")
if err != nil {
    // invalid name or version
}
client := eoscloud.NewClient(realmID, userAgent, /* auth option */)
```

### Authentication

#### Client Credentials Flow

Server-to-server authentication using a client ID and secret. No user interaction required. Tokens are cached in the system keyring and refreshed automatically.

```go
client := eoscloud.NewClient(
    realmID,
    api.WithClientCredentialsFlow(
        ctx,
        clientID,
        clientSecret,
        organizationID,
        realmID,
        auth.WithClientCredentialsAuthBaseURL("https://auth.eoscloud.dev"), // optional
        auth.WithClientCredentialsKeyringServiceName(auth.DefaultKeyringServiceName), // optional
    ),
)
```

#### Device Flow

For devices without a browser or with limited input capabilities. Displays a URL and user code — the user authenticates on another device.

```go
client := eoscloud.NewClient(
    realmID,
    api.WithDeviceFlow(
        ctx,
        organizationID,
        realmID,
        auth.WithDeviceFlowBaseURL("https://auth.eoscloud.dev"), // optional
        auth.WithDeviceFlowClientID("your-client-id"), // optional
        auth.WithDeviceFlowKeyringServiceName(auth.DefaultKeyringServiceName), // optional
    ),
)
```

### Additional Options

```go
// Custom base URL (e.g. for local development)
api.WithBaseURL("http://localhost:8000")

// Custom HTTP client
api.WithHTTPClient(myHTTPClient)
```

## Compute Resources

Once you have a client, all compute resources are available under `client.Compute`:

```go
ctx := context.Background()
params := url.Values{}

// Virtual Machines
vms, err := client.Compute.VM.List(ctx, params)
vm, err := client.Compute.VM.Get(ctx, vmID, params)

// VM Console (WebSocket)
readErrCh, writeErrCh, err := client.Compute.VM.Console(ctx, &vmID, &model.VMConsoleRequestRequest{Force: true}, rwc, http.Header{})

// VM Images, Snapshots, Templates
images, err := client.Compute.VMImage.List(ctx, params)
snapshots, err := client.Compute.VMSnapshot.List(ctx, params)
templates, err := client.Compute.VMTemplate.List(ctx, params)

// Networks
networks, err := client.Compute.Network.List(ctx, params)

// Block Storage & Volumes
blocks, err := client.Compute.Block.List(ctx, params)
volumes, err := client.Compute.Volume.List(ctx, params)
volume, err := client.Compute.Volume.Get(ctx, volumeID, params)

// ACLs
acls, err := client.Compute.ACL.List(ctx, params)
rules, err := client.Compute.ACLRule.List(ctx, params)

// Public IPs & Port Forwarding
pips, err := client.Compute.PublicIP.List(ctx, params)
forwards, err := client.Compute.Forward.List(ctx, params)

// Products
vmProducts, err := client.Compute.VMProduct.List(ctx, params)
volumeProducts, err := client.Compute.VolumeProduct.List(ctx, params)
ipProducts, err := client.Compute.PublicIPProduct.List(ctx, params)
lbProducts, err := client.Compute.LoadBalancerProduct.List(ctx, params)
```

### Error Handling

Use `api.UnexpectedHttpStatusError` to inspect HTTP error responses:

```go
import "errors"

vms, err := client.Compute.VM.List(ctx, url.Values{})
if err != nil {
    var httpErr *api.UnexpectedHttpStatusError
    if errors.As(err, &httpErr) {
        fmt.Printf("HTTP error: %s\n", httpErr)
        return
    }
    panic(err)
}
```

## Contributing

Bug reports, feature requests and pull requests are welcome, see
[CONTRIBUTING.md](CONTRIBUTING.md). Please report security issues privately as
described in [SECURITY.md](SECURITY.md).

## License

Apache License 2.0, see [LICENSE](LICENSE) and [NOTICE](NOTICE).
