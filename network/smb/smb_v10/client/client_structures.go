package client

import (
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/securitymode"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// Client represents an SMB v1.0 client
type Client struct {
	// Transport is the transport layer for the client
	Transport transport.Transport

	// TreeConnect is the tree connect for the client
	TreeConnect *TreeConnect

	// Session is the session for the client
	Session *Session

	// Connection is the connection for the client
	Connection *Connection
}

// Connection represents an established SMB connection between the client and server
type Connection struct {
	Server *Server

	// ClientNextSendSequenceNumber is the sequence number for the next signed request being sent
	ClientNextSendSequenceNumber uint32

	// ClientResponseSequenceNumber is the expected sequence numbers for responses of outstanding signed requests, indexed by PID and MID
	ClientResponseSequenceNumber map[uint32]uint32

	// ConnectionlessSessionID is the SMB Connection identifier for connectionless transport
	ConnectionlessSessionID uint32

	// IsSigningActive indicates whether message signing is active
	IsSigningActive bool

	// NegotiateSent indicates whether an SMB_COM_NEGOTIATE request has been sent
	NegotiateSent bool

	// NTLMChallenge is the cryptographic challenge received from the server during negotiation
	NTLMChallenge []byte

	// OpenTable is the list of Opens, allowing lookups based on FID
	OpenTable map[uint16]interface{}

	// PIDMIDList is the list of outstanding SMB commands
	PIDMIDList []interface{}

	// SearchOpenTable is the list of SearchOpens representing open file searches
	SearchOpenTable []interface{}

	// SelectedDialect is the SMB Protocol dialect selected for this connection
	SelectedDialect string

	// MaxMpxCount is the maximum number of commands permitted to be outstanding
	MaxMpxCount uint16

	// SessionTable is the list of authenticated sessions established on this connection
	SessionTable map[uint16]*Session

	// ShareLevelAccessControl indicates whether the server requires share passwords instead of user accounts
	ShareLevelAccessControl bool

	// SigningChallengeResponse is the challenge response used for signing
	SigningChallengeResponse []byte

	// SigningSessionKey is the session key used for signing packets
	SigningSessionKey []byte

	// TreeConnectTable is the list of tree connects over this SMB connection
	TreeConnectTable map[uint16]interface{}
}

// Server represents the server for the client
type Server struct {
	// Host is the IP address of the server
	Host net.IP

	// Port is the port number of the server
	Port int

	// Name is the name of the server
	Name string

	// SecurityMode is the security mode of the server
	SecurityMode securitymode.SecurityMode

	// SigningState is the signing policy of the server (Disabled, Enabled, or Required)
	SigningState string

	// Capabilities is the capabilities of the server
	Capabilities uint32

	// MaxBufferSize is the negotiated maximum size for SMB messages sent to server
	MaxBufferSize uint32

	// ChallengeResponse indicates whether server supports challenge/response authentication
	ChallengeResponse bool

	// SessionKey is the session key value returned by the server in negotiate response
	SessionKey uint32

	// SystemTime is the system time of the server
	SystemTime types.SMB_TIME

	// TimeZone is the time zone of the server
	TimeZone int16

	// DomainName is the domain name of the server
	DomainName string
}
