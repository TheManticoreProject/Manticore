// Package security implements NTLMSSP per-message security: the message integrity
// (signing) and confidentiality (sealing) services used by GSS_WrapEx / GSS_GetMICEx
// once an NTLM authentication exchange has produced an exported session key.
//
// It implements the extended session security (NTLMv2) variant of MS-NLMP sections
// 3.4.3 (SEAL), 3.4.4 (MAC / message signature), and 3.4.5 (SIGNKEY / SEALKEY key
// derivation). This is the machinery a connection-oriented DCE/RPC client needs to
// sign (PKT_INTEGRITY) and seal (PKT_PRIVACY) request stubs per [MS-RPCE] 3.3.
//
// References:
//   - [MS-NLMP] 3.4 Session Security:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/0c2fa6f4-baf3-4f49-9a72-7c7d7c4c1bc8
//   - [MS-NLMP] 4.2.4.4 GSS_WrapEx Examples (the known-answer test vectors):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/9d9ed1f4-7c91-4d54-b4b2-b87ac1f5b64a
package security

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/rc4"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

// SignatureSize is the size, in bytes, of an NTLMSSP_MESSAGE_SIGNATURE
// (version + checksum + sequence number), the value carried in the RPC auth_verifier.
const SignatureSize = 16

// Magic constants for signing- and sealing-key derivation (MS-NLMP 3.4.5.2/3.4.5.3).
// Each includes its terminating NUL, which is part of the MD5 input.
const (
	clientSigningMagic = "session key to client-to-server signing key magic constant\x00"
	serverSigningMagic = "session key to server-to-client signing key magic constant\x00"
	clientSealingMagic = "session key to client-to-server sealing key magic constant\x00"
	serverSealingMagic = "session key to server-to-client sealing key magic constant\x00"
)

// Context holds the per-direction signing keys, sealing RC4 stream ciphers, and
// sequence numbers for one authenticated NTLM session. A client uses the "client"
// keys to protect outbound PDUs and the "server" keys to verify inbound PDUs.
//
// A Context is not safe for concurrent use: the RC4 handles and sequence numbers are
// stateful and advance with every message, so calls must be serialized in the same
// order the PDUs go on (and come off) the wire.
type Context struct {
	negFlg flags.NegotiateFlags

	clientSigningKey []byte
	serverSigningKey []byte

	clientSeal *rc4.RC4
	serverSeal *rc4.RC4

	outSeq uint32 // sequence number for outbound (client-to-server) messages
	inSeq  uint32 // sequence number for inbound (server-to-client) messages
}

// NewContext derives the signing and sealing keys from exportedSessionKey and the
// negotiated flags, and initializes the client and server RC4 sealing handles. With
// extended session security each direction gets its own keys, so client and server
// signatures never collide. negFlg must be the flags actually negotiated for the
// session (those carried in the AUTHENTICATE message).
func NewContext(exportedSessionKey []byte, negFlg flags.NegotiateFlags) (*Context, error) {
	if !negFlg.HasFlag(flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY) {
		return nil, fmt.Errorf("ntlm security context: only extended session security (NTLMv2) is supported")
	}
	if len(exportedSessionKey) == 0 {
		return nil, fmt.Errorf("ntlm security context: empty exported session key")
	}

	clientSealKey := deriveKey(sealingKeyInput(exportedSessionKey, negFlg), clientSealingMagic)
	serverSealKey := deriveKey(sealingKeyInput(exportedSessionKey, negFlg), serverSealingMagic)

	clientSeal, err := rc4.NewRC4WithKey(clientSealKey)
	if err != nil {
		return nil, fmt.Errorf("ntlm security context: client sealing key: %w", err)
	}
	serverSeal, err := rc4.NewRC4WithKey(serverSealKey)
	if err != nil {
		return nil, fmt.Errorf("ntlm security context: server sealing key: %w", err)
	}

	return &Context{
		negFlg:           negFlg,
		clientSigningKey: deriveKey(exportedSessionKey, clientSigningMagic),
		serverSigningKey: deriveKey(exportedSessionKey, serverSigningMagic),
		clientSeal:       clientSeal,
		serverSeal:       serverSeal,
	}, nil
}

// Sign returns the NTLMSSP_MESSAGE_SIGNATURE over message for an outbound
// PKT_INTEGRITY PDU. The message is left in cleartext; only the signature is produced.
// It advances the outbound sequence number.
func (c *Context) Sign(message []byte) [SignatureSize]byte {
	sig := c.mac(c.clientSigningKey, c.clientSeal, c.outSeq, message)
	c.outSeq++
	return sig
}

