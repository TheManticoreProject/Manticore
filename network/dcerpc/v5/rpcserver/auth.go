package rpcserver

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/security"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// NTLMAuth configures the NTLM-authenticated binds an endpoint accepts
// (RPC_C_AUTHN_WINNT, auth_type 10).
type NTLMAuth struct {
	// CredentialLookup resolves a claimed identity to its NT hash; false means
	// the identity is unknown and the bind's AUTH3 fails.
	CredentialLookup func(domain, username string) (ntHash [16]byte, ok bool)

	// TargetName is the name advertised in the CHALLENGE, and TargetType says
	// whether it names a domain or a server.
	TargetName string
	TargetType challenge.TargetType

	// TargetInfo is the AV_PAIR list advertised in the CHALLENGE (see
	// targetinfo.BuildServerTargetInfo). Clients such as impacket require
	// MsvAvDnsComputerName in it.
	TargetInfo []byte
}

// EnableNTLM makes the endpoint accept NTLM-authenticated binds at the
// connect, packet, integrity and privacy levels. Call it before the dispatcher
// serves any association.
func (d *Dispatcher) EnableNTLM(auth NTLMAuth) {
	d.ntlm = &auth
}

// authState is what an authenticated bind established on an association.
type authState struct {
	authType  uint8
	authLevel uint8
	contextID uint32
	acceptor  *spnego.AcceptContext
	// sec protects requests and responses once the AUTH3 has been verified; it
	// is nil until then, and stays nil when the AUTH3 fails.
	sec *security.Context
	// established records a verified AUTH3.
	established bool
}

// protectsMessages reports whether the level signs (and possibly seals) every
// request and response.
func (s *authState) protectsMessages() bool {
	return s.authLevel >= pdu.AuthLevelPkt
}

// startAuth answers the NTLM NEGOTIATE carried by an authenticated bind with a
// CHALLENGE, and returns the state the bind establishes.
func (d *Dispatcher) startAuth(trailer *pdu.SecTrailer, token []byte) (*authState, []byte, error) {
	if d.ntlm == nil {
		return nil, nil, fmt.Errorf("this endpoint accepts no authenticated bind")
	}
	if trailer.AuthType != pdu.AuthTypeNTLMSSP {
		return nil, nil, fmt.Errorf("auth_type %d is not supported", trailer.AuthType)
	}
	// The call level (3) is refused rather than given connect-level treatment:
	// a client asking for it expects every call protected, and the in-repo
	// client promotes it to the packet level for the same reason.
	if trailer.AuthLevel != pdu.AuthLevelConnect && (trailer.AuthLevel < pdu.AuthLevelPkt || trailer.AuthLevel > pdu.AuthLevelPktPrivacy) {
		return nil, nil, fmt.Errorf("auth_level %d is not supported", trailer.AuthLevel)
	}
	acceptor := spnego.NewAcceptContext(d.ntlm.TargetName, d.ntlm.TargetType, d.ntlm.TargetInfo)
	acceptor.CredentialLookup = d.ntlm.CredentialLookup
	if _, err := acceptor.AcceptNegotiateToken(token); err != nil {
		return nil, nil, err
	}
	state := &authState{
		authType:  trailer.AuthType,
		authLevel: trailer.AuthLevel,
		contextID: trailer.AuthContextID,
		acceptor:  acceptor,
	}
	// auth_type 10 carries bare NTLMSSP messages, so the CHALLENGE goes back
	// without the SPNEGO wrapping AcceptNegotiateToken returns.
	return state, acceptor.ChallengeMessageBytes, nil
}

