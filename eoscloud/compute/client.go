package compute

import (
	"context"
	"strings"

	"go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/acl"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/acl/rule"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/block"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/district"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/forward"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/ip"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/loadbalancer"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/network"
	peeringGroups "go.eoscloud.io/go-eoscloud/eoscloud/compute/peering/groups"
	ipProduct "go.eoscloud.io/go-eoscloud/eoscloud/compute/product/ip"
	loadBalancerProduct "go.eoscloud.io/go-eoscloud/eoscloud/compute/product/loadbalancer"
	vmProduct "go.eoscloud.io/go-eoscloud/eoscloud/compute/product/vm"
	volumeProduct "go.eoscloud.io/go-eoscloud/eoscloud/compute/product/volume"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/vm"
	vmGroups "go.eoscloud.io/go-eoscloud/eoscloud/compute/vm/groups"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/vm/image"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/vm/snapshot"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/vm/template"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/volume"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

const serviceName = "compute"

// Client represents the compute service client
type Client struct {
	Network             network.APIClient
	PeeringGroup        peeringGroups.APIClient
	ACL                 acl.APIClient
	ACLRule             rule.APIClient
	PublicIP            ip.APIClient
	Forward             forward.APIClient
	LoadBalancer        loadbalancer.APIClient
	VMProduct           vmProduct.APIClient
	VolumeProduct       volumeProduct.APIClient
	PublicIPProduct     ipProduct.APIClient
	LoadBalancerProduct loadBalancerProduct.APIClient
	Block               block.APIClient
	District            district.APIClient
	Volume              volume.APIClient
	VM                  vm.APIClient
	VMImage             image.APIClient
	VMGroup             vmGroups.APIClient
	VMSnapshot          snapshot.APIClient
	VMTemplate          template.APIClient
	statsFetcher        api.NoInputExecutor[model.Statistics]
}

func (c *Client) Stats(ctx context.Context) (result *model.Statistics, err error) {
	return c.statsFetcher.Execute(ctx, nil)
}

// NewClient returns a compute service client eoscloud
func NewClient(c api.Caller, baseGlobalPathPattern, baseRealmPathPattern string) *Client {
	return &Client{
		Network:             network.NewClient(c, serviceName, baseRealmPathPattern),
		PeeringGroup:        peeringGroups.NewClient(c, serviceName, baseRealmPathPattern),
		ACL:                 acl.NewClient(c, serviceName, baseRealmPathPattern),
		ACLRule:             rule.NewClient(c, serviceName, baseRealmPathPattern),
		PublicIP:            ip.NewClient(c, serviceName, baseRealmPathPattern),
		Forward:             forward.NewClient(c, serviceName, baseRealmPathPattern),
		LoadBalancer:        loadbalancer.NewClient(c, serviceName, baseRealmPathPattern),
		VM:                  vm.NewClient(c, serviceName, baseRealmPathPattern),
		VMProduct:           vmProduct.NewClient(c, serviceName, baseGlobalPathPattern),
		VolumeProduct:       volumeProduct.NewClient(c, serviceName, baseGlobalPathPattern),
		PublicIPProduct:     ipProduct.NewClient(c, serviceName, baseGlobalPathPattern),
		LoadBalancerProduct: loadBalancerProduct.NewClient(c, serviceName, baseGlobalPathPattern),
		VMImage:             image.NewClient(c, serviceName, baseGlobalPathPattern),
		VMGroup:             vmGroups.NewClient(c, serviceName, baseRealmPathPattern),
		VMSnapshot:          snapshot.NewClient(c, serviceName, baseRealmPathPattern),
		Block:               block.NewClient(c, serviceName, baseGlobalPathPattern),
		District:            district.NewClient(c, serviceName, baseGlobalPathPattern),
		VMTemplate:          template.NewClient(c, serviceName, baseRealmPathPattern),
		Volume:              volume.NewClient(c, serviceName, baseRealmPathPattern),
		statsFetcher: api.NewNoInputExecutor[model.Statistics](
			c,
			serviceName,
			strings.Replace(baseRealmPathPattern, "{kind}", "", 1),
			map[string]string{
				"version": "v1",
			},
			"stats",
			api.GET,
			api.IsHTTPStatusOK,
		),
	}
}
