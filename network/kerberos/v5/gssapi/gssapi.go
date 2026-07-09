// Package gssapi implements the Kerberos V5 GSS-API mechanism (RFC 1964 /
// RFC 4121) context-establishment tokens: the GSS InitialContextToken framing,
// the 0x8003 authenticator checksum that carries channel bindings and service
// flags, building the KRB_AP_REQ token from a service ticket (InitSecContext),
// and verifying the KRB_AP_REP mutual-authentication reply.
//
// This is the mechanism that plugs into SPNEGO (as the Kerberos mechToken) to
// authenticate SMB, RPC, and LDAP. Per-message tokens (MIC/Wrap, RFC 4121 §4.2)
// are a separate layer built on the SecContext produced here.
package gssapi

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// MechOIDKerberos5 is the GSS-API mechanism OID for Kerberos V5
// (1.2.840.113554.1.2.2), used as thisMech in the InitialContextToken and as a
// mechType in SPNEGO negotiation.
var MechOIDKerberos5 = asn1.ObjectIdentifier{1, 2, 840, 113554, 1, 2, 2}

// GSS-API token identifiers (RFC 1964 §1.1): the 2-byte TOK_ID prefixing the
// Kerberos message inside the InitialContextToken.
var (
	TokIDAPReq = [2]byte{0x01, 0x00} // KRB_AP_REQ
	TokIDAPRep = [2]byte{0x02, 0x00} // KRB_AP_REP
	TokIDError = [2]byte{0x03, 0x00} // KRB_ERROR
)

// GSS context-establishment flags (RFC 1964 §1.1.1), as they appear in the
// 0x8003 checksum Flags field.
const (
	GSSDelegFlag    = 1
	GSSMutualFlag   = 2
	GSSReplayFlag   = 4
	GSSSequenceFlag = 8
	GSSConfFlag     = 16
	GSSIntegFlag    = 32
	// GSSDCEStyleFlag (GSS_C_DCE_STYLE) selects the DCE RPC three-leg mutual
	// authentication style. Windows RPC requires it for per-message protection
	// (PKT_INTEGRITY / PKT_PRIVACY); with it the acceptor's AP-REP is followed by
	// a third leg carrying the initiator's own AP-REP.
	GSSDCEStyleFlag = 0x1000
)

// ChecksumTypeGSSAPI is the Kerberos checksum type (0x8003) used for the GSS-API
// authenticator checksum.
const ChecksumTypeGSSAPI = 0x8003

// GSSChecksumValue builds the 0x8003 authenticator checksum value (RFC 1964
// §1.1.1): Lgth (16, little-endian) | Bnd | Flags (little-endian). Bnd is the
// MD5 of the channel bindings, or 16 zero bytes for GSS_C_NO_BINDINGS (nil).
func GSSChecksumValue(flags uint32, channelBindings []byte) []byte {
	out := make([]byte, 24)
	binary.LittleEndian.PutUint32(out[0:], 16) // Lgth
	if len(channelBindings) > 0 {
		sum := md5.Sum(channelBindings)
		copy(out[4:20], sum[:])
	} // else 16 zero bytes
	binary.LittleEndian.PutUint32(out[20:], flags)
	return out
}

// WrapToken wraps a Kerberos message in the GSS-API InitialContextToken framing
// (RFC 1964 §1.1): [APPLICATION 0] { kerberos5-OID, TOK_ID | krbMessage }.
func WrapToken(tokID [2]byte, krbMessage []byte) ([]byte, error) {
	oidBytes, err := asn1.Marshal(MechOIDKerberos5)
	if err != nil {
		return nil, err
	}
	inner := make([]byte, 0, len(oidBytes)+2+len(krbMessage))
	inner = append(inner, oidBytes...)
	inner = append(inner, tokID[0], tokID[1])
	inner = append(inner, krbMessage...)
	return asn1.Marshal(asn1.RawValue{Class: asn1.ClassApplication, Tag: 0, IsCompound: true, Bytes: inner})
}

// UnwrapToken parses a GSS-API InitialContextToken, verifies the kerberos5 mech
// OID, and returns the TOK_ID and the enclosed Kerberos message bytes.
func UnwrapToken(data []byte) (tokID [2]byte, krbMessage []byte, err error) {
	var outer asn1.RawValue
	if _, err = asn1.Unmarshal(data, &outer); err != nil {
		return tokID, nil, fmt.Errorf("gssapi: parse InitialContextToken: %w", err)
	}
	if outer.Class != asn1.ClassApplication || outer.Tag != 0 {
		return tokID, nil, fmt.Errorf("gssapi: not an InitialContextToken (class=%d tag=%d)", outer.Class, outer.Tag)
	}
	var oid asn1.ObjectIdentifier
	rest, err := asn1.Unmarshal(outer.Bytes, &oid)
	if err != nil {
		return tokID, nil, fmt.Errorf("gssapi: parse mech OID: %w", err)
	}
	if !oid.Equal(MechOIDKerberos5) {
		return tokID, nil, fmt.Errorf("gssapi: unexpected mech OID %v", oid)
	}
	if len(rest) < 2 {
		return tokID, nil, fmt.Errorf("gssapi: token too short for TOK_ID")
	}
	tokID[0], tokID[1] = rest[0], rest[1]
	return tokID, rest[2:], nil
}

