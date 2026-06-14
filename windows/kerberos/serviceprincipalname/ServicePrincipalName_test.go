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
