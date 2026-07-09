package client

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/security"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// rpcNTLMNegotiateFlags are the flags proposed in the NTLM NEGOTIATE message for an
// authenticated bind. They request extended session security (NTLMv2) with signing and
// sealing at 128/56-bit strength, matching what the AUTHENTICATE message commits to.
// Key exchange is deliberately not negotiated, so the exported session key equals the
// NTLMv2 session base key and the signature checksum is not re-encrypted.
const rpcNTLMNegotiateFlags = flags.NTLMSSP_NEGOTIATE_UNICODE |
	flags.NTLMSSP_REQUEST_TARGET |
	flags.NTLMSSP_NEGOTIATE_SIGN |
	flags.NTLMSSP_NEGOTIATE_SEAL |
	flags.NTLMSSP_NEGOTIATE_NTLM |
	flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
	flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
	flags.NTLMSSP_NEGOTIATE_128 |
	flags.NTLMSSP_NEGOTIATE_56

// SetAuth enables RPC-level authentication for subsequent binds. authType selects the
// security provider (only NTLM, pdu.AuthTypeNTLMSSP, is supported) and authLevel the
// protection applied to each PDU: pdu.AuthLevelConnect authenticates the bind only;
// pdu.AuthLevelCall and pdu.AuthLevelPkt attach a per-PDU authenticity verifier;
// pdu.AuthLevelPktIntegrity additionally signs the request data; and
// pdu.AuthLevelPktPrivacy signs and seals it. creds may carry a cleartext password or,
// for pass-the-hash, an NT hash with no password. It must be called before Bind.
//
// AuthLevelCall is promoted to AuthLevelPkt: it has no distinct meaning for the
// connection-oriented protocol, where the runtime uses packet-level protection instead
// ([MS-RPCE] 2.2.1.1.8).
func (c *Client) SetAuth(authType, authLevel uint8, creds *credentials.Credentials) error {
	if authType != pdu.AuthTypeNTLMSSP {
		return fmt.Errorf("dcerpc auth: unsupported auth_type 0x%02x (only NTLM 0x%02x is supported)", authType, pdu.AuthTypeNTLMSSP)
	}
	if authLevel == pdu.AuthLevelCall {
		authLevel = pdu.AuthLevelPkt
	}
	switch authLevel {
	case pdu.AuthLevelConnect, pdu.AuthLevelPkt, pdu.AuthLevelPktIntegrity, pdu.AuthLevelPktPrivacy:
	default:
		return fmt.Errorf("dcerpc auth: unsupported auth_level %d", authLevel)
	}
	if creds == nil {
		return fmt.Errorf("dcerpc auth: nil credentials")
	}
	c.authType = authType
	c.authLevel = authLevel
	c.creds = creds
	return nil
}

// SetAuthProvider configures an authenticated bind with a caller-supplied single-leg
// security provider, for SSPs whose session key is established out of band and whose bind
// completes in one round trip (no challenge/auth3) — Netlogon (auth_type 0x44) in
// particular. sec protects and unprotects each PDU; bindToken is the auth_value carried in
// the bind PDU (for Netlogon, a marshalled NL_AUTH_MESSAGE), or nil. This is the seam that
// keeps the client free of any specific SSP's dependencies: the provider and its bind token
// are built by the caller. Unlike SetAuth (NTLM), the security context is active
// immediately, so protected calls work as soon as Bind returns. Call before Bind.
func (c *Client) SetAuthProvider(authType, authLevel uint8, sec SecurityContext, bindToken []byte) error {
	if sec == nil {
		return fmt.Errorf("dcerpc auth: nil security provider")
	}
	if authType == pdu.AuthTypeNone {
		return fmt.Errorf("dcerpc auth: auth_type must not be none")
	}
	// Netlogon authenticates the bind with an NL_AUTH_MESSAGE token; without it the bind
	// carries a sec_trailer with an empty auth_value and the DC rejects the association,
	// which would otherwise surface only as an opaque bind or first-call failure.
	if authType == pdu.AuthTypeNetlogon && len(bindToken) == 0 {
		return fmt.Errorf("dcerpc auth: netlogon (0x%02x) requires a bind token (NL_AUTH_MESSAGE)", pdu.AuthTypeNetlogon)
	}
	if authLevel == pdu.AuthLevelCall {
		authLevel = pdu.AuthLevelPkt
	}
	switch authLevel {
	case pdu.AuthLevelConnect, pdu.AuthLevelPkt, pdu.AuthLevelPktIntegrity, pdu.AuthLevelPktPrivacy:
	default:
		return fmt.Errorf("dcerpc auth: unsupported auth_level %d", authLevel)
	}
	c.authType = authType
	c.authLevel = authLevel
	c.sec = sec
	c.bindToken = bindToken
	return nil
}

