package client

import (
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// Connection represents an established SMB connection between the client and server
type Connection struct {
	Host net.IP
	Port int
	//
	ClientNextSendSequenceNumber uint32                 // Sequence number for the next signed request being sent
	ClientResponseSequenceNumber map[uint32]uint32      // Expected sequence numbers for responses of outstanding signed requests, indexed by PID and MID
	ConnectionlessSessionID      uint32                 // SMB Connection identifier for connectionless transport
	IsSigningActive              bool                   // Indicates whether message signing is active
	NegotiateSent                bool                   // Indicates whether an SMB_COM_NEGOTIATE request has been sent
	NTLMChallenge                []byte                 // Cryptographic challenge received from the server during negotiation
	OpenTable                    map[uint16]interface{} // List of Opens, allowing lookups based on FID
	PIDMIDList                   []interface{}          // List of outstanding SMB commands
	SearchOpenTable              []interface{}          // List of SearchOpens representing open file searches
	SelectedDialect              string                 // SMB Protocol dialect selected for this connection
	MaxMpxCount                  uint16                 // Maximum number of commands permitted to be outstanding
	SessionTable                 map[uint16]*Session    // List of authenticated sessions established on this connection
	ShareLevelAccessControl      bool                   // Whether server requires share passwords instead of user accounts
	SigningChallengeResponse     []byte                 // Challenge response used for signing
	SigningSessionKey            []byte                 // Session key used for signing packets
	TreeConnectTable             map[uint16]interface{} // List of tree connects over this SMB connection

	Server struct {
		Name              string         // Name of the server
		SigningState      string         // Signing policy of the server (Disabled, Enabled, or Required)
		Capabilities      uint32         // Capabilities of the server
		MaxBufferSize     uint32         // Negotiated maximum size for SMB messages sent to server
		ChallengeResponse bool           // Indicates whether server supports challenge/response authentication
		SessionKey        uint32         // Session key value returned by the server in negotiate response
		SystemTime        types.SMB_TIME // System time of the server
		TimeZone          int16          // Time zone of the server
		DomainName        string         // Domain name of the server
	}
}
