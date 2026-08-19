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

// Ticket import (pass-the-ticket)
//
// These methods are the read-side counterpart of export.go: they load a TGT
// (or a single service ticket) that was obtained by another tool or a prior
// run — from a .kirbi (DER KRB-CRED) or an MIT ccache — and wire it into the
// client so GetTGS and the GSSAPI/SMB/RPC/LDAP transports operate with no
// password. A .kirbi/ccache stores, in the clear, everything pass-the-ticket
// needs: the raw APPLICATION[1] ticket to re-emit verbatim in a downstream
// AP-REQ, plus the session key and enctype used to seal that AP-REQ's
// Authenticator (RFC 4120 Section 5.8.1 / MIT ccache v4).

// LoadTGT wires a Ticket Granting Ticket held in a parsed KRB-CRED (.kirbi) into
// the client, enabling pass-the-ticket: subsequent GetTGS / SPNEGO calls reuse
// the ticket and its session key with no password.
//
// The KRB-CRED must carry an unencrypted enc-part (etype 0, the .kirbi
// convention) so the session key and ticket metadata can be read. When the
// credential holds multiple tickets the krbtgt/REALM entry is selected; if none
// is a TGT the first entry is used. The client's username and realm are set from
// the ticket's client principal, so a client created with NewClient("", "",
// kdcHost) becomes fully usable after import.
func (c *KerberosClient) LoadTGT(cred *messages.KRBCred) error {
	ticketRaw, info, err := selectCredInfo(cred, true)
	if err != nil {
		return err
	}
	return c.applyTGT(ticketRaw, info)
}

// LoadTGTFromKirbiBytes parses .kirbi (DER KRB-CRED) bytes and loads the TGT they
// carry via LoadTGT.
func (c *KerberosClient) LoadTGTFromKirbiBytes(data []byte) error {
	cred, err := kirbi.Parse(data)
	if err != nil {
		return err
	}
	return c.LoadTGT(cred)
}

// LoadTGTFromKirbiFile reads a .kirbi file and loads the TGT it carries.
func (c *KerberosClient) LoadTGTFromKirbiFile(path string) error {
	cred, err := kirbi.Load(path)
	if err != nil {
		return err
	}
	return c.LoadTGT(cred)
}

// LoadTGTFromCCache wires a Ticket Granting Ticket held in a parsed MIT ccache
// into the client for pass-the-ticket. The krbtgt/REALM credential is selected;
// if the cache holds none, the first credential is used. Username and realm are
// taken from the credential's client principal.
func (c *KerberosClient) LoadTGTFromCCache(cc *ccache.CCache) error {
	ticketRaw, info, err := selectCCacheCred(cc, true)
	if err != nil {
		return err
	}
	return c.applyTGT(ticketRaw, info)
}

// LoadTGTFromCCacheBytes parses MIT ccache (v4) bytes and loads the TGT they
// carry.
func (c *KerberosClient) LoadTGTFromCCacheBytes(data []byte) error {
	cc, err := ccache.Unmarshal(data)
	if err != nil {
		return err
	}
	return c.LoadTGTFromCCache(cc)
}

// LoadTGTFromCCacheFile reads an MIT ccache file and loads the TGT it carries.
func (c *KerberosClient) LoadTGTFromCCacheFile(path string) error {
	cc, err := ccache.Load(path)
	if err != nil {
		return err
	}
	return c.LoadTGTFromCCache(cc)
}

// LoadTGTFromCCacheEnv loads the TGT from the ccache named by the KRB5CCNAME
// environment variable. A leading "FILE:" type prefix (the only cache type this
// package implements) is accepted and stripped. It errors if KRB5CCNAME is unset
// or names a non-FILE cache.
func (c *KerberosClient) LoadTGTFromCCacheEnv() error {
	path, err := ccachePathFromEnv()
	if err != nil {
		return err
	}
	return c.LoadTGTFromCCacheFile(path)
}

