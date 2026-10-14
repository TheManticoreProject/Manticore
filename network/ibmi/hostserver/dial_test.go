package hostserver_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"net"
	"sync/atomic"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/servicemapper"
)

// testCA is a throwaway certificate authority generated per test run.
type testCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pool *x509.CertPool
}

func newTestCA(t *testing.T) *testCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Manticore IBM i test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, _ := x509.ParseCertificate(der)
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return &testCA{cert: cert, key: key, pool: pool}
}

// leaf issues a server certificate for dnsName and 127.0.0.1.
func (ca *testCA) leaf(t *testing.T, dnsName string, notAfter time.Time) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: dnsName},
		DNSNames:     []string{dnsName},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:    notAfter.Add(-2 * time.Hour),
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// startServer listens on loopback and answers one host-server request per
// connection with a 0x8003 reply. With cert nil it speaks cleartext.
func startServer(t *testing.T, cert *tls.Certificate) (port int, accepted *atomic.Int32) {
	t.Helper()
	var ln net.Listener
	var err error
	if cert != nil {
		ln, err = tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{*cert}, MinVersion: tls.VersionTLS12})
	} else {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	accepted = &atomic.Int32{}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			accepted.Add(1)
			go func() {
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(5 * time.Second))
				var hb [hostserver.HeaderLength]byte
				if _, err := io.ReadFull(c, hb[:]); err != nil {
					return
				}
				body := make([]byte, binary.BigEndian.Uint32(hb[:])-hostserver.HeaderLength)
				if _, err := io.ReadFull(c, body); err != nil {
					return
				}
				reply, _ := hostserver.Frame{Header: hostserver.Header{
					ServerID: hostserver.ServerIDRemoteCommand, CorrelationID: binary.BigEndian.Uint32(hb[12:]), RequestID: 0x8003,
				}}.Marshal(0)
				_, _ = c.Write(reply)
			}()
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, accepted
}

func tlsEndpoint(port int) ibminet.Endpoint {
	return ibminet.Endpoint{Service: ibminet.ServiceRemoteCommand, Name: "as-rmtcmd-s", Host: "127.0.0.1", Port: port, TLS: true, Source: ibminet.EndpointExplicit}
}

func TestDialTLSVerified(t *testing.T) {
	ca := newTestCA(t)
	cert := ca.leaf(t, "ibmi.test", time.Now().Add(time.Hour))
	port, _ := startServer(t, &cert)
	d := &hostserver.Dialer{Timeout: 5 * time.Second, TLS: hostserver.TLSOptions{RootCAs: ca.pool, ServerName: "ibmi.test"}}
	c, err := d.DialEndpoint(context.Background(), tlsEndpoint(port))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, _, err := c.Exchange(context.Background(), request(1), expect1); err != nil {
		t.Fatalf("Exchange over TLS: %v", err)
	}
	// With no ServerName, the endpoint host (an IP SAN here) is verified.
	d.TLS.ServerName = ""
	c2, err := d.DialEndpoint(context.Background(), tlsEndpoint(port))
	if err != nil {
		t.Fatalf("IP-SAN verification: %v", err)
	}
	c2.Close()
}

func TestDialTLSFailuresNeverFallBack(t *testing.T) {
	ca := newTestCA(t)
	otherCA := newTestCA(t)
	good := ca.leaf(t, "ibmi.test", time.Now().Add(time.Hour))
	expired := ca.leaf(t, "ibmi.test", time.Now().Add(-time.Hour))
	goodPort, _ := startServer(t, &good)
	expiredPort, _ := startServer(t, &expired)
	plainPort, plainAccepted := startServer(t, nil)

	isUnknownAuthority := func(err error) bool { var e x509.UnknownAuthorityError; return errors.As(err, &e) }
	isHostname := func(err error) bool { var e x509.HostnameError; return errors.As(err, &e) }
	isExpired := func(err error) bool {
		var e x509.CertificateInvalidError
		return errors.As(err, &e) && e.Reason == x509.Expired
	}
	// A cleartext peer may answer garbage or nothing at all; either way the
	// handshake must fail rather than continue without TLS.
	anyFailure := func(err error) bool { return err != nil }
	for name, tc := range map[string]struct {
		port int
		opts hostserver.TLSOptions
		want func(error) bool
	}{
		"unknown CA":       {goodPort, hostserver.TLSOptions{RootCAs: otherCA.pool, ServerName: "ibmi.test"}, isUnknownAuthority},
		"wrong name":       {goodPort, hostserver.TLSOptions{RootCAs: ca.pool, ServerName: "other.test"}, isHostname},
		"expired":          {expiredPort, hostserver.TLSOptions{RootCAs: ca.pool, ServerName: "ibmi.test"}, isExpired},
		"cleartext server": {plainPort, hostserver.TLSOptions{RootCAs: ca.pool, ServerName: "ibmi.test"}, anyFailure},
	} {
		t.Run(name, func(t *testing.T) {
			d := &hostserver.Dialer{Timeout: 500 * time.Millisecond, TLS: tc.opts}
			c, err := d.DialEndpoint(context.Background(), tlsEndpoint(tc.port))
			if err == nil {
				c.Close()
				t.Fatal("DialEndpoint succeeded")
			}
			if c != nil || !tc.want(err) {
				t.Fatalf("error = %v (%T)", err, errors.Unwrap(err))
			}
		})
	}
	// The cleartext server saw exactly one connection: the failed TLS
	// attempt. No cleartext retry followed.
	time.Sleep(50 * time.Millisecond)
	if n := plainAccepted.Load(); n != 1 {
		t.Errorf("cleartext server accepted %d connections; want 1 (no fallback)", n)
	}
}

