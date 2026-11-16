package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv1"
	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/network/tcp"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
)

// Offsets into a raw SMB1 frame (header included) of the fields these tests read.
const (
	rawHeaderSize    = 32
	rawWordCountAt   = rawHeaderSize
	rawStatusAt      = 5
	rawFlags2At      = 10
	rawUIDAt         = 28
	negotiateWords   = 17
	negotiateCapsAt  = rawWordCountAt + 1 + 19 // DialectIndex..SessionKey precede it
	negotiateChalLen = rawWordCountAt + 1 + 33 // last byte of the 17 words
)

// nonExtendedNegotiation is what a client learns from a NEGOTIATE that did not
// ask for extended security.
type nonExtendedNegotiation struct {
	raw          []byte
	flags2       uint16
	capabilities uint32
	challenge    [8]byte
	domainName   []byte
	serverName   []byte
}

// negotiateWithoutExtendedSecurity sends a NEGOTIATE with
// SMB_FLAGS2_EXTENDED_SECURITY clear and decodes the reply from its raw bytes,
// following the [MS-CIFS] 2.2.4.52.2 layout.
func negotiateWithoutExtendedSecurity(t *testing.T, client *tcp.TCPTransport) *nonExtendedNegotiation {
	t.Helper()

	request := newRequest(codes.SMB_COM_NEGOTIATE)
	negotiate := commands.NewNegotiateRequest()
	negotiate.Dialects.AddDialect(dialects.DIALECT_NT_LM_0_12)
	request.AddCommand(negotiate)
	sendRequest(t, client, request)

	raw, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	if status := binary.LittleEndian.Uint32(raw[rawStatusAt:]); status != 0 {
		t.Fatalf("NEGOTIATE answered 0x%08X, want success", status)
	}
	if wc := raw[rawWordCountAt]; wc != negotiateWords {
		t.Fatalf("NEGOTIATE WordCount = %d, want %d", wc, negotiateWords)
	}

	result := &nonExtendedNegotiation{
		raw:          raw,
		flags2:       binary.LittleEndian.Uint16(raw[rawFlags2At:]),
		capabilities: binary.LittleEndian.Uint32(raw[negotiateCapsAt:]),
	}

	challengeLength := int(raw[negotiateChalLen])
	if challengeLength != 8 {
		t.Fatalf("ChallengeLength = %d, want 8 (data block % x)", challengeLength, raw[rawWordCountAt+1+34:])
	}

	byteCountAt := rawWordCountAt + 1 + 2*negotiateWords
	byteCount := int(binary.LittleEndian.Uint16(raw[byteCountAt:]))
	dataBlock := raw[byteCountAt+2:]
	if len(dataBlock) != byteCount {
		t.Fatalf("ByteCount = %d but %d data bytes follow", byteCount, len(dataBlock))
	}
	copy(result.challenge[:], dataBlock[:8])

	rest := dataBlock[8:]
	result.domainName, rest = splitUTF16Terminated(t, rest, "DomainName")
	result.serverName, rest = splitUTF16Terminated(t, rest, "ServerName")
	if len(rest) != 0 {
		t.Fatalf("%d bytes follow ServerName: % x", len(rest), rest)
	}
	return result
}

// splitUTF16Terminated cuts a NUL-terminated UTF-16LE string off the front of a
// buffer, returning the string without its terminator and the bytes after it.
func splitUTF16Terminated(t *testing.T, data []byte, field string) ([]byte, []byte) {
	t.Helper()
	for i := 0; i+1 < len(data); i += 2 {
		if data[i] == 0 && data[i+1] == 0 {
			return data[:i], data[i+2:]
		}
	}
	t.Fatalf("%s is not NUL-terminated: % x", field, data)
	return nil, nil
}

// nonExtendedSetup is the content of a [MS-CIFS] 2.2.4.53.1 session setup.
type nonExtendedSetup struct {
	oemPassword     []byte
	unicodePassword []byte
	account         string
	domain          string
	requestSigning  bool
}

