package objects

import (
	"fmt"
	"testing"

	"github.com/go-ldap/ldap/v3"
)

// queryRecorder is an LdapSessionInterface that records the filters it is asked to
// run and returns no entries.
type queryRecorder struct {
	queries []string
	bases   []string
	// rootDSE, when set, is returned by GetRootDSE.
	rootDSE *ldap.Entry
}

func (r *queryRecorder) Connect() (bool, error)   { return true, nil }
func (r *queryRecorder) ReConnect() (bool, error) { return true, nil }
func (r *queryRecorder) Close()                   {}

func (r *queryRecorder) Query(searchBase string, query string, attributes []string, scope int) ([]*ldap.Entry, error) {
	r.queries = append(r.queries, query)
	r.bases = append(r.bases, searchBase)
	return nil, nil
}

// GetRootDSE records a read of the RootDSE as a query on the base "(RootDSE)".
func (r *queryRecorder) GetRootDSE() (*ldap.Entry, error) {
	r.queries = append(r.queries, "")
	r.bases = append(r.bases, "(RootDSE)")
	if r.rootDSE == nil {
		return nil, fmt.Errorf("no RootDSE")
	}
	return r.rootDSE, nil
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

		// The RootDSE read comes first and finds nothing here, so the domain
		// head search is the last query.
		if len(recorder.queries) == 0 || recorder.bases[len(recorder.bases)-1] != distinguishedName {
			t.Fatalf("no domain head search for %q (queries %q on bases %q)", distinguishedName, recorder.queries, recorder.bases)
		}
		filter := recorder.queries[len(recorder.queries)-1]

		packet, err := ldap.CompileFilter(filter)
		if err != nil {
			t.Errorf("filter %q for %q does not compile: %s", filter, distinguishedName, err)
			continue
		}

		// The equality assertion value must decode back to the DN unchanged.
		if got := string(packet.Children[1].Data.Bytes()); got != distinguishedName {
			t.Errorf("filter %q asserts %q, want %q", filter, got, distinguishedName)
		}
	}
}

// TestGetDomainFunctionalityLevelUsesRootDSE checks the level comes from the
// RootDSE's domainFunctionality without searching the domain head when the
// RootDSE describes this domain, and that another domain's RootDSE is not used.
func TestGetDomainFunctionalityLevelUsesRootDSE(t *testing.T) {
	rootDSE := ldap.NewEntry("", map[string][]string{
		"defaultNamingContext": {"DC=corp,DC=local"},
		"domainFunctionality":  {"7"},
	})

	recorder := &queryRecorder{rootDSE: rootDSE}
	domain := &Domain{LdapSession: recorder, DistinguishedName: "dc=CORP,dc=local"}
	level, err := domain.GetDomainFunctionalityLevel()
	if err != nil || level != 7 {
		t.Fatalf("level=%d err=%v, want 7 from the RootDSE", level, err)
	}
	if len(recorder.queries) != 1 || recorder.bases[0] != "(RootDSE)" {
		t.Errorf("queries on bases %q, want only the RootDSE read", recorder.bases)
	}

	recorder = &queryRecorder{rootDSE: rootDSE}
	other := &Domain{LdapSession: recorder, DistinguishedName: "DC=child,DC=corp,DC=local"}
	other.GetDomainFunctionalityLevel()
	if len(recorder.bases) != 2 || recorder.bases[1] != "DC=child,DC=corp,DC=local" {
		t.Errorf("queries on bases %q, want the RootDSE then the child domain head", recorder.bases)
	}
}
