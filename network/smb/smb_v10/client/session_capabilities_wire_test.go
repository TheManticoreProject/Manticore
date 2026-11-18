package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func TestNTLMSessionSetupWireCarriesClientOwnedPropertiesOnBothLegs(t *testing.T) {
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
	wantCapabilities := capabilities.CAP_UNICODE |
		capabilities.CAP_LARGE_FILES |
		capabilities.CAP_NT_SMBS |
		capabilities.CAP_STATUS32 |
		capabilities.CAP_LEVEL_II_OPLOCKS |
		capabilities.CAP_NT_FIND |
		capabilities.CAP_LARGE_READX |
		capabilities.CAP_LARGE_WRITEX |
		capabilities.CAP_EXTENDED_SECURITY

	transport := &ntlmSetupServer{uid: 0x0A00}
	c := ntlmSetupClient(client.SigningStateDisabled)
	c.Connection.Server.Capabilities = serverCapabilities
	c.Connection.MaxMpxCount = 512
	c.Transport = transport

	if err := c.SessionSetup(&credentials.Credentials{Domain: "LAB", Username: "alice", Password: "pw"}); err != nil {
		t.Fatalf("SessionSetup: %v", err)
	}
	if len(transport.setups) != 2 {
		t.Fatalf("captured %d SESSION_SETUP_ANDX requests, want both NTLM legs", len(transport.setups))
	}
	for index, setup := range transport.setups {
		if setup.capabilities != wantCapabilities {
			t.Errorf("leg %d capabilities = 0x%08x, want client mask 0x%08x", index+1, setup.capabilities, wantCapabilities)
		}
		if setup.capabilities == serverCapabilities {
			t.Errorf("leg %d echoed the server capabilities 0x%08x", index+1, setup.capabilities)
		}
		if setup.maxMpxCount != client.DefaultMaxMpxCount {
			t.Errorf("leg %d MaxMpxCount = %d, want client limit %d", index+1, setup.maxMpxCount, client.DefaultMaxMpxCount)
		}
	}
}
