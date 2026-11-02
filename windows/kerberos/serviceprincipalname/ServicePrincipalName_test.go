package serviceprincipalname

import (
	"testing"
)

func TestFromString_Valid(t *testing.T) {
	tests := []struct {
		name string
		spn  string
		want ServicePrincipalName
	}{
		{
			name: "three-part canonical",
			spn:  "ldap/dc-01.fabrikam.com/fabrikam.com",
			want: ServicePrincipalName{
				ServiceClass: "ldap",
				Hostname:     "dc-01.fabrikam.com",
				ServiceName:  "fabrikam.com",
			},
		},
		{
			name: "two-part",
			spn:  "cifs/host.domain.com",
			want: ServicePrincipalName{
				ServiceClass: "cifs",
				Hostname:     "host.domain.com",
			},
		},
		{
			name: "numeric port suffix",
			spn:  "MSSQLSvc/host.domain.com:1433",
			want: ServicePrincipalName{
				ServiceClass: "MSSQLSvc",
				Hostname:     "host.domain.com",
				Port:         1433,
				HasPort:      true,
			},
		},
		{
			name: "non-numeric instance suffix",
			spn:  "MSSQLSvc/host.domain.com:INST01",
			want: ServicePrincipalName{
				ServiceClass: "MSSQLSvc",
				Hostname:     "host.domain.com",
				InstanceName: "INST01",
			},
		},
		{
			name: "port and service name",
			spn:  "http/web:8080/realm",
			want: ServicePrincipalName{
				ServiceClass: "http",
				Hostname:     "web",
				Port:         8080,
				HasPort:      true,
				ServiceName:  "realm",
			},
		},
		{
			name: "instance name and service name",
			spn:  "MSSQLSvc/db:INST01/CONTOSO.LOCAL",
			want: ServicePrincipalName{
				ServiceClass: "MSSQLSvc",
				Hostname:     "db",
				InstanceName: "INST01",
				ServiceName:  "CONTOSO.LOCAL",
			},
		},
		{
			name: "max port",
			spn:  "svc/host:65535",
			want: ServicePrincipalName{
				ServiceClass: "svc",
				Hostname:     "host",
				Port:         65535,
				HasPort:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromString(tt.spn)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.spn, err)
			}
			if *got != tt.want {
				t.Errorf("FromString(%q) = %+v, want %+v", tt.spn, *got, tt.want)
			}
			// Round-trip: String must reproduce the input.
			if rendered := got.String(); rendered != tt.spn {
				t.Errorf("round-trip mismatch: String() = %q, want %q", rendered, tt.spn)
			}
			if !got.IsValid() {
				t.Errorf("IsValid() = false for parsed SPN %q", tt.spn)
			}
		})
	}
}

