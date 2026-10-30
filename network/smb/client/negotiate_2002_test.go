package client

import (
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	smb2dialects "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	smb2server "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
)

// startSMB2Server serves the in-repo SMB2 server offering exactly offered on a
// loopback port and returns that port.
func startSMB2Server(t *testing.T, offered []smb2dialects.Dialect) int {
	t.Helper()
	srv, err := smb2server.NewServer(smb2server.Config{Dialects: offered})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(listener)
	t.Cleanup(func() { srv.Close() })
	return listener.Addr().(*net.TCPAddr).Port
}

// TestDialHighestInSetSelects2002 checks that the multi-protocol negotiate
// reaches SMB 2.0.2 on a 2.0.2-only server when a higher dialect is also wanted
// (the "SMB 2.???" wildcard alone is not recognised by such a server), and the
// highest dialect of the set on a server supporting them all.
func TestDialHighestInSetSelects2002(t *testing.T) {
	cases := []struct {
		name   string
		server []smb2dialects.Dialect
		prefs  []smb.SMBProtocolVersion
		want   smb.SMBProtocolVersion
	}{
		{
			"2.0.2-only server, 2.1+2.0.2 wanted",
			[]smb2dialects.Dialect{smb2dialects.SMB2_DIALECT_2_0_2},
			[]smb.SMBProtocolVersion{smb.SMB_VERSION_2_1, smb.SMB_VERSION_2_0_2},
			smb.SMB_VERSION_2_0_2,
		},
		{
			"2.0.2-only server, default preference",
			[]smb2dialects.Dialect{smb2dialects.SMB2_DIALECT_2_0_2},
			[]smb.SMBProtocolVersion{smb.SMB_VERSION_3_1_1, smb.SMB_VERSION_3_0_2, smb.SMB_VERSION_3_0, smb.SMB_VERSION_2_1, smb.SMB_VERSION_2_0_2},
			smb.SMB_VERSION_2_0_2,
		},
		{
			"full server, 2.1+2.0.2 wanted",
			nil,
			[]smb.SMBProtocolVersion{smb.SMB_VERSION_2_1, smb.SMB_VERSION_2_0_2},
			smb.SMB_VERSION_2_1,
		},
		{
			"full server, 2.0.2 only wanted",
			nil,
			[]smb.SMBProtocolVersion{smb.SMB_VERSION_2_0_2},
			smb.SMB_VERSION_2_0_2,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			port := startSMB2Server(t, c.server)
			cl, err := Dial("127.0.0.1", port, Options{Policy: PolicyHighestInSet, Preferred: c.prefs})
			if err != nil {
				t.Fatalf("Dial: %v", err)
			}
			defer cl.Disconnect()
			if got := cl.Dialect(); got != c.want {
				t.Errorf("Dialect() = %s, want %s", got, c.want)
			}
		})
	}
}
