package client

import (
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb"
)

func TestEngineSupportsSMB2(t *testing.T) {
	supported := []smb.SMBProtocolVersion{
		smb.SMB_VERSION_2_0, smb.SMB_VERSION_2_0_2, smb.SMB_VERSION_2_1,
		smb.SMB_VERSION_3_0, smb.SMB_VERSION_3_0_2, smb.SMB_VERSION_3_1_1,
	}
	for _, v := range supported {
		if !engineSupportsSMB2(v) {
			t.Errorf("engineSupportsSMB2(%s) = false, want true", v)
		}
	}
	if engineSupportsSMB2(smb.SMB_VERSION_1_0) {
		t.Errorf("engineSupportsSMB2(SMB1) = true, want false")
	}
}

func TestWantedFamilies(t *testing.T) {
	cases := []struct {
		name                             string
		prefs                            []smb.SMBProtocolVersion
		wantSMB1, wantSMB2, needWildcard bool
	}{
		{"empty", nil, false, false, false},
		{"smb1 only", []smb.SMBProtocolVersion{smb.SMB_VERSION_1_0}, true, false, false},
		{"smb2 2.0.2 only", []smb.SMBProtocolVersion{smb.SMB_VERSION_2_0_2}, false, true, false},
		{"both smb1+2.0.2", []smb.SMBProtocolVersion{smb.SMB_VERSION_1_0, smb.SMB_VERSION_2_0_2}, true, true, false},
		{"3.1.1 needs wildcard", []smb.SMBProtocolVersion{smb.SMB_VERSION_3_1_1}, false, true, true},
		{"2.1 needs wildcard", []smb.SMBProtocolVersion{smb.SMB_VERSION_2_1}, false, true, true},
		{"2.0.2+3.0 needs wildcard", []smb.SMBProtocolVersion{smb.SMB_VERSION_2_0_2, smb.SMB_VERSION_3_0}, false, true, true},
		{"default preference -> all true", defaultPreference(), true, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s1, s2, wc := wantedFamilies(c.prefs)
			if s1 != c.wantSMB1 || s2 != c.wantSMB2 || wc != c.needWildcard {
				t.Errorf("wantedFamilies(%v) = (%v,%v,%v), want (%v,%v,%v)",
					c.prefs, s1, s2, wc, c.wantSMB1, c.wantSMB2, c.needWildcard)
			}
		})
	}
}

func TestOptionsDialTimeout(t *testing.T) {
	cases := []struct {
		name string
		in   time.Duration
		want time.Duration
	}{
		{"zero applies default", 0, DefaultDialTimeout},
		{"negative disables", -1, 0},
		{"explicit value kept", 3 * time.Second, 3 * time.Second},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := (Options{DialTimeout: c.in}).dialTimeout(); got != c.want {
				t.Errorf("Options{DialTimeout: %v}.dialTimeout() = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

// TestDialStrictOrderFailsOnSilentServer reproduces the scenario of a server
// that accepts the TCP connection but never answers a negotiate probe (e.g. an
// SMB1-only host receiving an SMB2-framed request). Dial must fail within the
// per-attempt bound instead of blocking forever.
func TestDialStrictOrderFailsOnSilentServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	// Accept every probe connection but never write a response.
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			defer c.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	done := make(chan error, 1)
	go func() {
		_, err := Dial(addr.IP.String(), addr.Port, Options{
			Preferred:   []smb.SMBProtocolVersion{smb.SMB_VERSION_2_0_2, smb.SMB_VERSION_1_0},
			Policy:      PolicyStrictOrder,
			DialTimeout: 200 * time.Millisecond,
		})
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Dial should fail against a server that never answers, got nil error")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Dial blocked past the per-attempt bound against a silent server")
	}
}

// TestDialHighestInSetFailsOnSilentServer is the PolicyHighestInSet variant:
// the single multi-protocol negotiate must also fail within the bound.
func TestDialHighestInSetFailsOnSilentServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			defer c.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	done := make(chan error, 1)
	go func() {
		_, err := Dial(addr.IP.String(), addr.Port, Options{
			Preferred:   []smb.SMBProtocolVersion{smb.SMB_VERSION_2_0_2, smb.SMB_VERSION_1_0},
			Policy:      PolicyHighestInSet,
			DialTimeout: 200 * time.Millisecond,
		})
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Dial should fail against a server that never answers, got nil error")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Dial blocked past the per-attempt bound against a silent server")
	}
}
