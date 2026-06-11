package client

import (
	"net"
	"sync"

	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// Client represents an SMB 2.0 client. It mirrors the SMB 1.0 client
// (network/smb/smb_v10/client) but drives the SMB2 wire layer.
type Client struct {
	// Transport is the byte-stream transport (TCP / NBT) to the server.
	Transport transport.Transport

	// Connection holds per-connection negotiated state.
	Connection *Connection

	// Session is the authenticated session, once established.
	Session *Session

	// ClientGuid is a per-client GUID sent in SMB2 NEGOTIATE.
	ClientGuid [16]byte

	// Workstation is the client workstation name used during NTLM authentication.
	Workstation string

	// asyncMu guards the identifiers of the request currently awaiting an async
	// (STATUS_PENDING) completion, which Cancel reads to target a pending
	// operation (for example a blocked CHANGE_NOTIFY). Only one operation is
	// outstanding at a time on this synchronous client.
	asyncMu          sync.Mutex
	pendingMessageId uint64
	pendingAsyncId   uint64
	hasPendingAsync  bool
}

// setPendingAsync records the in-flight async operation so a concurrent Cancel
// can target it.
func (c *Client) setPendingAsync(messageId, asyncId uint64) {
	c.asyncMu.Lock()
	c.pendingMessageId, c.pendingAsyncId, c.hasPendingAsync = messageId, asyncId, true
	c.asyncMu.Unlock()
}

// clearPendingAsync forgets any recorded in-flight async operation.
func (c *Client) clearPendingAsync() {
	c.asyncMu.Lock()
	c.hasPendingAsync = false
	c.asyncMu.Unlock()
}

// pendingAsync returns the recorded in-flight async operation's identifiers.
func (c *Client) pendingAsync() (messageId, asyncId uint64, ok bool) {
	c.asyncMu.Lock()
	defer c.asyncMu.Unlock()
	return c.pendingMessageId, c.pendingAsyncId, c.hasPendingAsync
}

// Connection represents an established SMB2 connection between client and server.
type Connection struct {
	// Server holds the negotiated properties of the peer.
	Server *Server

	// MessageId is the value to use for the next request's 64-bit MessageId. SMB2
	// requires each request on a connection to use a unique, monotonically
	// increasing MessageId.
	MessageId uint64

	// Credits is the number of credits the server has granted that the client may
	// still spend. SMB 2.0.2 uses a simple one-credit-per-request model.
	Credits uint16

	// Dialect is the SMB2 dialect selected during negotiation.
	Dialect dialects.Dialect

	// SessionTable holds authenticated sessions on this connection, keyed by the
	// server-assigned 64-bit SessionId.
	SessionTable map[uint64]*Session

	// TreeConnectTable holds tree connects on this connection, keyed by TreeId.
	TreeConnectTable map[uint32]*TreeConnect
}

// Server holds the peer properties negotiated via SMB2 NEGOTIATE.
type Server struct {
	// Host is the server IP address.
	Host net.IP

	// Port is the server TCP port.
	Port int

	// SecurityMode is the server's signing policy.
	SecurityMode securitymode.SecurityMode

	// Capabilities is the server's capability set.
	Capabilities capabilities.Capabilities

	// MaxTransactSize / MaxReadSize / MaxWriteSize are the negotiated maxima.
	MaxTransactSize uint32
	MaxReadSize     uint32
	MaxWriteSize    uint32

	// ServerGuid is the server's GUID from the NEGOTIATE response.
	ServerGuid [16]byte

	// SystemTime / ServerStartTime are FILETIME values from the NEGOTIATE response.
	SystemTime      uint64
	ServerStartTime uint64

	// SecurityBuffer is the GSS (SPNEGO) token returned in the NEGOTIATE response,
	// used to seed authentication.
	SecurityBuffer []byte

	// Server identity advertised in the NTLM CHALLENGE during authentication: the
	// NetBIOS/DNS computer and domain names and the OS version. Populated by
	// SessionSetup; empty/zero before a successful authentication.
	NetBIOSComputerName string
	NetBIOSDomainName   string
	DNSComputerName     string
	DNSDomainName       string
	OSVersionMajor      uint8
	OSVersionMinor      uint8
	OSVersionBuild      uint16

	// SupportsNTLMv2 reports whether the server's NTLM CHALLENGE advertised
	// extended session security (NTLM2 / NTLMv2). Populated by SessionSetup.
	SupportsNTLMv2 bool
}

// Session represents an authenticated SMB2 session.
type Session struct {
	// Client is the owning client.
	Client *Client

	// SessionId is the server-assigned 64-bit session identifier.
	SessionId uint64

	// TreeId is the 32-bit tree connect currently selected for file operations.
	TreeId uint32

	// SessionKey is the session key derived during authentication.
	SessionKey []byte

	// SigningKey is the key used to sign/verify messages on this session. For the
	// SMB 2.0.2 and 2.1 dialects it is the session key itself.
	SigningKey []byte

	// SigningActive indicates that requests on this session are signed and
	// responses are verified.
	SigningActive bool

	// Credentials are the credentials used to authenticate the session.
	Credentials *credentials.Credentials
}

// TreeConnect represents a connection to a share on the server.
type TreeConnect struct {
	// Connection is the owning connection.
	Connection *Connection

	// Session is the session on which the tree connect was established.
	Session *Session

	// ShareName is the share this tree connect targets.
	ShareName string

	// TreeId is the server-assigned 32-bit tree identifier.
	TreeId uint32

	// ShareType is the type of share (disk/pipe/print) from the TREE_CONNECT response.
	ShareType uint8
}