// ServiceTicket is a service ticket recovered from a .kirbi or ccache, in the
// shape GetTGS returns: the parsed Ticket, the raw APPLICATION[1] bytes for
// verbatim re-emission in an AP-REQ, and the associated session key. It supports
// silver-ticket-style reuse — presenting a captured service ticket to a single
// service without contacting the KDC.
type ServiceTicket struct {
	// Ticket is the parsed service ticket.
	Ticket messages.Ticket
	// TicketRaw is the raw APPLICATION[1] ticket TLV (feed to messages.APReq{TicketRaw}).
	TicketRaw []byte
	// SessionKey seals the AP-REQ Authenticator presented to the service.
	SessionKey []byte
	// SessionEType is the encryption type of SessionKey.
	SessionEType int
	// Client is the client principal the ticket was issued to.
	Client messages.PrincipalName
	// CRealm is the client's realm.
	CRealm string
	// SName is the service principal the ticket is for.
	SName messages.PrincipalName
	// SRealm is the service's realm.
	SRealm string
}

// LoadServiceTicketFromKirbiBytes parses .kirbi bytes and returns the service
// ticket they carry for pass-the-ticket to a single service. When the credential
// holds several tickets, spn selects one by service principal name ("service/host",
// matched case-insensitively); pass "" to take the first ticket.
func LoadServiceTicketFromKirbiBytes(data []byte, spn string) (*ServiceTicket, error) {
	cred, err := kirbi.Parse(data)
	if err != nil {
		return nil, err
	}
	return serviceTicketFromKirbi(cred, spn)
}

// LoadServiceTicketFromKirbiFile reads a .kirbi file and returns the selected
// service ticket. See LoadServiceTicketFromKirbiBytes for spn matching.
func LoadServiceTicketFromKirbiFile(path, spn string) (*ServiceTicket, error) {
	cred, err := kirbi.Load(path)
	if err != nil {
		return nil, err
	}
	return serviceTicketFromKirbi(cred, spn)
}

// LoadServiceTicketFromCCacheFile reads an MIT ccache file and returns the
// selected service ticket. See LoadServiceTicketFromKirbiBytes for spn matching.
func LoadServiceTicketFromCCacheFile(path, spn string) (*ServiceTicket, error) {
	cc, err := ccache.Load(path)
	if err != nil {
		return nil, err
	}
	return serviceTicketFromCCache(cc, spn)
}

// LoadServiceTicketFromCCacheBytes parses MIT ccache (v4) bytes and returns the
// selected service ticket. See LoadServiceTicketFromKirbiBytes for spn matching.
func LoadServiceTicketFromCCacheBytes(data []byte, spn string) (*ServiceTicket, error) {
	cc, err := ccache.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return serviceTicketFromCCache(cc, spn)
}

// LoadServiceTicket wires a service ticket (a forged silver ticket, or one
// recovered from a .kirbi/ccache) into the client so a subsequent GetTGS for its
// SPN returns it verbatim with no TGT and no KDC round-trip. Combined with the
// SPNEGO mechanism this is silver-ticket pass-the-ticket against SMB/RPC/LDAP.
func (c *KerberosClient) LoadServiceTicket(st *ServiceTicket) error {
	if st == nil {
		return fmt.Errorf("kerberos: nil service ticket")
	}
	if len(st.SessionKey) == 0 {
		return fmt.Errorf("kerberos: service ticket has no session key")
	}
	if len(st.SName.NameString) < 2 {
		return fmt.Errorf("kerberos: service ticket has no service principal name")
	}
	if c.preloadedTGS == nil {
		c.preloadedTGS = make(map[string]preloadedServiceTicket)
	}
	sessionKey := append([]byte(nil), st.SessionKey...)
	// Adopt the ticket's client identity when the client has none, so the AP-REQ
	// authenticator's cname/crealm match the ticket (else the service rejects with
	// KRB_AP_ERR_BADMATCH). A client created with only a KDC host thus becomes
	// usable straight after import.
	if c.username == "" && len(st.Client.NameString) > 0 {
		c.username = st.Client.NameString[0]
	}
	if c.realm == "" && st.CRealm != "" {
		c.realm = strings.ToUpper(st.CRealm)
	}
	spn := st.SName.NameString[0] + "/" + st.SName.NameString[1]
	c.preloadedTGS[normalizeSPN(spn)] = preloadedServiceTicket{
		ticket:       st.Ticket,
		ticketRaw:    st.TicketRaw,
		sessionKey:   sessionKey,
		sessionEType: st.SessionEType,
	}
	// Also retain it for export (harvest→save): a captured/forged ticket wired in
	// here can be written back out via ExportServiceTicketKirbi/CCache. The
	// ServiceTicket carries no flags/times, so those stay zero.
	c.cacheServiceTicket(st.TicketRaw, messages.EncTGSRepPart{
		Key:    messages.EncryptionKey{KeyType: st.SessionEType, KeyValue: sessionKey},
		SRealm: st.SRealm,
		SName:  st.SName,
	})
	return nil
}

