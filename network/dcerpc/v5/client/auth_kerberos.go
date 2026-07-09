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
//     alter_context with the initiator's own AP-REP) establishes the context, and
//     an RC4-HMAC service ticket is requested (the RFC 4757 per-message tokens
//     Windows expects here).
//   - PKT_PRIVACY (sealing) additionally needs a GSS Wrap-IOV encrypt and is not
//     yet supported.
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
	case pdu.AuthLevelConnect, pdu.AuthLevelPkt, pdu.AuthLevelPktIntegrity:
	case pdu.AuthLevelPktPrivacy:
		return fmt.Errorf("dcerpc auth: kerberos PKT_PRIVACY (sealing) is not yet supported; use PKT_INTEGRITY")
	default:
		return fmt.Errorf("dcerpc auth: unsupported auth_level %d", authLevel)
	}
	perMessage := authLevel >= pdu.AuthLevelPkt

	if !kc.HasTGT() {
		if err := kc.GetTGT(); err != nil {
			return fmt.Errorf("dcerpc auth: kerberos GetTGT: %w", err)
		}
	}
	if perMessage {
		// Windows RPC's DCE-style per-message protection interoperates with RC4 on
		// this class of server; request an RC4 service ticket.
		kc.PreferRC4ServiceTicket()
	}
	ticket, ticketRaw, key, err := kc.GetTGS(spn, true)
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
		SessionEType: ticket.EncPart.EType,
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
		subKey := make([]byte, kerbcrypto.KeyLen(ticket.EncPart.EType))
		if _, err := rand.Read(subKey); err != nil {
			return err
		}
		initOpts.SubKey = subKey
		initOpts.SubKeyEType = ticket.EncPart.EType
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
// interface. The per-PDU verifier is a GSS MIC token (RFC 4121 §4.2.6.1) over the
// request/response stub; the stub itself travels in the clear (integrity only).
// It also implements bindCompleter to complete the mutual-auth handshake.
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

// AuthValueLen is the GSS MIC token length: a 16-byte token header plus the
// etype's checksum. It is independent of the seal flag (sealing is unsupported).
func (k *kerberosSecurityContext) AuthValueLen(bool) int {
	return k.ctx.MICTokenLen()
}

// ProtectRequest signs the request stub with a GSS MIC token and returns it as
// the auth_value. Per MS-RPCE the MIC covers only the stub (pduData), not the
// PDU header or sec_trailer, when PFC_SUPPORT_HEADER_SIGN is not negotiated.
func (k *kerberosSecurityContext) ProtectRequest(signedRegion, stub []byte, seal bool) ([]byte, []byte, error) {
	if seal {
		return nil, nil, fmt.Errorf("dcerpc auth: kerberos sealing is not supported")
	}
	mic, err := k.ctx.MakeMIC(stub)
	if err != nil {
		return nil, nil, err
	}
	return stub, mic, nil
}

// UnprotectResponse verifies the server's GSS MIC over the response stub and
// returns the stub unchanged (integrity only).
func (k *kerberosSecurityContext) UnprotectResponse(signedRegion, stub, authValue []byte, seal bool) ([]byte, error) {
	if seal {
		return nil, fmt.Errorf("dcerpc auth: kerberos sealing is not supported")
	}
	if err := k.ctx.VerifyMIC(stub, authValue); err != nil {
		return nil, err
	}
	return stub, nil
}
