package ibmi_test

import (
	"errors"
	"testing"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
)

func TestMapperNames(t *testing.T) {
	for _, tt := range []struct {
		s    ibminet.Service
		tls  bool
		name string
	}{
		{ibminet.ServiceSignon, false, "as-signon"},
		{ibminet.ServiceSignon, true, "as-signon-s"},
		{ibminet.ServiceRemoteCommand, false, "as-rmtcmd"},
		{ibminet.ServiceRemoteCommand, true, "as-rmtcmd-s"},
	} {
		if got := tt.s.MapperName(tt.tls); got != tt.name {
			t.Errorf("%s.MapperName(%t) = %q, want %q", tt.s, tt.tls, got, tt.name)
		}
		s, tls, ok := ibminet.ServiceForMapperName(tt.name)
		if !ok || s != tt.s || tls != tt.tls {
			t.Errorf("ServiceForMapperName(%q) = %s, %t, %t", tt.name, s, tls, ok)
		}
	}
	for _, name := range []string{"", "as-svrmap", "as-file", "AS-SIGNON", "as-signon-s ", "as-signon\x00"} {
		if _, _, ok := ibminet.ServiceForMapperName(name); ok {
			t.Errorf("ServiceForMapperName(%q) accepted", name)
		}
	}
	if ibminet.Service(0).MapperName(true) != "" || ibminet.Service(9).Valid() {
		t.Error("invalid services must have no mapper name")
	}
}

func TestTransportPolicy(t *testing.T) {
	for _, p := range []ibminet.TransportPolicy{ibminet.TransportUnset, 3, 255} {
		if !errors.Is(p.Validate(), ibminet.ErrTransportPolicy) {
			t.Errorf("%s.Validate() accepted", p)
		}
	}
	if ibminet.TransportTLSRequired.Validate() != nil || !ibminet.TransportTLSRequired.TLS() {
		t.Error("TLSRequired must validate and require TLS")
	}
	if ibminet.TransportCleartextExplicit.Validate() != nil || ibminet.TransportCleartextExplicit.TLS() {
		t.Error("CleartextExplicit must validate and not use TLS")
	}
}

func TestEndpointValidate(t *testing.T) {
	ok := ibminet.Endpoint{Service: ibminet.ServiceRemoteCommand, Name: "as-rmtcmd-s", Host: "ibmi.example", Port: 9475, TLS: true, Source: ibminet.EndpointServiceMapper}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	if ok.Address() != "ibmi.example:9475" {
		t.Errorf("Address() = %q", ok.Address())
	}
	v6 := ok
	v6.Host = "2001:db8::1"
	if v6.Address() != "[2001:db8::1]:9475" {
		t.Errorf("IPv6 Address() = %q", v6.Address())
	}
	for name, mutate := range map[string]func(*ibminet.Endpoint){
		"bad service":       func(e *ibminet.Endpoint) { e.Service = 0 },
		"name/tls mismatch": func(e *ibminet.Endpoint) { e.TLS = false },
		"wrong name":        func(e *ibminet.Endpoint) { e.Name = "as-signon-s" },
		"empty host":        func(e *ibminet.Endpoint) { e.Host = "" },
		"port 0":            func(e *ibminet.Endpoint) { e.Port = 0 },
		"port too large":    func(e *ibminet.Endpoint) { e.Port = 65536 },
		"no source":         func(e *ibminet.Endpoint) { e.Source = 0 },
	} {
		e := ok
		mutate(&e)
		if !errors.Is(e.Validate(), ibminet.ErrInvalidEndpoint) {
			t.Errorf("%s: Validate accepted %+v", name, e)
		}
	}
}

func TestServiceCatalogEndpoint(t *testing.T) {
	c := ibminet.ServiceCatalog{Signon: ibminet.Endpoint{Service: ibminet.ServiceSignon}, Missing: []ibminet.Service{ibminet.ServiceRemoteCommand}}
	if _, ok := c.Endpoint(ibminet.ServiceSignon); !ok {
		t.Error("resolved Signon not reported")
	}
	if _, ok := c.Endpoint(ibminet.ServiceRemoteCommand); ok {
		t.Error("missing RemoteCommand reported as resolved")
	}
}
