package serviceprincipalname

import "testing"

// builtInList is the verbatim built-in SPN list from the setspn documentation
// (https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/setspn),
// used as an independent known-answer check against the package's own set.
var builtInList = []string{
	"alerter", "appmgmt", "browser", "cifs", "cisvc", "clipsrv",
	"dcom", "dhcp", "dmserver", "dns", "dnscache", "eventlog",
	"eventsystem", "fax", "http", "ias", "iisadmin", "mcsvc",
	"messenger", "msiserver", "netdde", "netddedsm", "netlogon", "netman",
	"nmagent", "oakley", "plugplay", "policyagent", "protectedstorage", "rasman",
	"remoteaccess", "replicator", "rpc", "rpclocator", "rpcss", "rsvp",
	"samss", "scardsvr", "scesrv", "schedule", "scm", "seclogon",
	"snmp", "spooler", "tapisrv", "time", "trksvr", "trkwks",
	"ups", "w3svc", "wins", "www",
}

func TestBuiltInServiceClassesMatchesDoc(t *testing.T) {
	got := BuiltInServiceClasses()
	if len(got) != len(builtInList) {
		t.Fatalf("BuiltInServiceClasses len = %d, want %d", len(got), len(builtInList))
	}
	// BuiltInServiceClasses is documented to return a sorted slice; builtInList is
	// already sorted, so compare element-wise.
	for i := range builtInList {
		if got[i] != builtInList[i] {
			t.Errorf("BuiltInServiceClasses[%d] = %q, want %q", i, got[i], builtInList[i])
		}
	}
}

func TestIsBuiltInServiceClass(t *testing.T) {
	for _, class := range builtInList {
		if !IsBuiltInServiceClass(class) {
			t.Errorf("IsBuiltInServiceClass(%q) = false, want true", class)
		}
		// SPNs are not case sensitive: uppercase must match too.
		up := ""
		for _, r := range class {
			if r >= 'a' && r <= 'z' {
				r -= 'a' - 'A'
			}
			up += string(r)
		}
		if !IsBuiltInServiceClass(up) {
			t.Errorf("IsBuiltInServiceClass(%q) = false, want true (case-insensitive)", up)
		}
	}

	for _, class := range []string{"HOST", "host", "MSSQLSvc", "ldap", "", "htttp"} {
		if IsBuiltInServiceClass(class) {
			t.Errorf("IsBuiltInServiceClass(%q) = true, want false", class)
		}
	}
}

func TestHostIsNotBuiltIn(t *testing.T) {
	// HOST is the substituting class, not itself a built-in service class.
	if IsBuiltInServiceClass(HostServiceClass) {
		t.Errorf("IsBuiltInServiceClass(%q) = true, want false", HostServiceClass)
	}
}

func TestSPNBuiltInMethods(t *testing.T) {
	tests := []struct {
		spn     string
		builtIn bool
		host    bool
		covered bool
	}{
		{"cifs/server.corp.local", true, false, true},
		{"HTTP/web.corp.local", true, false, true}, // case-insensitive
		{"HOST/server.corp.local", false, true, false},
		{"host/server.corp.local", false, true, false},
		{"MSSQLSvc/db.corp.local:1433", false, false, false},
		{"ldap/dc.corp.local/corp.local", false, false, false},
	}
	for _, tt := range tests {
		s, err := FromString(tt.spn)
		if err != nil {
			t.Fatalf("FromString(%q): %v", tt.spn, err)
		}
		if got := s.IsBuiltInServiceClass(); got != tt.builtIn {
			t.Errorf("%q IsBuiltInServiceClass = %v, want %v", tt.spn, got, tt.builtIn)
		}
		if got := s.IsHostServiceClass(); got != tt.host {
			t.Errorf("%q IsHostServiceClass = %v, want %v", tt.spn, got, tt.host)
		}
		if got := s.CoveredByHostSPN(); got != tt.covered {
			t.Errorf("%q CoveredByHostSPN = %v, want %v", tt.spn, got, tt.covered)
		}
	}
}

func TestNilSPNBuiltInMethods(t *testing.T) {
	var s *ServicePrincipalName
	if s.IsBuiltInServiceClass() || s.IsHostServiceClass() || s.CoveredByHostSPN() {
		t.Error("nil SPN methods should all report false")
	}
}
