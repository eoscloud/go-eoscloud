package model_test

import (
	"encoding/json"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
)

func TestACLRuleSelectorsRoundTrip(t *testing.T) {
	in := `{"acl_id":"a","sources":[{"type":"ip","ip":"10.0.0.1"},{"type":"vm","id":"v"}],"destinations":[{"type":"group","id":"g"}]}`

	var rule model.ACLRule
	require.NoError(t, json.Unmarshal([]byte(in), &rule))

	require.Len(t, rule.Sources, 2)
	assert.Equal(t, model.NewACLSelectorIP("10.0.0.1"), rule.Sources[0].ACLSelector)
	assert.Equal(t, model.NewACLSelectorVM("v"), rule.Sources[1].ACLSelector)
	assert.Equal(t, model.NewACLSelectorGroup("g"), rule.Destinations[0].ACLSelector)

	out, err := json.Marshal(model.ACLRuleRequest{ACL: "a", Sources: &rule.Sources, Destinations: &rule.Destinations})
	require.NoError(t, err)
	assert.JSONEq(t, in, string(out))
}

func TestACLSelectorUnknownTypeRoundTrips(t *testing.T) {
	in := `{"type":"subnet","cidr":"10.0.0.0/8","extra":1}`
	var v model.ACLSelectorValue
	require.NoError(t, json.Unmarshal([]byte(in), &v))

	u, ok := v.ACLSelector.(model.ACLSelectorUnknown)
	require.True(t, ok)
	assert.Equal(t, model.ACLSelectorType("subnet"), u.SelectorType())

	out, err := json.Marshal(v)
	require.NoError(t, err)
	assert.JSONEq(t, in, string(out))
}

func TestACLRuleRequestEmptySourcesClears(t *testing.T) {
	empty := []model.ACLSelectorValue{}
	b, err := json.Marshal(model.PatchedACLRuleRequest{Sources: &empty})
	require.NoError(t, err)
	assert.JSONEq(t, `{"sources":[]}`, string(b))

	b, err = json.Marshal(model.PatchedACLRuleRequest{})
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(b))
}

func TestPortForwardTarget(t *testing.T) {
	var fw model.PortForward
	require.NoError(t, json.Unmarshal([]byte(`{"target":{"type":"vm","id":"v"}}`), &fw))
	require.NotNil(t, fw.Target)
	assert.Equal(t, model.NewPortForwardTargetVM("v"), fw.Target.PortForwardTarget)

	fw = model.PortForward{}
	require.NoError(t, json.Unmarshal([]byte(`{"target":null}`), &fw))
	assert.Nil(t, fw.Target)

	// Types this client does not know are kept, not rejected.
	require.NoError(t, json.Unmarshal([]byte(`{"target":{"type":"group","id":"g"}}`), &fw))
	assert.Equal(t, model.PortForwardTargetType("group"), fw.Target.TargetType())

	// A literal without Type still serializes with the discriminator.
	b, err := json.Marshal(model.PortForwardRequest{Target: model.NewPortForwardTarget(model.PortForwardTargetIP{IP: "1.2.3.4"})})
	require.NoError(t, err)
	assert.Contains(t, string(b), `"target":{"type":"ip","ip":"1.2.3.4"}`)
}

func TestSelectorMarshalPinsType(t *testing.T) {
	// A wrong or empty Type on a literal must not leak into the payload.
	b, err := json.Marshal(model.ACLSelectorVM{Type: "ip", ID: "v"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"vm","id":"v"}`, string(b))

	b, err = json.Marshal(model.ACLSelectorGroup{ID: "g"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"group","id":"g"}`, string(b))
}

func TestSelectorYAMLOutputIsFlat(t *testing.T) {
	r := model.ACLRuleRequest{ACL: "a", Sources: &[]model.ACLSelectorValue{model.NewACLSelector(model.NewACLSelectorIP("10.0.0.1"))}}
	b, err := yaml.Marshal(r)
	require.NoError(t, err)
	assert.Equal(t, "acl_id: a\nsources:\n    - type: ip\n      ip: 10.0.0.1\n", string(b))

	f := model.PatchedPortForwardRequest{Target: model.NewPortForwardTarget(model.NewPortForwardTargetVM("v"))}
	b, err = yaml.Marshal(f)
	require.NoError(t, err)
	assert.Equal(t, "target:\n    type: vm\n    id: v\n", string(b))
}

func TestClearPortForwardTargetOnPatch(t *testing.T) {
	b, err := json.Marshal(model.PatchedPortForwardRequest{Target: model.NullPortForwardTarget()})
	require.NoError(t, err)
	assert.JSONEq(t, `{"target":null}`, string(b))

	b, err = json.Marshal(model.PatchedPortForwardRequest{})
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(b))
}

