package client

import (
	"crypto/rand"
	"encoding/asn1"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// SetAuthKerberos configures an authenticated bind using native Kerberos. It
// acquires a TGT (if the client does not already hold one) and a service ticket
// for spn, builds the KRB_AP_REQ carried in the bind's auth verifier, and
// installs a per-message security context. Mutual authentication is always
// requested so the acceptor's KRB_AP_REP can be verified.
//
// The auth level selects both the RPC security provider and the per-PDU
// protection:
//
//   - CONNECT authenticates the bind only, as raw Kerberos (auth_type 0x10) with
//     a GSS-wrapped AP-REQ and a single mutual-auth round.
//   - PKT / PKT_INTEGRITY protect each PDU with a GSS MIC. Windows requires these
//     to be negotiated through SPNEGO (auth_type 0x09, RPC_C_AUTHN_GSS_NEGOTIATE)
//     with GSS_C_DCE_STYLE: a three-leg handshake (bind AP-REQ, bind_ack AP-REP,
//     alter_context with the initiator's own AP-REP) establishes the context. The
//     MIC token type follows the negotiated session key: an RFC 4121 CFX MIC
//     (tok_id 04 04) for an AES ticket, or the RFC 4757 MIC (tok_id 01 01) for an
//     RC4 ticket.
//   - PKT_PRIVACY (sealing) uses the same SPNEGO/DCE-style handshake and, matching
//     the MIC path, seals the stub with the enctype-appropriate GSS Wrap token: an
//     RFC 4121 CFX Wrap (tok_id 05 04) for AES or the RFC 4757 Wrap (tok_id 02 01)
//     for RC4. The stub is encrypted in place while the PDU header and sec_trailer
//     stay sign-only, and the Wrap token travels in the auth_value.
//
// Call before Bind.
func (c *Client) SetAuthKerberos(authLevel uint8, kc *kerberos.KerberosClient, spn string) error {
	if kc == nil {
		return fmt.Errorf("dcerpc auth: nil kerberos client")
	}
	if spn == "" {
		return fmt.Errorf("dcerpc auth: kerberos requires a service principal name")
	}
	if authLevel == pdu.AuthLevelCall {
		authLevel = pdu.AuthLevelPkt
	}
	switch authLevel {
	case pdu.AuthLevelConnect, pdu.AuthLevelPkt, pdu.AuthLevelPktIntegrity, pdu.AuthLevelPktPrivacy:
	default:
		return fmt.Errorf("dcerpc auth: unsupported auth_level %d", authLevel)
	}
	perMessage := authLevel >= pdu.AuthLevelPkt

	if !kc.HasTGT() {
		if err := kc.GetTGT(); err != nil {
			return fmt.Errorf("dcerpc auth: kerberos GetTGT: %w", err)
		}
	}
	// The per-message token type follows the negotiated service-ticket session
	// key: an AES ticket yields RFC 4121 CFX tokens, an RC4 ticket the RFC 4757
	// tokens. No enctype is forced here — the client's TGS-REQ prefers AES, so an
	// AES-capable target produces AES per-message protection, and a target that
	// can only issue an RC4 session key transparently falls back to RC4 tokens.
	_, ticketRaw, key, keyEType, err := kc.GetTGS(spn, true)
	if err != nil {
		return fmt.Errorf("dcerpc auth: kerberos GetTGS %q: %w", spn, err)
	}

	// Windows requires GSS_C_DCE_STYLE for per-PDU protection (the three-leg
	// handshake). Connect-level authentication does not use it.
	flags := uint32(gssapi.GSSMutualFlag | gssapi.GSSReplayFlag | gssapi.GSSSequenceFlag | gssapi.GSSConfFlag | gssapi.GSSIntegFlag)
	if perMessage {
		flags |= gssapi.GSSDCEStyleFlag
	}
	initOpts := gssapi.InitOptions{
		TicketRaw:    ticketRaw,
		SessionKey:   key,
		SessionEType: keyEType,
		ClientName:   messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{kc.Username()}},
		ClientRealm:  kc.Realm(),
		Flags:        flags,
		Mutual:       true,
		// DCE/RPC seeds the acceptor's per-message receive sequence from the
		// authenticator sequence number and starts its per-PDU counter at 0.
		ZeroSeqNumber: perMessage,
	}
	if !perMessage {
		// Connect level asserts an initiator subkey (as generic Windows GSS clients
		// do). The DCE-style per-message path omits it so the acceptor chooses the
		// subkey, matching the Windows RPC client.
		subKey := make([]byte, kerbcrypto.KeyLen(keyEType))
		if _, err := rand.Read(subKey); err != nil {
			return err
		}
		initOpts.SubKey = subKey
		initOpts.SubKeyEType = keyEType
	}
	apReq, ctx, err := gssapi.InitSecContext(initOpts)
	if err != nil {
		return fmt.Errorf("dcerpc auth: kerberos InitSecContext: %w", err)
	}

	sec := &kerberosSecurityContext{ctx: ctx, spnego: perMessage}
	c.authLevel = authLevel
	c.sec = sec
	// A non-zero auth_context_id identifies the security context across the bind,
	// alter_context, and request PDUs (Windows correlates the per-PDU verifier to
	// the established context by this id; a zero id is not honored).
	c.authContextID = 79231
	if perMessage {
		// RPC_C_AUTHN_GSS_NEGOTIATE: the bind carries a SPNEGO NegTokenInit that
		// advertises the (MS) Kerberos mechanism and wraps the AP-REQ as its
		// mechToken.
		c.authType = pdu.AuthTypeGSSNegotiate
		bindToken, err := (&spnego.NegTokenInit{
			MechTypes: []asn1.ObjectIdentifier{spnego.MSKerberosOID},
			MechToken: apReq,
		}).Marshal()
		if err != nil {
			return fmt.Errorf("dcerpc auth: marshal SPNEGO NegTokenInit: %w", err)
		}
		c.bindToken = bindToken
	} else {
		// RPC_C_AUTHN_GSS_KERBEROS: the bind carries the bare GSS Kerberos token.
		c.authType = pdu.AuthTypeGSSKerberos
		c.bindToken = apReq
	}
	return nil
}

