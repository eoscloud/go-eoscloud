package model

import "time"

// Load balancer models (tag: load-balancers)

type LoadBalancer struct {
	Self             string                   `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID               string                   `json:"id,omitempty" yaml:"id,omitempty"`
	Name             string                   `json:"name,omitempty" yaml:"name,omitempty"`
	Listener         string                   `json:"listener,omitempty" yaml:"listener,omitempty"`
	Protocol         LoadBalancerProtocolEnum `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	SourcePorts      []string                 `json:"source_ports,omitempty" yaml:"source_ports,omitempty"`
	DestinationPorts []string                 `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty"`
	DistrictID       string                   `json:"district_id,omitempty" yaml:"district_id,omitempty"`
	TargetGroupID    string                   `json:"target_group_id,omitempty" yaml:"target_group_id,omitempty"`
	ProductID        string                   `json:"product_id,omitempty" yaml:"product_id,omitempty"`
	Description      string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Status           StatusEnum               `json:"status,omitempty" yaml:"status,omitempty"`
	Health           LoadBalancerHealthEnum   `json:"health,omitempty" yaml:"health,omitempty"`
	CreatedAt        time.Time                `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt        time.Time                `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions          []string                 `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type LoadBalancerRequest struct {
	Name             string                   `json:"name,omitempty" yaml:"name,omitempty"`
	Protocol         LoadBalancerProtocolEnum `json:"protocol" yaml:"protocol" shortflag:"P"`
	SourcePorts      []string                 `json:"source_ports" yaml:"source_ports" shortflag:"S"`
	DestinationPorts []string                 `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty" shortflag:"D"`
	DistrictID       string                   `json:"district_id" yaml:"district_id" shortflag:"d"`
	TargetGroupID    string                   `json:"target_group_id" yaml:"target_group_id" shortflag:"T"`
	IpFamily         IpFamilyEnum             `json:"ip_family" yaml:"ip_family"`
	Description      string                   `json:"description,omitempty" yaml:"description,omitempty"`
}

type PatchedLoadBalancerRequest struct {
	Name             string                   `json:"name,omitempty" yaml:"name,omitempty"`
	Protocol         LoadBalancerProtocolEnum `json:"protocol,omitempty" yaml:"protocol,omitempty" shortflag:"P"`
	SourcePorts      []string                 `json:"source_ports,omitempty" yaml:"source_ports,omitempty" shortflag:"S"`
	DestinationPorts []string                 `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty" shortflag:"D"`
	Description      string                   `json:"description,omitempty" yaml:"description,omitempty"`
}