func TestSelectorYAMLInput(t *testing.T) {
	in := "acl_id: a\nsources:\n  - type: ip\n    ip: 10.0.0.1\n  - type: vm\n    id: v\n  - type: group\n    id: g\n  - type: subnet\n    cidr: 10.0.0.0/8\n"
	var r model.ACLRuleRequest
	require.NoError(t, yaml.Unmarshal([]byte(in), &r))
	require.NotNil(t, r.Sources)
	require.Len(t, *r.Sources, 4)

	src := *r.Sources
	assert.Equal(t, model.NewACLSelectorIP("10.0.0.1"), src[0].ACLSelector)
	assert.Equal(t, model.NewACLSelectorVM("v"), src[1].ACLSelector)
	assert.Equal(t, model.NewACLSelectorGroup("g"), src[2].ACLSelector)
	assert.Equal(t, model.ACLSelectorUnknown{Fields: map[string]any{"type": "subnet", "cidr": "10.0.0.0/8"}}, src[3].ACLSelector)

	// Same payload must come out as JSON the API accepts.
	b, err := json.Marshal(r)
	require.NoError(t, err)
	assert.JSONEq(t, `{"acl_id":"a","sources":[{"type":"ip","ip":"10.0.0.1"},{"type":"vm","id":"v"},{"type":"group","id":"g"},{"type":"subnet","cidr":"10.0.0.0/8"}]}`, string(b))
}

func TestPortForwardTargetYAMLInput(t *testing.T) {
	var f model.PortForwardRequest
	require.NoError(t, yaml.Unmarshal([]byte("network_id: n\nlistener_id: l\ntarget:\n  type: vm\n  id: v\n"), &f))
	require.NotNil(t, f.Target)
	assert.Equal(t, model.NewPortForwardTargetVM("v"), f.Target.PortForwardTarget)

	var p model.PatchedPortForwardRequest
	require.NoError(t, yaml.Unmarshal([]byte("target: null\n"), &p))
	// yaml.v3 sets the pointer to nil for null, which omits the field on
	// PATCH; use NullPortForwardTarget to clear it explicitly.
	assert.Nil(t, p.Target)
}

func TestACLRuleRequestValidate(t *testing.T) {
	ok := []model.ACLSelectorValue{
		model.NewACLSelector(model.NewACLSelectorIP("10.0.0.1")),
		model.NewACLSelector(model.NewACLSelectorVM("v")),
		model.NewACLSelector(model.NewACLSelectorGroup("g")),
		model.NewACLSelector(model.ACLSelectorIP{IP: "x"}), // Type unset is fine
		model.NewACLSelector(model.ACLSelectorUnknown{Fields: map[string]any{"type": "subnet"}}),
	}
	assert.NoError(t, model.ACLRuleRequest{ACL: "a", Sources: &ok, Destinations: &ok}.Validate())
	assert.NoError(t, model.PatchedACLRuleRequest{}.Validate())

	cases := map[string]struct {
		sel  model.ACLSelectorValue
		want string
	}{
		"empty wrapper":   {model.ACLSelectorValue{}, "sources[1]: selector is empty"},
		"ip missing":      {model.NewACLSelector(model.ACLSelectorIP{}), `sources[1]: selector "ip": ip is required`},
		"vm id missing":   {model.NewACLSelector(model.ACLSelectorVM{}), `sources[1]: selector "vm": id is required`},
		"group missing":   {model.NewACLSelector(model.ACLSelectorGroup{}), `sources[1]: selector "group": id is required`},
		"type mismatch":   {model.NewACLSelector(model.ACLSelectorVM{Type: "ip", ID: "v"}), `sources[1]: selector "vm": type is "ip"`},
		"unknown no type": {model.NewACLSelector(model.ACLSelectorUnknown{Fields: map[string]any{"cidr": "x"}}), "sources[1]: selector: type is required"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			src := []model.ACLSelectorValue{ok[0], tc.sel}
			assert.EqualError(t, model.ACLRuleRequest{ACL: "a", Sources: &src}.Validate(), tc.want)
			assert.EqualError(t, model.PatchedACLRuleRequest{Sources: &src}.Validate(), tc.want)
		})
	}

	// Destinations are reported under their own name.
	bad := []model.ACLSelectorValue{{}}
	assert.EqualError(t, model.ACLRuleRequest{ACL: "a", Destinations: &bad}.Validate(), "destinations[0]: selector is empty")
}

func TestPortForwardRequestValidate(t *testing.T) {
	assert.NoError(t, model.PortForwardRequest{}.Validate())
	assert.NoError(t, model.PortForwardRequest{Target: model.NullPortForwardTarget()}.Validate())
	assert.NoError(t, model.PortForwardRequest{Target: model.NewPortForwardTarget(model.NewPortForwardTargetIP("1.2.3.4"))}.Validate())
	assert.NoError(t, model.PatchedPortForwardRequest{Target: model.NewPortForwardTarget(model.PortForwardTargetVM{ID: "v"})}.Validate())

	cases := map[string]struct {
		target model.PortForwardTarget
		want   string
	}{
		"ip missing":      {model.PortForwardTargetIP{}, `target: target "ip": ip is required`},
		"vm id missing":   {model.PortForwardTargetVM{}, `target: target "vm": id is required`},
		"type mismatch":   {model.PortForwardTargetIP{Type: "vm", IP: "x"}, `target: target "ip": type is "vm"`},
		"unknown no type": {model.PortForwardTargetUnknown{Fields: map[string]any{}}, "target: target: type is required"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.EqualError(t, model.PortForwardRequest{Target: model.NewPortForwardTarget(tc.target)}.Validate(), tc.want)
			assert.EqualError(t, model.PatchedPortForwardRequest{Target: model.NewPortForwardTarget(tc.target)}.Validate(), tc.want)
		})
	}
}
