package model

// Shared enums and types used across multiple models

// StatusEnum represents resource lifecycle status
// Values: creating, ready, destroying, updating
// Kept as string to match OpenAPI enum

type StatusEnum string

const (
	StatusCreating   StatusEnum = "creating"
	StatusReady      StatusEnum = "ready"
	StatusDestroying StatusEnum = "destroying"
	StatusUpdating   StatusEnum = "updating"
)

// ActionEnum represents ACL rule action
// Values: allow, drop

type ActionEnum string

const (
	ActionAllow ActionEnum = "allow"
	ActionDrop  ActionEnum = "drop"
)

// ACLRuleProtocolEnum represents protocol for ACL rules
// Values: tcp, udp, icmp4, icmp6, any

type ACLRuleProtocolEnum string

const (
	ProtocolTCP   ACLRuleProtocolEnum = "tcp"
	ProtocolUDP   ACLRuleProtocolEnum = "udp"
	ProtocolICMP4 ACLRuleProtocolEnum = "icmp4"
	ProtocolICMP6 ACLRuleProtocolEnum = "icmp6"
	ProtocolAny   ACLRuleProtocolEnum = "any"
)

// PortForwardProtocolEnum represents protocol for port forwarding
// Values: tcp, udp, any

type PortForwardProtocolEnum string

const (
	PortForwardProtocolTCP PortForwardProtocolEnum = "tcp"
	PortForwardProtocolUDP PortForwardProtocolEnum = "udp"
	PortForwardProtocolAny PortForwardProtocolEnum = "any"
)

// IpFamilyEnum is 4 or 6 (integer enum in OpenAPI)
// Values: 4, 6

type IpFamilyEnum int

const (
	IpFamily4 IpFamilyEnum = 4
	IpFamily6 IpFamilyEnum = 6
)

// LoadBalancerProtocolEnum represents protocol for load balancers
// Values: tcp, udp

type LoadBalancerProtocolEnum string

const (
	LoadBalancerProtocolTCP LoadBalancerProtocolEnum = "tcp"
	LoadBalancerProtocolUDP LoadBalancerProtocolEnum = "udp"
)

// PowerStateEnum represents VM power state
// Values: started, stopped, starting, stopping, rebooting

type PowerStateEnum string

const (
	PowerStateStarted   PowerStateEnum = "started"
	PowerStateStopped   PowerStateEnum = "stopped"
	PowerStateStarting  PowerStateEnum = "starting"
	PowerStateStopping  PowerStateEnum = "stopping"
	PowerStateRebooting PowerStateEnum = "rebooting"
)

// Additional enums from compute-api.yaml

// ArchitectureEnum represents CPU architecture
type ArchitectureEnum string

const (
	ArchitectureAmd64 ArchitectureEnum = "amd64"
	ArchitectureArm64 ArchitectureEnum = "arm64"
)

// OptimizationEnum represents product/block optimization
type OptimizationEnum string

const (
	OptimizationGeneral OptimizationEnum = "d"
	OptimizationGPU     OptimizationEnum = "g"
	OptimizationArm     OptimizationEnum = "a"
	OptimizationCompute OptimizationEnum = "c"
	OptimizationMemory  OptimizationEnum = "m"
)
