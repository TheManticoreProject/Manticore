package server

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/securitymode"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// noDialectSelected is the DialectIndex a server returns when it supports none of
// the dialects the client offered ([MS-CIFS] 2.2.4.52.2).
const noDialectSelected = 0xFFFF

// defaultMaxRawSize is the MaxRawSize reported in NEGOTIATE. Windows (2008
// through 2012 R2) reports 65536; the value is informational here because
// CAP_RAW_MODE is not advertised, so no raw transfer is ever negotiated.
const defaultMaxRawSize = 65536

// serverCapabilities are the capabilities advertised in the NEGOTIATE response.
//
// Each one is a promise, so the set is deliberately limited to what the server
// actually honours. CAP_RAW_MODE is absent because raw transfers are not
// implemented; advertising it would have clients issue commands that are then
// refused. CAP_LOCK_AND_READ is advertised because SMB_COM_LOCK_AND_READ and
// SMB_COM_WRITE_AND_UNLOCK are implemented (handler_lock_and_read.go).
//
// CAP_LEVEL_II_OPLOCKS says a client may be granted a level II oplock and told
// when it is broken. Level II is the only level granted: it permits a client to
// cache reads and nothing else, so withdrawing it needs no acknowledgement, and
// the break can be sent without holding up the operation that caused it. A
// request for an exclusive or batch oplock is answered with level II, which
// [MS-SMB] section 3.3.5.1.2 sanctions.
//
// CAP_LARGE_READX and CAP_LARGE_WRITEX let one read or write exceed the
// negotiated MaxBufferSize, bounded by Config.MaxLargeTransfer. Both are a
// two-sided agreement: [MS-SMB] section 2.2.4.5.2.1 has the capability take
// effect only when the client sets it in its session setup too, so a client that
// does not ask keeps the old bound.
// CAP_INFOLEVEL_PASSTHROUGH is advertised because the pass-through information
// range is served for both files and volumes. A client may ask in that range
// without being told it can, so advertising it changes nothing about what is
// answered — but a client that checks first would otherwise take the long way
// round to information the server already has.
const serverCapabilities = capabilities.CAP_UNICODE |
	capabilities.CAP_LARGE_FILES |
	capabilities.CAP_NT_SMBS |
	capabilities.CAP_NT_STATUS |
	capabilities.CAP_NT_FIND |
	capabilities.CAP_LOCK_AND_READ |
	capabilities.CAP_LARGE_READX |
	capabilities.CAP_LARGE_WRITEX |
	capabilities.CAP_LEVEL_II_OPLOCKS |
	capabilities.CAP_INFOLEVEL_PASSTHROUGH |
	capabilities.CAP_EXTENDED_SECURITY

// serverTimeZoneBias returns the SMB_COM_NEGOTIATE ServerTimeZone for the given
// time: the number of minutes to add to local time to reach UTC ([MS-CIFS]
// 2.2.4.52.2). Go's Zone reports the offset as seconds east of UTC, so the bias
// is its negation in minutes (UTC+1 -> -60, UTC-5 -> +300, UTC -> 0).
func serverTimeZoneBias(t time.Time) types.SHORT {
	_, offsetSeconds := t.Zone()
	return types.SHORT(-offsetSeconds / 60)
}

