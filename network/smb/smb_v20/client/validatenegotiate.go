package client

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
)

// validateNegotiateInfoResponseSize is the fixed size of a VALIDATE_NEGOTIATE_INFO
// response: Capabilities(4) ServerGuid(16) SecurityMode(2) Dialect(2)
// ([MS-SMB2] 2.2.32.6).
const validateNegotiateInfoResponseSize = 24

// requiresSecureNegotiateValidation reports whether this connection must validate
// its negotiate exchange after a tree connect.
//
// [MS-SMB2] 3.2.5.5 requires it when the negotiated dialect is in the SMB 3.x
// family but is *not* 3.1.1. The exclusion is not an oversight: 3.1.1 carries
// pre-authentication integrity, which already covers the NEGOTIATE exchange, and
// [MS-SMB2] 3.3.5.15.12 has a 3.1.1 server terminate the connection outright on
// receiving the request. Sending it there would break the session it was meant to
// protect.
//
// The request must be signed to be worth anything — an unprotected response can
// simply be forged by whoever forged the negotiate — so it is skipped when neither
// signing nor encryption is in force.
func (c *Client) requiresSecureNegotiateValidation() bool {
	if c.Session == nil {
		return false
	}
	if !isSMB3Dialect(c.Connection.Dialect) || c.Connection.Dialect == dialects.SMB2_DIALECT_3_1_1 {
		return false
	}
	return c.Session.SigningActive || c.Session.EncryptData
}

// validateNegotiateInfo asks the server to restate the negotiate exchange under the
// protection of the established session, and checks that what comes back matches
// what this client actually negotiated.
//
// A downgrade attacker who tampered with the unprotected NEGOTIATE cannot produce a
// matching signed answer, so a mismatch means the negotiation was altered and the
// connection must not be used ([MS-SMB2] 3.2.5.14.12).
func (c *Client) validateNegotiateInfo() error {
	input := buildValidateNegotiateInfoRequest(
		c.Connection.ClientCapabilities,
		c.ClientGuid,
		c.Connection.ClientSecurityMode,
		c.Connection.OfferedDialects,
	)

	output, err := c.Ioctl(validateNegotiateFileId(), filesystem.FSCTL_VALIDATE_NEGOTIATE_INFO, input, true, validateNegotiateInfoResponseSize)
	if err != nil {
		return fmt.Errorf("secure negotiate validation failed: %w", err)
	}
	if len(output) < validateNegotiateInfoResponseSize {
		return fmt.Errorf("secure negotiate validation returned %d bytes, want %d", len(output), validateNegotiateInfoResponseSize)
	}

	gotCapabilities := capabilities.Capabilities(binary.LittleEndian.Uint32(output[0:4]))
	var gotServerGuid [16]byte
	copy(gotServerGuid[:], output[4:20])
	gotSecurityMode := securitymode.SecurityMode(binary.LittleEndian.Uint16(output[20:22]))
	gotDialect := dialects.Dialect(binary.LittleEndian.Uint16(output[22:24]))

	server := c.Connection.Server
	if gotCapabilities != server.Capabilities {
		return fmt.Errorf("secure negotiate validation: server capabilities %#08x do not match the negotiated %#08x", uint32(gotCapabilities), uint32(server.Capabilities))
	}
	if !bytes.Equal(gotServerGuid[:], server.ServerGuid[:]) {
		return fmt.Errorf("secure negotiate validation: server GUID % x does not match the negotiated % x", gotServerGuid, server.ServerGuid)
	}
	if gotSecurityMode != server.SecurityMode {
		return fmt.Errorf("secure negotiate validation: server security mode %#04x does not match the negotiated %#04x", uint16(gotSecurityMode), uint16(server.SecurityMode))
	}
	if gotDialect != c.Connection.Dialect {
		return fmt.Errorf("secure negotiate validation: server dialect %s does not match the negotiated %s", gotDialect, c.Connection.Dialect)
	}

	return nil
}

// buildValidateNegotiateInfo builds the VALIDATE_NEGOTIATE_INFO request body
// ([MS-SMB2] 2.2.31.4). The values are the ones this client put in its NEGOTIATE
// request, not the ones the server answered with: the point is to have the server
// confirm what it saw against what was sent.
func buildValidateNegotiateInfoRequest(caps capabilities.Capabilities, clientGuid [16]byte, mode securitymode.SecurityMode, offered []dialects.Dialect) []byte {
	body := make([]byte, 0, 24+2*len(offered))
	body = binary.LittleEndian.AppendUint32(body, uint32(caps))
	body = append(body, clientGuid[:]...)
	body = binary.LittleEndian.AppendUint16(body, uint16(mode))
	body = binary.LittleEndian.AppendUint16(body, uint16(len(offered)))
	for _, dialect := range offered {
		body = binary.LittleEndian.AppendUint16(body, uint16(dialect))
	}
	return body
}

// validateNegotiateFileId is the reserved handle an FSCTL that addresses the
// connection rather than a file is sent on.
func validateNegotiateFileId() types.SMB2_FILEID {
	return types.SMB2_FILEID{Persistent: 0xFFFFFFFFFFFFFFFF, Volatile: 0xFFFFFFFFFFFFFFFF}
}
