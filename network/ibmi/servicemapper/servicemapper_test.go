package servicemapper_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/servicemapper"
)

// scriptedMapper answers each connection from a reply table keyed by the
// exact request bytes, stays silent for unknown requests, and records every
// request.
type scriptedMapper struct {
	ln       net.Listener
	replies  map[string][]byte
	mu       sync.Mutex
	requests []string
	wg       sync.WaitGroup
}

func newScriptedMapper(t *testing.T, replies map[string][]byte) *scriptedMapper {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	m := &scriptedMapper{ln: ln, replies: replies}
	m.wg.Add(1)
	go m.serve()
	t.Cleanup(func() { ln.Close(); m.wg.Wait() })
	return m
}

func (m *scriptedMapper) serve() {
	defer m.wg.Done()
	for {
		c, err := m.ln.Accept()
		if err != nil {
			return
		}
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			defer c.Close()
			_ = c.SetDeadline(time.Now().Add(5 * time.Second))
			buf := make([]byte, 64)
			n, _ := c.Read(buf)
			req := string(buf[:n])
			m.mu.Lock()
			m.requests = append(m.requests, req)
			m.mu.Unlock()
			if reply, ok := m.replies[req]; ok {
				_, _ = c.Write(reply)
				return
			}
			// Unknown request: stay silent until the client gives up.
			_, _ = c.Read(buf)
		}()
	}
}

func (m *scriptedMapper) port() int { return m.ln.Addr().(*net.TCPAddr).Port }

func (m *scriptedMapper) seen() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.requests...)
}

func portReply(port uint32) []byte {
	return []byte{'+', byte(port >> 24), byte(port >> 16), byte(port >> 8), byte(port)}
}

func TestEncodeRequestByteExact(t *testing.T) {
	for _, name := range []string{"as-signon", "as-signon-s", "as-rmtcmd", "as-rmtcmd-s"} {
		got, err := servicemapper.EncodeRequest(name)
		if err != nil || !bytes.Equal(got, []byte(name)) {
			t.Errorf("EncodeRequest(%q) = %q, %v", name, got, err)
		}
	}
	for _, name := range []string{"", "as-file", "as-signon\n", "AS-SIGNON", "as-signon\x00"} {
		if _, err := servicemapper.EncodeRequest(name); !errors.Is(err, servicemapper.ErrUnsupportedService) {
			t.Errorf("EncodeRequest(%q) error = %v", name, err)
		}
	}
}

func TestDecodeReply(t *testing.T) {
	if port, err := servicemapper.DecodeReply("as-rmtcmd", []byte{0x2B, 0x00, 0x00, 0x22, 0x23}); err != nil || port != 8739 {
		t.Errorf("DecodeReply = %d, %v; want 8739", port, err)
	}
	if port, err := servicemapper.DecodeReply("as-rmtcmd", portReply(65535)); err != nil || port != 65535 {
		t.Errorf("DecodeReply(65535) = %d, %v", port, err)
	}
	var le *servicemapper.LookupError
	if _, err := servicemapper.DecodeReply("as-rmtcmd", []byte{'-'}); !errors.Is(err, servicemapper.ErrNotFound) || !errors.As(err, &le) || le.Status != '-' {
		t.Errorf("'-' reply error = %v", err)
	}
	for name, b := range map[string][]byte{
		"empty":      nil,
		"short":      {'+', 0, 0, 1},
		"long":       {'+', 0, 0, 1, 0, 9},
		"port 0":     portReply(0),
		"port 65536": portReply(65536),
		"port max":   portReply(0xFFFFFFFF),
	} {
		if _, err := servicemapper.DecodeReply("as-rmtcmd", b); !errors.Is(err, servicemapper.ErrProtocol) {
			t.Errorf("%s: error = %v; want ErrProtocol", name, err)
		}
	}
}

func TestLookup(t *testing.T) {
	m := newScriptedMapper(t, map[string][]byte{
		"as-rmtcmd":   portReply(8475),
		"as-signon":   {'-'},
		"as-rmtcmd-s": {'+', 0, 0},
	})
	c := &servicemapper.Client{Timeout: 5 * time.Second}
	ctx := context.Background()
	if port, err := c.Lookup(ctx, "127.0.0.1", m.port(), "as-rmtcmd"); err != nil || port != 8475 {
		t.Fatalf("Lookup = %d, %v", port, err)
	}
	if _, err := c.Lookup(ctx, "127.0.0.1", m.port(), "as-signon"); !errors.Is(err, servicemapper.ErrNotFound) {
		t.Errorf("not-found error = %v", err)
	}
	if _, err := c.Lookup(ctx, "127.0.0.1", m.port(), "as-rmtcmd-s"); !errors.Is(err, servicemapper.ErrProtocol) {
		t.Errorf("truncated reply error = %v", err)
	}
	if _, err := c.Lookup(ctx, "127.0.0.1", m.port(), "as-file"); !errors.Is(err, servicemapper.ErrUnsupportedService) {
		t.Errorf("unsupported name error = %v", err)
	}
	got := m.seen()
	want := map[string]bool{"as-rmtcmd": true, "as-signon": true, "as-rmtcmd-s": true}
	if len(got) != 3 {
		t.Fatalf("mapper saw %d connections %q; want 3 (one per lookup, none for the rejected name)", len(got), got)
	}
	for _, r := range got {
		if !want[r] {
			t.Errorf("mapper received %q; requests must be the bare ASCII name", r)
		}
	}
}

