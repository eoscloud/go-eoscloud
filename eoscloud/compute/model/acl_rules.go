package model

import (
	"encoding/json"
	"fmt"
	"time"

	"go.yaml.in/yaml/v3"
)

// ACL Rule models (tag: acl-rules)

type ACLRule struct {
	Self             string              `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID               string              `json:"id,omitempty" yaml:"id,omitempty"`
	ACL              string              `json:"acl_id" yaml:"acl_id"`
	Name             string              `json:"name,omitempty" yaml:"name,omitempty"`
	Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
	Status           StatusEnum          `json:"status,omitempty" yaml:"status,omitempty"`
	Action           ActionEnum          `json:"action,omitempty" yaml:"action,omitempty"`
	Protocol         ACLRuleProtocolEnum `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Sources          []ACLSelectorValue  `json:"sources,omitempty" yaml:"sources,omitempty"`
	Destinations     []ACLSelectorValue  `json:"destinations,omitempty" yaml:"destinations,omitempty"`
	SourcePorts      []string            `json:"source_ports,omitempty" yaml:"source_ports,omitempty"`
	DestinationPorts []string            `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty"`
	ICMPType         string              `json:"icmp_type,omitempty" yaml:"icmp_type,omitempty"`
	ICMPCode         string              `json:"icmp_code,omitempty" yaml:"icmp_code,omitempty"`
	CreatedAt        time.Time           `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt        time.Time           `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions          []string            `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

// ACLRuleList is an alias for ACLRule used for list views
// The API returns the same fields for list and detail; keep a distinct type
// to satisfy generic L parameter while reusing the underlying structure.
type ACLRuleList = ACLRule

type ACLRuleRequest struct {
	ACL         string              `json:"acl_id" yaml:"acl_id"`
	Name        string              `json:"name,omitempty" yaml:"name,omitempty"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	Action      ActionEnum          `json:"action,omitempty" yaml:"action,omitempty"`
	Protocol    ACLRuleProtocolEnum `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	// Sources, Destinations, SourcePorts and DestinationPorts are pointers so that
	// an explicitly empty list ([]) can be distinguished from an unset field: a nil
	// pointer is omitted from the JSON payload (leaves the value untouched on PATCH),
	// while a pointer to an empty slice serializes as [] (clears the value).
	Sources          *[]ACLSelectorValue `json:"sources,omitempty" yaml:"sources,omitempty"`
	Destinations     *[]ACLSelectorValue `json:"destinations,omitempty" yaml:"destinations,omitempty"`
	SourcePorts      *[]string           `json:"source_ports,omitempty" yaml:"source_ports,omitempty"`
	DestinationPorts *[]string           `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty"`
	ICMPType         string              `json:"icmp_type,omitempty" yaml:"icmp_type,omitempty"`
	ICMPCode         string              `json:"icmp_code,omitempty" yaml:"icmp_code,omitempty"`
}