// SecContext holds the state a context initiator needs to verify the AP-REP and
// to produce/verify per-message tokens.
type SecContext struct {
	// SessionKey / SessionEType are the service-ticket session key.
	SessionKey   []byte
	SessionEType int
	// SubKey is the client-chosen subkey placed in the authenticator (optional).
	SubKey      []byte
	SubKeyEType int
	// SeqNumber is the initial sequence number sent in the authenticator.
	SeqNumber int
	// ctime/cusec are retained to match the AP-REP echo (mutual authentication).
	ctime time.Time
	cusec int
	// sendSeq is the running sequence number for outgoing per-message tokens.
	sendSeq uint64
	// acceptorSubkey records that the base key is an acceptor-asserted subkey,
	// so per-message tokens must set the AcceptorSubkey flag.
	acceptorSubkey bool
	// acceptorSeq is the sequence number the acceptor placed in its AP-REP. The
	// DCE-style third-leg AP-REP must echo it (krb5_rd_rep_dce checks it).
	acceptorSeq int
}

// InitOptions configures InitSecContext.
type InitOptions struct {
	// TicketRaw is the raw APPLICATION[1] service ticket (from GetTGS).
	TicketRaw []byte
	// SessionKey / SessionEType are that ticket's session key.
	SessionKey   []byte
	SessionEType int
	// ClientName / ClientRealm identify the initiator (the ticket's client).
	ClientName  messages.PrincipalName
	ClientRealm string
	// Flags are the GSS context-establishment flags for the 0x8003 checksum.
	Flags uint32
	// ChannelBindings is hashed into the checksum Bnd field (nil = no bindings).
	ChannelBindings []byte
	// Mutual requests mutual authentication (sets AP-options mutual-required and
	// the GSS mutual flag), so the acceptor returns an AP-REP.
	Mutual bool
	// SubKey optionally supplies a client subkey (used to key per-message tokens);
	// if set, SubKeyEType must be set too.
	SubKey      []byte
	SubKeyEType int
	// ZeroSeqNumber sets the authenticator sequence number to 0 instead of a
	// random value. DCE/RPC requires it (the acceptor's per-message receive
	// sequence is seeded from it, and the per-PDU counter starts at 0).
	ZeroSeqNumber bool
}

// InitSecContext builds the initiator's KRB_AP_REQ GSS token from a service
// ticket, returning the wrapped token and the SecContext for AP-REP/per-message
// handling. The authenticator carries the 0x8003 checksum and is encrypted with
// the ticket session key at key usage 11 (AP-REQ authenticator).
func InitSecContext(opts InitOptions) ([]byte, *SecContext, error) {
	if len(opts.TicketRaw) == 0 || len(opts.SessionKey) == 0 {
		return nil, nil, fmt.Errorf("gssapi: InitSecContext requires a ticket and session key")
	}

	flags := opts.Flags
	if opts.Mutual {
		flags |= GSSMutualFlag
	}

	now := time.Now().UTC()
	cusec := now.Nanosecond() / 1000

	// The initiator's authenticator sequence number seeds the acceptor's expected
	// per-message receive sequence. DCE/RPC (which starts its per-PDU counter at 0)
	// requires it to be 0; other callers use a random value.
	seqNum := 0
	if !opts.ZeroSeqNumber {
		var seqBuf [4]byte
		if _, err := rand.Read(seqBuf[:]); err != nil {
			return nil, nil, err
		}
		seqNum = int(binary.BigEndian.Uint32(seqBuf[:]) & 0x7fffffff)
	}

	cksum := &messages.Checksum{
		CKSumType: ChecksumTypeGSSAPI,
		Checksum:  GSSChecksumValue(flags, opts.ChannelBindings),
	}
	auth := &messages.Authenticator{
		AVno:      messages.KerberosV5,
		CRealm:    opts.ClientRealm,
		CName:     opts.ClientName,
		Cksum:     cksum,
		CUSec:     cusec,
		CTime:     now,
		SeqNumber: seqNum,
	}
	if len(opts.SubKey) > 0 {
		auth.SubKey = &messages.EncryptionKey{KeyType: opts.SubKeyEType, KeyValue: opts.SubKey}
	}

	authBytes, err := auth.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("gssapi: marshal authenticator: %w", err)
	}
	encAuth, err := kerbcrypto.Encrypt(opts.SessionEType, opts.SessionKey, kerbcrypto.KeyUsageAPReqAuthen, authBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("gssapi: encrypt authenticator: %w", err)
	}

	apOptions := asn1.BitString{Bytes: []byte{0x00, 0x00, 0x00, 0x00}, BitLength: 32}
	if opts.Mutual {
		// mutual-required is APOptions bit 2 -> byte 0, 0x20.
		apOptions.Bytes[0] |= 0x20
	}
	apReq := &messages.APReq{
		PVNO:          messages.KerberosV5,
		MsgType:       messages.MsgTypeAPReq,
		APOptions:     apOptions,
		TicketRaw:     opts.TicketRaw,
		Authenticator: messages.EncryptedData{EType: opts.SessionEType, Cipher: encAuth},
	}
	apReqBytes, err := apReq.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("gssapi: marshal AP-REQ: %w", err)
	}

	token, err := WrapToken(TokIDAPReq, apReqBytes)
	if err != nil {
		return nil, nil, err
	}

	ctx := &SecContext{
		SessionKey:   opts.SessionKey,
		SessionEType: opts.SessionEType,
		SubKey:       opts.SubKey,
		SubKeyEType:  opts.SubKeyEType,
		SeqNumber:    seqNum,
		ctime:        now.Truncate(time.Second),
		cusec:        cusec,
		sendSeq:      uint64(seqNum),
	}
	return token, ctx, nil
}