// rawNonExtendedSessionSetup builds a WordCount-13 SESSION_SETUP_ANDX by hand,
// with the Unicode strings aligned on a 2-byte boundary from the start of the SMB
// header as [MS-CIFS] 2.2.4.53.1 requires, sends it and returns the raw reply.
func rawNonExtendedSessionSetup(t *testing.T, client *tcp.TCPTransport, setup nonExtendedSetup) []byte {
	t.Helper()

	request := newRequest(codes.SMB_COM_SESSION_SETUP_ANDX)
	request.Header.UID = 0
	if setup.requestSigning {
		request.Header.Flags2 |= flags2.FLAGS2_SECURITY_SIGNATURE
	}
	headerBytes, err := request.Header.Marshal()
	if err != nil {
		t.Fatalf("Header.Marshal() error = %v", err)
	}

	words := make([]byte, 26)
	words[0] = 0xFF                                 // AndXCommand: none
	binary.LittleEndian.PutUint16(words[4:], 16644) // MaxBufferSize
	binary.LittleEndian.PutUint16(words[6:], 50)    // MaxMpxCount
	binary.LittleEndian.PutUint16(words[8:], 0)     // VcNumber
	binary.LittleEndian.PutUint32(words[10:], 0)    // SessionKey
	binary.LittleEndian.PutUint16(words[14:], uint16(len(setup.oemPassword)))
	binary.LittleEndian.PutUint16(words[16:], uint16(len(setup.unicodePassword)))
	binary.LittleEndian.PutUint32(words[22:], uint32(capabilities.CAP_UNICODE|capabilities.CAP_NT_STATUS|capabilities.CAP_LARGE_FILES))

	data := []byte{}
	data = append(data, setup.oemPassword...)
	data = append(data, setup.unicodePassword...)
	const dataOffset = rawHeaderSize + 1 + 26 + 2
	if (dataOffset+len(data))%2 != 0 {
		data = append(data, 0x00) // Pad
	}
	for _, s := range []string{setup.account, setup.domain, "Unix", "Samba"} {
		data = append(data, utf16.EncodeUTF16LE(s)...)
		data = append(data, 0x00, 0x00)
	}

	frame := append([]byte{}, headerBytes...)
	frame = append(frame, 13)
	frame = append(frame, words...)
	frame = binary.LittleEndian.AppendUint16(frame, uint16(len(data)))
	frame = append(frame, data...)

	if _, err := client.Send(frame); err != nil {
		t.Fatalf("client Send() error = %v", err)
	}
	raw, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	return raw
}

func rawStatus(raw []byte) uint32 { return binary.LittleEndian.Uint32(raw[rawStatusAt:]) }
func rawUID(raw []byte) uint16    { return binary.LittleEndian.Uint16(raw[rawUIDAt:]) }

// ntlmv2Response computes the NTLMv2 and LMv2 responses a client sends against
// the negotiate challenge.
func ntlmv2Response(t *testing.T, domain, username, password string, challenge [8]byte) (nt, lm, responseKeyNT []byte) {
	t.Helper()
	ctx, err := ntlmv2.NewNTLMv2CtxWithPassword(domain, username, password, challenge, [8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	if err != nil {
		t.Fatalf("NewNTLMv2CtxWithPassword() error = %v", err)
	}
	timestamp := []byte{0x00, 0x80, 0x3e, 0xd5, 0xde, 0xb1, 0x9d, 0x01}
	nt, _, err = ctx.ComputeNTChallengeResponse(timestamp, []byte{0, 0, 0, 0})
	if err != nil {
		t.Fatalf("ComputeNTChallengeResponse() error = %v", err)
	}
	return nt, ctx.ComputeLMChallengeResponse(false), ctx.ResponseKeyNT[:]
}

// TestNegotiateWithoutExtendedSecurityCarriesChallenge asserts a client that does
// not set SMB_FLAGS2_EXTENDED_SECURITY gets the [MS-CIFS] 2.2.4.52.2 body: an
// 8-byte challenge followed by DomainName and ServerName, with
// CAP_EXTENDED_SECURITY clear so the header, capabilities and body agree.
func TestNegotiateWithoutExtendedSecurityCarriesChallenge(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)
	client := dialServer(t, addr)

	got := negotiateWithoutExtendedSecurity(t, client)

	if got.flags2&uint16(flags2.FLAGS2_EXTENDED_SECURITY) != 0 {
		t.Errorf("reply Flags2 0x%04X sets EXTENDED_SECURITY for a client that did not", got.flags2)
	}
	if got.capabilities&uint32(capabilities.CAP_EXTENDED_SECURITY) != 0 {
		t.Errorf("Capabilities 0x%08X advertise CAP_EXTENDED_SECURITY in a non-extended reply", got.capabilities)
	}
	if got.challenge == [8]byte{} {
		t.Error("the challenge is all zeroes")
	}
	if want := utf16.EncodeUTF16LE("LAB"); !bytes.Equal(got.domainName, want) {
		t.Errorf("DomainName = % x, want % x", got.domainName, want)
	}
	if want := utf16.EncodeUTF16LE("MANTICORE"); !bytes.Equal(got.serverName, want) {
		t.Errorf("ServerName = % x, want % x", got.serverName, want)
	}

	// A second connection gets a different challenge.
	other := negotiateWithoutExtendedSecurity(t, dialServer(t, addr))
	if other.challenge == got.challenge {
		t.Errorf("two connections were issued the same challenge %x", got.challenge)
	}
}

// TestNegotiateWithExtendedSecurityIsUnchanged asserts the extended-security
// reply keeps its ServerGUID + SPNEGO layout and CAP_EXTENDED_SECURITY.
func TestNegotiateWithExtendedSecurityIsUnchanged(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)
	client := dialServer(t, addr)

	request := newRequest(codes.SMB_COM_NEGOTIATE)
	request.Header.Flags2 |= flags2.FLAGS2_EXTENDED_SECURITY
	negotiate := commands.NewNegotiateRequest()
	negotiate.Dialects.AddDialect(dialects.DIALECT_NT_LM_0_12)
	request.AddCommand(negotiate)
	sendRequest(t, client, request)

	raw, err := client.Receive()
	if err != nil {
		t.Fatalf("client Receive() error = %v", err)
	}
	if caps := binary.LittleEndian.Uint32(raw[negotiateCapsAt:]); caps&uint32(capabilities.CAP_EXTENDED_SECURITY) == 0 {
		t.Fatalf("Capabilities 0x%08X lack CAP_EXTENDED_SECURITY", caps)
	}
	if l := raw[negotiateChalLen]; l != 0 {
		t.Fatalf("ChallengeLength = %d, want 0", l)
	}
	byteCountAt := rawWordCountAt + 1 + 2*negotiateWords
	dataBlock := raw[byteCountAt+2:]
	// ServerGUID(16) then a GSS InitialContextToken (0x60).
	if len(dataBlock) < 17 || dataBlock[16] != 0x60 {
		t.Fatalf("data block does not hold ServerGUID + SPNEGO: % x", dataBlock)
	}
}