// handleAuth3 completes an authenticated bind with the client's AUTHENTICATE.
// An AUTH3 is never answered. One that fails verification leaves the
// association unauthenticated, so every request on it is faulted.
func (a *Association) handleAuth3(request []byte) ([]byte, error) {
	auth3 := &pdu.Auth3{}
	if _, err := auth3.Unmarshal(request); err != nil {
		logger.Debugf("rpcserver: an AUTH3 PDU would not decode: %v", err)
		return nil, nil
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	state := a.auth
	if state == nil || state.established {
		logger.Debugf("rpcserver: an AUTH3 with no authenticated bind pending")
		return nil, nil
	}
	if auth3.SecTrailer.AuthType != state.authType || auth3.SecTrailer.AuthContextID != state.contextID {
		logger.Debugf("rpcserver: an AUTH3 for another security context")
		return nil, nil
	}
	if err := state.acceptor.AcceptAuthenticateToken(auth3.AuthValue); err != nil {
		logger.Debugf("rpcserver: the AUTH3 token was refused: %v", err)
		return nil, nil
	}
	if err := state.acceptor.Verify(); err != nil {
		domain, user, _ := state.acceptor.Identity()
		logger.Debugf("rpcserver: NTLM authentication of %s\\%s failed: %v", domain, user, err)
		return nil, nil
	}
	negotiated := state.acceptor.Authenticate.NegotiateFlags
	// Sealing with a 40- or 56-bit key is refused, as Windows does by default.
	if state.authLevel == pdu.AuthLevelPktPrivacy && !negotiated.HasFlag(flags.NTLMSSP_NEGOTIATE_128) {
		logger.Debugf("rpcserver: refusing privacy-level authentication without 128-bit keys")
		return nil, nil
	}
	sec, err := security.NewAcceptorContext(state.acceptor.SessionKey, negotiated)
	if err != nil {
		logger.Debugf("rpcserver: no security context for the session: %v", err)
		return nil, nil
	}
	state.sec, state.established = sec, true
	return nil, nil
}

// unprotectRequest verifies (and, at the privacy level, decrypts) one request
// fragment and returns its plaintext stub without the auth padding. The
// signature covers the whole PDU up to the auth_value, computed over the
// plaintext stub, the same rule the client signs requests by.
func (s *authState) unprotectRequest(fragment []byte, header *pdu.Header, objectUUID bool) ([]byte, error) {
	trailer, authValue, err := pdu.ExtractAuthVerifier(fragment)
	if err != nil {
		return nil, err
	}
	if trailer == nil {
		return nil, fmt.Errorf("the request carries no auth verifier")
	}
	if trailer.AuthType != s.authType || trailer.AuthLevel != s.authLevel || trailer.AuthContextID != s.contextID {
		return nil, fmt.Errorf("the request's sec_trailer does not match the bound security context")
	}
	if len(authValue) != security.SignatureSize {
		return nil, fmt.Errorf("auth_length %d, want %d", len(authValue), security.SignatureSize)
	}

	fragLength := int(header.FragLength)
	authLength := int(header.AuthLength)
	stubStart := pdu.HeaderSize + 8
	if objectUUID {
		stubStart += 16
	}
	stubEnd := fragLength - authLength - pdu.SecTrailerSize
	if fragLength > len(fragment) || stubEnd < stubStart {
		return nil, fmt.Errorf("authenticated request stub bounds invalid")
	}

	// Work on a copy so the caller's buffer is left alone; the stub is a
	// sub-slice of the signed region, so decrypting it in place updates both.
	buffer := append([]byte(nil), fragment[:fragLength-authLength]...)
	stub := buffer[stubStart:stubEnd]
	if s.authLevel == pdu.AuthLevelPktPrivacy {
		s.sec.DecryptInbound(stub)
	}
	var signature [security.SignatureSize]byte
	copy(signature[:], authValue)
	if err := s.sec.VerifySignature(buffer, signature); err != nil {
		return nil, err
	}
	if int(trailer.AuthPadLength) > len(stub) {
		return nil, fmt.Errorf("auth_pad_length %d exceeds the stub", trailer.AuthPadLength)
	}
	return append([]byte(nil), stub[:len(stub)-int(trailer.AuthPadLength)]...), nil
}

// stripConnectVerifier checks the sec_trailer of a connect-level request names
// the bound security context and removes its auth padding from stub.
func (s *authState) stripConnectVerifier(fragment, stub []byte) ([]byte, error) {
	trailer, _, err := pdu.ExtractAuthVerifier(fragment)
	if err != nil {
		return nil, err
	}
	if trailer.AuthType != s.authType || trailer.AuthContextID != s.contextID {
		return nil, fmt.Errorf("the request's sec_trailer does not match the bound security context")
	}
	if int(trailer.AuthPadLength) > len(stub) {
		return nil, fmt.Errorf("auth_pad_length %d exceeds the stub", trailer.AuthPadLength)
	}
	return stub[:len(stub)-int(trailer.AuthPadLength)], nil
}

// secTrailerAlign is the boundary the sec_trailer is padded to, measured from
// the start of the stub.
const secTrailerAlign = 16

// protectResponse frames one response fragment with an auth verifier: the stub
// is padded so the sec_trailer is 16-byte aligned, the PDU is signed, and at the
// privacy level the padded stub is encrypted.
func (s *authState) protectResponse(header pdu.Header, allocHint uint32, contextID uint16, stub []byte) ([]byte, error) {
	pad := (secTrailerAlign - len(stub)%secTrailerAlign) % secTrailerAlign
	padded := make([]byte, len(stub)+pad)
	copy(padded, stub)

	body := make([]byte, 8)
	binary.LittleEndian.PutUint32(body[0:4], allocHint)
	binary.LittleEndian.PutUint16(body[4:6], contextID)

	trailer := pdu.SecTrailer{AuthType: s.authType, AuthLevel: s.authLevel, AuthPadLength: uint8(pad), AuthContextID: s.contextID}
	header.PacketType = pdu.PacketTypeResponse
	header.AuthLength = security.SignatureSize
	header.FragLength = uint16(pdu.HeaderSize + len(body) + len(padded) + pdu.SecTrailerSize + security.SignatureSize)
	encodedHeader, err := header.Marshal()
	if err != nil {
		return nil, err
	}

	signed := make([]byte, 0, int(header.FragLength))
	signed = append(signed, encodedHeader...)
	signed = append(signed, body...)
	signed = append(signed, padded...)
	signed = append(signed, trailer.Marshal()...)

	onWire, signature := padded, [security.SignatureSize]byte{}
	if s.authLevel == pdu.AuthLevelPktPrivacy {
		onWire, signature = s.sec.SealWith(signed, padded)
	} else {
		signature = s.sec.Sign(signed)
	}

	out := make([]byte, 0, int(header.FragLength))
	out = append(out, encodedHeader...)
	out = append(out, body...)
	out = append(out, onWire...)
	out = append(out, trailer.Marshal()...)
	out = append(out, signature[:]...)
	return out, nil
}

// authOverhead is the most a protected response fragment adds to its stub:
// up to 15 bytes of padding, the sec_trailer and the signature.
const authOverhead = secTrailerAlign - 1 + pdu.SecTrailerSize + security.SignatureSize
