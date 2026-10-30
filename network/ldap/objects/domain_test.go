package objects

import (
	"testing"

	"github.com/go-ldap/ldap/v3"
)

// queryRecorder is an LdapSessionInterface that records the filters it is asked to
// run and returns no entries.
type queryRecorder struct {
	queries []string
}

func (r *queryRecorder) Connect() (bool, error)   { return true, nil }
func (r *queryRecorder) ReConnect() (bool, error) { return true, nil }
func (r *queryRecorder) Close()                   {}

func (r *queryRecorder) Query(searchBase string, query string, attributes []string, scope int) ([]*ldap.Entry, error) {
	r.queries = append(r.queries, query)
	return nil, nil
}

func (r *queryRecorder) QueryWholeSubtree(searchBase string, query string, attributes []string) ([]*ldap.Entry, error) {
	return r.Query(searchBase, query, attributes, ldap.ScopeWholeSubtree)
}

func (r *queryRecorder) QueryBaseObject(searchBase string, query string, attributes []string) ([]*ldap.Entry, error) {
	return r.Query(searchBase, query, attributes, ldap.ScopeBaseObject)
}

func (r *queryRecorder) QuerySingleLevel(searchBase string, query string, attributes []string) ([]*ldap.Entry, error) {
	return r.Query(searchBase, query, attributes, ldap.ScopeSingleLevel)
}

func (r *queryRecorder) QueryChildren(searchBase string, query string, attributes []string) ([]*ldap.Entry, error) {
	return r.Query(searchBase, query, attributes, ldap.ScopeChildren)
}

func (r *queryRecorder) GetDomain(distinguishedName string) (*Domain, error) { return nil, nil }

// TestGetDomainFunctionalityLevelFilterCompiles checks that the filter built from the
// domain's distinguished name is a valid RFC 4515 filter when the DN carries filter
// metacharacters, and that it still matches that exact DN.
func TestGetDomainFunctionalityLevelFilterCompiles(t *testing.T) {
	distinguishedNames := []string{
		`DC=corp,DC=local`,
		`OU=Smith\, John,DC=corp,DC=local`,
		`OU=Svc (prod),DC=corp,DC=local`,
		`OU=Star*Name,DC=corp,DC=local`,
	}

	for _, distinguishedName := range distinguishedNames {
		recorder := &queryRecorder{}
		domain := &Domain{LdapSession: recorder, DistinguishedName: distinguishedName}
		domain.GetDomainFunctionalityLevel()

		if len(recorder.queries) != 1 {
			t.Fatalf("expected 1 query for %q, got %d", distinguishedName, len(recorder.queries))
		}

		packet, err := ldap.CompileFilter(recorder.queries[0])
		if err != nil {
			t.Errorf("filter %q for %q does not compile: %s", recorder.queries[0], distinguishedName, err)
			continue
		}

		// The equality assertion value must decode back to the DN unchanged.
		if got := string(packet.Children[1].Data.Bytes()); got != distinguishedName {
			t.Errorf("filter %q asserts %q, want %q", recorder.queries[0], got, distinguishedName)
		}
	}
}
