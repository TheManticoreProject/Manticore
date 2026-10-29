package server

import (
	"testing"
	"time"
)

func TestNewServerAppliesConnectionDefaults(t *testing.T) {
	s, err := NewServer(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if s.config.MaxConnections != DefaultMaxConnections || cap(s.slots) != DefaultMaxConnections {
		t.Errorf("MaxConnections = %d with %d slots, want %d", s.config.MaxConnections, cap(s.slots), DefaultMaxConnections)
	}
	if s.config.HandshakeTimeout != DefaultHandshakeTimeout || s.config.IdleTimeout != DefaultIdleTimeout {
		t.Errorf("HandshakeTimeout, IdleTimeout = %v, %v; want %v, %v", s.config.HandshakeTimeout, s.config.IdleTimeout, DefaultHandshakeTimeout, DefaultIdleTimeout)
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

func TestConnectionReadTimeout(t *testing.T) {
	const handshake, idle = time.Minute, 15 * time.Minute
	for _, tc := range []struct {
		name                     string
		handshake, idle, timeout time.Duration
		established, holdsAnOpen bool
		want                     time.Duration
	}{
		{"unauthenticated", handshake, idle, 0, false, false, handshake},
		{"authenticated and idle", handshake, idle, 0, true, false, idle},
		{"authenticated with an open", handshake, idle, 0, true, true, 0},
		{"Timeout caps the handshake", handshake, idle, time.Second, false, false, time.Second},
		{"Timeout above the handshake", handshake, idle, time.Hour, false, false, handshake},
		{"Timeout alone bounds a busy connection", handshake, idle, time.Hour, true, true, time.Hour},
		{"handshake bound disabled", -1, idle, 0, false, false, 0},
		{"idle bound disabled", handshake, -1, 0, true, false, 0},
		{"idle bound disabled, Timeout kept", handshake, -1, time.Hour, true, false, time.Hour},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{config: Config{HandshakeTimeout: tc.handshake, IdleTimeout: tc.idle, Timeout: tc.timeout}}
			c := newConnection(s, nil, nil)
			if tc.established {
				c.sessions[1] = &Session{established: true}
			} else {
				c.sessions[1] = &Session{} // still authenticating
			}
			if tc.holdsAnOpen {
				c.opens[1] = &Open{}
			}
			if got := c.readTimeout(); got != tc.want {
				t.Errorf("readTimeout() = %v, want %v", got, tc.want)
			}
		})
	}
}
