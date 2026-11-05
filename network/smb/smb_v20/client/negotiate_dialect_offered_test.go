package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/fingerprint"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// cannedNegotiateResponseWithDialect builds an SMB2 NEGOTIATE response that
// advertises an arbitrary DialectRevision, so a test can present a dialect the
// client did not offer.
func cannedNegotiateResponseWithDialect(t *testing.T, dialect dialects.Dialect) []byte {
	t.Helper()
	resp := commands.NewNegotiateResponse()
	resp.DialectRevision = dialect
	resp.SecurityMode = securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED
	resp.MaxTransactSize = 0x10000
	resp.MaxReadSize = 0x10000
	resp.MaxWriteSize = 0x10000
	resp.SystemTime = 0x01D0000000000000
	resp.SecurityBuffer = []byte{0x60, 0x40, 0x06, 0x06}

	m := message.NewMessage()
	m.Header.AddFlags(flags.SMB2_FLAGS_SERVER_TO_REDIR)
	m.SetCommand(resp)
	wire, err := m.Marshal()
	if err != nil {
		t.Fatalf("building canned response: %v", err)
	}
	return wire
}

// TestNegotiateRejectsUnofferedDialect checks the [MS-SMB2] 3.2.5.2 rule that the
// DialectRevision in the response must be one the client offered: a server that
// selects a dialect outside the offered set (here a downgrade to 2.0.2 when only
// 2.1 was offered) must make Negotiate fail rather than silently accept it.
func TestNegotiateRejectsUnofferedDialect(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{cannedNegotiateResponseWithDialect(t, dialects.SMB2_DIALECT_2_0_2)}}
	c := newTestClient(ft)
	c.Profile = fingerprint.Default()
	c.Profile.Dialects = []dialects.Dialect{dialects.SMB2_DIALECT_2_1_0}

	if err := c.Negotiate(); err == nil {
		t.Fatalf("Negotiate accepted dialect %s, which was not offered (offered only 2.1)",
			c.Connection.Dialect)
	}
}

// TestNegotiateRejectsWildcardDialect checks that a server answering the engine's
// native NEGOTIATE with the wildcard revision 0x02FF is rejected: the wildcard is
// valid only in reply to a multi-protocol negotiate and is never in the engine's
// offered list.
func TestNegotiateRejectsWildcardDialect(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{cannedNegotiateResponseWithDialect(t, dialects.SMB2_DIALECT_WILDCARD)}}
	c := newTestClient(ft)
	c.Profile = fingerprint.Default()
	c.Profile.Dialects = []dialects.Dialect{dialects.SMB2_DIALECT_2_1_0, dialects.SMB2_DIALECT_3_1_1}

	if err := c.Negotiate(); err == nil {
		t.Fatal("Negotiate accepted the wildcard revision 0x02FF on a native negotiate")
	}
}

// TestNegotiateAcceptsOfferedDialect is the positive control: a dialect that is in
// the offered set is accepted and recorded.
func TestNegotiateAcceptsOfferedDialect(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{cannedNegotiateResponseWithDialect(t, dialects.SMB2_DIALECT_2_1_0)}}
	c := newTestClient(ft)
	c.Profile = fingerprint.Default()
	c.Profile.Dialects = []dialects.Dialect{dialects.SMB2_DIALECT_2_0_2, dialects.SMB2_DIALECT_2_1_0}

	if err := c.Negotiate(); err != nil {
		t.Fatalf("Negotiate rejected an offered dialect: %v", err)
	}
	if c.Connection.Dialect != dialects.SMB2_DIALECT_2_1_0 {
		t.Errorf("Connection.Dialect = %s, want SMB 2.1", c.Connection.Dialect)
	}
}
