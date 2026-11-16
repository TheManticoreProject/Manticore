package server

import (
	"crypto/subtle"
	"errors"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv1"
	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// sessionSetupNonExtendedWordCount is the WordCount of a SESSION_SETUP_ANDX
// request made without extended security ([MS-CIFS] 2.2.4.53.1).
const sessionSetupNonExtendedWordCount = 0x0D

// sessionSetupNonExtendedResponseDataOffset is where the data block of the
// non-extended SESSION_SETUP_ANDX response begins, measured from the start of the
// SMB header: SMB_HEADER_SIZE(32) + WordCount(1) + 3 words(6) + ByteCount(2). It
// is odd, so a Unicode NativeOS needs one Pad byte ([MS-CIFS] 2.2.4.53.2).
const sessionSetupNonExtendedResponseDataOffset = 41

// ntlmv1ResponseLength is the length of an NTLMv1 (and LM) challenge response.
// Anything longer in the Unicode password field is an NTLMv2 response.
const ntlmv1ResponseLength = 24

// nonExtendedAuthentication answers a session setup made without extended
// security ([MS-CIFS] 3.3.5.43). It is a single leg: the client already holds the
// challenge from the negotiate response, so the request carries the account name
// and its responses, and the reply either establishes the session or refuses it.
//
// The NT response in UnicodePassword is verified against the credential store:
// a 24-byte response as NTLMv1, a longer one as NTLMv2. The LM response in
// OEMPassword is not used. The admission rules for unknown identities, null
// sessions and signing are the ones the extended path applies.
func (c *Connection) nonExtendedAuthentication(
	w ResponseWriter,
	req *message.Message,
	request *commands.SessionSetupAndxRequest,
) nt_status.NT_STATUS {
	decode := func(b []byte) string {
		if request.IsUnicode() {
			return utf16.DecodeUTF16LE(b)
		}
		return string(b)
	}
	username := decode(request.AccountName.Buffer)
	domain := decode(request.PrimaryDomain.Buffer)
	ntResponse := []byte(request.UnicodePassword)

	uid, err := c.uids.Allocate()
	if err != nil {
		logger.Warnf("SMB1 server: refusing a session setup from %s: %v", c.Remote, err)
		return nt_status.NT_STATUS_TOO_MANY_SESSIONS
	}

	session := &Session{
		UID:      uid,
		Domain:   domain,
		Username: username,
		Created:  time.Now().UTC(),
	}

	var macKey []byte
	sessionKey, verifyErr := c.verifyNonExtendedResponse(domain, username, ntResponse)
	switch {
	case verifyErr == nil:
		session.SessionKey = sessionKey
		// Without extended security the signature hashes the session key and
		// then the challenge response ahead of the message ([MS-CIFS] 3.1.4.1),
		// and 3.3.5.43 takes that response from UnicodePassword when it carries
		// the NT response. Their concatenation is the MAC key.
		macKey = append(append([]byte{}, sessionKey...), ntResponse...)

	case errors.Is(verifyErr, spnego.ErrUnknownIdentity), errors.Is(verifyErr, spnego.ErrCaptureOnly):
		// The same admission as the extended path: a null session only under
		// AllowAnonymous, any other unknown identity only under AllowGuest.
		anonymous := username == "" && len(ntResponse) == 0
		admitted := c.Server.config.AllowGuest
		if anonymous {
			admitted = c.Server.config.AllowAnonymous
		}
		if !admitted {
			logger.Debugf("SMB1 server: refusing %s from %s: %v", session.Account(), c.Remote, verifyErr)
			c.uids.Release(uid)
			return nt_status.NT_STATUS_LOGON_FAILURE
		}
		session.IsGuest = true
		session.IsAnonymous = anonymous

	default:
		logger.Debugf("SMB1 server: refusing %s from %s: %v", session.Account(), c.Remote, verifyErr)
		c.uids.Release(uid)
		return nt_status.NT_STATUS_LOGON_FAILURE
	}

	armSigning, status := c.signingDecision(req, session)
	if status != nt_status.NT_STATUS_SUCCESS {
		c.uids.Release(uid)
		return status
	}
	// Signing is armed once per connection, by the first exchange that can, and
	// at the same sequence numbers as the extended path ([MS-CIFS] 3.3.5.3).
	if armSigning && !c.SigningActive {
		c.SigningActive = true
		c.SigningKey = macKey
		c.ExpectedRequestSequenceNumber = signing.NextRequestSequenceNumber(authenticateRequestSequenceNumber)
		w.SignResponse(macKey, signing.ResponseSequenceNumber(authenticateRequestSequenceNumber))

		logger.Debugf("SMB1 server: signing armed for %s on UID 0x%04X", c.Remote, uid)
	}

	response := c.newSessionSetupResponse()
	response.NonExtendedSecurity = true
	if c.UseUnicode && sessionSetupNonExtendedResponseDataOffset%2 != 0 {
		response.Pad = []types.UCHAR{0x00}
	}
	response.PrimaryDomain = encodeNativeString(c.Server.config.DomainName, c.UseUnicode)
	if session.IsGuest {
		response.Action = types.USHORT(SMB_SETUP_GUEST)
	}

	c.addSession(session)
	logger.Infof("SMB1 server: %s authenticated as %s%s", c.Remote, session.Account(), admissionSuffix(session))

	w.SetResponseUID(uid)
	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to complete the session setup for %s: %v", c.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// verifyNonExtendedResponse checks an NT challenge response against the
// challenge this connection issued in its NEGOTIATE response, and returns the
// session key it yields.
//
// It reports spnego.ErrCaptureOnly when no Authenticator is configured,
// spnego.ErrUnknownIdentity when the Authenticator does not know the identity,
// and spnego.ErrBadResponse when the response does not verify.
func (c *Connection) verifyNonExtendedResponse(domain, username string, ntResponse []byte) ([]byte, error) {
	lookup := c.Server.config.Authenticator
	if lookup == nil {
		return nil, spnego.ErrCaptureOnly
	}
	ntHash, ok := lookup(domain, username)
	if !ok {
		return nil, spnego.ErrUnknownIdentity
	}

	switch {
	case len(ntResponse) == ntlmv1ResponseLength:
		// NTLMv1: DESL(NTOWFv1, ServerChallenge), session key MD4(NTOWFv1)
		// ([MS-NLMP] 3.3.1).
		v1, err := ntlmv1.NewNTLMv1CtxWithNTHash(domain, username, ntHash, c.NegotiateChallenge)
		if err != nil {
			return nil, err
		}
		expected, err := v1.ComputeNtChallengeResponse()
		if err != nil {
			return nil, err
		}
		if subtle.ConstantTimeCompare(expected[:], ntResponse) != 1 {
			return nil, spnego.ErrBadResponse
		}
		key := v1.SessionBaseKey()
		return key[:], nil

	case len(ntResponse) > ntlmv1ResponseLength:
		// NTLMv2: NTProofStr over the challenge and the client's blob, session
		// key HMAC_MD5(NTOWFv2, NTProofStr) ([MS-NLMP] 3.3.2).
		v2, err := ntlmv2.NewNTLMv2CtxWithNTHash(domain, username, ntHash, c.NegotiateChallenge, [8]byte{})
		if err != nil {
			return nil, err
		}
		if !ntlmv2.VerifyNTChallengeResponse(v2.ResponseKeyNT[:], c.NegotiateChallenge, ntResponse) {
			return nil, spnego.ErrBadResponse
		}
		return ntlmv2.SessionBaseKeyFromResponse(v2.ResponseKeyNT[:], ntResponse), nil
	}

	// No NT response, or one too short to be either form: LM-only and plaintext
	// logons are not served.
	return nil, spnego.ErrBadResponse
}
