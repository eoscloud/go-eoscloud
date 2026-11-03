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

`NewClient` takes one or more options, typically an authentication option. Realm and
organization are read from the access token.

```go
client := eoscloud.NewClient(/* auth option */)
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
client := eoscloud.NewClient(userAgent, /* auth option */)
```

### Authentication

#### Client Credentials Flow

Server-to-server authentication using a client ID and secret. No user interaction required.
Access tokens are kept in memory only and requested again when they expire.

```go
client := eoscloud.NewClient(
    api.WithClientCredentialsFlow(
        ctx,
        clientID,
        clientSecret,
        auth.WithClientCredentialsAuthBaseURL(auth.DefaultAuthBaseURL), // optional
    ),
)
```

#### Device Flow

For devices without a browser or with limited input capabilities. Displays a URL and user code — the user authenticates on another device.

```go
client := eoscloud.NewClient(
    api.WithDeviceFlow(
        ctx,
        auth.WithDeviceFlowAuthBaseURL(auth.DefaultAuthBaseURL), // optional
        auth.WithDeviceFlowClientID("your-client-id"), // optional
        auth.WithDeviceFlowKeyringServiceName(auth.DefaultKeyringServiceName), // optional
        auth.WithDeviceFlowPromptWriter(os.Stderr), // optional, default os.Stderr
        auth.WithDeviceFlowBrowserOpener(nil), // optional, nil disables opening a browser
    ),
)
```

The library never writes to stdout. Prompts go to the prompt writer (default stderr), so the
device flow is safe to use from stdio-based tools such as MCP servers. A browser is only opened
for a verification URL on the configured auth host.

#### Token storage

Device flow tokens (including the refresh token) are stored in the OS keyring (macOS Keychain,
Windows Credential Manager, Secret Service on Linux). If no keyring is available, they are
stored **unencrypted** as JSON in `<user config dir>/eosctl/tokens` (see `os.UserConfigDir`, e.g.
`~/.config/eosctl/tokens` on Linux),
in a directory with mode `0700` and files with mode `0600`. The library refuses to use a token
file that is a symlink, not owned by the current user or has a different mode. If neither a
keyring nor a home/config directory is available, token storage fails with
`auth.ErrNoTokenDir`; set a directory explicitly with `auth.WithDeviceFlowFallbackTokenStoreDir`.

### Additional Options

```go
// Custom base URL. Must use https; http is only accepted for loopback hosts
// (localhost, 127.0.0.0/8, ::1), e.g. for local development.
api.WithBaseURL("https://{service}.example.com")

// Custom base HTTP client. Its transport and timeout are kept; authentication is
// layered on top regardless of option order. Redirects are only followed to the
// same scheme and host.
api.WithHTTPClient(myHTTPClient)

// Debugging only: prints a fingerprint of the access token and selected claims
// (eos, exp, iss, sub) to stderr. Do not enable in production.
api.WithDebugTokenPrinting()
```

TLS verification cannot be disabled through library options.

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
