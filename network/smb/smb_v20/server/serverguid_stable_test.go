package server

import (
	"bytes"
	"testing"

	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
)

// TestServerGUIDStableAndSharedWithSMB1 checks the SMB2 server's auto-generated
// ServerGUID is stable across restarts and equals the shared host-derived value
// the SMB1 server uses, so the two servers present one consistent identity rather
// than a fresh GUID each start.
func TestServerGUIDStableAndSharedWithSMB1(t *testing.T) {
	s1, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	s2, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if s1.config.ServerGUID != s2.config.ServerGUID {
		t.Fatalf("SMB2 ServerGUID changed across restarts: % x vs % x", s1.config.ServerGUID, s2.config.ServerGUID)
	}

	shared := smb1.StableServerGUID()
	want := shared.ToBytes()
	if !bytes.Equal(s1.config.ServerGUID[:], want) {
		t.Fatalf("SMB2 ServerGUID % x does not match the shared host GUID % x", s1.config.ServerGUID, want)
	}
}