// authConfigured reports whether an authenticated bind was selected, by either SetAuth
// (NTLM, which supplies creds and derives the context during Bind) or SetAuthProvider (a
// single-leg provider, which supplies the context up front).
func (c *Client) authConfigured() bool {
	return c.authType != pdu.AuthTypeNone && (c.creds != nil || c.sec != nil)
}

// protectsRequests reports whether each request PDU must carry a per-PDU auth verifier.
// PKT, PKT_INTEGRITY, and PKT_PRIVACY all attach one (NTLM produces the same signature
// for the first two; only PKT_PRIVACY additionally seals the stub); CONNECT
// authenticates the bind alone. The levels are an ascending scale ([MS-RPCE]
// 2.2.1.1.8), so a single threshold separates the per-PDU levels from CONNECT.
func (c *Client) protectsRequests() bool {
	return c.sec != nil && c.authLevel >= pdu.AuthLevelPkt
}

// authVerifierOverhead is the worst-case number of bytes an authenticated request PDU
// adds after the stub: up to 3 padding bytes to 4-byte-align the trailer, the 8-byte
// sec_trailer, and the auth_value token. The token size is provider-specific (NTLM: 16;
// Netlogon: larger), so it is taken from the active security context at its sealing level.
func (c *Client) authVerifierOverhead() int {
	// A negative stub length requests the largest token the provider can emit, so
	// the reservation is safe for any fragment stub size.
	return 3 + pdu.SecTrailerSize + c.sec.AuthValueLen(c.authLevel == pdu.AuthLevelPktPrivacy, -1)
}

// negotiateToken builds the auth_value carried in the bind's auth verifier. For a single-leg
// provider (SetAuthProvider) it is the caller-supplied bind token; for NTLM it is a freshly
// built NEGOTIATE message.
func (c *Client) negotiateToken() ([]byte, error) {
	if c.authType != pdu.AuthTypeNTLMSSP {
		return c.bindToken, nil
	}
	neg, err := negotiate.CreateNegotiateMessage("", c.workstation, rpcNTLMNegotiateFlags, nil)
	if err != nil {
		return nil, err
	}
	return neg.Marshal()
}

// buildAuthenticate constructs the NTLM AUTHENTICATE message for the configured
// credentials. When the credentials carry an NT hash but no password it authenticates
// by pass-the-hash; otherwise it uses the cleartext password.
func (c *Client) buildAuthenticate(chal *challenge.ChallengeMessage) (*authenticate.AuthenticateMessage, error) {
	if c.creds.GetPassword() == "" && c.creds.CanPassTheHash() {
		raw, err := hex.DecodeString(c.creds.GetNTHash())
		if err != nil {
			return nil, fmt.Errorf("decode NT hash: %w", err)
		}
		if len(raw) != 16 {
			return nil, fmt.Errorf("NT hash must be 16 bytes, got %d", len(raw))
		}
		var ntHash [16]byte
		copy(ntHash[:], raw)
		return authenticate.CreateAuthenticateMessageWithNTHash(chal, c.creds.GetUsername(), ntHash, c.creds.GetDomain(), c.workstation)
	}
	return authenticate.CreateAuthenticateMessage(chal, c.creds.GetUsername(), c.creds.GetPassword(), c.creds.GetDomain(), c.workstation)
}