// TestNonExtendedSessionSetupWithNTLMv2 asserts a non-extended session setup
// carrying an NTLMv2 response to the negotiate challenge is verified against the
// credential store and establishes a session, answered with the WordCount-3
// response of [MS-CIFS] 2.2.4.53.2.
func TestNonExtendedSessionSetupWithNTLMv2(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)
	client := dialServer(t, addr)
	neg := negotiateWithoutExtendedSecurity(t, client)

	ntResp, lmResp, _ := ntlmv2Response(t, captureDomain, captureUsername, capturePassword, neg.challenge)
	raw := rawNonExtendedSessionSetup(t, client, nonExtendedSetup{
		oemPassword:     lmResp,
		unicodePassword: ntResp,
		account:         captureUsername,
		domain:          captureDomain,
	})
	if status := rawStatus(raw); status != 0 {
		t.Fatalf("session setup answered 0x%08X, want success", status)
	}
	if wc := raw[rawWordCountAt]; wc != 3 {
		t.Fatalf("session setup reply WordCount = %d, want 3", wc)
	}
	action := binary.LittleEndian.Uint16(raw[rawWordCountAt+1+4:])
	if action&SMB_SETUP_GUEST != 0 {
		t.Fatalf("Action = 0x%04X, a verified logon was reported as guest", action)
	}
	// NativeOS must start on an even offset from the header.
	dataAt := rawWordCountAt + 1 + 6 + 2
	nativeOS := raw[dataAt:]
	if dataAt%2 != 0 {
		nativeOS = nativeOS[1:]
	}
	if want := utf16.EncodeUTF16LE(DefaultNativeOS + "\x00"); !bytes.HasPrefix(nativeOS, want) {
		t.Fatalf("NativeOS is not aligned UTF-16: % x", raw[dataAt:])
	}
	uid := rawUID(raw)
	if uid == 0 {
		t.Fatal("no UID was assigned")
	}

	// The session is usable: LOGOFF_ANDX needs a session to name.
	request := newRequest(codes.SMB_COM_LOGOFF_ANDX)
	request.Header.UID = uid
	request.AddCommand(commands.NewLogoffAndxRequest())
	sendRequest(t, client, request)
	response, _ := receiveResponse(t, client)
	if response.Header.Status != 0 {
		t.Fatalf("LOGOFF_ANDX on the new UID answered 0x%08X, want success", response.Header.Status)
	}
}

// TestNonExtendedSessionSetupWithNTLMv1 asserts a 24-byte NTLM response to the
// negotiate challenge is verified as well.
func TestNonExtendedSessionSetupWithNTLMv1(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, nil)
	client := dialServer(t, addr)
	neg := negotiateWithoutExtendedSecurity(t, client)

	ctx, err := ntlmv1.NewNTLMv1CtxWithPassword(captureDomain, captureUsername, capturePassword, neg.challenge)
	if err != nil {
		t.Fatalf("NewNTLMv1CtxWithPassword() error = %v", err)
	}
	ntResp, err := ctx.ComputeNtChallengeResponse()
	if err != nil {
		t.Fatalf("ComputeNtChallengeResponse() error = %v", err)
	}
	raw := rawNonExtendedSessionSetup(t, client, nonExtendedSetup{
		oemPassword:     ntResp[:],
		unicodePassword: ntResp[:],
		account:         captureUsername,
		domain:          captureDomain,
	})
	if status := rawStatus(raw); status != 0 {
		t.Fatalf("session setup answered 0x%08X, want success", status)
	}
}

