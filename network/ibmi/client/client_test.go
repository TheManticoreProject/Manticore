package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
)

const (
	testProfile  = "CANARYUSR"
	testPassword = "CANARYPW7"
)

func newServer() *hosttest.Server {
	return &hosttest.Server{Profile: testProfile, Password: testPassword}
}

func newClient(t *testing.T, srv *hosttest.Server, mutate ...func(*Config)) *Client {
	t.Helper()
	cfg := Config{
		Host:              "ibmi.test",
		Transport:         ibminet.TransportCleartextExplicit,
		SignonPort:        hosttest.SignonPort,
		RemoteCommandPort: hosttest.RemoteCommandPort,
		DataQueuePort:     hosttest.DataQueuePort,
		FilePort:          hosttest.FilePort,
		Dial:              srv.Dial,
		CleanupTimeout:    time.Second,
	}
	for _, m := range mutate {
		m(&cfg)
	}
	c, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func credential(t *testing.T, password string) *ibminet.PasswordCredential {
	t.Helper()
	cred, err := ibminet.NewPasswordCredential(testProfile, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	return cred
}

// waitGoroutines fails if the goroutine count does not return to base.
func waitGoroutines(t *testing.T, base int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > base {
		if time.Now().After(deadline) {
			t.Fatalf("goroutines leaked: %d running, %d before", runtime.NumGoroutine(), base)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestConnectAndClose(t *testing.T) {
	base := runtime.NumGoroutine()
	srv := newServer()
	srv.PasswordLevel = 2
	srv.RemoteCommandLevel = 10
	c := newClient(t, srv)
	cred := credential(t, testPassword)
	if err := c.Connect(context.Background(), cred); err != nil {
		t.Fatal(err)
	}
	cred.Destroy() // the client must not need the credential after Connect

	info := c.SessionInfo()
	if !info.Connected || info.VRM.String() != "V5R4M0" || info.PasswordLevel != 2 || info.PasswordFamily != ibminet.PasswordFamilySHA1 ||
		info.SystemCCSID != 37 || info.JobCCSID != 37 || info.RemoteCommandLevel != 10 || !info.Capabilities.ExplicitCommandCCSID ||
		info.ServerJob != "QZRCSRVS" || info.SignonEndpoint.Port != hosttest.SignonPort || info.RemoteEndpoint.Port != hosttest.RemoteCommandPort {
		t.Errorf("SessionInfo = %+v", info)
	}

	// Sign-on and Remote Command authenticate with independent seed pairs.
	var signonReq, rmtReq hosttest.Request
	for _, r := range srv.Requests() {
		switch r.RequestID {
		case 0x7003:
			signonReq.ClientSeed = r.ClientSeed
		case 0x7004:
			signonReq.Substitute = r.Substitute
		case 0x7001:
			rmtReq.ClientSeed = r.ClientSeed
		case 0x7002:
			rmtReq.Substitute = r.Substitute
		}
	}
	if signonReq.ClientSeed == rmtReq.ClientSeed || string(signonReq.Substitute) == string(rmtReq.Substitute) {
		t.Error("Sign-on and Remote Command reused a seed or substitute")
	}

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("second Close = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	if srv.Count(0x1004) != 1 || srv.Count(0x7006) != 1 {
		t.Errorf("end requests: 0x1004 x%d, 0x7006 x%d", srv.Count(0x1004), srv.Count(0x7006))
	}
	if c.SessionInfo().Connected {
		t.Error("SessionInfo still Connected after Close")
	}
	waitGoroutines(t, base)
}

func TestConnectOnlyOnce(t *testing.T) {
	srv := newServer()
	c := newClient(t, srv)
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.Connect(context.Background(), credential(t, testPassword)); !errors.Is(err, ErrAlreadyConnected) {
		t.Errorf("second Connect = %v", err)
	}
}

func TestFailedConnectClosesEverything(t *testing.T) {
	base := runtime.NumGoroutine()
	srv := newServer()
	c := newClient(t, srv)
	err := c.Connect(context.Background(), credential(t, "WRONGPW1"))
	if !errors.Is(err, signon.ErrAuthenticationRejected) {
		t.Fatalf("Connect = %v", err)
	}
	if srv.Count(0x7004) != 1 || srv.Count(0x7001) != 0 {
		t.Errorf("requests after rejection: %+v", srv.Requests())
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	if err := c.Connect(context.Background(), credential(t, testPassword)); !errors.Is(err, ErrAlreadyConnected) {
		t.Errorf("Connect after failure = %v; a new Client is required", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close after failed Connect = %v", err)
	}
	for _, leak := range []string{"CANARY", "WRONGPW1"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("error leaks %q", leak)
		}
	}
	waitGoroutines(t, base)
}

func TestRemoteCommandFailureClosesSockets(t *testing.T) {
	srv := newServer()
	srv.StartRC = 0x00020002 // profile disabled at start-server
	c := newClient(t, srv)
	err := c.Connect(context.Background(), credential(t, testPassword))
	var se *signon.Error
	if !errors.As(err, &se) || se.Reason() != signon.ReasonProfileDisabled {
		t.Fatalf("Connect = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestConnectCancellation(t *testing.T) {
	srv := newServer()
	srv.SilentOn = 0x1001
	c := newClient(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := c.Connect(ctx, credential(t, testPassword)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Connect = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestSessionInfoIsASnapshot(t *testing.T) {
	srv := newServer()
	c := newClient(t, srv, func(cfg *Config) { cfg.JobCCSIDOverride = 1140 })
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	info := c.SessionInfo()
	if info.JobCCSID != 1140 || info.NegotiatedJobCCSID != 37 || len(info.Warnings) != 1 {
		t.Fatalf("info = %+v", info)
	}
	info.Warnings[0] = "mutated"
	if c.SessionInfo().Warnings[0] == "mutated" {
		t.Error("SessionInfo aliases client state")
	}
	out := fmt.Sprintf("%+v %#v", info, info)
	if strings.Contains(out, "CANARY") {
		t.Errorf("SessionInfo leaks the profile: %s", out)
	}
}

func TestDiscoveryThroughMapper(t *testing.T) {
	srv := newServer()
	mapped := 0
	dial := func(ctx context.Context, network, address string) (net.Conn, error) {
		if !strings.HasSuffix(address, ":449") {
			return srv.Dial(ctx, network, address)
		}
		mapped++
		client, server := net.Pipe()
		go func() {
			defer server.Close()
			buf := make([]byte, 16)
			n, _ := server.Read(buf)
			port := map[string]uint32{"as-signon": hosttest.SignonPort, "as-rmtcmd": hosttest.RemoteCommandPort}[string(buf[:n])]
			_, _ = server.Write(binary.BigEndian.AppendUint32([]byte{'+'}, port))
		}()
		return client, nil
	}
	c := newClient(t, srv, func(cfg *Config) { cfg.SignonPort, cfg.RemoteCommandPort, cfg.Dial = 0, 0, dial })
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	info := c.SessionInfo()
	if mapped != 2 || info.SignonEndpoint.Source != ibminet.EndpointServiceMapper || info.RemoteEndpoint.Port != hosttest.RemoteCommandPort {
		t.Errorf("mapper lookups %d, info %+v", mapped, info)
	}
}

func TestConfigValidation(t *testing.T) {
	ok := Config{Host: "ibmi.test", Transport: ibminet.TransportCleartextExplicit}
	c, err := New(ok)
	if err != nil {
		t.Fatal(err)
	}
	if c.cfg.MapperPort != 449 || c.cfg.DialTimeout != DefaultDialTimeout || c.cfg.OperationTimeout != DefaultOperationTimeout ||
		c.cfg.CleanupTimeout != DefaultCleanupTimeout || c.cfg.MaxFrame != 16<<20 || c.cfg.NLV != "2924" {
		t.Errorf("defaults = %+v", c.cfg)
	}
	for name, mutate := range map[string]func(*Config){
		"no host":                 func(c *Config) { c.Host = "" },
		"unset transport":         func(c *Config) { c.Transport = ibminet.TransportUnset },
		"mapper port":             func(c *Config) { c.MapperPort = 70000 },
		"negative signon port":    func(c *Config) { c.SignonPort = -1 },
		"negative dial timeout":   func(c *Config) { c.DialTimeout = -time.Second },
		"huge operation timeout":  func(c *Config) { c.OperationTimeout = 2 * time.Hour },
		"cleanup above operation": func(c *Config) { c.OperationTimeout, c.CleanupTimeout = time.Second, 2*time.Second },
		"max frame":               func(c *Config) { c.MaxFrame = 10 },
		"NLV":                     func(c *Config) { c.NLV = "29x4" },
		"job CCSID override":      func(c *Config) { c.JobCCSIDOverride = 273 },
		"legacy TLS without flag": func(c *Config) { c.Transport = ibminet.TransportTLSRequired; c.TLS.MinVersion = 0x0301 },
	} {
		cfg := ok
		mutate(&cfg)
		if _, err := New(cfg); !errors.Is(err, ErrConfig) {
			t.Errorf("%s: New = %v", name, err)
		}
	}
}

func TestCloseBeforeConnect(t *testing.T) {
	c := newClient(t, newServer())
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := c.Connect(context.Background(), credential(t, testPassword)); !errors.Is(err, ErrAlreadyConnected) {
		t.Errorf("Connect after Close = %v", err)
	}
}

func connected(t *testing.T) (*Client, *hosttest.Server) {
	t.Helper()
	srv := newServer()
	c := newClient(t, srv)
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	return c, srv
}

func TestExclusiveSerializes(t *testing.T) {
	c, _ := connected(t)
	defer c.Close()
	counter := 0 // unsynchronized on purpose: -race flags any overlap
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.exclusive(context.Background(), func(r *reservation) error {
				v := counter
				time.Sleep(time.Millisecond)
				counter = v + 1
				return nil
			})
		}()
	}
	wg.Wait()
	if counter != 20 {
		t.Errorf("counter = %d; exclusive sections interleaved", counter)
	}
}

func TestCloseWaitsForExclusive(t *testing.T) {
	c, srv := connected(t)
	entered := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		_ = c.exclusive(context.Background(), func(r *reservation) error {
			close(entered)
			time.Sleep(100 * time.Millisecond)
			if s, err := r.Session(); err != nil || s.Err() != nil {
				t.Errorf("session closed underneath an exclusive section: %v", err)
			}
			close(finished)
			return nil
		})
	}()
	<-entered
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
	default:
		t.Fatal("Close returned before the exclusive section finished")
	}
	if err := c.exclusive(context.Background(), func(*reservation) error { return nil }); !errors.Is(err, ErrNotReady) {
		t.Errorf("exclusive after Close = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestExclusivePanicReleasesLock(t *testing.T) {
	c, _ := connected(t)
	defer c.Close()
	func() {
		defer func() {
			if r := recover(); r != "boom" {
				t.Errorf("recovered %v; the panic must propagate unchanged", r)
			}
		}()
		_ = c.exclusive(context.Background(), func(*reservation) error { panic("boom") })
	}()
	if c.getState() != stateReady {
		t.Errorf("state after panic = %s", c.getState())
	}
	if err := c.exclusive(context.Background(), func(*reservation) error { return nil }); err != nil {
		t.Errorf("lock not released after panic: %v", err)
	}
}

func TestReservationUseAfterReturn(t *testing.T) {
	c, _ := connected(t)
	defer c.Close()
	var kept *reservation
	if err := c.exclusive(context.Background(), func(r *reservation) error {
		kept = r
		_, err := r.Session()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := kept.Session(); !errors.Is(err, ErrReservationReleased) {
		t.Errorf("use after return = %v", err)
	}
}

func TestExclusiveHonorsContextWhileWaiting(t *testing.T) {
	c, _ := connected(t)
	defer c.Close()
	release := make(chan struct{})
	held := make(chan struct{})
	go func() {
		_ = c.exclusive(context.Background(), func(*reservation) error { close(held); <-release; return nil })
	}()
	<-held
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := c.exclusive(ctx, func(*reservation) error { return nil }); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("waiting exclusive = %v", err)
	}
	close(release)
}