func TestFromString_Invalid(t *testing.T) {
	tests := []struct {
		name string
		spn  string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"single token", "ldap"},
		{"four parts", "a/b/c/d"},
		{"empty service class", "/host"},
		{"empty host", "ldap//x"},
		{"empty host two-part", "ldap/"},
		{"empty suffix", "cifs/host:"},
		{"empty host with port", "cifs/:1433"},
		{"port out of range", "svc/host:99999"},
		{"port way out of range", "svc/host:70000"},
		{"empty service name", "ldap/host/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromString(tt.spn)
			if err == nil {
				t.Errorf("FromString(%q) = %+v, want error", tt.spn, got)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spn     ServicePrincipalName
		wantErr bool
	}{
		{"valid two-part", ServicePrincipalName{ServiceClass: "cifs", Hostname: "host"}, false},
		{"valid with port", ServicePrincipalName{ServiceClass: "svc", Hostname: "host", Port: 1433}, false},
		{"empty service class", ServicePrincipalName{Hostname: "host"}, true},
		{"empty host", ServicePrincipalName{ServiceClass: "cifs"}, true},
		{"port and instance both set", ServicePrincipalName{ServiceClass: "svc", Hostname: "host", Port: 1, InstanceName: "x"}, true},
		{"separator in service class", ServicePrincipalName{ServiceClass: "ci/fs", Hostname: "host"}, true},
		{"colon in service name", ServicePrincipalName{ServiceClass: "cifs", Hostname: "host", ServiceName: "re:alm"}, true},
		{"slash in host", ServicePrincipalName{ServiceClass: "cifs", Hostname: "ho/st"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spn.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.spn.IsValid() != (err == nil) {
				t.Errorf("IsValid() inconsistent with Validate()")
			}
		})
	}

	var nilSPN *ServicePrincipalName
	if err := nilSPN.Validate(); err == nil {
		t.Errorf("Validate() on nil = nil, want error")
	}
}

func TestHasServiceName(t *testing.T) {
	three, _ := FromString("ldap/dc.fabrikam.com/fabrikam.com")
	if !three.HasServiceName() {
		t.Errorf("HasServiceName() = false for three-part SPN")
	}
	two, _ := FromString("cifs/host.domain.com")
	if two.HasServiceName() {
		t.Errorf("HasServiceName() = true for two-part SPN")
	}
}

func TestEqual(t *testing.T) {
	a, _ := FromString("CIFS/Host.Domain.Com")
	b, _ := FromString("cifs/host.domain.com")
	if !a.Equal(b) {
		t.Errorf("Equal() = false for case-differing equivalent SPNs")
	}

	c, _ := FromString("ldap/host.domain.com")
	if a.Equal(c) {
		t.Errorf("Equal() = true for differing service classes")
	}

	d, _ := FromString("svc/host:1433")
	e, _ := FromString("svc/host:1434")
	if d.Equal(e) {
		t.Errorf("Equal() = true for differing ports")
	}

	if a.Equal(nil) {
		t.Errorf("Equal(nil) = true, want false")
	}
	var n1, n2 *ServicePrincipalName
	if !n1.Equal(n2) {
		t.Errorf("nil.Equal(nil) = false, want true")
	}
}

// TestFromString_IPv6Host checks a bracketed IPv6 literal host is kept whole,
// with the port or instance name after the "]", and renders back bracketed.
func TestFromString_IPv6Host(t *testing.T) {
	for _, tc := range []struct {
		spn  string
		want ServicePrincipalName
	}{
		{"host/[fe80::1]", ServicePrincipalName{ServiceClass: "host", Hostname: "fe80::1"}},
		{"cifs/[fd2e:5387:e712:3333::1]:445", ServicePrincipalName{ServiceClass: "cifs", Hostname: "fd2e:5387:e712:3333::1", Port: 445, HasPort: true}},
		{"MSSQLSvc/[::1]:SQLEXPRESS/corp.local", ServicePrincipalName{ServiceClass: "MSSQLSvc", Hostname: "::1", InstanceName: "SQLEXPRESS", ServiceName: "corp.local"}},
	} {
		got, err := FromString(tc.spn)
		if err != nil {
			t.Fatalf("FromString(%q): %v", tc.spn, err)
		}
		if *got != tc.want {
			t.Errorf("FromString(%q) = %+v, want %+v", tc.spn, *got, tc.want)
		}
		if s := got.String(); s != tc.spn {
			t.Errorf("String() = %q, want %q", s, tc.spn)
		}
	}
	for _, bad := range []string{"host/[fe80::1", "host/[fe80::1]x", "host/[host.local]", "host/[fe80::1]:"} {
		if _, err := FromString(bad); err == nil {
			t.Errorf("FromString(%q) accepted a malformed bracketed host", bad)
		}
	}
}

// TestFromString_ExplicitZeroPortRoundTrips checks an explicit ":0" port is kept
// through FromString and String, and distinguishes the SPN from one with no port.
func TestFromString_ExplicitZeroPortRoundTrips(t *testing.T) {
	got, err := FromString("svc/host:0")
	if err != nil {
		t.Fatal(err)
	}
	if !got.HasPort || got.Port != 0 {
		t.Errorf("FromString(svc/host:0) = %+v, want HasPort with port 0", *got)
	}
	if s := got.String(); s != "svc/host:0" {
		t.Errorf("String() = %q, want svc/host:0", s)
	}
	plain, _ := FromString("svc/host")
	if got.Equal(plain) {
		t.Error("svc/host:0 compares equal to svc/host")
	}
}

// TestValidate_HostColonOnlyForIPv6 checks a hand-built host name with a colon is
// accepted only when it is an IPv6 literal, which String brackets.
func TestValidate_HostColonOnlyForIPv6(t *testing.T) {
	if err := (&ServicePrincipalName{ServiceClass: "svc", Hostname: "fe80::1", Port: 88}).Validate(); err != nil {
		t.Errorf("IPv6 host rejected: %v", err)
	}
	if err := (&ServicePrincipalName{ServiceClass: "svc", Hostname: "host:x"}).Validate(); err == nil {
		t.Error("host name with a colon accepted")
	}
	// Brackets are rendering syntax, not part of a host name: "[x" would render
	// as "svc/[x", which no longer parses.
	for _, h := range []string{"[x", "x]", "[::1]"} {
		if err := (&ServicePrincipalName{ServiceClass: "svc", Hostname: h}).Validate(); err == nil {
			t.Errorf("host name %q with a bracket accepted", h)
		}
	}
}