// LoadForgedServiceTicket wires a forged (silver) ticket into the client for
// pass-the-ticket, the forging-side counterpart of LoadServiceTicket.
func (c *KerberosClient) LoadForgedServiceTicket(ft *ForgedTicket) error {
	if ft == nil {
		return fmt.Errorf("kerberos: nil forged ticket")
	}
	return c.LoadServiceTicket(&ServiceTicket{
		Ticket:       ft.Ticket,
		TicketRaw:    ft.TicketRaw,
		SessionKey:   ft.SessionKey,
		SessionEType: ft.SessionEType,
		Client:       ft.CredInfo.PName,
		CRealm:       ft.CredInfo.PRealm,
		SName:        ft.CredInfo.SName,
		SRealm:       ft.CredInfo.SRealm,
	})
}

// hasPreloadedServiceTicket reports whether a service ticket has been preloaded
// for the given SPN.
func (c *KerberosClient) hasPreloadedServiceTicket(spn string) bool {
	_, ok := c.preloadedTGS[normalizeSPN(spn)]
	return ok
}

// normalizeSPN reduces an SPN to a lowercase "service/host" key (dropping any
// @REALM suffix) for preloaded-ticket lookup.
func normalizeSPN(spn string) string {
	if at := strings.IndexByte(spn, '@'); at >= 0 {
		spn = spn[:at]
	}
	return strings.ToLower(spn)
}

// ── internal helpers ──────────────────────────────────────────────────────────

// applyTGT parses the raw ticket and populates the TGT state used by GetTGS and
// buildAPReq (realm, username, tgtTicket/tgtTicketRaw, sessionKey/sessionEType,
// tgtEnc, hasTGT). It mirrors processASRep for the imported-ticket path.
func (c *KerberosClient) applyTGT(ticketRaw []byte, info messages.KrbCredInfo) error {
	if len(info.Key.KeyValue) == 0 {
		return fmt.Errorf("kerberos: imported ticket has no session key (encrypted enc-part?)")
	}
	var tkt messages.Ticket
	if _, err := tkt.Unmarshal(ticketRaw); err != nil {
		return fmt.Errorf("kerberos: parse imported ticket: %w", err)
	}

	// The client identity is the ticket's client principal; adopt it so a client
	// created with only a KDC host is usable for pass-the-ticket.
	if info.PRealm != "" {
		c.realm = info.PRealm
	}
	if len(info.PName.NameString) > 0 {
		c.username = info.PName.NameString[0]
	}

	sessionKey := append([]byte(nil), info.Key.KeyValue...)
	info.Key.KeyValue = sessionKey
	c.tgtTicket = tkt
	c.tgtTicketRaw = append([]byte(nil), ticketRaw...)
	c.sessionKey = sessionKey
	c.sessionEType = info.Key.KeyType
	// Reconstruct the AS-REP enc-part so re-export (tgtCredInfo) and downstream
	// consumers see the same times/flags/service name the ticket was issued with.
	c.tgtEnc = messages.EncASRepPart{
		Key:       info.Key,
		Flags:     info.Flags,
		AuthTime:  info.AuthTime,
		StartTime: info.StartTime,
		EndTime:   info.EndTime,
		RenewTill: info.RenewTill,
		SRealm:    info.SRealm,
		SName:     info.SName,
	}
	c.hasTGT = true
	return nil
}

// selectCredInfo returns the raw ticket and ticket-info for the entry chosen from
// a KRB-CRED: when wantTGT is set the krbtgt/REALM entry is preferred, otherwise
// the first entry is used. The enc-part must be unencrypted (.kirbi convention).
func selectCredInfo(cred *messages.KRBCred, wantTGT bool) ([]byte, messages.KrbCredInfo, error) {
	infos, err := kirbi.TicketInfo(cred)
	if err != nil {
		return nil, messages.KrbCredInfo{}, err
	}
	if len(infos) == 0 || len(cred.TicketsRaw) == 0 {
		return nil, messages.KrbCredInfo{}, fmt.Errorf("kerberos: KRB-CRED carries no tickets")
	}

	idx := 0
	if wantTGT {
		for i := range infos {
			if i < len(cred.TicketsRaw) && isTGTName(infos[i].SName) {
				idx = i
				break
			}
		}
	}
	if idx >= len(cred.TicketsRaw) {
		return nil, messages.KrbCredInfo{}, fmt.Errorf("kerberos: KRB-CRED ticket-info/ticket count mismatch")
	}
	return cred.TicketsRaw[idx], infos[idx], nil
}