// kerberosSecurityContext adapts a GSS-API SecContext to the RPC SecurityContext
// interface. The per-PDU verifier is a GSS token over the request/response stub:
// a MIC (integrity only, stub in the clear) for PKT/PKT_INTEGRITY, or a Wrap token
// that seals the stub for PKT_PRIVACY. The token family (RFC 4121 CFX for AES,
// RFC 4757 for RC4) follows the negotiated session/subkey enctype. It also
// implements bindCompleter to complete the mutual-auth handshake.
type kerberosSecurityContext struct {
	ctx *gssapi.SecContext
	// spnego is set for the per-message SPNEGO/DCE-style path: the acceptor's
	// tokens are SPNEGO NegTokenResp envelopes carrying bare Kerberos AP-REP
	// messages, and a third leg (the initiator's own AP-REP) is required.
	spnego bool
}

// CompleteBind completes the mutual-authentication handshake from the bind_ack's
// auth_value and returns an optional continuation token for the third leg. For
// the SPNEGO/DCE-style path it unwraps the NegTokenResp, verifies the acceptor's
// bare AP-REP, and returns its own AP-REP wrapped in a NegTokenResp (sent in an
// alter_context). For connect-level raw Kerberos it verifies the GSS-wrapped
// AP-REP and returns no continuation.
func (k *kerberosSecurityContext) CompleteBind(bindAckAuthValue []byte) ([]byte, error) {
	if len(bindAckAuthValue) == 0 {
		return nil, nil
	}
	if !k.spnego {
		return nil, k.ctx.AcceptAPRep(bindAckAuthValue)
	}

	// The acceptor's NegTokenResp is carried in a [1] context tag; unwrap it to
	// the bare SEQUENCE the parser expects.
	var wrapped asn1.RawValue
	if _, err := asn1.Unmarshal(bindAckAuthValue, &wrapped); err != nil {
		return nil, fmt.Errorf("dcerpc auth: parse SPNEGO response wrapper: %w", err)
	}
	respBytes := bindAckAuthValue
	if wrapped.Class == asn1.ClassContextSpecific && wrapped.Tag == 1 {
		respBytes = wrapped.Bytes
	}
	var resp spnego.NegTokenResp
	if _, err := resp.Unmarshal(respBytes); err != nil {
		return nil, fmt.Errorf("dcerpc auth: parse SPNEGO NegTokenResp: %w", err)
	}
	if resp.NegState == spnego.NegStateReject {
		return nil, fmt.Errorf("dcerpc auth: SPNEGO negotiation rejected")
	}
	// resp.ResponseToken is a bare KRB_AP_REP (SPNEGO carries the raw mechanism
	// token in continuation exchanges).
	if err := k.ctx.AcceptAPRepRaw(resp.ResponseToken); err != nil {
		return nil, err
	}
	apRep, err := k.ctx.MakeAPRep()
	if err != nil {
		return nil, err
	}
	// Per-PDU GSS tokens use a zero-based sequence counter after the DCE-style
	// handshake completes.
	k.ctx.ResetSendSeq(0)
	// The third leg is the initiator's AP-REP in a NegTokenResp carrying only the
	// responseToken (no negState, no mechListMIC), wrapped in the [1] tag.
	legSeq, err := (&spnego.NegTokenResp{ResponseToken: apRep, SuppressNegState: true}).Marshal()
	if err != nil {
		return nil, fmt.Errorf("dcerpc auth: marshal SPNEGO leg-3: %w", err)
	}
	return asn1.Marshal(asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 1, IsCompound: true, Bytes: legSeq})
}

