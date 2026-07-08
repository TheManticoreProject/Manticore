package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/TheManticoreProject/Manticore/crypto/aes/cfb8"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// MessageSecurity produces and verifies the per-message Netlogon security tokens used when
// Netlogon acts as its own RPC security provider (RPC_C_AUTHN_NETLOGON) with the AES cipher
// suite negotiated ([MS-NRPC] 3.3.4.2.1). It signs (integrity) or seals (privacy) request
// stubs with an NL_AUTH_SHA2_SIGNATURE token: an HMAC-SHA256 checksum, a per-message
// sequence number, and — when sealing — a random confounder, with the sequence number,
// confounder, and stub encrypted under AES-128 in 8-bit CFB mode.
//
// A MessageSecurity is stateful and not safe for concurrent use: the client sequence number
// advances with every Sign/Seal call and calls must be serialized in the order the PDUs go
// on the wire. The legacy (non-AES) RC4/HMAC-MD5 suite is not implemented here.
type MessageSecurity struct {
	sessionKey [16]byte
	clientSeq  uint64
	// confounderSrc supplies the sealing confounder; when nil, crypto/rand is used. It is a
	// seam for deterministic testing and is never set in production code.
	confounderSrc io.Reader
}

// NewMessageSecurityAES returns a MessageSecurity for the AES cipher suite, keyed by the
// 16-byte Netlogon session key derived by the secure-channel handshake
// (ComputeSessionKeyAES). The sequence number starts at zero.
func NewMessageSecurityAES(sessionKey [16]byte) *MessageSecurity {
	return &MessageSecurity{sessionKey: sessionKey, confounderSrc: rand.Reader}
}

// Sign builds an integrity-only NL_AUTH_SHA2_SIGNATURE token over data ([MS-NRPC] 3.3.4.2.1
// with Confidentiality not requested): the checksum covers the token header and the
// plaintext stub, and no confounder is included. data is left in the clear. The returned
// token is 48 octets. The client sequence number is consumed and advanced.
func (m *MessageSecurity) Sign(data []byte) ([]byte, error) {
	sig := &msnrpc.NL_AUTH_SHA2_SIGNATURE{
		SignatureAlgorithm: msnrpc.NlSignatureHMACSHA256,
		SealAlgorithm:      msnrpc.NlSealNotEncrypted,
		Pad:                0xffff,
	}
	checksum := computeSignatureAES(sig.Header(), nil, data, m.sessionKey)
	copy(sig.Checksum[:], checksum)

	derived := deriveSequenceNumber(m.clientSeq, true)
	seqField, err := cryptSequenceNumberAES(derived, sig.Checksum[:], m.sessionKey, false)
	if err != nil {
		return nil, fmt.Errorf("netlogon sign: %w", err)
	}
	copy(sig.SequenceNumber[:], seqField)

	m.clientSeq++
	return sig.Marshal(), nil
}

// Seal builds a sealing NL_AUTH_SHA2_SIGNATURE token over data ([MS-NRPC] 3.3.4.2.1 with
// Confidentiality requested) and returns the encrypted stub alongside it. The checksum is
// computed over the token header, the plaintext confounder, and the plaintext stub; the
// confounder and stub are then encrypted as one continuous AES-128-CFB8 stream keyed by the
// session key XOR 0xF0 with IV = derivedSeq||derivedSeq. The returned token is 56 octets.
// The client sequence number is consumed and advanced.
func (m *MessageSecurity) Seal(data []byte) (sealed, token []byte, err error) {
	src := m.confounderSrc
	if src == nil {
		src = rand.Reader
	}
	var confounder [8]byte
	if _, err = io.ReadFull(src, confounder[:]); err != nil {
		return nil, nil, fmt.Errorf("netlogon seal: read confounder: %w", err)
	}

	sig := &msnrpc.NL_AUTH_SHA2_SIGNATURE{
		SignatureAlgorithm: msnrpc.NlSignatureHMACSHA256,
		SealAlgorithm:      msnrpc.NlSealAES128,
		Pad:                0xffff,
	}
	// The checksum is computed over the plaintext, before any encryption.
	checksum := computeSignatureAES(sig.Header(), confounder[:], data, m.sessionKey)
	copy(sig.Checksum[:], checksum)

	derived := deriveSequenceNumber(m.clientSeq, true)
	seqField, err := cryptSequenceNumberAES(derived, sig.Checksum[:], m.sessionKey, false)
	if err != nil {
		return nil, nil, fmt.Errorf("netlogon seal: %w", err)
	}
	copy(sig.SequenceNumber[:], seqField)

	// Seal the confounder and then the stub as a single continuous CFB8 stream: the shift
	// register state after the 8 confounder octets carries into the stub, so they must not
	// be encrypted with separate stream objects ([MS-NRPC] 3.3.4.2.1 step 8).
	stream, err := newSealStream(m.sessionKey, derived, false)
	if err != nil {
		return nil, nil, fmt.Errorf("netlogon seal: %w", err)
	}
	var encConfounder [8]byte
	stream.XORKeyStream(encConfounder[:], confounder[:])
	sealed = make([]byte, len(data))
	stream.XORKeyStream(sealed, data)
	copy(sig.Confounder[:], encConfounder[:])

	m.clientSeq++
	return sealed, sig.Marshal(), nil
}