// serviceTicketFromKirbi selects a service ticket from a KRB-CRED by SPN (or the
// first, when spn is "") and returns it in ServiceTicket form.
func serviceTicketFromKirbi(cred *messages.KRBCred, spn string) (*ServiceTicket, error) {
	infos, err := kirbi.TicketInfo(cred)
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 || len(cred.TicketsRaw) == 0 {
		return nil, fmt.Errorf("kerberos: KRB-CRED carries no tickets")
	}
	idx := -1
	for i := range infos {
		if i >= len(cred.TicketsRaw) {
			break
		}
		if spn == "" || snameMatchesSPN(infos[i].SName, spn) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("kerberos: no ticket for SPN %q in KRB-CRED", spn)
	}
	info := infos[idx]
	if len(info.Key.KeyValue) == 0 {
		return nil, fmt.Errorf("kerberos: imported ticket has no session key (encrypted enc-part?)")
	}
	var tkt messages.Ticket
	if _, err := tkt.Unmarshal(cred.TicketsRaw[idx]); err != nil {
		return nil, fmt.Errorf("kerberos: parse imported ticket: %w", err)
	}
	return &ServiceTicket{
		Ticket:       tkt,
		TicketRaw:    cred.TicketsRaw[idx],
		SessionKey:   info.Key.KeyValue,
		SessionEType: info.Key.KeyType,
		Client:       info.PName,
		CRealm:       info.PRealm,
		SName:        info.SName,
		SRealm:       info.SRealm,
	}, nil
}

// selectCCacheCred returns the raw ticket and a synthesized KrbCredInfo for the
// credential chosen from a ccache: when wantTGT is set the krbtgt/REALM entry is
// preferred, otherwise the first credential is used. Configuration entries (the
// MIT convention of a "X-CACHECONF:" realm) are skipped.
func selectCCacheCred(cc *ccache.CCache, wantTGT bool) ([]byte, messages.KrbCredInfo, error) {
	creds := realCCacheCreds(cc)
	if len(creds) == 0 {
		return nil, messages.KrbCredInfo{}, fmt.Errorf("kerberos: ccache holds no credentials")
	}
	idx := 0
	if wantTGT {
		for i := range creds {
			sname := messages.PrincipalName{
				NameType:   int(creds[i].Server.NameType),
				NameString: creds[i].Server.Components,
			}
			if isTGTName(sname) {
				idx = i
				break
			}
		}
	}
	return creds[idx].Ticket, credInfoFromCCache(creds[idx]), nil
}

// serviceTicketFromCCache selects a service ticket from a ccache by SPN (or the
// first non-config credential, when spn is "").
func serviceTicketFromCCache(cc *ccache.CCache, spn string) (*ServiceTicket, error) {
	creds := realCCacheCreds(cc)
	if len(creds) == 0 {
		return nil, fmt.Errorf("kerberos: ccache holds no credentials")
	}
	idx := -1
	for i := range creds {
		sname := messages.PrincipalName{
			NameType:   int(creds[i].Server.NameType),
			NameString: creds[i].Server.Components,
		}
		if spn == "" || snameMatchesSPN(sname, spn) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("kerberos: no ticket for SPN %q in ccache", spn)
	}
	info := credInfoFromCCache(creds[idx])
	if len(info.Key.KeyValue) == 0 {
		return nil, fmt.Errorf("kerberos: imported ticket has no session key")
	}
	var tkt messages.Ticket
	if _, err := tkt.Unmarshal(creds[idx].Ticket); err != nil {
		return nil, fmt.Errorf("kerberos: parse imported ticket: %w", err)
	}
	return &ServiceTicket{
		Ticket:       tkt,
		TicketRaw:    creds[idx].Ticket,
		SessionKey:   info.Key.KeyValue,
		SessionEType: info.Key.KeyType,
		Client:       info.PName,
		CRealm:       info.PRealm,
		SName:        info.SName,
		SRealm:       info.SRealm,
	}, nil
}

