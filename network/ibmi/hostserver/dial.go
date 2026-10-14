package hostserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
)

// DefaultDialTimeout bounds connection setup, TLS handshake included, when
// Dialer.Timeout is 0.
const DefaultDialTimeout = 10 * time.Second

// ErrTLSOptions is matched (via errors.Is) by invalid TLS options.
var ErrTLSOptions = errors.New("hostserver: invalid TLS options")

// TLSOptions are the only TLS settings a caller controls. There is
// deliberately no way to skip certificate or host-name verification: the
// library builds a fresh tls.Config from these values for every connection.
type TLSOptions struct {
	// RootCAs verifies the server; nil uses the system roots. It is cloned.
	RootCAs *x509.CertPool
	// ServerName is the name verified against the certificate; empty uses
	// the endpoint host.
	ServerName string
	// Certificates are optional client certificates. The slice is copied.
	Certificates []tls.Certificate
	// MinVersion and MaxVersion bound the protocol version. MinVersion 0
	// means TLS 1.2. Versions below TLS 1.2 require AllowLegacy.
	MinVersion uint16
	MaxVersion uint16
	// CipherSuites restricts TLS 1.0-1.2 suites; nil uses Go's defaults.
	// Suites from tls.InsecureCipherSuites require AllowLegacy.
	CipherSuites []uint16
	// AllowLegacy permits TLS 1.0/1.1 and the insecure suites the compiled
	// Go toolchain still implements, for old IBM i releases. Certificates
	// and server names are still verified, and process-wide GODEBUG is never
	// changed.
	AllowLegacy bool
}

// Config returns a new verified tls.Config for connecting to host.
func (o TLSOptions) Config(host string) (*tls.Config, error) {
	min := o.MinVersion
	if min == 0 {
		min = tls.VersionTLS12
	}
	if min < tls.VersionTLS10 || (o.MaxVersion != 0 && o.MaxVersion < min) {
		return nil, fmt.Errorf("%w: version range 0x%04X-0x%04X", ErrTLSOptions, min, o.MaxVersion)
	}
	if min < tls.VersionTLS12 && !o.AllowLegacy {
		return nil, fmt.Errorf("%w: minimum version 0x%04X below TLS 1.2 requires AllowLegacy", ErrTLSOptions, min)
	}
	insecure := map[uint16]bool{}
	for _, s := range tls.InsecureCipherSuites() {
		insecure[s.ID] = true
	}
	known := map[uint16]bool{}
	for _, s := range tls.CipherSuites() {
		known[s.ID] = true
	}
	for _, id := range o.CipherSuites {
		switch {
		case insecure[id] && !o.AllowLegacy:
			return nil, fmt.Errorf("%w: insecure cipher suite %s requires AllowLegacy", ErrTLSOptions, tls.CipherSuiteName(id))
		case !insecure[id] && !known[id]:
			return nil, fmt.Errorf("%w: cipher suite 0x%04X is not implemented by this Go toolchain", ErrTLSOptions, id)
		}
	}
	name := o.ServerName
	if name == "" {
		name = host
	}
	if name == "" {
		return nil, fmt.Errorf("%w: no server name to verify", ErrTLSOptions)
	}
	cfg := &tls.Config{
		ServerName:   name,
		MinVersion:   min,
		MaxVersion:   o.MaxVersion,
		Certificates: append([]tls.Certificate(nil), o.Certificates...),
		CipherSuites: append([]uint16(nil), o.CipherSuites...),
	}
	if o.RootCAs != nil {
		cfg.RootCAs = o.RootCAs.Clone()
	}
	return cfg, nil
}

// Dialer opens host-server connections to resolved endpoints.
type Dialer struct {
	// Dial opens the TCP connection; nil uses net.Dialer.DialContext.
	Dial ibminet.DialFunc
	// Timeout bounds TCP connect plus TLS handshake; 0 means
	// DefaultDialTimeout.
	Timeout time.Duration
	TLS     TLSOptions
	Conn    Options
}

// DialEndpoint connects to ep. When ep.TLS is set it completes a verified
// TLS handshake before returning and never falls back to cleartext; on any
// failure the socket is closed.
func (d *Dialer) DialEndpoint(ctx context.Context, ep ibminet.Endpoint) (*Conn, error) {
	if err := ep.Validate(); err != nil {
		return nil, err
	}
	if d.Timeout < 0 {
		return nil, fmt.Errorf("hostserver: negative dial timeout %s", d.Timeout)
	}
	var cfg *tls.Config
	if ep.TLS {
		var err error
		if cfg, err = d.TLS.Config(ep.Host); err != nil {
			return nil, err
		}
	}
	timeout := d.Timeout
	if timeout == 0 {
		timeout = DefaultDialTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dial := d.Dial
	if dial == nil {
		dial = (&net.Dialer{}).DialContext
	}
	raw, err := dial(ctx, "tcp", ep.Address())
	if err != nil {
		return nil, fmt.Errorf("hostserver: dial %s (%s): %w", ep.Name, ep.Address(), err)
	}
	conn := raw
	if cfg != nil {
		tc := tls.Client(raw, cfg)
		if err := tc.HandshakeContext(ctx); err != nil {
			_ = raw.Close()
			return nil, fmt.Errorf("hostserver: TLS handshake with %s (%s): %w", ep.Name, ep.Address(), err)
		}
		conn = tc
	}
	c, err := NewConn(conn, d.Conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return c, nil
}
