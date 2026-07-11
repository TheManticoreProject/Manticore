package kerberos

import (
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
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

// cacheServiceTicket records a service ticket (raw APPLICATION[1] bytes plus the
// enc-part it was issued with) under its SPN so ExportServiceTicketKirbi/CCache
// can serialize it later. It is called on every successful GetTGS and by
// LoadServiceTicket. The client principal is this client's identity (the
// principal the ticket was issued to).
func (c *KerberosClient) cacheServiceTicket(ticketRaw []byte, enc messages.EncTGSRepPart) {
	if len(enc.SName.NameString) == 0 {
		return
	}
	if c.serviceTickets == nil {
		c.serviceTickets = make(map[string]cachedServiceTicket)
	}
	spn := strings.Join(enc.SName.NameString, "/")
	c.serviceTickets[normalizeSPN(spn)] = cachedServiceTicket{
		ticketRaw: ticketRaw,
		credInfo: messages.KrbCredInfo{
			Key:       enc.Key,
			PRealm:    c.realm,
			PName:     messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{c.username}},
			Flags:     enc.Flags,
			AuthTime:  enc.AuthTime,
			StartTime: enc.StartTime,
			EndTime:   enc.EndTime,
			RenewTill: enc.RenewTill,
			SRealm:    enc.SRealm,
			SName:     enc.SName,
		},
	}
}

// serviceTicketFor returns the cached service ticket for spn: an exact normalized
// "service/host" match first, then a match on the service principal name (so a
// bare "cifs" or a differently spelled host component still resolves). It errors
// if no service ticket for spn has been obtained.
func (c *KerberosClient) serviceTicketFor(spn string) (cachedServiceTicket, error) {
	if ct, ok := c.serviceTickets[normalizeSPN(spn)]; ok {
		return ct, nil
	}
	for _, ct := range c.serviceTickets {
		if snameMatchesSPN(ct.credInfo.SName, spn) {
			return ct, nil
		}
	}
	return cachedServiceTicket{}, fmt.Errorf("kerberos: no service ticket for SPN %q: call GetTGS first", spn)
}

// ExportServiceTicketKirbi returns the cached service ticket for spn as .kirbi
// (DER KRB-CRED) bytes, the service-ticket counterpart of ExportTGTKirbi. The
// ticket must have been obtained via GetTGS (or wired in via LoadServiceTicket)
// beforehand.
func (c *KerberosClient) ExportServiceTicketKirbi(spn string) ([]byte, error) {
	ct, err := c.serviceTicketFor(spn)
	if err != nil {
		return nil, err
	}
	cred, err := kirbi.New(ct.ticketRaw, ct.credInfo)
	if err != nil {
		return nil, err
	}
	return kirbi.Bytes(cred)
}

// ExportServiceTicketCCache returns the cached service ticket for spn as an MIT
// ccache (v4) holding a single credential, the service-ticket counterpart of
// ExportTGTCCache. The ticket must have been obtained via GetTGS (or wired in via
// LoadServiceTicket) beforehand.
func (c *KerberosClient) ExportServiceTicketCCache(spn string) ([]byte, error) {
	ct, err := c.serviceTicketFor(spn)
	if err != nil {
		return nil, err
	}
	info := ct.credInfo
	client := ccache.Principal{
		NameType:   uint32(info.PName.NameType),
		Realm:      info.PRealm,
		Components: append([]string(nil), info.PName.NameString...),
	}
	server := ccache.Principal{
		NameType:   uint32(info.SName.NameType),
		Realm:      info.SRealm,
		Components: append([]string(nil), info.SName.NameString...),
	}
	start := info.StartTime
	if start.IsZero() {
		start = info.AuthTime
	}
	cc := &ccache.CCache{
		DefaultPrincipal: client,
		Credentials: []ccache.Credential{{
			Client:      client,
			Server:      server,
			Key:         ccache.Keyblock{EType: uint16(info.Key.KeyType), KeyValue: info.Key.KeyValue},
			AuthTime:    unixU32(info.AuthTime),
			StartTime:   unixU32(start),
			EndTime:     unixU32(info.EndTime),
			RenewTill:   unixU32(info.RenewTill),
			TicketFlags: flagsToUint32(info.Flags),
			Ticket:      ct.ticketRaw,
		}},
	}
	return cc.Marshal()
}

// ExportServiceTicketKirbiToFile writes the cached service ticket for spn to path
// in .kirbi (DER KRB-CRED) form, mode 0600.
func (c *KerberosClient) ExportServiceTicketKirbiToFile(spn, path string) error {
	blob, err := c.ExportServiceTicketKirbi(spn)
	if err != nil {
		return err
	}
	return os.WriteFile(path, blob, 0o600)
}

// ExportServiceTicketCCacheToFile writes the cached service ticket for spn to path
// in MIT ccache (v4) form, mode 0600.
func (c *KerberosClient) ExportServiceTicketCCacheToFile(spn, path string) error {
	blob, err := c.ExportServiceTicketCCache(spn)
	if err != nil {
		return err
	}
	return os.WriteFile(path, blob, 0o600)
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
