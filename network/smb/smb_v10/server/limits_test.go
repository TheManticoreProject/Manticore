package server

import (
	"errors"
	"net"
	"testing"
	"time"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// TestNewServerAppliesConnectionDefaults asserts the zero config bounds the
// connections served and how long they may stay quiet, and that a negative
// MaxConnections lifts the bound.
func TestNewServerAppliesConnectionDefaults(t *testing.T) {
	srv, err := NewServer(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if srv.config.MaxConnections != DefaultMaxConnections || cap(srv.slots) != DefaultMaxConnections {
		t.Errorf("MaxConnections = %d with %d slots, want %d", srv.config.MaxConnections, cap(srv.slots), DefaultMaxConnections)
	}
	if srv.config.HandshakeTimeout != DefaultHandshakeTimeout || srv.config.IdleTimeout != DefaultIdleTimeout {
		t.Errorf("HandshakeTimeout, IdleTimeout = %v, %v; want %v, %v",
			srv.config.HandshakeTimeout, srv.config.IdleTimeout, DefaultHandshakeTimeout, DefaultIdleTimeout)
	}

	unbounded, err := NewServer(Config{MaxConnections: -1})
	if err != nil {
		t.Fatal(err)
	}
	if unbounded.slots != nil {
		t.Error("a negative MaxConnections still bounds the connections")
	}
	if _, err := NewServer(Config{Timeout: -time.Second}); err == nil {
		t.Error("a negative Timeout was accepted")
	}
}

// TestConnectionReadTimeout asserts the bound on the next read follows what the
// connection holds: an unauthenticated one gets HandshakeTimeout, an idle
// authenticated one IdleTimeout, and one holding an open or a deferred request
// none, with Timeout capping all three.
func TestConnectionReadTimeout(t *testing.T) {
	const handshake, idle = time.Minute, 15 * time.Minute
	for _, tc := range []struct {
		name                          string
		handshake, idle, timeout      time.Duration
		authenticated, open, deferred bool
		want                          time.Duration
	}{
		{"unauthenticated", handshake, idle, 0, false, false, false, handshake},
		{"authenticated and idle", handshake, idle, 0, true, false, false, idle},
		{"authenticated with an open", handshake, idle, 0, true, true, false, 0},
		{"authenticated with a deferred request", handshake, idle, 0, true, false, true, 0},
		{"Timeout caps the handshake", handshake, idle, time.Second, false, false, false, time.Second},
		{"Timeout above the handshake", handshake, idle, time.Hour, false, false, false, handshake},
		{"Timeout alone bounds a busy connection", handshake, idle, time.Hour, true, true, false, time.Hour},
		{"handshake bound disabled", -1, idle, 0, false, false, false, 0},
		{"idle bound disabled", handshake, -1, 0, true, false, false, 0},
		{"idle bound disabled, Timeout kept", handshake, -1, time.Hour, true, false, false, time.Hour},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{config: Config{HandshakeTimeout: tc.handshake, IdleTimeout: tc.idle, Timeout: tc.timeout}}
			c := newConnection(srv, nil, nil)
			// A logon still in progress is not a session yet.
			c.pendingAuth[1] = nil
			if tc.authenticated {
				c.addSession(&Session{UID: 2})
			}
			if tc.open {
				c.opens[1] = &Open{}
			}
			if tc.deferred {
				c.responders = map[*asyncResponder]struct{}{{}: {}}
			}
			if got := c.readTimeout(); got != tc.want {
				t.Errorf("readTimeout() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestServerDropsUnauthenticatedAndIdleConnections asserts, end to end, that a
// connection that never logs on and one that logs on but holds nothing are
// closed once quiet for their bound, while one holding an open is kept.
func TestServerDropsUnauthenticatedAndIdleConnections(t *testing.T) {
	const bound = 300 * time.Millisecond

	config := conformanceConfig(SigningDisabled)
	config.HandshakeTimeout = bound
	config.IdleTimeout = bound
	srv, err := NewServer(config)
	if err != nil {
		t.Fatal(err)
	}
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("hello.txt", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := srv.AddShare(&Share{Name: fileShareName, Type: ShareTypeDisk, FS: fs}); err != nil {
		t.Fatal(err)
	}
	listener := listenLoopback(t)
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { srv.Close() })
	addr := listener.Addr().(*net.TCPAddr)

	// closedWithin reports whether the server closes conn within d.
	closedWithin := func(conn net.Conn, d time.Duration) bool {
		conn.SetReadDeadline(time.Now().Add(d))
		_, err := conn.Read(make([]byte, 1))
		var timeout net.Error
		return err != nil && !(errors.As(err, &timeout) && timeout.Timeout())
	}
	logon := func() *smb1client.Client {
		client := smb1client.NewClientUsingTCPTransport(addr.IP, addr.Port)
		if err := client.Connect(addr.IP, addr.Port); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { client.Disconnect() })
		creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := client.SessionSetup(creds); err != nil {
			t.Fatal(err)
		}
		if err := client.TreeConnect(fileShareName); err != nil {
			t.Fatal(err)
		}
		return client
	}

	silent, err := net.Dial("tcp", addr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer silent.Close()
	if !closedWithin(silent, 5*time.Second) {
		t.Error("a connection that never logged on was kept")
	}

	idle := logon()
	time.Sleep(3 * bound)
	if fid, err := idle.OpenFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE); err == nil {
		t.Errorf("an authenticated connection holding nothing was kept past IdleTimeout (opened FID 0x%04X)", fid)
	}

	busy := logon()
	fid, err := busy.OpenFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ,
		fileflags.FILE_OPEN, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * bound)
	if _, err := busy.ReadFile(fid, 0, 5); err != nil {
		t.Errorf("a connection holding an open was dropped while quiet: %v", err)
	}
}