// TestNonExtendedSessionSetupWrongPasswordIsRefused asserts a response computed
// from the wrong password is refused with STATUS_LOGON_FAILURE.
func TestNonExtendedSessionSetupWrongPasswordIsRefused(t *testing.T) {
	_, addr := serverWithAccount(t, SigningDisabled, func(c *Config) { c.AllowGuest = true })
	client := dialServer(t, addr)
	neg := negotiateWithoutExtendedSecurity(t, client)

	ntResp, lmResp, _ := ntlmv2Response(t, captureDomain, captureUsername, "wrong", neg.challenge)
	raw := rawNonExtendedSessionSetup(t, client, nonExtendedSetup{
		oemPassword:     lmResp,
		unicodePassword: ntResp,
		account:         captureUsername,
		domain:          captureDomain,
	})
	if status := rawStatus(raw); status != uint32(nt_status.NT_STATUS_LOGON_FAILURE) {
		t.Fatalf("session setup answered 0x%08X, want NT_STATUS_LOGON_FAILURE", status)
	}
}

// TestNonExtendedNullSessionFollowsAllowAnonymous asserts a session setup with
// no account and no response is a null session, admitted only under
// AllowAnonymous and then reported as a guest.
func TestNonExtendedNullSessionFollowsAllowAnonymous(t *testing.T) {
	for _, allow := range []bool{false, true} {
		_, addr := serverWithAccount(t, SigningDisabled, func(c *Config) { c.AllowAnonymous = allow })
		client := dialServer(t, addr)
		negotiateWithoutExtendedSecurity(t, client)

		raw := rawNonExtendedSessionSetup(t, client, nonExtendedSetup{oemPassword: []byte{0}})
		status := rawStatus(raw)
		if !allow {
			if status != uint32(nt_status.NT_STATUS_LOGON_FAILURE) {
				t.Fatalf("AllowAnonymous=false: null session answered 0x%08X, want NT_STATUS_LOGON_FAILURE", status)
			}
			continue
		}
		if status != 0 {
			t.Fatalf("AllowAnonymous=true: null session answered 0x%08X, want success", status)
		}
		action := binary.LittleEndian.Uint16(raw[rawWordCountAt+1+4:])
		if action&SMB_SETUP_GUEST == 0 {
			t.Fatalf("Action = 0x%04X, a null session is not reported as guest", action)
		}
	}
}

// TestNonExtendedSessionSigns asserts that under SigningRequired a non-extended
// NTLMv2 logon arms signing with the MAC key [MS-CIFS] 3.1.4.1 hashes ahead of
// the message, SessionKey || NT response, and that later requests and responses
// verify under it.
func TestNonExtendedSessionSigns(t *testing.T) {
	_, addr := serverWithAccount(t, SigningRequired, nil)
	client := dialServer(t, addr)
	neg := negotiateWithoutExtendedSecurity(t, client)

	ntResp, lmResp, responseKeyNT := ntlmv2Response(t, captureDomain, captureUsername, capturePassword, neg.challenge)
	raw := rawNonExtendedSessionSetup(t, client, nonExtendedSetup{
		oemPassword:     lmResp,
		unicodePassword: ntResp,
		account:         captureUsername,
		domain:          captureDomain,
		requestSigning:  true,
	})
	if status := rawStatus(raw); status != 0 {
		t.Fatalf("session setup answered 0x%08X, want success", status)
	}

	mac := hmac.New(md5.New, responseKeyNT)
	mac.Write(ntResp[:16])
	macKey := append(mac.Sum(nil), ntResp...)

	if !signing.Verify(macKey, raw, 1) {
		t.Fatal("the session setup response is not signed at sequence 1 with SessionKey || NT response")
	}

	session := &authSession{client: client, uid: types.USHORT(rawUID(raw)), key: macKey, signing: true, sequence: 2}
	echo := commands.NewEchoRequest()
	echo.EchoCount = 1
	echo.Data = []byte("signed")
	response := sendOnSession(t, session, codes.SMB_COM_ECHO, echo)
	if response.Header.Status != 0 {
		t.Fatalf("signed echo answered 0x%08X, want success", response.Header.Status)
	}
}