// completeAuth finishes the authenticated bind after a bind_ack. For a single-leg provider
// (SetAuthProvider, e.g. Netlogon) the security context is already active and the session
// key was established out of band, so there is no auth3 and nothing to derive — the
// bind_ack (which for Netlogon carries the server's NL_AUTH_MESSAGE) is simply accepted.
// For NTLM it parses the server's CHALLENGE from the bind_ack's auth_value, sends an auth3
// carrying the AUTHENTICATE token (to which the server sends no reply), and builds the
// per-message security context from the derived session key.
//
// It returns an optional continuation token: under Kerberos GSS_C_DCE_STYLE the
// handshake has a third leg (the initiator's AP-REP), which the caller sends in
// an alter_context PDU. NTLM and single-leg providers return a nil continuation.
func (c *Client) completeAuth(bindAckFrag []byte) ([]byte, error) {
	if c.authType != pdu.AuthTypeNTLMSSP {
		// A single-leg provider whose context needs the bind_ack's auth_value to
		// finish establishing (Kerberos mutual auth: verify the KRB_AP_REP and adopt
		// any acceptor subkey). Providers without this need (Netlogon) are unchanged.
		if bc, ok := c.sec.(bindCompleter); ok {
			_, authValue, err := pdu.ExtractAuthVerifier(bindAckFrag)
			if err != nil {
				return nil, fmt.Errorf("read bind_ack auth verifier: %w", err)
			}
			return bc.CompleteBind(authValue)
		}
		return nil, nil
	}
	_, challengeBytes, err := pdu.ExtractAuthVerifier(bindAckFrag)
	if err != nil {
		return nil, fmt.Errorf("read challenge: %w", err)
	}
	if len(challengeBytes) == 0 {
		return nil, fmt.Errorf("server returned no NTLM challenge in bind_ack")
	}

	var chal challenge.ChallengeMessage
	if _, err := chal.Unmarshal(challengeBytes); err != nil {
		return nil, fmt.Errorf("parse challenge: %w", err)
	}

	auth, err := c.buildAuthenticate(&chal)
	if err != nil {
		return nil, fmt.Errorf("build authenticate: %w", err)
	}
	authBytes, err := auth.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal authenticate: %w", err)
	}

	a3 := &pdu.Auth3{
		SecTrailer: pdu.SecTrailer{AuthType: c.authType, AuthLevel: c.authLevel, AuthContextID: c.authContextID},
		AuthValue:  authBytes,
	}
	a3.Header = pdu.NewHeader(pdu.PacketTypeAuth3, pdu.PFCFirstFrag|pdu.PFCLastFrag, c.callID)
	raw, err := a3.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal auth3: %w", err)
	}
	if err := c.transport.Send(raw); err != nil {
		return nil, fmt.Errorf("send auth3: %w", err)
	}

	if len(auth.SessionKey) == 0 {
		return nil, fmt.Errorf("no session key derived (NTLMv1 is not supported for RPC auth)")
	}
	sec, err := security.NewContext(auth.SessionKey, auth.NegotiateFlags)
	if err != nil {
		return nil, fmt.Errorf("init security context: %w", err)
	}
	c.sec = &ntlmSecurityContext{ctx: sec}
	c.sessionKey = append([]byte(nil), auth.SessionKey...)
	return nil, nil
}