// Unseal reverses Seal: it decrypts the sealed stub using the token's own encrypted
// sequence number and confounder, then recomputes the checksum over the recovered plaintext
// and verifies it. It is self-contained (the sequence number is recovered from the token,
// not from a counter), so a token produced by Seal round-trips through Unseal under the same
// session key.
func (m *MessageSecurity) Unseal(sealed, token []byte) ([]byte, error) {
	var sig msnrpc.NL_AUTH_SHA2_SIGNATURE
	if err := sig.Unmarshal(token); err != nil {
		return nil, fmt.Errorf("netlogon unseal: %w", err)
	}
	if !sig.Sealed() {
		return nil, fmt.Errorf("netlogon unseal: token is integrity-only, not sealed")
	}

	derived, err := cryptSequenceNumberAES(sig.SequenceNumber[:], sig.Checksum[:], m.sessionKey, true)
	if err != nil {
		return nil, fmt.Errorf("netlogon unseal: %w", err)
	}

	stream, err := newSealStream(m.sessionKey, derived, true)
	if err != nil {
		return nil, fmt.Errorf("netlogon unseal: %w", err)
	}
	var confounder [8]byte
	stream.XORKeyStream(confounder[:], sig.Confounder[:])
	data := make([]byte, len(sealed))
	stream.XORKeyStream(data, sealed)

	want := computeSignatureAES(sig.Header(), confounder[:], data, m.sessionKey)
	if !hmac.Equal(want, sig.Checksum[:]) {
		return nil, fmt.Errorf("netlogon unseal: checksum mismatch")
	}
	return data, nil
}

// VerifySignature checks an integrity-only token against data by recomputing the HMAC-SHA256
// checksum ([MS-NRPC] 3.3.4.2.1 step 7). It is for signed (not sealed) messages; for a
// sealing token use Unseal, which verifies the checksum over the recovered plaintext.
func (m *MessageSecurity) VerifySignature(data, token []byte) error {
	var sig msnrpc.NL_AUTH_SHA2_SIGNATURE
	if err := sig.Unmarshal(token); err != nil {
		return fmt.Errorf("netlogon verify: %w", err)
	}
	if sig.Sealed() {
		return fmt.Errorf("netlogon verify: token is a sealing token, use Unseal")
	}
	want := computeSignatureAES(sig.Header(), nil, data, m.sessionKey)
	if !hmac.Equal(want, sig.Checksum[:]) {
		return fmt.Errorf("netlogon verify: checksum mismatch")
	}
	return nil
}

// computeSignatureAES computes the Netlogon AES signature ([MS-NRPC] 3.3.4.2.1 step 7):
// HMAC-SHA256(sessionKey, header[0:8] || confounder || message), truncated to 8 octets. The
// confounder is nil for integrity-only tokens.
func computeSignatureAES(header, confounder, message []byte, sessionKey [16]byte) []byte {
	mac := hmac.New(sha256.New, sessionKey[:])
	mac.Write(header[:nlAuthHeader8])
	mac.Write(confounder)
	mac.Write(message)
	return mac.Sum(nil)[:8]
}

// deriveSequenceNumber renders the 64-bit client sequence number into the 8 wire octets
// ([MS-NRPC] 3.3.4.2.1 step 5): the low 32 bits big-endian, then the high 32 bits
// big-endian. When client is true the high-order 0x80000000 bit is set to mark a
// client-originated token; server tokens leave it clear.
func deriveSequenceNumber(seq uint64, client bool) []byte {
	low := uint32(seq & 0xffffffff)
	high := uint32((seq >> 32) & 0xffffffff)
	if client {
		high |= 0x80000000
	}
	out := make([]byte, 8)
	binary.BigEndian.PutUint32(out[0:4], low)
	binary.BigEndian.PutUint32(out[4:8], high)
	return out
}

// cryptSequenceNumberAES encrypts (decrypt=false) or decrypts (decrypt=true) the 8-octet
// sequence number with AES-128-CFB8 under the plain session key, using an IV of the first 8
// checksum octets repeated twice ([MS-NRPC] 3.3.4.2.1 step 9).
func cryptSequenceNumberAES(seq, checksum []byte, sessionKey [16]byte, decrypt bool) ([]byte, error) {
	block, err := aes.NewCipher(sessionKey[:])
	if err != nil {
		return nil, err
	}
	var iv [16]byte
	copy(iv[0:8], checksum[:8])
	copy(iv[8:16], checksum[:8])
	var stream cipher.Stream
	if decrypt {
		stream = cfb8.NewDecrypter(block, iv[:])
	} else {
		stream = cfb8.NewEncrypter(block, iv[:])
	}
	out := make([]byte, len(seq))
	stream.XORKeyStream(out, seq)
	return out, nil
}

// newSealStream builds the AES-128-CFB8 stream used to seal (decrypt=false) or unseal
// (decrypt=true) the confounder and stub ([MS-NRPC] 3.3.4.2.1 step 8): keyed by the session
// key with every octet XORed by 0xF0, with IV = derivedSeq||derivedSeq. A single stream is
// returned so the caller can process the confounder and stub as one continuous segment.
func newSealStream(sessionKey [16]byte, derivedSeq []byte, decrypt bool) (cipher.Stream, error) {
	var encKey [16]byte
	for i := range sessionKey {
		encKey[i] = sessionKey[i] ^ 0xf0
	}
	block, err := aes.NewCipher(encKey[:])
	if err != nil {
		return nil, err
	}
	var iv [16]byte
	copy(iv[0:8], derivedSeq)
	copy(iv[8:16], derivedSeq)
	if decrypt {
		return cfb8.NewDecrypter(block, iv[:]), nil
	}
	return cfb8.NewEncrypter(block, iv[:]), nil
}

// nlAuthHeader8 is the number of leading token-header octets fed into the checksum.
const nlAuthHeader8 = 8
