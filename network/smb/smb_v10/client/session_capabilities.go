package client

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// defaultSessionSetupCapabilities is the feature set this client reports in an
// SMB_COM_SESSION_SETUP_ANDX request. It deliberately contains only client-side
// behavior implemented by the SMB1 engine; server-only features such as remote
// APIs, DFS, passthrough information levels and LWIO must not be echoed back.
// It starts with the compact Windows SMB1 client profile and includes the file,
// search and large-transfer operations implemented by this engine.
const defaultSessionSetupCapabilities capabilities.Capabilities = capabilities.CAP_UNICODE |
	capabilities.CAP_LARGE_FILES |
	capabilities.CAP_NT_SMBS |
	capabilities.CAP_STATUS32 |
	capabilities.CAP_LEVEL_II_OPLOCKS |
	capabilities.CAP_NT_FIND |
	capabilities.CAP_LARGE_READX |
	capabilities.CAP_LARGE_WRITEX |
	capabilities.CAP_EXTENDED_SECURITY

// sessionSetupCapabilities returns the client's capabilities, restricted to
// features the server advertised as required by the SESSION_SETUP_ANDX wire
// contract.
func sessionSetupCapabilities(server capabilities.Capabilities) capabilities.Capabilities {
	return defaultSessionSetupCapabilities & server
}

// sessionSetupMaxMpxCount reports the client's own pending-request limit while
// respecting the upper bound negotiated by the server.
func sessionSetupMaxMpxCount(serverMax uint16) uint16 {
	if serverMax < DefaultMaxMpxCount {
		return serverMax
	}
	return DefaultMaxMpxCount
}

// setSessionSetupClientProperties applies the client-owned fields shared by
// every SESSION_SETUP_ANDX authentication mechanism. Keeping this assignment in
// one place prevents the NTLM and Kerberos wire profiles from drifting apart.
func setSessionSetupClientProperties(cmd *commands.SessionSetupAndxRequest, connection *Connection) {
	cmd.MaxMpxCount = types.USHORT(sessionSetupMaxMpxCount(connection.MaxMpxCount))
	cmd.Capabilities = sessionSetupCapabilities(connection.Server.Capabilities)
}