// sendAuthContinuation sends the DCE-style third-leg continuation token (the
// initiator's KRB_AP_REP, wrapped in a SPNEGO NegTokenResp) in an alter_context
// PDU carrying the same presentation contexts, and consumes the
// alter_context_response that completes context establishment. Windows negotiates
// the Kerberos DCE-style third leg via alter_context (which the server answers),
// not auth3.
func (c *Client) sendAuthContinuation(contexts []pdu.ContextElement, authValue []byte) error {
	alter := &pdu.Bind{
		MaxXmitFrag:  c.transport.MaxXmitFrag(),
		MaxRecvFrag:  c.transport.MaxRecvFrag(),
		AssocGroupID: 0,
		ContextList:  contexts,
		SecTrailer:   pdu.SecTrailer{AuthType: c.authType, AuthLevel: c.authLevel, AuthContextID: c.authContextID},
		AuthValue:    authValue,
	}
	alter.Header = pdu.NewHeader(pdu.PacketTypeAlterContext, pdu.PFCFirstFrag|pdu.PFCLastFrag, c.callID)
	raw, err := alter.Marshal()
	if err != nil {
		return fmt.Errorf("marshal alter_context: %w", err)
	}
	// Bind.Marshal forces PTYPE=bind; an alter_context PDU is wire-identical to a
	// bind but with PTYPE=alter_context (offset 2 of the common header).
	if len(raw) > 2 {
		raw[2] = byte(pdu.PacketTypeAlterContext)
	}
	if err := c.transport.Send(raw); err != nil {
		return fmt.Errorf("send alter_context: %w", err)
	}
	respFrag, err := c.readFragment(&fragmentReader{t: c.transport})
	if err != nil {
		return fmt.Errorf("recv alter_context_response: %w", err)
	}
	hdr, err := pdu.PeekHeader(respFrag)
	if err != nil {
		return fmt.Errorf("alter_context_response: %w", err)
	}
	switch hdr.PacketType {
	case pdu.PacketTypeAlterContextResp:
		return nil
	case pdu.PacketTypeBindNak:
		var nak pdu.BindNak
		if _, err := nak.Unmarshal(respFrag); err != nil {
			return fmt.Errorf("parse bind_nak: %w", err)
		}
		return fmt.Errorf("alter_context rejected: reject_reason=%d", nak.RejectReason)
	default:
		return fmt.Errorf("unexpected alter_context response PDU type %s", hdr.PacketType)
	}
}

// marshalProtectedRequest serializes a request PDU with a per-PDU auth verifier. The stub
// is padded so the sec_trailer starts on a 4-byte boundary; the security context signs the
// PDU (for PKT_INTEGRITY) or signs and seals the stub (for PKT_PRIVACY). The auth_value
// token format and which bytes are signed are provider-specific — the client supplies both
// the whole signed region (header, body, padded stub, sec_trailer, over plaintext) and the
// padded stub, and the provider uses whichever it needs.
func (c *Client) marshalProtectedRequest(req *pdu.Request) ([]byte, error) {
	if req.ObjectUUID != nil {
		return nil, fmt.Errorf("authenticated request with object UUID is not supported")
	}

	seal := c.authLevel == pdu.AuthLevelPktPrivacy

	allocHint := req.AllocHint
	if allocHint == 0 {
		allocHint = uint32(len(req.Stub))
	}
	bodyHdr := make([]byte, 8)
	binary.LittleEndian.PutUint32(bodyHdr[0:4], allocHint)
	binary.LittleEndian.PutUint16(bodyHdr[4:6], req.ContextID)
	binary.LittleEndian.PutUint16(bodyHdr[6:8], req.Opnum)

	// The 16-byte header plus the 8-byte body keep the stub on a 4-byte boundary, so the
	// pad needed before the sec_trailer depends only on the stub length.
	pad := (4 - (len(req.Stub) % 4)) % 4
	stubPad := make([]byte, len(req.Stub)+pad)
	copy(stubPad, req.Stub)

	// The auth_value length can depend on the padded stub length (the AES Kerberos
	// Wrap token grows with the stub's block padding), so it is computed here.
	tokenLen := c.sec.AuthValueLen(seal, len(stubPad))

	st := pdu.SecTrailer{
		AuthType:      c.authType,
		AuthLevel:     c.authLevel,
		AuthPadLength: uint8(pad),
		AuthContextID: c.authContextID,
	}

	req.Header.PacketType = pdu.PacketTypeRequest
	req.Header.AuthLength = uint16(tokenLen)
	fragLen := pdu.HeaderSize + len(bodyHdr) + len(stubPad) + pdu.SecTrailerSize + tokenLen
	req.Header.FragLength = uint16(fragLen)
	hdrBytes, err := req.Header.Marshal()
	if err != nil {
		return nil, err
	}

	// The signed region is the whole PDU up to (but not including) the auth_value,
	// computed over the plaintext stub.
	signedRegion := make([]byte, 0, fragLen-tokenLen)
	signedRegion = append(signedRegion, hdrBytes...)
	signedRegion = append(signedRegion, bodyHdr...)
	signedRegion = append(signedRegion, stubPad...)
	signedRegion = append(signedRegion, st.Marshal()...)

	onWireStub, authValue, err := c.sec.ProtectRequest(signedRegion, stubPad, seal)
	if err != nil {
		return nil, err
	}
	if len(authValue) != tokenLen {
		return nil, fmt.Errorf("auth token length %d does not match reserved length %d", len(authValue), tokenLen)
	}

	// A sealing provider may expand the stub (the Kerberos RC4 GSS Wrap seals the
	// stub in an 8-octet block, so the on-wire stub can be longer than the padded
	// plaintext). The fragment length must reflect the actual on-wire stub;
	// re-marshal the header with the corrected frag_length. The RC4 token does not
	// cover the PDU header, so adjusting it after signing is safe.
	if len(onWireStub) != len(stubPad) {
		fragLen = pdu.HeaderSize + len(bodyHdr) + len(onWireStub) + pdu.SecTrailerSize + tokenLen
		req.Header.FragLength = uint16(fragLen)
		if hdrBytes, err = req.Header.Marshal(); err != nil {
			return nil, err
		}
	}

	out := make([]byte, 0, fragLen)
	out = append(out, hdrBytes...)
	out = append(out, bodyHdr...)
	out = append(out, onWireStub...)
	out = append(out, st.Marshal()...)
	out = append(out, authValue...)
	return out, nil
}

