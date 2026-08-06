package ldap

import (
	"strings"
	"testing"

	goldapv3 "github.com/go-ldap/ldap/v3"
)

// TestBuildModifyRequestEmptyReplaceClearsAttribute is the regression test for an
// empty (non-nil) ReplaceValues: it must still emit a Replace change with no values,
// which is how an attribute is cleared. Previously the value-count guard dropped it,
// producing an empty modify request the directory rejects.
func TestBuildModifyRequestEmptyReplaceClearsAttribute(t *testing.T) {
	mr := &ModifyRequest{
		DistinguishedName: "cn=obj,dc=example,dc=com",
		Attributes: []*Action{
			{Attribute: "msDS-KeyCredentialLink", ReplaceValues: []string{}},
		},
	}

	m := buildModifyRequest(mr)

	if len(m.Changes) != 1 {
		t.Fatalf("Changes = %d; want 1 (empty replace must clear the attribute, not be dropped)", len(m.Changes))
	}
	change := m.Changes[0]
	if change.Operation != goldapv3.ReplaceAttribute {
		t.Errorf("Operation = %d; want ReplaceAttribute (%d)", change.Operation, goldapv3.ReplaceAttribute)
	}
	if change.Modification.Type != "msDS-KeyCredentialLink" {
		t.Errorf("Type = %q; want %q", change.Modification.Type, "msDS-KeyCredentialLink")
	}
	if len(change.Modification.Vals) != 0 {
		t.Errorf("Vals = %v; want empty", change.Modification.Vals)
	}
}

// TestBuildModifyRequestNilReplaceIsNoOp confirms that an unset (nil) ReplaceValues
// with no other values emits no change, so the fix does not turn an unset action
// into an accidental attribute clear.
func TestBuildModifyRequestNilReplaceIsNoOp(t *testing.T) {
	mr := &ModifyRequest{
		DistinguishedName: "cn=obj,dc=example,dc=com",
		Attributes:        []*Action{{Attribute: "description"}},
	}

	m := buildModifyRequest(mr)

	if len(m.Changes) != 0 {
		t.Fatalf("Changes = %d; want 0 (an unset action must not emit an operation)", len(m.Changes))
	}
}

// TestBuildModifyRequestOperationSelection checks that a populated value list still
// maps to its matching operation.
func TestBuildModifyRequestOperationSelection(t *testing.T) {
	cases := []struct {
		name     string
		action   *Action
		wantOp   uint
		wantVals []string
	}{
		{"add", &Action{Attribute: "member", AddValues: []string{"a"}}, goldapv3.AddAttribute, []string{"a"}},
		{"delete", &Action{Attribute: "member", DelValues: []string{"b"}}, goldapv3.DeleteAttribute, []string{"b"}},
		{"replace", &Action{Attribute: "member", ReplaceValues: []string{"c", "d"}}, goldapv3.ReplaceAttribute, []string{"c", "d"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := buildModifyRequest(&ModifyRequest{
				DistinguishedName: "cn=obj,dc=example,dc=com",
				Attributes:        []*Action{tc.action},
			})
			if len(m.Changes) != 1 {
				t.Fatalf("Changes = %d; want 1", len(m.Changes))
			}
			change := m.Changes[0]
			if change.Operation != tc.wantOp {
				t.Errorf("Operation = %d; want %d", change.Operation, tc.wantOp)
			}
			if len(change.Modification.Vals) != len(tc.wantVals) {
				t.Fatalf("Vals = %v; want %v", change.Modification.Vals, tc.wantVals)
			}
			for i, v := range tc.wantVals {
				if change.Modification.Vals[i] != v {
					t.Errorf("Vals[%d] = %q; want %q", i, change.Modification.Vals[i], v)
				}
			}
		})
	}
}

// A request that selects no operation at all cannot change anything, and sending it
// spends a round trip to come back as a generic "Unwilling To Perform" that describes
// neither the cause nor the fix. Modify has to refuse it before the wire.
//
// The guard returns before the connection is used, so a zero-valued Session is enough
// here: reaching the connection on any of these inputs would panic instead.
func TestModifyRejectsRequestWithNoChanges(t *testing.T) {
	cases := []struct {
		name       string
		attributes []*Action
	}{
		{name: "nil attributes", attributes: nil},
		{name: "empty attributes", attributes: []*Action{}},
		{name: "action with every value list unset", attributes: []*Action{{Attribute: "description"}}},
		{name: "empty AddValues", attributes: []*Action{{Attribute: "description", AddValues: []string{}}}},
		{name: "empty DelValues", attributes: []*Action{{Attribute: "description", DelValues: []string{}}}},
		{name: "empty IncrementValues", attributes: []*Action{{Attribute: "description", IncrementValues: []string{}}}},
		{name: "several unset actions", attributes: []*Action{{Attribute: "description"}, {Attribute: "displayName"}}},
	}

	session := &Session{}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := session.Modify(&ModifyRequest{
				DistinguishedName: "cn=obj,dc=example,dc=com",
				Attributes:        tt.attributes,
			})
			if err == nil {
				t.Fatal("error = nil, want a refusal")
			}
			if !strings.Contains(err.Error(), "no changes") {
				t.Errorf("error = %q, want it to say the request carries no changes", err)
			}
			if !strings.Contains(err.Error(), "cn=obj,dc=example,dc=com") {
				t.Errorf("error = %q, want it to name the target", err)
			}
		})
	}
}

// The guard is about the request as a whole, not the individual actions: an action
// that selects nothing alongside one that does must still reach the wire.
func TestBuildModifyRequestKeepsRealChangeBesideNoOpAction(t *testing.T) {
	mr := &ModifyRequest{
		DistinguishedName: "cn=obj,dc=example,dc=com",
		Attributes: []*Action{
			{Attribute: "description", ReplaceValues: []string{"value"}},
			{Attribute: "displayName"},
		},
	}

	m := buildModifyRequest(mr)

	if len(m.Changes) != 1 {
		t.Fatalf("Changes = %d; want 1 (the real change survives, the no-op adds nothing)", len(m.Changes))
	}
	if m.Changes[0].Operation != goldapv3.ReplaceAttribute {
		t.Errorf("Operation = %d; want ReplaceAttribute", m.Changes[0].Operation)
	}
}
