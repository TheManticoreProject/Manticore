package client

import (
	"context"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
)

// TestConnectPasswordLevel4 drives Sign-on and Remote Command at password
// level 4 against the scripted servers, which verify every substitute with
// the oracle-checked algorithm. Level 4 is not validated against a live
// system yet.
func TestConnectPasswordLevel4(t *testing.T) {
	for _, tc := range []struct {
		level     byte
		family    ibminet.PasswordFamily
		credType  byte
		seedAttrs byte
	}{
		{0, ibminet.PasswordFamilyDES, 0x01, 0x01},
		{4, ibminet.PasswordFamilySHA512, 0x07, 0x03},
	} {
		srv := newServer()
		srv.PasswordLevel = tc.level
		c := newClient(t, srv)
		if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
			t.Fatalf("level %d: %v", tc.level, err)
		}
		if info := c.SessionInfo(); info.PasswordFamily != tc.family || info.PasswordLevel != tc.level {
			t.Errorf("level %d: session %+v", tc.level, info)
		}
		for _, r := range srv.Requests() {
			switch r.RequestID {
			case 0x7004, 0x7002:
				if r.CredentialType != tc.credType {
					t.Errorf("level %d: 0x%04X credential type 0x%02X, want 0x%02X", tc.level, r.RequestID, r.CredentialType, tc.credType)
				}
				if want := map[byte]int{0x01: 8, 0x07: 64}[tc.credType]; len(r.Substitute) != want {
					t.Errorf("level %d: 0x%04X substitute of %d bytes", tc.level, r.RequestID, len(r.Substitute))
				}
			case 0x7001:
				if r.Attributes != tc.seedAttrs {
					t.Errorf("level %d: 0x7001 attributes 0x%02X, want 0x%02X", tc.level, r.Attributes, tc.seedAttrs)
				}
			}
		}
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
		if err := srv.WaitClosed(2 * time.Second); err != nil {
			t.Fatal(err)
		}
	}
}
