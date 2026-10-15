package ibmi

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

// TransportPolicy is the caller's explicit choice between verified TLS and
// cleartext. The zero value is deliberately unusable.
type TransportPolicy uint8

const (
	// TransportUnset is the zero policy; it is always rejected.
	TransportUnset TransportPolicy = iota
	// TransportTLSRequired uses the "-s" services over verified TLS and never
	// falls back to cleartext.
	TransportTLSRequired
	// TransportCleartextExplicit uses the cleartext services. Credentials and
	// data cross the network unprotected; use it only on an isolated lab
	// network.
	TransportCleartextExplicit
)

func (p TransportPolicy) String() string {
	switch p {
	case TransportUnset:
		return "unset"
	case TransportTLSRequired:
		return "tls-required"
	case TransportCleartextExplicit:
		return "cleartext-explicit"
	}
	return fmt.Sprintf("TransportPolicy(%d)", uint8(p))
}

// ErrTransportPolicy is matched (via errors.Is) when a transport policy is
// unset or undefined.
var ErrTransportPolicy = errors.New("ibmi: transport policy must be TLSRequired or CleartextExplicit")

// Validate rejects the zero and undefined policies.
func (p TransportPolicy) Validate() error {
	if p == TransportTLSRequired || p == TransportCleartextExplicit {
		return nil
	}
	return fmt.Errorf("%w (got %s)", ErrTransportPolicy, p)
}

// TLS reports whether the policy requires TLS service sockets.
func (p TransportPolicy) TLS() bool { return p == TransportTLSRequired }

// EndpointSource records how an endpoint's port was determined.
type EndpointSource uint8

const (
	// EndpointExplicit means the caller configured the port.
	EndpointExplicit EndpointSource = iota + 1
	// EndpointServiceMapper means the TCP/449 service mapper returned it.
	EndpointServiceMapper
)

func (s EndpointSource) String() string {
	switch s {
	case EndpointExplicit:
		return "explicit"
	case EndpointServiceMapper:
		return "service-mapper"
	}
	return fmt.Sprintf("EndpointSource(%d)", uint8(s))
}

// Endpoint is a resolved host-server address.
type Endpoint struct {
	Service Service
	// Name is the service-mapper name, for example "as-rmtcmd-s". It always
	// agrees with Service and TLS.
	Name   string
	Host   string
	Port   int
	TLS    bool
	Source EndpointSource
}

// ErrInvalidEndpoint is matched (via errors.Is) by Endpoint.Validate errors.
var ErrInvalidEndpoint = errors.New("ibmi: invalid endpoint")

// Validate checks that the endpoint is complete and self-consistent.
func (e Endpoint) Validate() error {
	switch {
	case !e.Service.Valid():
		return fmt.Errorf("%w: service %s", ErrInvalidEndpoint, e.Service)
	case e.Name != e.Service.MapperName(e.TLS):
		return fmt.Errorf("%w: name %q does not match %s with TLS=%t", ErrInvalidEndpoint, e.Name, e.Service, e.TLS)
	case e.Host == "":
		return fmt.Errorf("%w: empty host", ErrInvalidEndpoint)
	case e.Port < 1 || e.Port > 65535:
		return fmt.Errorf("%w: port %d outside 1-65535", ErrInvalidEndpoint, e.Port)
	case e.Source != EndpointExplicit && e.Source != EndpointServiceMapper:
		return fmt.Errorf("%w: source %s", ErrInvalidEndpoint, e.Source)
	}
	return nil
}

// Address returns "host:port", bracketing IPv6 hosts.
func (e Endpoint) Address() string { return net.JoinHostPort(e.Host, strconv.Itoa(e.Port)) }

// ServiceCatalog is the outcome of service discovery. Missing lists the
// services that could not be resolved; their Endpoint fields are zero.
type ServiceCatalog struct {
	Signon        Endpoint
	RemoteCommand Endpoint
	DataQueue     Endpoint
	File          Endpoint
	Missing       []Service
}

// Endpoint returns the catalog entry for s and whether it was resolved.
func (c ServiceCatalog) Endpoint(s Service) (Endpoint, bool) {
	var e Endpoint
	switch s {
	case ServiceSignon:
		e = c.Signon
	case ServiceRemoteCommand:
		e = c.RemoteCommand
	case ServiceDataQueue:
		e = c.DataQueue
	case ServiceFile:
		e = c.File
	}
	return e, e.Service.Valid()
}
