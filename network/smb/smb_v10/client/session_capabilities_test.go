package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
)

func TestSessionSetupCapabilitiesDescribeTheClient(t *testing.T) {
	serverCapabilities := capabilities.CAP_UNICODE |
		capabilities.CAP_LARGE_FILES |
		capabilities.CAP_NT_SMBS |
		capabilities.CAP_RPC_REMOTE_APIS |
		capabilities.CAP_STATUS32 |
		capabilities.CAP_LEVEL_II_OPLOCKS |
		capabilities.CAP_NT_FIND |
		capabilities.CAP_DFS |
		capabilities.CAP_INFOLEVEL_PASSTHROUGH |
		capabilities.CAP_LARGE_READX |
		capabilities.CAP_LARGE_WRITEX |
		capabilities.CAP_LWIO |
		capabilities.CAP_DYNAMIC_REAUTH |
		capabilities.CAP_EXTENDED_SECURITY

	want := capabilities.CAP_UNICODE |
		capabilities.CAP_LARGE_FILES |
		capabilities.CAP_NT_SMBS |
		capabilities.CAP_STATUS32 |
		capabilities.CAP_LEVEL_II_OPLOCKS |
		capabilities.CAP_NT_FIND |
		capabilities.CAP_LARGE_READX |
		capabilities.CAP_LARGE_WRITEX |
		capabilities.CAP_EXTENDED_SECURITY

	if got := sessionSetupCapabilities(serverCapabilities); got != want {
		t.Errorf("session setup capabilities = 0x%08x, want client mask 0x%08x", got, want)
	}
	if got := sessionSetupCapabilities(serverCapabilities); got == serverCapabilities {
		t.Errorf("session setup echoed the server capabilities 0x%08x", got)
	}
}

func TestSessionSetupCapabilitiesRemainAServerSubset(t *testing.T) {
	serverCapabilities := capabilities.CAP_UNICODE | capabilities.CAP_STATUS32
	got := sessionSetupCapabilities(serverCapabilities)
	if got&^serverCapabilities != 0 {
		t.Errorf("client advertised capabilities absent from the server: client=0x%08x server=0x%08x", got, serverCapabilities)
	}
}

func TestSessionSetupMaxMpxCountIsClientOwnedAndServerCapped(t *testing.T) {
	tests := []struct {
		name      string
		serverMax uint16
		want      uint16
	}{
		{name: "server allows the client default", serverMax: 512, want: DefaultMaxMpxCount},
		{name: "server limit is smaller", serverMax: 2, want: 2},
		{name: "server limit equals default", serverMax: DefaultMaxMpxCount, want: DefaultMaxMpxCount},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := sessionSetupMaxMpxCount(test.serverMax); got != test.want {
				t.Errorf("session setup MaxMpxCount = %d, want %d", got, test.want)
			}
		})
	}
}
