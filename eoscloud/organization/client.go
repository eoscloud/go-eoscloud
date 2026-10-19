package organization

import (
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/clientcredential"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/invitation"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/member"
	orgresource "go.eoscloud.io/go-eoscloud/eoscloud/organization/organization"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/realm"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/role"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

const serviceName = "organization"

// Client represents the organization service client.
type Client struct {
	Organization     orgresource.APIClient
	Realm            realm.APIClient
	Member           member.APIClient
	Role             role.APIClient
	Invitation       invitation.APIClient
	ClientCredential clientcredential.APIClient
}

// NewClient returns an organization service client.
func NewClient(c api.Caller, baseGlobalPathPattern, baseRealmPathPattern string) *Client {
	return &Client{
		Organization:     orgresource.NewClient(c, serviceName, baseGlobalPathPattern),
		Realm:            realm.NewClient(c, serviceName, baseGlobalPathPattern),
		Member:           member.NewClient(c, serviceName, baseGlobalPathPattern),
		Role:             role.NewClient(c, serviceName, baseGlobalPathPattern),
		Invitation:       invitation.NewClient(c, serviceName, baseGlobalPathPattern),
		ClientCredential: clientcredential.NewClient(c, serviceName, baseGlobalPathPattern),
	}
}