// AuthValueLen is the GSS token length carried in the auth_value: for sealing
// (PKT_PRIVACY) the Wrap token, otherwise the MIC token. The length depends on
// the negotiated enctype (CFX header plus the etype checksum for AES, or the
// fixed RFC 4757 layout for RC4).
func (k *kerberosSecurityContext) AuthValueLen(seal bool, stubLen int) int {
	if seal {
		return k.ctx.WrapTokenLen(stubLen)
	}
	return k.ctx.MICTokenLen()
}

// ProtectRequest protects the request stub. For PKT_PRIVACY it seals the stub
// with a GSS Wrap token and returns the encrypted stub plus the Wrap token; for
// PKT/PKT_INTEGRITY it signs the stub with a GSS MIC token and returns the stub
// unchanged. Per MS-RPCE the Kerberos per-message tokens cover only the stub
// (pduData), not the PDU header or sec_trailer, so signedRegion is unused.
func (k *kerberosSecurityContext) ProtectRequest(signedRegion, stub []byte, seal bool) ([]byte, []byte, error) {
	if seal {
		sealed, token, err := k.ctx.Seal(stub)
		if err != nil {
			return nil, nil, err
		}
		return sealed, token, nil
	}
	mic, err := k.ctx.MakeMIC(stub)
	if err != nil {
		return nil, nil, err
	}
	return stub, mic, nil
}

// UnprotectResponse recovers the response stub. For PKT_PRIVACY it decrypts and
// verifies the GSS Wrap token, returning the plaintext stub; otherwise it
// verifies the GSS MIC over the (cleartext) stub and returns it unchanged.
func (k *kerberosSecurityContext) UnprotectResponse(signedRegion, stub, authValue []byte, seal bool) ([]byte, error) {
	if seal {
		return k.ctx.Unseal(stub, authValue)
	}
	if err := k.ctx.VerifyMIC(stub, authValue); err != nil {
		return nil, err
	}
	return stub, nil
}
