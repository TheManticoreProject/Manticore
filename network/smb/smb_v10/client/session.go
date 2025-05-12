package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// Session represents an established session between the client and server
type Session struct {
	// The SMB connection associated with this session
	Connection *Client

	// The cryptographic session key associated with this session
	SessionKey []byte

	// The 2-byte UID for this session
	SessionUID uint16

	// Opaque implementation-specific entity that identifies the credentials
	UserCredentials interface{}
}

func (c *Client) SessionSetup() error {
	if !c.Transport.IsConnected() {
		return fmt.Errorf("transport is not connected")
	}
	request_msg := message.NewMessage()
	session_setup_cmd := commands.NewSessionSetupAndxRequest()

	// Here put the common logic for all session setup commands
	request_msg.Header.Command = codes.SMB_COM_SESSION_SETUP_ANDX
	request_msg.Header.Flags = 0x18    // SMB_FLAGS_CANONICAL_PATHNAMES | SMB_FLAGS_CASELESS_PATHNAMES
	request_msg.Header.Flags2 = 0x0001 // SMB_FLAGS2_KNOWS_LONG_NAMES

	// Add Unicode support if server supports it
	if c.Connection.Server.Capabilities&0x00000004 != 0 { // CAP_UNICODE
		request_msg.Header.Flags2 |= 0x8000 // SMB_FLAGS2_UNICODE
	}

	// Add extended security if server supports it
	if c.Connection.Server.Capabilities&0x80000000 != 0 { // CAP_EXTENDED_SECURITY
		request_msg.Header.Flags2 |= 0x0800 // SMB_FLAGS2_EXTENDED_SECURITY
	}

	// Set message signing flags based on server security mode
	if c.Connection.Server.SecurityMode.IsSecuritySignatureEnabled() {
		request_msg.Header.Flags2 |= 0x0004 // SMB_FLAGS2_SECURITY_SIGNATURE
	}

	// Set process ID and multiplex ID
	request_msg.Header.SetPID(0x0001)
	request_msg.Header.MID = 0x0002
	request_msg.Header.TID = 0x0000
	request_msg.Header.UID = 0x0000

	session_setup_cmd.MaxBufferSize = types.USHORT(c.Connection.Server.MaxBufferSize)
	session_setup_cmd.MaxMpxCount = c.Connection.MaxMpxCount
	session_setup_cmd.Capabilities = c.Connection.Server.Capabilities

	// Check if we're using share level access control
	if c.Connection.Server.SecurityMode.SupportsShareLevelAccessControl() {
		// Share level access control is required by the server
		// If no authentication has been performed on the SMB connection, use anonymous authentication

		// Parameters
		session_setup_cmd.VcNumber = types.USHORT(0x0000)
		session_setup_cmd.SessionKey = c.Connection.Server.SessionKey
		session_setup_cmd.OEMPasswordLen = types.USHORT(0x0000)
		session_setup_cmd.UnicodePasswordLen = types.USHORT(0x0000)

		// Data section - for null session, use empty strings
		session_setup_cmd.OEMPassword = []types.UCHAR{}
		session_setup_cmd.UnicodePassword = []types.UCHAR{}
		session_setup_cmd.AccountName.SetString("")
		session_setup_cmd.PrimaryDomain.SetString("")
		session_setup_cmd.NativeOS.SetString("")
		session_setup_cmd.NativeLanMan.SetString("")
	} else {
		// User level access control is required by the server
		// TODO: Look up Session from Client.Connection.SessionTable where Session.UserCredentials matches
		// the application-supplied UserCredentials and reuse if found

		// Handle authentication based on server capabilities
		if !c.Connection.Server.SecurityMode.SupportsChallengeResponseAuth() {
			// Server doesn't support challenge/response authentication

			// Use plaintext authentication
			session_setup_cmd.VcNumber = types.USHORT(0x0000)
			session_setup_cmd.SessionKey = c.Connection.Server.SessionKey

			// Check if Unicode is supported
			if c.Connection.Server.Capabilities&0x00000004 != 0 { // CAP_UNICODE
				// Send password in Unicode
				session_setup_cmd.UnicodePassword = []types.UCHAR(utf16.EncodeUTF16LE("UnicodePassword"))
				session_setup_cmd.UnicodePasswordLen = types.USHORT(len(session_setup_cmd.UnicodePassword))
				session_setup_cmd.OEMPasswordLen = types.USHORT(0x0000)
				session_setup_cmd.OEMPassword = []types.UCHAR{}
			} else {
				// Send password in OEM format
				session_setup_cmd.OEMPassword = []types.UCHAR("OEMPassword")
				session_setup_cmd.OEMPasswordLen = types.USHORT(len(session_setup_cmd.OEMPassword))
				session_setup_cmd.UnicodePasswordLen = types.USHORT(0x0000)
				session_setup_cmd.UnicodePassword = []types.UCHAR{}
			}

			session_setup_cmd.AccountName.SetString("AccountName")
			session_setup_cmd.PrimaryDomain.SetString("DomainName")
			session_setup_cmd.NativeOS.SetString("")
			session_setup_cmd.NativeLanMan.SetString("")
		} else {
			// Server supports challenge/response authentication
			// Determine authentication type based on policies

			session_setup_cmd.VcNumber = types.USHORT(0x0000)
			session_setup_cmd.SessionKey = c.Connection.Server.SessionKey

			// TODO: Implement proper LM/NTLM/LMv2/NTLMv2 response selection based on policies
			// For now, we'll use placeholder values

			// LM or LMv2 response in OEMPassword field
			session_setup_cmd.OEMPassword = []types.UCHAR("LMResponse")
			session_setup_cmd.OEMPasswordLen = types.USHORT(len(session_setup_cmd.OEMPassword))

			// NTLM or NTLMv2 response in UnicodePassword field
			session_setup_cmd.UnicodePassword = []types.UCHAR("NTLMResponse")
			session_setup_cmd.UnicodePasswordLen = types.USHORT(len(session_setup_cmd.UnicodePassword))

			session_setup_cmd.AccountName.SetString("AccountName")
			session_setup_cmd.PrimaryDomain.SetString("DomainName")
			session_setup_cmd.NativeOS.SetString("")
			session_setup_cmd.NativeLanMan.SetString("")
		}
	}

	request_msg.AddCommand(session_setup_cmd)

	marshalled_message, err := request_msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal negotiate message: %v", err)
	}

	// Send the message
	_, err = c.Transport.Send(marshalled_message)
	if err != nil {
		return fmt.Errorf("failed to send negotiate message: %v", err)
	}

	// Receive the response
	raw_response_message, err := c.Transport.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive response message: %v", err)
	}

	response_msg := message.NewMessage()
	err = response_msg.Unmarshal(raw_response_message)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response message: %v", err)
	}

	if response_msg.Header.Command != codes.SMB_COM_SESSION_SETUP_ANDX {
		return fmt.Errorf("unexpected response command: %d", response_msg.Header.Command)
	}

	_ = response_msg.Command.(*commands.SessionSetupAndxResponse)

	// TODO

	return nil
}
