package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/api/users"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
)

// faultConn cuts a connection after a byte budget: the write or read that
// crosses it moves only the bytes up to the budget, then the connection is
// closed and an error returned.
type faultConn struct {
	net.Conn
	mu        sync.Mutex
	failWrite bool
	budget    int // -1: unlimited
	written   int
	read      int
	broken    bool
}

var errInjected = errors.New("injected connection failure")

func (c *faultConn) cut(n int, count *int) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.broken {
		return 0, true
	}
	if c.budget < 0 || *count+n <= c.budget {
		*count += n
		return n, false
	}
	allowed := c.budget - *count
	*count = c.budget
	c.broken = true
	return allowed, true
}

func (c *faultConn) Write(b []byte) (int, error) {
	if !c.failWrite {
		n, err := c.Conn.Write(b)
		c.mu.Lock()
		c.written += n
		c.mu.Unlock()
		return n, err
	}
	allowed, fail := c.cut(len(b), &c.written)
	if !fail {
		return c.Conn.Write(b)
	}
	n, _ := c.Conn.Write(b[:allowed])
	c.Conn.Close()
	return n, errInjected
}

func (c *faultConn) Read(b []byte) (int, error) {
	if c.failWrite {
		n, err := c.Conn.Read(b)
		c.mu.Lock()
		c.read += n
		c.mu.Unlock()
		return n, err
	}
	c.mu.Lock()
	if c.broken {
		c.mu.Unlock()
		return 0, errInjected
	}
	if c.budget >= 0 && c.read+len(b) > c.budget {
		b = b[:c.budget-c.read]
	}
	c.mu.Unlock()
	if len(b) == 0 {
		c.mu.Lock()
		c.broken = true
		c.mu.Unlock()
		c.Conn.Close()
		return 0, errInjected
	}
	n, err := c.Conn.Read(b)
	c.mu.Lock()
	c.read += n
	c.mu.Unlock()
	return n, err
}

// scenario is one full read-only session with a fault on one service.
type scenario struct {
	port      int
	failWrite bool
	budget    int
}

type scenarioResult struct {
	errs          []error
	written, read int
	srv           *hosttest.Server
}

func runScenario(t *testing.T, sc scenario) scenarioResult {
	t.Helper()
	srv := newServer()
	q := &qgyWireServer{n: 7}
	srv.Handle = q.handle
	var fc *faultConn
	dial := func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := srv.Dial(ctx, network, address)
		if err != nil {
			return nil, err
		}
		_, p, _ := net.SplitHostPort(address)
		if port, _ := strconv.Atoi(p); port == sc.port {
			fc = &faultConn{Conn: conn, failWrite: sc.failWrite, budget: sc.budget}
			return fc, nil
		}
		return conn, nil
	}
	c := newClient(t, srv, func(cfg *Config) {
		cfg.Dial = dial
		cfg.OperationTimeout = 2 * time.Second
		cfg.CleanupTimeout = time.Second
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var res scenarioResult
	res.srv = srv
	record := func(err error) {
		if err != nil {
			res.errs = append(res.errs, err)
		}
	}
	if err := c.Connect(ctx, credential(t, testPassword)); err != nil {
		record(err)
	} else {
		_, err := c.RunCommand(ctx, "DSPJOB", rmtcmd.RunCommandOptions{})
		record(err)
		info := c.SessionInfo()
		u, err := users.New(c, users.Options{SystemCCSID: info.SystemCCSID, JobCCSID: info.JobCCSID})
		record(err)
		if u != nil {
			_, err = u.List(ctx, users.ListOptions{PageSize: 4})
			record(err)
		}
	}
	record(c.Close())
	if ctx.Err() != nil {
		t.Fatalf("%+v: the scenario hung", sc)
	}
	if fc != nil {
		res.written, res.read = fc.written, fc.read
	}
	return res
}

// TestFaultInjectionAtEveryByte cuts the Sign-on or Remote Command
// connection after every possible number of written or read bytes. Every
// cut must surface as an error (never a panic or hang), every socket must be
// closed, no goroutine may leak, and no error may reveal the credential.
func TestFaultInjectionAtEveryByte(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive fault injection")
	}
	base := runtime.NumGoroutine()
	for _, port := range []int{hosttest.SignonPort, hosttest.RemoteCommandPort} {
		clean := runScenario(t, scenario{port: port, budget: -1})
		if len(clean.errs) != 0 {
			t.Fatalf("port %d baseline: %v", port, clean.errs)
		}
		for _, failWrite := range []bool{true, false} {
			total := clean.read
			if failWrite {
				total = clean.written
			}
			for k := 0; k < total; k++ {
				sc := scenario{port: port, failWrite: failWrite, budget: k}
				res := runScenario(t, sc)
				name := fmt.Sprintf("port %d, %s cut at byte %d of %d", port, map[bool]string{true: "write", false: "read"}[failWrite], k, total)
				// The Sign-on 0x7006 end request (the last 20 written bytes)
				// is one-way and best effort: the credential is already
				// validated, so cutting it must not fail the session.
				endServer := port == hosttest.SignonPort && failWrite && k >= total-20
				switch {
				case endServer && len(res.errs) != 0:
					t.Errorf("%s: a failed best-effort 0x7006 failed the session: %v", name, res.errs)
				case !endServer && len(res.errs) == 0:
					t.Errorf("%s: no error surfaced", name)
				}
				if err := res.srv.WaitClosed(2 * time.Second); err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				for _, err := range res.errs {
					if s := err.Error(); strings.Contains(s, "CANARY") {
						t.Fatalf("%s: error reveals the credential: %s", name, s)
					}
				}
			}
			t.Logf("port %d: %d %s boundaries", port, total, map[bool]string{true: "write", false: "read"}[failWrite])
		}
	}
	waitGoroutines(t, base)
}
