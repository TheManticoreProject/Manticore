package server_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
)

func TestNegotiateAdvertisesDFSCapability(t *testing.T) {
	tests := []struct {
		dialect dialects.Dialect
		want    capabilities.Capabilities
	}{
		{dialects.SMB2_DIALECT_2_0_2, capabilities.SMB2_GLOBAL_CAP_DFS},
		{dialects.SMB2_DIALECT_2_1_0, capabilities.SMB2_GLOBAL_CAP_DFS | capabilities.SMB2_GLOBAL_CAP_LEASING | capabilities.SMB2_GLOBAL_CAP_LARGE_MTU},
		{dialects.SMB2_DIALECT_3_0_0, capabilities.SMB2_GLOBAL_CAP_DFS | capabilities.SMB2_GLOBAL_CAP_LEASING | capabilities.SMB2_GLOBAL_CAP_LARGE_MTU | capabilities.SMB2_GLOBAL_CAP_ENCRYPTION},
		{dialects.SMB2_DIALECT_3_0_2, capabilities.SMB2_GLOBAL_CAP_DFS | capabilities.SMB2_GLOBAL_CAP_LEASING | capabilities.SMB2_GLOBAL_CAP_LARGE_MTU | capabilities.SMB2_GLOBAL_CAP_ENCRYPTION},
		{dialects.SMB2_DIALECT_3_1_1, capabilities.SMB2_GLOBAL_CAP_DFS | capabilities.SMB2_GLOBAL_CAP_LEASING | capabilities.SMB2_GLOBAL_CAP_LARGE_MTU},
	}

	for _, tc := range tests {
		t.Run(tc.dialect.String(), func(t *testing.T) {
			_, port := startServer(t, server.Config{})
			c, err := dial(t, port, tc.dialect, "Passw0rd!")
			if err != nil {
				t.Fatalf("SessionSetup: %v", err)
			}
			if got := c.Connection.Server.Capabilities; got != tc.want {
				t.Fatalf("NEGOTIATE capabilities = %#08x, want %#08x", uint32(got), uint32(tc.want))
			}
		})
	}
}