// AcceptAPRep verifies a KRB_AP_REP GSS token returned by the acceptor for
// mutual authentication: it unwraps the token, decrypts the EncAPRepPart with
// the ticket session key (key usage 12), and checks that the echoed ctime/cusec
// match those sent in the initiator's authenticator. On success it adopts an
// acceptor subkey into the context if one was supplied.
func (ctx *SecContext) AcceptAPRep(token []byte) error {
	tokID, krbMsg, err := UnwrapToken(token)
	if err != nil {
		return err
	}
	if tokID != TokIDAPRep {
		return fmt.Errorf("gssapi: expected AP-REP token (02 00), got %02x %02x", tokID[0], tokID[1])
	}
	return ctx.acceptAPRepMessage(krbMsg)
}

// AcceptAPRepRaw verifies a bare KRB_AP_REP (APPLICATION[15]) that is not wrapped
// in a GSS InitialContextToken. DCE RPC (GSS_C_DCE_STYLE) carries the acceptor's
// AP-REP this way, since the GSS framing OID appears only on the first token.
func (ctx *SecContext) AcceptAPRepRaw(apRep []byte) error {
	return ctx.acceptAPRepMessage(apRep)
}

// acceptAPRepMessage decrypts and verifies a KRB_AP_REP message body: it checks
// the echoed ctime/cusec against the initiator's authenticator (mutual
// authentication) and adopts an acceptor subkey into the context if present.
func (ctx *SecContext) acceptAPRepMessage(krbMsg []byte) error {
	var apRep messages.APRep
	if _, err := apRep.Unmarshal(krbMsg); err != nil {
		return fmt.Errorf("gssapi: parse AP-REP: %w", err)
	}
	plain, err := kerbcrypto.Decrypt(ctx.SessionEType, ctx.SessionKey, kerbcrypto.KeyUsageAPRepEncPart, apRep.EncPart.Cipher)
	if err != nil {
		return fmt.Errorf("gssapi: decrypt AP-REP enc-part: %w", err)
	}
	var enc messages.EncAPRepPart
	if _, err := enc.Unmarshal(plain); err != nil {
		return fmt.Errorf("gssapi: parse EncAPRepPart: %w", err)
	}
	if !enc.CTime.Equal(ctx.ctime) || enc.CUSec != ctx.cusec {
		return fmt.Errorf("gssapi: AP-REP ctime/cusec mismatch (mutual authentication failed)")
	}
	if enc.SubKey != nil {
		ctx.SubKey = enc.SubKey.KeyValue
		ctx.SubKeyEType = enc.SubKey.KeyType
		ctx.acceptorSubkey = true
	}
	ctx.acceptorSeq = enc.SeqNumber
	return nil
}

// MakeAPRep builds the initiator's own KRB_AP_REP as a bare APPLICATION[15]
// message (no GSS wrapper), for the third leg of the DCE-style mutual-auth
// handshake (GSS_C_DCE_STYLE). Following the Windows/MIT behaviour, it uses a
// fresh timestamp (not the echoed authenticator time), carries the sequence
// number the acceptor sent in its AP-REP, has no subkey, and is encrypted with
// the ticket session key (key usage 12).
func (ctx *SecContext) MakeAPRep() ([]byte, error) {
	now := time.Now().UTC()
	enc := messages.EncAPRepPart{
		CTime: now,
		CUSec: now.Nanosecond() / 1000,
		// The DCE-style acceptor validates that the third-leg AP-REP echoes the
		// sequence number it sent in its own AP-REP (krb5_rd_rep_dce).
		SeqNumber: ctx.acceptorSeq,
	}
	encBytes, err := enc.Marshal()
	if err != nil {
		return nil, fmt.Errorf("gssapi: marshal EncAPRepPart: %w", err)
	}
	cipher, err := kerbcrypto.Encrypt(ctx.SessionEType, ctx.SessionKey, kerbcrypto.KeyUsageAPRepEncPart, encBytes)
	if err != nil {
		return nil, fmt.Errorf("gssapi: encrypt EncAPRepPart: %w", err)
	}
	apRep := messages.APRep{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeAPRep,
		EncPart: messages.EncryptedData{EType: ctx.SessionEType, Cipher: cipher},
	}
	return apRep.Marshal()
}