func TestLookupTimesOutOnSilentMapper(t *testing.T) {
	m := newScriptedMapper(t, map[string][]byte{}) // never answers
	c := &servicemapper.Client{Timeout: 200 * time.Millisecond}
	start := time.Now()
	_, err := c.Lookup(context.Background(), "127.0.0.1", m.port(), "as-rmtcmd")
	if !errors.Is(err, servicemapper.ErrProtocol) || !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("silent mapper error = %v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Error("lookup ignored its timeout")
	}
}

func TestLookupHonorsCancellation(t *testing.T) {
	m := newScriptedMapper(t, map[string][]byte{})
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)
	_, err := (&servicemapper.Client{}).Lookup(ctx, "127.0.0.1", m.port(), "as-rmtcmd")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("cancelled lookup error = %v", err)
	}
}

func TestDiscover(t *testing.T) {
	m := newScriptedMapper(t, map[string][]byte{
		"as-signon":   portReply(8476),
		"as-rmtcmd":   portReply(8475),
		"as-signon-s": portReply(9476),
		"as-rmtcmd-s": portReply(9475),
	})
	c := &servicemapper.Client{Timeout: 5 * time.Second}
	ctx := context.Background()

	cat, err := c.Discover(ctx, servicemapper.DiscoverOptions{Host: "127.0.0.1", MapperPort: m.port(), Policy: ibminet.TransportTLSRequired})
	if err != nil {
		t.Fatal(err)
	}
	for _, ep := range []ibminet.Endpoint{cat.Signon, cat.RemoteCommand} {
		if !ep.TLS || ep.Source != ibminet.EndpointServiceMapper || ep.Validate() != nil {
			t.Errorf("TLS discovery endpoint = %+v", ep)
		}
	}
	if cat.Signon.Port != 9476 || cat.RemoteCommand.Port != 9475 || len(cat.Missing) != 0 {
		t.Errorf("TLS catalog = %+v", cat)
	}

	cat, err = c.Discover(ctx, servicemapper.DiscoverOptions{Host: "127.0.0.1", MapperPort: m.port(), Policy: ibminet.TransportCleartextExplicit})
	if err != nil || cat.Signon.Port != 8476 || cat.RemoteCommand.Port != 8475 || cat.Signon.TLS || cat.RemoteCommand.TLS {
		t.Errorf("cleartext catalog = %+v, %v", cat, err)
	}

	if _, err := c.Discover(ctx, servicemapper.DiscoverOptions{Host: "127.0.0.1", MapperPort: m.port()}); !errors.Is(err, ibminet.ErrTransportPolicy) {
		t.Errorf("unset policy error = %v", err)
	}
}