// SealWith encrypts messageToEncrypt for an outbound PKT_PRIVACY PDU and computes the
// signature over messageToSign. In connection-oriented RPC the two differ: the signed
// region is the whole PDU (header, stub, padding, sec_trailer) computed over the
// plaintext, while only the stub and its padding are encrypted ([MS-RPCE] 3.3, NTLM2).
// The plaintext is encrypted first so that, when key exchange is negotiated, the
// checksum is encrypted with the subsequent keystream (MS-NLMP 3.4.3). It advances the
// outbound sequence number once and returns the ciphertext for the encrypted region.
func (c *Context) SealWith(messageToSign, messageToEncrypt []byte) ([]byte, [SignatureSize]byte) {
	sealed := make([]byte, len(messageToEncrypt))
	c.clientSeal.XORKeyStream(sealed, messageToEncrypt)
	sig := c.mac(c.clientSigningKey, c.clientSeal, c.outSeq, messageToSign)
	c.outSeq++
	return sealed, sig
}

// VerifySignature checks the signature of an inbound PDU over messageToSign and
// advances the inbound sequence number. For PKT_INTEGRITY messageToSign is the
// cleartext PDU minus its trailing auth_value; for PKT_PRIVACY the caller must have
// already decrypted the stub with DecryptInbound so the MAC is computed over plaintext.
func (c *Context) VerifySignature(messageToSign []byte, sig [SignatureSize]byte) error {
	expected := c.mac(c.serverSigningKey, c.serverSeal, c.inSeq, messageToSign)
	c.inSeq++
	if !hmac.Equal(expected[:], sig[:]) {
		return fmt.Errorf("ntlm signature verification failed")
	}
	return nil
}

// DecryptInbound decrypts an inbound sealed region (a PKT_PRIVACY response stub plus
// padding) in place using the server sealing stream. It neither advances the sequence
// number nor verifies the signature; call VerifySignature afterwards over the now-
// plaintext PDU. Decryption must precede verification so the MAC covers plaintext.
func (c *Context) DecryptInbound(buf []byte) {
	c.serverSeal.XORKeyStream(buf, buf)
}

// mac computes the NTLMSSP_MESSAGE_SIGNATURE with extended session security
// (MS-NLMP 3.4.4.1): version (1) || HMAC_MD5(signingKey, seq || message)[0:8] ||
// seq. When key exchange is negotiated the 8-byte checksum is itself encrypted with
// the sealing handle. The caller owns the sequence-number bookkeeping.
func (c *Context) mac(signingKey []byte, seal *rc4.RC4, seq uint32, message []byte) [SignatureSize]byte {
	var seqBytes [4]byte
	binary.LittleEndian.PutUint32(seqBytes[:], seq)

	h := hmac.New(md5.New, signingKey)
	h.Write(seqBytes[:])
	h.Write(message)
	checksum := h.Sum(nil)[:8]

	if c.negFlg.HasFlag(flags.NTLMSSP_NEGOTIATE_KEY_EXCH) {
		enc := make([]byte, 8)
		seal.XORKeyStream(enc, checksum)
		checksum = enc
	}

	var sig [SignatureSize]byte
	binary.LittleEndian.PutUint32(sig[0:4], 1) // version
	copy(sig[4:12], checksum)
	binary.LittleEndian.PutUint32(sig[12:16], seq)
	return sig
}

// deriveKey returns MD5(key || magic), the signing/sealing key derivation used by
// extended session security (MS-NLMP 3.4.5.2/3.4.5.3).
func deriveKey(key []byte, magic string) []byte {
	h := md5.New()
	h.Write(key)
	h.Write([]byte(magic))
	return h.Sum(nil)
}

// sealingKeyInput selects the portion of the exported session key used as the sealing
// key derivation input, per the negotiated key strength (MS-NLMP 3.4.5.3): the full
// 16 bytes for 128-bit, the first 7 for 56-bit, otherwise the first 5 (40-bit).
func sealingKeyInput(exportedSessionKey []byte, negFlg flags.NegotiateFlags) []byte {
	switch {
	case negFlg.HasFlag(flags.NTLMSSP_NEGOTIATE_128):
		return exportedSessionKey
	case negFlg.HasFlag(flags.NTLMSSP_NEGOTIATE_56):
		return exportedSessionKey[:7]
	default:
		return exportedSessionKey[:5]
	}
}
