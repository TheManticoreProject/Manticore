package kerberos

import (
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/ccache"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/kirbi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// tgtCredInfo builds the KrbCredInfo describing the currently held TGT from the
// stored AS-REP enc-part (session key, flags, times, service name).
func (c *KerberosClient) tgtCredInfo() messages.KrbCredInfo {
	return messages.KrbCredInfo{
		Key:       messages.EncryptionKey{KeyType: c.sessionEType, KeyValue: c.sessionKey},
		PRealm:    c.realm,
		PName:     messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{c.username}},
		Flags:     c.tgtEnc.Flags,
		AuthTime:  c.tgtEnc.AuthTime,
		StartTime: c.tgtEnc.StartTime,
		EndTime:   c.tgtEnc.EndTime,
		RenewTill: c.tgtEnc.RenewTill,
		SRealm:    c.tgtEnc.SRealm,
		SName:     c.tgtEnc.SName,
	}
}

// ExportTGTKirbi returns the current TGT as .kirbi (DER KRB-CRED) bytes, suitable
// for pass-the-ticket with Rubeus. GetTGT must have succeeded first.
func (c *KerberosClient) ExportTGTKirbi() ([]byte, error) {
	if !c.hasTGT {
		return nil, fmt.Errorf("kerberos: no TGT: call GetTGT first")
	}
	cred, err := kirbi.New(c.tgtTicketRaw, c.tgtCredInfo())
	if err != nil {
		return nil, err
	}
	return kirbi.Bytes(cred)
}

// ExportTGTCCache returns the current TGT as an MIT ccache (v4) holding a single
// credential. GetTGT must have succeeded first.
func (c *KerberosClient) ExportTGTCCache() (*ccache.CCache, error) {
	if !c.hasTGT {
		return nil, fmt.Errorf("kerberos: no TGT: call GetTGT first")
	}
	client := ccache.Principal{
		NameType:   uint32(messages.NameTypePrincipal),
		Realm:      c.realm,
		Components: []string{c.username},
	}
	server := ccache.Principal{
		NameType:   uint32(c.tgtEnc.SName.NameType),
		Realm:      c.tgtEnc.SRealm,
		Components: append([]string(nil), c.tgtEnc.SName.NameString...),
	}
	start := c.tgtEnc.StartTime
	if start.IsZero() {
		start = c.tgtEnc.AuthTime
	}
	return &ccache.CCache{
		DefaultPrincipal: client,
		Credentials: []ccache.Credential{{
			Client:      client,
			Server:      server,
			Key:         ccache.Keyblock{EType: uint16(c.sessionEType), KeyValue: c.sessionKey},
			AuthTime:    unixU32(c.tgtEnc.AuthTime),
			StartTime:   unixU32(start),
			EndTime:     unixU32(c.tgtEnc.EndTime),
			RenewTill:   unixU32(c.tgtEnc.RenewTill),
			TicketFlags: flagsToUint32(c.tgtEnc.Flags),
			Ticket:      c.tgtTicketRaw,
		}},
	}, nil
}

// unixU32 returns t as Unix seconds in a uint32, or 0 for the zero time (rather
// than the wrapped negative value uint32(-6795...) would give).
func unixU32(t time.Time) uint32 {
	if t.IsZero() {
		return 0
	}
	return uint32(t.Unix())
}

// flagsToUint32 converts a KerberosFlags BIT STRING (bit 0 = MSB of the first
// octet) to the big-endian uint32 that the ccache ticket_flags field uses. A
// short or over-long byte slice is padded/truncated to exactly 4 bytes.
func flagsToUint32(flags asn1.BitString) uint32 {
	var b [4]byte
	copy(b[:], flags.Bytes)
	return binary.BigEndian.Uint32(b[:])
}