type PatchedACLRuleRequest struct {
	Name             string              `json:"name,omitempty" yaml:"name,omitempty"`
	Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
	Action           ActionEnum          `json:"action,omitempty" yaml:"action,omitempty"`
	Protocol         ACLRuleProtocolEnum `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Sources          *[]ACLSelectorValue `json:"sources,omitempty" yaml:"sources,omitempty"`
	Destinations     *[]ACLSelectorValue `json:"destinations,omitempty" yaml:"destinations,omitempty"`
	SourcePorts      *[]string           `json:"source_ports,omitempty" yaml:"source_ports,omitempty"`
	DestinationPorts *[]string           `json:"destination_ports,omitempty" yaml:"destination_ports,omitempty"`
	ICMPType         string              `json:"icmp_type,omitempty" yaml:"icmp_type,omitempty"`
	ICMPCode         string              `json:"icmp_code,omitempty" yaml:"icmp_code,omitempty"`
}

// Validate checks the selectors of the request. The clients do not call it;
// callers validate before sending when they want errors before the round trip.
func (r ACLRuleRequest) Validate() error {
	return validateSelectors(r.Sources, r.Destinations)
}

// Validate checks the selectors of the request. The clients do not call it;
// callers validate before sending when they want errors before the round trip.
func (r PatchedACLRuleRequest) Validate() error {
	return validateSelectors(r.Sources, r.Destinations)
}

func validateSelectors(sources, destinations *[]ACLSelectorValue) error {
	lists := []struct {
		name string
		list *[]ACLSelectorValue
	}{{"sources", sources}, {"destinations", destinations}}
	for _, l := range lists {
		if l.list == nil {
			continue
		}
		for i, sel := range *l.list {
			if err := sel.Validate(); err != nil {
				return fmt.Errorf("%s[%d]: %w", l.name, i, err)
			}
		}
	}
	return nil
}

// ACLSelectorType is the discriminator of an ACLSelector.
// Values: ip, vm, group
type ACLSelectorType string

const (
	ACLSelectorTypeIP    ACLSelectorType = "ip"
	ACLSelectorTypeVM    ACLSelectorType = "vm"
	ACLSelectorTypeGroup ACLSelectorType = "group"
)

// ACLSelector selects the traffic peers of an ACL rule, used for both sources
// and destinations. It is one of ACLSelectorIP, ACLSelectorVM or
// ACLSelectorGroup (spec: ACLEndpoint).
type ACLSelector interface {
	SelectorType() ACLSelectorType
	// Validate reports a selector the API would reject: a missing payload
	// field, or a Type that is set but does not match the Go type.
	Validate() error
}

// The Type field mirrors the spec's discriminator. Use the constructors to
// set it; MarshalJSON pins it to the right constant either way.
type ACLSelectorIP struct {
	Type ACLSelectorType `json:"type" yaml:"type"`
	IP   string          `json:"ip" yaml:"ip"`
}

type ACLSelectorVM struct {
	Type ACLSelectorType `json:"type" yaml:"type"`
	ID   string          `json:"id" yaml:"id"`
}

type ACLSelectorGroup struct {
	Type ACLSelectorType `json:"type" yaml:"type"`
	ID   string          `json:"id" yaml:"id"`
}

func NewACLSelectorIP(ip string) ACLSelectorIP {
	return ACLSelectorIP{Type: ACLSelectorTypeIP, IP: ip}
}

func NewACLSelectorVM(id string) ACLSelectorVM {
	return ACLSelectorVM{Type: ACLSelectorTypeVM, ID: id}
}

func NewACLSelectorGroup(id string) ACLSelectorGroup {
	return ACLSelectorGroup{Type: ACLSelectorTypeGroup, ID: id}
}

func (ACLSelectorIP) SelectorType() ACLSelectorType    { return ACLSelectorTypeIP }
func (ACLSelectorVM) SelectorType() ACLSelectorType    { return ACLSelectorTypeVM }
func (ACLSelectorGroup) SelectorType() ACLSelectorType { return ACLSelectorTypeGroup }

func (s ACLSelectorIP) Validate() error {
	if err := checkSelectorType(s.Type, ACLSelectorTypeIP); err != nil {
		return err
	}
	if s.IP == "" {
		return fmt.Errorf("selector %q: ip is required", ACLSelectorTypeIP)
	}
	return nil
}

func (s ACLSelectorVM) Validate() error {
	if err := checkSelectorType(s.Type, ACLSelectorTypeVM); err != nil {
		return err
	}
	if s.ID == "" {
		return fmt.Errorf("selector %q: id is required", ACLSelectorTypeVM)
	}
	return nil
}

func (s ACLSelectorGroup) Validate() error {
	if err := checkSelectorType(s.Type, ACLSelectorTypeGroup); err != nil {
		return err
	}
	if s.ID == "" {
		return fmt.Errorf("selector %q: id is required", ACLSelectorTypeGroup)
	}
	return nil
}

func checkSelectorType(got, want ACLSelectorType) error {
	if got != "" && got != want {
		return fmt.Errorf("selector %q: type is %q", want, got)
	}
	return nil
}

// The local alias type avoids recursing into MarshalJSON.
func (s ACLSelectorIP) MarshalJSON() ([]byte, error) {
	type plain ACLSelectorIP
	s.Type = ACLSelectorTypeIP
	return json.Marshal(plain(s))
}

func (s ACLSelectorVM) MarshalJSON() ([]byte, error) {
	type plain ACLSelectorVM
	s.Type = ACLSelectorTypeVM
	return json.Marshal(plain(s))
}

func (s ACLSelectorGroup) MarshalJSON() ([]byte, error) {
	type plain ACLSelectorGroup
	s.Type = ACLSelectorTypeGroup
	return json.Marshal(plain(s))
}

// ACLSelectorUnknown holds a selector whose type this client does not know,
// so reads keep working when the API introduces new kinds. Fields carries the
// object as received, including "type", and is written back unchanged.
type ACLSelectorUnknown struct {
	Fields map[string]any
}

func (u ACLSelectorUnknown) SelectorType() ACLSelectorType {
	t, _ := u.Fields["type"].(string)
	return ACLSelectorType(t)
}

func (u ACLSelectorUnknown) Validate() error {
	if u.SelectorType() == "" {
		return fmt.Errorf("selector: type is required")
	}
	return nil
}

func (u ACLSelectorUnknown) MarshalJSON() ([]byte, error) { return json.Marshal(u.Fields) }
func (u ACLSelectorUnknown) MarshalYAML() (any, error)    { return u.Fields, nil }

// ACLSelectorValue wraps ACLSelector so encoding/json can decode into the
// interface. Use it as the element type of selector slices.
type ACLSelectorValue struct{ ACLSelector }

// NewACLSelector wraps a concrete selector.
func NewACLSelector(s ACLSelector) ACLSelectorValue { return ACLSelectorValue{s} }

// Validate rejects an empty wrapper, which would serialize as null, and
// otherwise delegates to the selector.
func (v ACLSelectorValue) Validate() error {
	if v.ACLSelector == nil {
		return fmt.Errorf("selector is empty")
	}
	return v.ACLSelector.Validate()
}

func (v ACLSelectorValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ACLSelector)
}

// MarshalYAML unwraps the embedded selector; yaml.v3 would otherwise nest it
// under an "aclselector" key.
func (v ACLSelectorValue) MarshalYAML() (any, error) {
	return v.ACLSelector, nil
}

// decode dispatches on the discriminator and fills v through into, which is
// either json.Unmarshal or yaml.Node.Decode bound to the input.
func (v *ACLSelectorValue) decode(typ string, into func(dst any) error) error {
	var err error
	switch ACLSelectorType(typ) {
	case ACLSelectorTypeIP:
		var s ACLSelectorIP
		err = into(&s)
		v.ACLSelector = s
	case ACLSelectorTypeVM:
		var s ACLSelectorVM
		err = into(&s)
		v.ACLSelector = s
	case ACLSelectorTypeGroup:
		var s ACLSelectorGroup
		err = into(&s)
		v.ACLSelector = s
	default:
		var u ACLSelectorUnknown
		err = into(&u.Fields)
		v.ACLSelector = u
	}
	return err
}

func (v *ACLSelectorValue) UnmarshalJSON(b []byte) error {
	if isJSONNull(b) {
		v.ACLSelector = nil
		return nil
	}
	t, err := peekType(b)
	if err != nil {
		return err
	}
	return v.decode(t, func(dst any) error { return json.Unmarshal(b, dst) })
}

func (v *ACLSelectorValue) UnmarshalYAML(node *yaml.Node) error {
	if isYAMLNull(node) {
		v.ACLSelector = nil
		return nil
	}
	t, err := peekYAMLType(node)
	if err != nil {
		return err
	}
	return v.decode(t, node.Decode)
}