func TestSpoofedSecureMapperCannotRedirect(t *testing.T) {
	ca := newTestCA(t)
	attackerCA := newTestCA(t)
	attacker := attackerCA.leaf(t, "ibmi.test", time.Now().Add(time.Hour))
	attackerPort, _ := startServer(t, &attacker)

	// A spoofed mapper answers as-rmtcmd-s with the attacker's port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			buf := make([]byte, 32)
			_, _ = c.Read(buf)
			reply := []byte{'+', 0, 0, byte(attackerPort >> 8), byte(attackerPort)}
			_, _ = c.Write(reply)
			c.Close()
		}
	}()

	cat, err := (&servicemapper.Client{Timeout: 2 * time.Second}).Discover(context.Background(), servicemapper.DiscoverOptions{
		Host: "127.0.0.1", MapperPort: ln.Addr().(*net.TCPAddr).Port, Policy: ibminet.TransportTLSRequired,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !cat.RemoteCommand.TLS || cat.RemoteCommand.Port != attackerPort {
		t.Fatalf("catalog = %+v", cat.RemoteCommand)
	}
	d := &hostserver.Dialer{Timeout: 2 * time.Second, TLS: hostserver.TLSOptions{RootCAs: ca.pool, ServerName: "ibmi.test"}}
	if c, err := d.DialEndpoint(context.Background(), cat.RemoteCommand); err == nil || c != nil {
		t.Fatalf("redirected dial succeeded: %v", err)
	}
}

func TestTLSOptionsConfig(t *testing.T) {
	ca := newTestCA(t)
	cfg, err := hostserver.TLSOptions{RootCAs: ca.pool}.Config("ibmi.example")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.InsecureSkipVerify || cfg.ServerName != "ibmi.example" || cfg.MinVersion != tls.VersionTLS12 || cfg.RootCAs == ca.pool {
		t.Errorf("Config = skip %t, name %q, min 0x%04X, pool shared %t", cfg.InsecureSkipVerify, cfg.ServerName, cfg.MinVersion, cfg.RootCAs == ca.pool)
	}

	insecureSuite := tls.InsecureCipherSuites()[0].ID
	for name, opts := range map[string]hostserver.TLSOptions{
		"TLS 1.0 without AllowLegacy": {MinVersion: tls.VersionTLS10},
		"insecure suite without flag": {CipherSuites: []uint16{insecureSuite}},
		"unknown suite":               {CipherSuites: []uint16{0xFFFF}},
		"max below min":               {MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS11},
		"version below TLS 1.0":       {MinVersion: 0x0300, AllowLegacy: true},
	} {
		if _, err := opts.Config("ibmi.example"); !errors.Is(err, hostserver.ErrTLSOptions) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	if _, err := (hostserver.TLSOptions{}).Config(""); !errors.Is(err, hostserver.ErrTLSOptions) {
		t.Errorf("no server name: error = %v", err)
	}

	legacy, err := hostserver.TLSOptions{MinVersion: tls.VersionTLS10, CipherSuites: []uint16{insecureSuite}, AllowLegacy: true}.Config("ibmi.example")
	if err != nil {
		t.Fatalf("AllowLegacy config: %v", err)
	}
	if legacy.InsecureSkipVerify || legacy.MinVersion != tls.VersionTLS10 {
		t.Errorf("legacy config = %+v; it must still verify", legacy)
	}

	suites := []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256}
	cfg, _ = hostserver.TLSOptions{CipherSuites: suites}.Config("ibmi.example")
	suites[0] = 0
	if cfg.CipherSuites[0] != tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 {
		t.Error("Config shares the caller's cipher-suite slice")
	}
}

func TestDialRejectsInvalidEndpoint(t *testing.T) {
	ep := tlsEndpoint(1234)
	ep.TLS = false // name still says as-rmtcmd-s
	if _, err := (&hostserver.Dialer{}).DialEndpoint(context.Background(), ep); !errors.Is(err, ibminet.ErrInvalidEndpoint) {
		t.Errorf("error = %v", err)
	}
}

func TestDialCleartextUsesInjectedDialer(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	var dialed string
	d := &hostserver.Dialer{Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		dialed = network + " " + address
		return client, nil
	}}
	ep := ibminet.Endpoint{Service: ibminet.ServiceSignon, Name: "as-signon", Host: "ibmi.example", Port: 8476, Source: ibminet.EndpointServiceMapper}
	c, err := d.DialEndpoint(context.Background(), ep)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if dialed != "tcp ibmi.example:8476" {
		t.Errorf("dialed %q", dialed)
	}
}
