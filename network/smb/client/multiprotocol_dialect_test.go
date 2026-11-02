package client

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb"
	smb2dialects "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
)

// dialAgainstReply scripts a server that answers the multi-protocol negotiate
// with an SMB2 NEGOTIATE response selecting dialect, then closes.
func dialAgainstReply(t *testing.T, dialect smb2dialects.Dialect, prefs []smb.SMBProtocolVersion) error {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(5 * time.Second))
		if _, err := readDirectTCPFrame(c); err != nil {
			return
		}
		writeDirectTCPFrame(c, smb2NegotiateResponse(t, dialect, 0))
		readDirectTCPFrame(c) // hold the connection until the client gives up
	}()
	addr := ln.Addr().(*net.TCPAddr)
	c, err := Dial(addr.IP.String(), addr.Port, Options{Preferred: prefs, Policy: PolicyHighestInSet, DialTimeout: 2 * time.Second})
	if err == nil {
		c.Disconnect()
	}
	return err
}

// TestMultiProtocolNegotiateRefusesUnofferedDialect checks that a reply to a
// multi-protocol negotiate is held to what it offered: only 2.0.2, when "SMB
// 2.002" was offered, or the wildcard, when "SMB 2.???" was ([MS-SMB2]
// 3.3.5.3.1). Any other concrete dialect, or a wildcard nobody asked for, is
// refused instead of being driven as if negotiated.
func TestMultiProtocolNegotiateRefusesUnofferedDialect(t *testing.T) {
	smb1And2002 := []smb.SMBProtocolVersion{smb.SMB_VERSION_2_0_2, smb.SMB_VERSION_1_0}
	smb1And21 := []smb.SMBProtocolVersion{smb.SMB_VERSION_2_1, smb.SMB_VERSION_1_0}
	for _, tc := range []struct {
		name    string
		dialect smb2dialects.Dialect
		prefs   []smb.SMBProtocolVersion
	}{
		{"3.1.1 when only 2.002 was offered", smb2dialects.SMB2_DIALECT_3_1_1, smb1And2002},
		{"2.1 when only 2.002 was offered", smb2dialects.SMB2_DIALECT_2_1_0, smb1And2002},
		{"wildcard when only 2.002 was offered", smb2dialects.SMB2_DIALECT_WILDCARD, smb1And2002},
		{"2.0.2 when only the wildcard was offered", smb2dialects.SMB2_DIALECT_2_0_2, smb1And21},
		{"3.1.1 directly instead of the wildcard", smb2dialects.SMB2_DIALECT_3_1_1, smb1And21},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The refusal must come from the dialect check, not from a later
			// step failing against a server that stops answering.
			err := dialAgainstReply(t, tc.dialect, tc.prefs)
			if err == nil || !strings.Contains(err.Error(), "did not offer") {
				t.Errorf("Dial with SMB2 dialect 0x%04x: err = %v, want a refusal of the unoffered dialect", uint16(tc.dialect), err)
			}
		})
	}
}

// TestMultiProtocolNegotiateAccepts2002 checks the legitimate concrete answer.
func TestMultiProtocolNegotiateAccepts2002(t *testing.T) {
	if err := dialAgainstReply(t, smb2dialects.SMB2_DIALECT_2_0_2, []smb.SMBProtocolVersion{smb.SMB_VERSION_2_0_2, smb.SMB_VERSION_1_0}); err != nil {
		t.Errorf("Dial refused 2.0.2 after offering \"SMB 2.002\": %v", err)
	}
}