// realCCacheCreds returns the credentials that carry an actual ticket, skipping
// MIT configuration entries (server realm "X-CACHECONF:") and any entry with no
// ticket bytes.
func realCCacheCreds(cc *ccache.CCache) []ccache.Credential {
	out := make([]ccache.Credential, 0, len(cc.Credentials))
	for i := range cc.Credentials {
		cr := cc.Credentials[i]
		if len(cr.Ticket) == 0 || cr.Server.Realm == "X-CACHECONF:" {
			continue
		}
		out = append(out, cr)
	}
	return out
}

// credInfoFromCCache converts a ccache Credential into the KrbCredInfo shape the
// import path consumes, translating the uint32 flags and Unix-seconds times back
// to their RFC 4120 forms.
func credInfoFromCCache(cr ccache.Credential) messages.KrbCredInfo {
	return messages.KrbCredInfo{
		Key:       messages.EncryptionKey{KeyType: int(cr.Key.EType), KeyValue: cr.Key.KeyValue},
		PRealm:    cr.Client.Realm,
		PName:     messages.PrincipalName{NameType: int(cr.Client.NameType), NameString: cr.Client.Components},
		Flags:     uint32ToFlags(cr.TicketFlags),
		AuthTime:  timeFromU32(cr.AuthTime),
		StartTime: timeFromU32(cr.StartTime),
		EndTime:   timeFromU32(cr.EndTime),
		RenewTill: timeFromU32(cr.RenewTill),
		SRealm:    cr.Server.Realm,
		SName:     messages.PrincipalName{NameType: int(cr.Server.NameType), NameString: cr.Server.Components},
	}
}

// isTGTName reports whether a service principal name is a krbtgt/REALM TGT.
func isTGTName(sname messages.PrincipalName) bool {
	return len(sname.NameString) >= 1 && strings.EqualFold(sname.NameString[0], "krbtgt")
}

// snameMatchesSPN reports whether a service principal name matches an SPN of the
// form "service/host" (or bare "service"), compared case-insensitively. A bare
// "service" matches on the service component alone.
func snameMatchesSPN(sname messages.PrincipalName, spn string) bool {
	if at := strings.IndexByte(spn, '@'); at >= 0 {
		spn = spn[:at]
	}
	want := strings.Split(spn, "/")
	if len(want) == 1 {
		return len(sname.NameString) >= 1 && strings.EqualFold(sname.NameString[0], want[0])
	}
	if len(sname.NameString) != len(want) {
		return false
	}
	for i := range want {
		if !strings.EqualFold(sname.NameString[i], want[i]) {
			return false
		}
	}
	return true
}

// uint32ToFlags converts a ccache ticket_flags big-endian uint32 back to the
// RFC 4120 32-bit KerberosFlags BIT STRING (bit 0 = MSB of the first octet). It
// is the inverse of flagsToUint32.
func uint32ToFlags(v uint32) asn1.BitString {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	return asn1.BitString{Bytes: b[:], BitLength: 32}
}

// timeFromU32 converts ccache Unix seconds back to a UTC time.Time, mapping 0
// (the ccache "unset" value) to the zero time so optional fields stay optional.
func timeFromU32(v uint32) time.Time {
	if v == 0 {
		return time.Time{}
	}
	return time.Unix(int64(v), 0).UTC()
}

// ccachePathFromEnv resolves the filesystem path from KRB5CCNAME, accepting a
// bare path or a "FILE:" prefix. Cache types other than FILE are unsupported.
func ccachePathFromEnv() (string, error) {
	name := os.Getenv("KRB5CCNAME")
	if name == "" {
		return "", fmt.Errorf("kerberos: KRB5CCNAME is not set")
	}
	if strings.HasPrefix(name, "FILE:") {
		return name[len("FILE:"):], nil
	}
	if i := strings.IndexByte(name, ':'); i >= 0 {
		// A type prefix other than FILE (e.g. DIR:, KEYRING:, KCM:).
		return "", fmt.Errorf("kerberos: unsupported KRB5CCNAME cache type %q (only FILE: is supported)", name[:i+1])
	}
	return name, nil
}
