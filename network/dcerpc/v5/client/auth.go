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

// authConfigured reports whether SetAuth selected an authenticated bind.
func (c *Client) authConfigured() bool { return c.creds != nil && c.authType != pdu.AuthTypeNone }

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
// sec_trailer, and the 16-byte signature.
const authVerifierOverhead = 3 + pdu.SecTrailerSize + security.SignatureSize

// negotiateToken builds the raw NTLM NEGOTIATE token carried in the bind's auth_value.
func (c *Client) negotiateToken() ([]byte, error) {
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

// completeAuth finishes the NTLM exchange after a bind_ack: it parses the server's
// CHALLENGE from the bind_ack's auth_value, sends an auth3 carrying the AUTHENTICATE
// token (to which the server sends no reply), and builds the per-message security
// context from the derived session key.
func (c *Client) completeAuth(bindAckFrag []byte) error {
	_, challengeBytes, err := pdu.ExtractAuthVerifier(bindAckFrag)
	if err != nil {
		return fmt.Errorf("read challenge: %w", err)
	}
	if len(challengeBytes) == 0 {
		return fmt.Errorf("server returned no NTLM challenge in bind_ack")
	}

	var chal challenge.ChallengeMessage
	if _, err := chal.Unmarshal(challengeBytes); err != nil {
		return fmt.Errorf("parse challenge: %w", err)
	}

	auth, err := c.buildAuthenticate(&chal)
	if err != nil {
		return fmt.Errorf("build authenticate: %w", err)
	}
	authBytes, err := auth.Marshal()
	if err != nil {
		return fmt.Errorf("marshal authenticate: %w", err)
	}

	a3 := &pdu.Auth3{
		SecTrailer: pdu.SecTrailer{AuthType: c.authType, AuthLevel: c.authLevel, AuthContextID: c.authContextID},
		AuthValue:  authBytes,
	}
	a3.Header = pdu.NewHeader(pdu.PacketTypeAuth3, pdu.PFCFirstFrag|pdu.PFCLastFrag, c.callID)
	raw, err := a3.Marshal()
	if err != nil {
		return fmt.Errorf("marshal auth3: %w", err)
	}
	if err := c.transport.Send(raw); err != nil {
		return fmt.Errorf("send auth3: %w", err)
	}

	if len(auth.SessionKey) == 0 {
		return fmt.Errorf("no session key derived (NTLMv1 is not supported for RPC auth)")
	}
	sec, err := security.NewContext(auth.SessionKey, auth.NegotiateFlags)
	if err != nil {
		return fmt.Errorf("init security context: %w", err)
	}
	c.sec = sec
	c.sessionKey = append([]byte(nil), auth.SessionKey...)
	return nil
}

// marshalProtectedRequest serializes a request PDU with an NTLM auth verifier. The stub
// is padded so the sec_trailer starts on a 4-byte boundary; for PKT_INTEGRITY the
// signature covers the whole PDU (header, body, padded stub, sec_trailer) in cleartext,
// and for PKT_PRIVACY the same region is signed over plaintext while only the stub and
// its padding are sealed ([MS-RPCE] 3.3, NTLM2).
func (c *Client) marshalProtectedRequest(req *pdu.Request) ([]byte, error) {
	if req.ObjectUUID != nil {
		return nil, fmt.Errorf("authenticated request with object UUID is not supported")
	}

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

	st := pdu.SecTrailer{
		AuthType:      c.authType,
		AuthLevel:     c.authLevel,
		AuthPadLength: uint8(pad),
		AuthContextID: c.authContextID,
	}

	req.Header.PacketType = pdu.PacketTypeRequest
	req.Header.AuthLength = uint16(security.SignatureSize)
	fragLen := pdu.HeaderSize + len(bodyHdr) + len(stubPad) + pdu.SecTrailerSize + security.SignatureSize
	req.Header.FragLength = uint16(fragLen)
	hdrBytes, err := req.Header.Marshal()
	if err != nil {
		return nil, err
	}

	// The signed region is the whole PDU up to (but not including) the auth_value,
	// computed over the plaintext stub.
	toSign := make([]byte, 0, fragLen-security.SignatureSize)
	toSign = append(toSign, hdrBytes...)
	toSign = append(toSign, bodyHdr...)
	toSign = append(toSign, stubPad...)
	toSign = append(toSign, st.Marshal()...)

	var onWireStub []byte
	var sig [security.SignatureSize]byte
	if c.authLevel == pdu.AuthLevelPktPrivacy {
		onWireStub, sig = c.sec.SealWith(toSign, stubPad)
	} else {
		sig = c.sec.Sign(toSign)
		onWireStub = stubPad
	}

	out := make([]byte, 0, fragLen)
	out = append(out, hdrBytes...)
	out = append(out, bodyHdr...)
	out = append(out, onWireStub...)
	out = append(out, st.Marshal()...)
	out = append(out, sig[:]...)
	return out, nil
}

// unprotectResponseStub recovers the cleartext stub from an authenticated response
// fragment: for PKT_PRIVACY it decrypts the stub in place, then it verifies the
// signature over the whole PDU (minus the auth_value) and strips the auth padding.
func (c *Client) unprotectResponseStub(frag []byte) ([]byte, error) {
	hdr, err := pdu.PeekHeader(frag)
	if err != nil {
		return nil, err
	}
	authLen := int(hdr.AuthLength)
	if authLen != security.SignatureSize {
		return nil, fmt.Errorf("unexpected auth_length %d, want %d", authLen, security.SignatureSize)
	}
	fragLen := int(hdr.FragLength)
	if fragLen > len(frag) {
		fragLen = len(frag)
	}

	st, authValue, err := pdu.ExtractAuthVerifier(frag)
	if err != nil {
		return nil, err
	}
	var sig [security.SignatureSize]byte
	copy(sig[:], authValue)

	stubStart := pdu.HeaderSize + 8 // response body: alloc_hint, ctx_id, cancel_count, reserved
	stubEnd := fragLen - authLen - pdu.SecTrailerSize
	if stubEnd < stubStart {
		return nil, fmt.Errorf("authenticated response stub bounds invalid")
	}

	if c.authLevel == pdu.AuthLevelPktPrivacy {
		c.sec.DecryptInbound(frag[stubStart:stubEnd])
	}
	// Verify over the whole PDU minus the trailing auth_value, now that the stub is
	// plaintext.
	if err := c.sec.VerifySignature(frag[:fragLen-authLen], sig); err != nil {
		return nil, err
	}

	if int(st.AuthPadLength) > stubEnd-stubStart {
		return nil, fmt.Errorf("auth_pad_length %d exceeds stub length", st.AuthPadLength)
	}
	stub := frag[stubStart : stubEnd-int(st.AuthPadLength)]
	return append([]byte(nil), stub...), nil
}