func TestDiscoverExplicitPortSkipsOnlyItsLookup(t *testing.T) {
	m := newScriptedMapper(t, map[string][]byte{"as-rmtcmd-s": portReply(9475)})
	c := &servicemapper.Client{Timeout: 5 * time.Second}
	cat, err := c.Discover(context.Background(), servicemapper.DiscoverOptions{
		Host: "127.0.0.1", MapperPort: m.port(), Policy: ibminet.TransportTLSRequired, SignonPort: 7000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cat.Signon.Port != 7000 || cat.Signon.Source != ibminet.EndpointExplicit || !cat.Signon.TLS || cat.Signon.Name != "as-signon-s" {
		t.Errorf("explicit Signon = %+v; the port override must not change the TLS policy", cat.Signon)
	}
	if cat.RemoteCommand.Port != 9475 || cat.RemoteCommand.Source != ibminet.EndpointServiceMapper {
		t.Errorf("RemoteCommand = %+v", cat.RemoteCommand)
	}
	if got := m.seen(); len(got) != 1 || got[0] != "as-rmtcmd-s" {
		t.Errorf("mapper requests = %q; want only as-rmtcmd-s", got)
	}
}

func TestDiscoverNeverFallsBackToCleartext(t *testing.T) {
	// Only cleartext services are registered; secure lookups fail.
	m := newScriptedMapper(t, map[string][]byte{
		"as-signon":   portReply(8476),
		"as-rmtcmd":   portReply(8475),
		"as-signon-s": {'-'},
		"as-rmtcmd-s": {'-'},
	})
	cat, err := (&servicemapper.Client{Timeout: 5 * time.Second}).Discover(context.Background(), servicemapper.DiscoverOptions{
		Host: "127.0.0.1", MapperPort: m.port(), Policy: ibminet.TransportTLSRequired,
	})
	if !errors.Is(err, servicemapper.ErrNotFound) {
		t.Fatalf("error = %v; want ErrNotFound", err)
	}
	if len(cat.Missing) != 2 || cat.Signon.Port != 0 || cat.RemoteCommand.Port != 0 {
		t.Errorf("catalog = %+v; failed TLS lookups must leave services missing", cat)
	}
	for _, r := range m.seen() {
		if r == "as-signon" || r == "as-rmtcmd" {
			t.Errorf("discovery queried cleartext service %q under TLSRequired", r)
		}
	}
}

func TestDiscoverPartialCatalog(t *testing.T) {
	m := newScriptedMapper(t, map[string][]byte{"as-signon": portReply(8476), "as-rmtcmd": {'-'}})
	cat, err := (&servicemapper.Client{Timeout: 5 * time.Second}).Discover(context.Background(), servicemapper.DiscoverOptions{
		Host: "127.0.0.1", MapperPort: m.port(), Policy: ibminet.TransportCleartextExplicit,
	})
	if !errors.Is(err, servicemapper.ErrNotFound) {
		t.Fatalf("error = %v", err)
	}
	if _, ok := cat.Endpoint(ibminet.ServiceSignon); !ok || cat.Signon.Port != 8476 {
		t.Errorf("Signon = %+v", cat.Signon)
	}
	if len(cat.Missing) != 1 || cat.Missing[0] != ibminet.ServiceRemoteCommand {
		t.Errorf("Missing = %v", cat.Missing)
	}
}

func TestLookupUsesInjectedDialer(t *testing.T) {
	client, server := net.Pipe()
	var dialed string
	c := &servicemapper.Client{Timeout: 5 * time.Second, Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		dialed = address
		return client, nil
	}}
	go func() {
		buf := make([]byte, len("as-signon-s"))
		_, _ = io.ReadFull(server, buf)
		_, _ = server.Write(portReply(9476))
		server.Close()
	}()
	port, err := c.Lookup(context.Background(), "ibmi.example", 0, "as-signon-s")
	if err != nil || port != 9476 {
		t.Fatalf("Lookup = %d, %v", port, err)
	}
	if dialed != "ibmi.example:"+strconv.Itoa(servicemapper.DefaultPort) {
		t.Errorf("dialed %q; want the default mapper port", dialed)
	}
}

func TestDiscoverSelectedServices(t *testing.T) {
	m := newScriptedMapper(t, map[string][]byte{"as-dtaq": portReply(8472)})
	c := &servicemapper.Client{Timeout: 5 * time.Second}
	cat, err := c.Discover(context.Background(), servicemapper.DiscoverOptions{
		Host: "127.0.0.1", MapperPort: m.port(), Policy: ibminet.TransportCleartextExplicit,
		Services: []ibminet.Service{ibminet.ServiceDataQueue},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ep, ok := cat.Endpoint(ibminet.ServiceDataQueue); !ok || ep.Port != 8472 || ep.Name != "as-dtaq" {
		t.Errorf("DataQueue = %+v", cat.DataQueue)
	}
	if got := m.seen(); len(got) != 1 || got[0] != "as-dtaq" {
		t.Errorf("mapper requests = %q; want only as-dtaq", got)
	}
	cat, err = c.Discover(context.Background(), servicemapper.DiscoverOptions{
		Host: "127.0.0.1", MapperPort: m.port(), Policy: ibminet.TransportTLSRequired,
		Services: []ibminet.Service{ibminet.ServiceDataQueue}, DataQueuePort: 9472,
	})
	if err != nil || cat.DataQueue.Port != 9472 || cat.DataQueue.Name != "as-dtaq-s" || cat.DataQueue.Source != ibminet.EndpointExplicit {
		t.Errorf("explicit DataQueue = %+v, %v", cat.DataQueue, err)
	}
	if _, err := c.Discover(context.Background(), servicemapper.DiscoverOptions{
		Host: "127.0.0.1", MapperPort: m.port(), Policy: ibminet.TransportCleartextExplicit, Services: []ibminet.Service{9},
	}); err == nil {
		t.Error("an unsupported service was accepted")
	}
}