// unprotectResponseStub recovers the cleartext stub from an authenticated response
// fragment: the security context decrypts the stub in place (for PKT_PRIVACY) and verifies
// its per-PDU token, after which the auth padding is stripped. The token length and the
// verified region are provider-specific, so both the signed region (whole PDU minus the
// auth_value) and the stub sub-slice are handed to the provider.
func (c *Client) unprotectResponseStub(frag []byte) ([]byte, error) {
	hdr, err := pdu.PeekHeader(frag)
	if err != nil {
		return nil, err
	}
	seal := c.authLevel == pdu.AuthLevelPktPrivacy
	authLen := int(hdr.AuthLength)
	// A signing token is a fixed size, so verify the auth_length up front. Sealed
	// (PKT_PRIVACY) Kerberos Wrap tokens vary in length with the sender's block
	// padding (the AES CFX EC filler), so their length is validated by the
	// provider's Unseal rather than compared to a fixed expectation here.
	if !seal {
		if want := c.sec.AuthValueLen(seal, 0); authLen != want {
			return nil, fmt.Errorf("unexpected auth_length %d, want %d", authLen, want)
		}
	}
	fragLen := int(hdr.FragLength)
	if fragLen > len(frag) {
		fragLen = len(frag)
	}

	st, authValue, err := pdu.ExtractAuthVerifier(frag)
	if err != nil {
		return nil, err
	}

	stubStart := pdu.HeaderSize + 8 // response body: alloc_hint, ctx_id, cancel_count, reserved
	stubEnd := fragLen - authLen - pdu.SecTrailerSize
	if stubEnd < stubStart {
		return nil, fmt.Errorf("authenticated response stub bounds invalid")
	}

	// signedRegion is the whole PDU up to the auth_value; stub is a sub-slice of it, so an
	// in-place decrypt by the provider is reflected in signedRegion before verification.
	signedRegion := frag[:fragLen-authLen]
	plainStub, err := c.sec.UnprotectResponse(signedRegion, frag[stubStart:stubEnd], authValue, seal)
	if err != nil {
		return nil, err
	}

	if int(st.AuthPadLength) > len(plainStub) {
		return nil, fmt.Errorf("auth_pad_length %d exceeds stub length", st.AuthPadLength)
	}
	stub := plainStub[:len(plainStub)-int(st.AuthPadLength)]
	return append([]byte(nil), stub...), nil
}
