package client

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/securitymode"
)

// TestRequiresSecureNegotiateValidationByDialect pins the MS-SMB2 3.2.5.5
// condition, whose exclusion of 3.1.1 is the part most easily got wrong: 3.1.1
// carries pre-authentication integrity, and a 3.1.1 server terminates the
// connection outright on receiving the request ([MS-SMB2] 3.3.5.15.12), so
// sending one there breaks the session it was meant to protect.
func TestRequiresSecureNegotiateValidationByDialect(t *testing.T) {
	tests := []struct {
		dialect dialects.Dialect
		want    bool
	}{
		{dialects.SMB2_DIALECT_2_0_2, false},
		{dialects.SMB2_DIALECT_2_1_0, false},
		{dialects.SMB2_DIALECT_3_0_0, true},
		{dialects.SMB2_DIALECT_3_0_2, true},
		{dialects.SMB2_DIALECT_3_1_1, false},
	}

	for _, tt := range tests {
		t.Run(tt.dialect.String(), func(t *testing.T) {
			c := newTestClient(&fakeTransport{})
			c.Connection.Dialect = tt.dialect
			c.Session = &Session{Client: c, SigningActive: true}

			if got := c.requiresSecureNegotiateValidation(); got != tt.want {
				t.Errorf("requiresSecureNegotiateValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSecureNegotiateValidationNeedsProtection checks the request is skipped when
// neither signing nor encryption is in force. An unprotected answer proves nothing:
// whoever could forge the negotiate could forge the confirmation too.
func TestSecureNegotiateValidationNeedsProtection(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	c.Connection.Dialect = dialects.SMB2_DIALECT_3_0_2

	c.Session = &Session{Client: c}
	if c.requiresSecureNegotiateValidation() {
		t.Error("validation attempted with neither signing nor encryption active")
	}

	c.Session = &Session{Client: c, EncryptData: true}
	if !c.requiresSecureNegotiateValidation() {
		t.Error("validation skipped although the session is encrypted")
	}

	c.Session = nil
	if c.requiresSecureNegotiateValidation() {
		t.Error("validation attempted with no session")
	}
}

// TestBuildValidateNegotiateInfoRequest pins the request body ([MS-SMB2] 2.2.31.4).
// The values must be the ones the client offered, not those the server answered
// with — echoing the server's own values back would confirm nothing.
func TestBuildValidateNegotiateInfoRequest(t *testing.T) {
	guid := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	offered := []dialects.Dialect{
		dialects.SMB2_DIALECT_2_0_2,
		dialects.SMB2_DIALECT_2_1_0,
		dialects.SMB2_DIALECT_3_0_0,
		dialects.SMB2_DIALECT_3_0_2,
	}

	body := buildValidateNegotiateInfoRequest(
		capabilities.SMB2_GLOBAL_CAP_LARGE_MTU,
		guid,
		securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED,
		offered,
	)

	if want := 24 + 2*len(offered); len(body) != want {
		t.Fatalf("body is %d bytes, want %d", len(body), want)
	}
	if got := binary.LittleEndian.Uint32(body[0:4]); got != uint32(capabilities.SMB2_GLOBAL_CAP_LARGE_MTU) {
		t.Errorf("Capabilities = %#08x, want %#08x", got, uint32(capabilities.SMB2_GLOBAL_CAP_LARGE_MTU))
	}
	if !bytes.Equal(body[4:20], guid[:]) {
		t.Errorf("ClientGuid = % x, want % x", body[4:20], guid)
	}
	if got := binary.LittleEndian.Uint16(body[20:22]); got != uint16(securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED) {
		t.Errorf("SecurityMode = %#04x, want %#04x", got, uint16(securitymode.SMB2_NEGOTIATE_SIGNING_ENABLED))
	}
	if got := binary.LittleEndian.Uint16(body[22:24]); int(got) != len(offered) {
		t.Errorf("DialectCount = %d, want %d", got, len(offered))
	}
	for i, dialect := range offered {
		off := 24 + 2*i
		if got := dialects.Dialect(binary.LittleEndian.Uint16(body[off : off+2])); got != dialect {
			t.Errorf("dialect %d = %s, want %s", i, got, dialect)
		}
	}
}

// TestNegotiateRetainsOfferedValues checks the values the validation replays are
// recorded during NEGOTIATE. Without them the request would have nothing to assert.
func TestNegotiateRetainsOfferedValues(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{cannedNegotiateResponse(t)}}
	c := newTestClient(ft)

	if err := c.Negotiate(); err != nil {
		t.Fatalf("Negotiate: %v", err)
	}
	if len(c.Connection.OfferedDialects) == 0 {
		t.Error("Negotiate did not retain the offered dialects")
	}
	if c.Connection.ClientCapabilities == 0 {
		t.Error("Negotiate did not retain the offered capabilities")
	}
	if c.Connection.ClientSecurityMode == 0 {
		t.Error("Negotiate did not retain the offered security mode")
	}
}