// handleNegotiate answers SMB_COM_NEGOTIATE: it selects a dialect from the list
// the client offered and describes the connection the two will share.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/a4229e1a-8a4e-489a-a2eb-11b7f360e60c
func handleNegotiate(conn *Connection, w ResponseWriter, req *message.Message) nt_status.NT_STATUS {
	request, ok := req.Command.(*commands.NegotiateRequest)
	if !ok {
		return nt_status.NT_STATUS_INVALID_SMB
	}

	// A connection negotiates once. Renegotiating would silently discard the
	// authentication state built on the first agreement.
	if conn.Negotiated {
		logger.Debugf("SMB1 server: %s sent a second NEGOTIATE on one connection", conn.Remote)
		return nt_status.NT_STATUS_INVALID_SMB
	}

	response := commands.NewNegotiateResponse()

	index, selected := selectDialect(request.Dialects)
	if selected == "" {
		// Refusing by index rather than by error status is what the protocol
		// specifies, and lets the client fall back to another dialect family.
		logger.Debugf("SMB1 server: %s offered no dialect this server supports (%v)",
			conn.Remote, request.Dialects.Dialects)
		response.DialectIndex = types.USHORT(noDialectSelected)
		if err := w.WriteResponse(response); err != nil {
			logger.Debugf("SMB1 server: failed to answer NEGOTIATE for %s: %v", conn.Remote, err)
		}
		return nt_status.NT_STATUS_SUCCESS
	}

	response.DialectIndex = types.USHORT(index)

	// User-level access control with challenge/response authentication, plus
	// whatever the signing policy allows. Only what the server will actually
	// honour is advertised: a client that is told signatures are available and
	// then receives an unsigned response abandons the connection.
	response.SecurityMode = securitymode.NEGOTIATE_USER_SECURITY | securitymode.NEGOTIATE_ENCRYPT_PASSWORDS
	switch conn.Server.config.SigningPolicy {
	case SigningEnabled:
		response.SecurityMode |= securitymode.NEGOTIATE_SECURITY_SIGNATURES_ENABLED
	case SigningRequired:
		// [MS-CIFS] 2.2.4.52.2: the REQUIRED bit MUST NOT be set on its own.
		response.SecurityMode |= securitymode.NEGOTIATE_SECURITY_SIGNATURES_ENABLED |
			securitymode.NEGOTIATE_SECURITY_SIGNATURES_REQUIRED
	}

	response.MaxMpxCount = types.USHORT(conn.Server.config.MaxMpxCount)
	// One virtual circuit per connection: multiplexed circuits are not served.
	response.MaxNumberVcs = types.USHORT(1)
	response.MaxBufferSize = types.ULONG(conn.Server.config.MaxBufferSize)
	// Windows Server 2008 through 2012 R2 report MaxRawSize 65536 even though they
	// do not advertise CAP_RAW_MODE; the field is only meaningful once CAP_RAW_MODE
	// is negotiated ([MS-CIFS] 2.2.4.52.2), which this server does not do, so the
	// value is informational and reporting 0 (unlike every Windows version) was a
	// server fingerprint. Matching the Windows value makes no behavioural promise:
	// no raw-mode capability is advertised, so no client issues a raw transfer.
	response.MaxRawSize = types.ULONG(defaultMaxRawSize)
	response.Capabilities = serverCapabilities

	sessionKey, err := randomUint32()
	if err != nil {
		logger.Errorf("SMB1 server: failed to generate a session key for %s: %v", conn.Remote, err)
		return nt_status.NT_STATUS_UNSUCCESSFUL
	}
	response.SessionKey = types.ULONG(sessionKey)

	now := time.Now()
	response.SystemTime = *msdtyp.NewFILETIMEFromTime(now.UTC())
	// [MS-CIFS] 2.2.4.52.2: ServerTimeZone is the number of minutes to add to
	// local time to reach UTC. Windows reports the host's real bias, so report the
	// host zone offset here (seconds east of UTC, negated into minutes) rather than
	// a constant 0, which looked like a UTC host regardless of the real zone.
	response.ServerTimeZone = serverTimeZoneBias(now)

	// Extended security is the modern path: the challenge travels inside a GSS
	// token during session setup rather than in the negotiate response. The data
	// block's layout follows from the CAP_EXTENDED_SECURITY bit above, which is
	// why there is no separate switch for it here.
	response.ServerGUID = conn.Server.config.ServerGUID
	response.ChallengeLength = types.UCHAR(0)
	response.Challenge = []types.UCHAR{}

	securityBlob, err := spnego.CreateNegTokenInit(nil)
	if err != nil {
		logger.Errorf("SMB1 server: failed to build the negotiate security blob for %s: %v", conn.Remote, err)
		return nt_status.NT_STATUS_UNSUCCESSFUL
	}
	response.SecurityBlob = []types.UCHAR(securityBlob)

	// DomainName and ServerName are carried only in the non-extended form; under
	// extended security the security blob and the session-setup TargetInfo carry
	// the server's identity instead.
	response.DomainName = []types.UCHAR{}
	response.ServerName = []types.UCHAR{}

	conn.Dialect = selected
	conn.Negotiated = true
	conn.UseUnicode = req.Header.Flags2.IsUnicode()
	conn.UseNTStatus = req.Header.Flags2.IsNTStatusErrorCodes()
	conn.ExtendedSecurity = req.Header.Flags2.IsExtendedSecurity()

	logger.Debugf("SMB1 server: negotiated %q with %s (unicode=%t nt_status=%t extended_security=%t)",
		selected, conn.Remote, conn.UseUnicode, conn.UseNTStatus, conn.ExtendedSecurity)

	if err := w.WriteResponse(response); err != nil {
		logger.Debugf("SMB1 server: failed to answer NEGOTIATE for %s: %v", conn.Remote, err)
	}
	return nt_status.NT_STATUS_SUCCESS
}

// selectDialect picks the dialect to use from the list a client offered, and
// returns its index in that list along with its name.
//
// The index is the client's, not the server's: [MS-CIFS] 2.2.4.52.2 defines
// DialectIndex as an offset into the request's list, which is how the client maps
// the answer back to a name.
//
// Parameters:
//   - offered: the dialects the client offered, in the order it offered them
//
// Returns:
//   - The index of the selected dialect in the offered list
//   - The selected dialect's name, or "" when none is supported
func selectDialect(offered dialects.Dialects) (int, string) {
	// NT LM 0.12 is the only dialect this server speaks: the earlier ones use
	// different message layouts for the same commands, and nothing here
	// implements them.
	for index, dialect := range offered.Dialects {
		if dialect == dialects.DIALECT_NT_LM_0_12 {
			return index, dialect
		}
	}
	return noDialectSelected, ""
}

// randomUint32 draws a value from the system's cryptographic random source, for
// the connection's SessionKey token.
func randomUint32() (uint32, error) {
	buffer := make([]byte, 4)
	if _, err := rand.Read(buffer); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buffer), nil
}
