package model

import (
	"encoding/json"
	"fmt"
	"time"

	"go.yaml.in/yaml/v3"
)

// Port forward models (tag: forwards)

type PortForward struct {
	Self             string                  `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID               string                  `json:"id,omitempty" yaml:"id,omitempty"`
	Network          string                  `json:"network_id" yaml:"network_id"`
	Name             string                  `json:"name,omitempty" yaml:"name,omitempty"`
	ListenIP         string                  `json:"listener_id" yaml:"listener_id"`
	Protocol         PortForwardProtocolEnum `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	SourcePorts      []string                `json:"source_ports,omitempty" yaml:"source_ports,omitempty"`
	DestinationPorts []string                `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty"`
	Target           *PortForwardTargetValue `json:"target" yaml:"target"`
	Description      string                  `json:"description,omitempty" yaml:"description,omitempty"`
	Status           StatusEnum              `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt        time.Time               `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt        time.Time               `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions          []string                `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type PortForwardRequest struct {
	Network          string                  `json:"network_id" yaml:"network_id" shortflag:"N"`
	Name             string                  `json:"name,omitempty" yaml:"name,omitempty"`
	ListenIP         string                  `json:"listener_id" yaml:"listener_id" shortflag:"L"`
	Protocol         PortForwardProtocolEnum `json:"protocol,omitempty" yaml:"protocol,omitempty" shortflag:"P"`
	SourcePorts      []string                `json:"source_ports,omitempty" yaml:"source_ports,omitempty"`
	DestinationPorts []string                `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty"`
	Target           *PortForwardTargetValue `json:"target,omitempty" yaml:"target,omitempty" shortflag:"T"`
	Description      string                  `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

type PatchedPortForwardRequest struct {
	Name             string                  `json:"name,omitempty" yaml:"name,omitempty"`
	ListenIP         string                  `json:"listener_id,omitempty" yaml:"listener_id,omitempty" shortflag:"L"`
	Protocol         PortForwardProtocolEnum `json:"protocol,omitempty" yaml:"protocol,omitempty" shortflag:"P"`
	SourcePorts      []string                `json:"source_ports,omitempty" yaml:"source_ports,omitempty"`
	DestinationPorts []string                `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty"`
	Target           *PortForwardTargetValue `json:"target,omitempty" yaml:"target,omitempty" shortflag:"T"`
	Description      string                  `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

// Validate checks the target of the request. The clients do not call it;
// callers validate before sending when they want errors before the round trip.
func (r PortForwardRequest) Validate() error {
	return validateTarget(r.Target)
}

// Validate checks the target of the request. The clients do not call it;
// callers validate before sending when they want errors before the round trip.
func (r PatchedPortForwardRequest) Validate() error {
	return validateTarget(r.Target)
}

func validateTarget(t *PortForwardTargetValue) error {
	if t == nil {
		return nil
	}
	if err := t.Validate(); err != nil {
		return fmt.Errorf("target: %w", err)
	}
	return nil
}

// PortForwardTargetType is the discriminator of a PortForwardTarget.
// Values: ip, vm
type PortForwardTargetType string

const (
	PortForwardTargetTypeIP PortForwardTargetType = "ip"
	PortForwardTargetTypeVM PortForwardTargetType = "vm"
)

// PortForwardTarget is where a forward sends traffic: PortForwardTargetIP or
// PortForwardTargetVM. Group targets are not supported by the API.
type PortForwardTarget interface {
	TargetType() PortForwardTargetType
	// Validate reports a target the API would reject: a missing payload
	// field, or a Type that is set but does not match the Go type.
	Validate() error
}

// The Type field mirrors the spec's discriminator. Use the constructors to
// set it; MarshalJSON pins it to the right constant either way.
type PortForwardTargetIP struct {
	Type PortForwardTargetType `json:"type" yaml:"type"`
	IP   string                `json:"ip" yaml:"ip"`
}

type PortForwardTargetVM struct {
	Type PortForwardTargetType `json:"type" yaml:"type"`
	ID   string                `json:"id" yaml:"id"`
}

func NewPortForwardTargetIP(ip string) PortForwardTargetIP {
	return PortForwardTargetIP{Type: PortForwardTargetTypeIP, IP: ip}
}

func NewPortForwardTargetVM(id string) PortForwardTargetVM {
	return PortForwardTargetVM{Type: PortForwardTargetTypeVM, ID: id}
}

func (PortForwardTargetIP) TargetType() PortForwardTargetType { return PortForwardTargetTypeIP }
func (PortForwardTargetVM) TargetType() PortForwardTargetType { return PortForwardTargetTypeVM }

func (t PortForwardTargetIP) Validate() error {
	if err := checkTargetType(t.Type, PortForwardTargetTypeIP); err != nil {
		return err
	}
	if t.IP == "" {
		return fmt.Errorf("target %q: ip is required", PortForwardTargetTypeIP)
	}
	return nil
}

func (t PortForwardTargetVM) Validate() error {
	if err := checkTargetType(t.Type, PortForwardTargetTypeVM); err != nil {
		return err
	}
	if t.ID == "" {
		return fmt.Errorf("target %q: id is required", PortForwardTargetTypeVM)
	}
	return nil
}

func checkTargetType(got, want PortForwardTargetType) error {
	if got != "" && got != want {
		return fmt.Errorf("target %q: type is %q", want, got)
	}
	return nil
}

// The local alias type avoids recursing into MarshalJSON.
func (t PortForwardTargetIP) MarshalJSON() ([]byte, error) {
	type plain PortForwardTargetIP
	t.Type = PortForwardTargetTypeIP
	return json.Marshal(plain(t))
}

func (t PortForwardTargetVM) MarshalJSON() ([]byte, error) {
	type plain PortForwardTargetVM
	t.Type = PortForwardTargetTypeVM
	return json.Marshal(plain(t))
}

// PortForwardTargetUnknown holds a target whose type this client does not
// know, so reads keep working when the API introduces new kinds. Fields
// carries the object as received, including "type", and is written back
// unchanged.
type PortForwardTargetUnknown struct {
	Fields map[string]any
}

func (u PortForwardTargetUnknown) TargetType() PortForwardTargetType {
	t, _ := u.Fields["type"].(string)
	return PortForwardTargetType(t)
}

func (u PortForwardTargetUnknown) Validate() error {
	if u.TargetType() == "" {
		return fmt.Errorf("target: type is required")
	}
	return nil
}

func (u PortForwardTargetUnknown) MarshalJSON() ([]byte, error) { return json.Marshal(u.Fields) }
func (u PortForwardTargetUnknown) MarshalYAML() (any, error)    { return u.Fields, nil }

// PortForwardTargetValue wraps PortForwardTarget so encoding/json can decode
// into the interface. On request types the field is a pointer with omitempty:
// a nil pointer omits "target", a non-nil wrapper with a nil PortForwardTarget
// sends an explicit null (see NullPortForwardTarget).
type PortForwardTargetValue struct{ PortForwardTarget }

// NewPortForwardTarget wraps a concrete target.
func NewPortForwardTarget(t PortForwardTarget) *PortForwardTargetValue {
	return &PortForwardTargetValue{t}
}

// NullPortForwardTarget returns a value that serializes as "target": null,
// which clears the target on PATCH. A nil pointer would omit the field instead.
func NullPortForwardTarget() *PortForwardTargetValue {
	return &PortForwardTargetValue{}
}

// Validate delegates to the target. An empty wrapper is valid: it is the
// explicit null that clears the target (see NullPortForwardTarget).
func (v PortForwardTargetValue) Validate() error {
	if v.PortForwardTarget == nil {
		return nil
	}
	return v.PortForwardTarget.Validate()
}

func (v PortForwardTargetValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.PortForwardTarget)
}

// MarshalYAML unwraps the embedded target; yaml.v3 would otherwise nest it
// under a "portforwardtarget" key.
func (v PortForwardTargetValue) MarshalYAML() (any, error) {
	return v.PortForwardTarget, nil
}

// decode dispatches on the discriminator and fills v through into, which is
// either json.Unmarshal or yaml.Node.Decode bound to the input.
func (v *PortForwardTargetValue) decode(typ string, into func(dst any) error) error {
	var err error
	switch PortForwardTargetType(typ) {
	case PortForwardTargetTypeIP:
		var s PortForwardTargetIP
		err = into(&s)
		v.PortForwardTarget = s
	case PortForwardTargetTypeVM:
		var s PortForwardTargetVM
		err = into(&s)
		v.PortForwardTarget = s
	default:
		var u PortForwardTargetUnknown
		err = into(&u.Fields)
		v.PortForwardTarget = u
	}
	return err
}

func (v *PortForwardTargetValue) UnmarshalJSON(b []byte) error {
	if isJSONNull(b) {
		v.PortForwardTarget = nil
		return nil
	}
	t, err := peekType(b)
	if err != nil {
		return err
	}
	return v.decode(t, func(dst any) error { return json.Unmarshal(b, dst) })
}

func (v *PortForwardTargetValue) UnmarshalYAML(node *yaml.Node) error {
	if isYAMLNull(node) {
		v.PortForwardTarget = nil
		return nil
	}
	t, err := peekYAMLType(node)
	if err != nil {
		return err
	}
	return v.decode(t, node.Decode)
}
