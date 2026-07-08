// Package securechannel implements the Netlogon secure channel ([MS-NRPC] 3.1.4): the
// challenge/response handshake and session establishment, the rolling per-call authenticator,
// the per-message sign/seal tokens, and the adapter that lets the DCE/RPC client use Netlogon
// as its own security provider (RPC_C_AUTHN_NETLOGON). It builds on the cryptographic
// primitives in the sibling crypto package and the NDR structures in the parent package, and
// invokes the interface opnums to run the handshake.
package securechannel

// IDL source: [MS-NRPC] — verified against the protocol's authoritative IDL
// (https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7).

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/TheManticoreProject/Manticore/crypto/aes/cfb8"
	"github.com/TheManticoreProject/Manticore/crypto/rc4"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// MessageSecurity produces and verifies the per-message Netlogon security tokens used when
// Netlogon acts as its own RPC security provider (RPC_C_AUTHN_NETLOGON) ([MS-NRPC]
// 3.3.4.2.1). With AES negotiated it emits NL_AUTH_SHA2_SIGNATURE tokens (HMAC-SHA256 +
// AES-128-CFB8); otherwise it emits legacy NL_AUTH_SIGNATURE tokens (HMAC-MD5 + RC4). Both
// carry a checksum, a per-message sequence number, and — when sealing — a confounder.
//
// A MessageSecurity is stateful and not safe for concurrent use: the client sequence number
// advances with every Sign/Seal call and calls must be serialized in the order the PDUs go
// on the wire.
type MessageSecurity struct {
	sessionKey [16]byte
	aes        bool
	clientSeq  uint64
	// confounderSrc supplies the sealing confounder; when nil, crypto/rand is used. It is a
	// seam for deterministic testing and is never set in production code.
	confounderSrc io.Reader
}

// NewMessageSecurityAES returns a MessageSecurity for the AES cipher suite, keyed by the
// 16-byte session key derived by crypto.ComputeSessionKeyAES. The sequence number starts at
// zero.
func NewMessageSecurityAES(sessionKey [16]byte) *MessageSecurity {
	return &MessageSecurity{sessionKey: sessionKey, aes: true, confounderSrc: rand.Reader}
}

// NewMessageSecurityRC4 returns a MessageSecurity for the legacy (non-AES) cipher suite,
// keyed by the 16-byte session key derived by crypto.ComputeSessionKeyStrongKey.
func NewMessageSecurityRC4(sessionKey [16]byte) *MessageSecurity {
	return &MessageSecurity{sessionKey: sessionKey, aes: false, confounderSrc: rand.Reader}
}

// Sign builds an integrity-only token over data ([MS-NRPC] 3.3.4.2.1 with Confidentiality
// not requested): the checksum covers the token header and the plaintext stub, no confounder
// is included, and data is left in the clear. The client sequence number is consumed and
// advanced.
func (m *MessageSecurity) Sign(data []byte) ([]byte, error) {
	_, token, err := m.protect(data, false)
	return token, err
}

// Seal builds a sealing token over data ([MS-NRPC] 3.3.4.2.1 with Confidentiality requested)
// and returns the encrypted stub alongside it. The checksum is computed over the plaintext
// (header, confounder, stub); the confounder and stub are then encrypted. The client
// sequence number is consumed and advanced.
func (m *MessageSecurity) Seal(data []byte) (sealed, token []byte, err error) {
	return m.protect(data, true)
}

// protect implements Sign (seal=false) and Seal (seal=true) for whichever cipher suite the
// context was created with.
func (m *MessageSecurity) protect(data []byte, seal bool) (sealed, token []byte, err error) {
	var confounder []byte
	if seal {
		confounder = make([]byte, 8)
		src := m.confounderSrc
		if src == nil {
			src = rand.Reader
		}
		if _, err = io.ReadFull(src, confounder); err != nil {
			return nil, nil, fmt.Errorf("netlogon protect: read confounder: %w", err)
		}
	}

	header := m.header(seal)
	checksum := m.computeChecksum(header, confounder, data)
	derived := deriveSequenceNumber(m.clientSeq, true)
	seqField, err := m.cryptSequenceNumber(derived, checksum, false)
	if err != nil {
		return nil, nil, fmt.Errorf("netlogon protect: %w", err)
	}

	if seal {
		encConfounder, encData, serr := m.seal(confounder, data, derived)
		if serr != nil {
			return nil, nil, fmt.Errorf("netlogon protect: %w", serr)
		}
		sealed = encData
		confounder = encConfounder
	} else {
		sealed = data
	}

	token = m.marshalToken(header, seqField, checksum, confounder, seal)
	m.clientSeq++
	return sealed, token, nil
}

// Unseal reverses Seal: it decrypts the sealed stub using the token's own encrypted sequence
// number and confounder, then recomputes and verifies the checksum over the recovered
// plaintext. It is self-contained (the sequence number is recovered from the token), so a
// Seal round-trips through Unseal under the same session key.
func (m *MessageSecurity) Unseal(sealed, token []byte) ([]byte, error) {
	header, seqField, checksum, encConfounder, sealedFlag, err := m.parseToken(token)
	if err != nil {
		return nil, fmt.Errorf("netlogon unseal: %w", err)
	}
	if !sealedFlag {
		return nil, fmt.Errorf("netlogon unseal: token is integrity-only, not sealed")
	}
	derived, err := m.cryptSequenceNumber(seqField, checksum, true)
	if err != nil {
		return nil, fmt.Errorf("netlogon unseal: %w", err)
	}
	confounder, data, err := m.unseal(encConfounder, sealed, derived)
	if err != nil {
		return nil, fmt.Errorf("netlogon unseal: %w", err)
	}
	want := m.computeChecksum(header, confounder, data)
	if !hmac.Equal(want, checksum) {
		return nil, fmt.Errorf("netlogon unseal: checksum mismatch")
	}
	return data, nil
}

// VerifySignature checks an integrity-only token against data by recomputing the checksum.
func (m *MessageSecurity) VerifySignature(data, token []byte) error {
	header, _, checksum, _, sealedFlag, err := m.parseToken(token)
	if err != nil {
		return fmt.Errorf("netlogon verify: %w", err)
	}
	if sealedFlag {
		return fmt.Errorf("netlogon verify: token is a sealing token, use Unseal")
	}
	if !hmac.Equal(m.computeChecksum(header, nil, data), checksum) {
		return fmt.Errorf("netlogon verify: checksum mismatch")
	}
	return nil
}

// header returns the 8-byte token header for the active cipher suite: the signature/seal
// algorithm identifiers, the 0xFFFF pad, and zero flags ([MS-NRPC] 3.3.4.2.1 steps 1-4).
func (m *MessageSecurity) header(seal bool) []byte {
	sigAlg, sealAlg := msnrpc.NlSignatureHMACMD5, msnrpc.NlSealRC4
	if m.aes {
		sigAlg, sealAlg = msnrpc.NlSignatureHMACSHA256, msnrpc.NlSealAES128
	}
	if !seal {
		sealAlg = msnrpc.NlSealNotEncrypted
	}
	h := make([]byte, 8)
	binary.LittleEndian.PutUint16(h[0:2], sigAlg)
	binary.LittleEndian.PutUint16(h[2:4], sealAlg)
	binary.LittleEndian.PutUint16(h[4:6], 0xffff)
	binary.LittleEndian.PutUint16(h[6:8], 0x0000)
	return h
}

// computeChecksum computes the signature checksum ([MS-NRPC] 3.3.4.2.1 step 7): for AES,
// HMAC-SHA256(Sk, header || confounder || message)[:8]; for the legacy suite,
// HMAC-MD5(Sk, MD5(0x00000000 || header || confounder || message))[:8]. confounder is nil
// for integrity-only tokens.
func (m *MessageSecurity) computeChecksum(header, confounder, message []byte) []byte {
	if m.aes {
		mac := hmac.New(sha256.New, m.sessionKey[:])
		mac.Write(header[:nlAuthHeader8])
		mac.Write(confounder)
		mac.Write(message)
		return mac.Sum(nil)[:8]
	}
	inner := md5.New()
	inner.Write([]byte{0, 0, 0, 0})
	inner.Write(header[:nlAuthHeader8])
	inner.Write(confounder)
	inner.Write(message)
	mac := hmac.New(md5.New, m.sessionKey[:])
	mac.Write(inner.Sum(nil))
	return mac.Sum(nil)[:8]
}

// cryptSequenceNumber encrypts (decrypt=false) or decrypts (decrypt=true) the 8-octet
// sequence number ([MS-NRPC] 3.3.4.2.1 step 9). AES: AES-128-CFB8 under the session key with
// IV = checksum[:8] repeated. Legacy: RC4 under HMAC-MD5(HMAC-MD5(Sk, 0x00000000), checksum),
// which is symmetric so decrypt and encrypt are the same operation.
func (m *MessageSecurity) cryptSequenceNumber(seq, checksum []byte, decrypt bool) ([]byte, error) {
	if m.aes {
		block, err := aes.NewCipher(m.sessionKey[:])
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
	encKey := hmacMD5(hmacMD5(m.sessionKey[:], []byte{0, 0, 0, 0}), checksum)
	return rc4Apply(encKey, seq)
}

// seal encrypts the confounder and then the stub ([MS-NRPC] 3.3.4.2.1 step 8). AES: one
// continuous AES-128-CFB8 stream keyed by Sk^0xF0 with IV = derivedSeq repeated. Legacy: RC4
// keyed by HMAC-MD5(HMAC-MD5(Sk^0xF0, 0x00000000), derivedSeq), re-initialised between the
// confounder and the stub.
func (m *MessageSecurity) seal(confounder, data, derivedSeq []byte) (encConfounder, encData []byte, err error) {
	if m.aes {
		stream, serr := newAESSealStream(m.sessionKey, derivedSeq, false)
		if serr != nil {
			return nil, nil, serr
		}
		encConfounder = make([]byte, len(confounder))
		stream.XORKeyStream(encConfounder, confounder)
		encData = make([]byte, len(data))
		stream.XORKeyStream(encData, data)
		return encConfounder, encData, nil
	}
	encKey := hmacMD5(hmacMD5(xorKey(m.sessionKey), []byte{0, 0, 0, 0}), derivedSeq)
	if encConfounder, err = rc4Apply(encKey, confounder); err != nil {
		return nil, nil, err
	}
	if encData, err = rc4Apply(encKey, data); err != nil { // RC4 re-initialised (fresh keystream)
		return nil, nil, err
	}
	return encConfounder, encData, nil
}

// unseal reverses seal for the active cipher suite.
func (m *MessageSecurity) unseal(encConfounder, encData, derivedSeq []byte) (confounder, data []byte, err error) {
	if m.aes {
		stream, serr := newAESSealStream(m.sessionKey, derivedSeq, true)
		if serr != nil {
			return nil, nil, serr
		}
		confounder = make([]byte, len(encConfounder))
		stream.XORKeyStream(confounder, encConfounder)
		data = make([]byte, len(encData))
		stream.XORKeyStream(data, encData)
		return confounder, data, nil
	}
	encKey := hmacMD5(hmacMD5(xorKey(m.sessionKey), []byte{0, 0, 0, 0}), derivedSeq)
	if confounder, err = rc4Apply(encKey, encConfounder); err != nil {
		return nil, nil, err
	}
	if data, err = rc4Apply(encKey, encData); err != nil {
		return nil, nil, err
	}
	return confounder, data, nil
}

// marshalToken assembles the on-wire token for the active cipher suite from its parts.
func (m *MessageSecurity) marshalToken(header, seqField, checksum, confounder []byte, seal bool) []byte {
	if m.aes {
		var sig msnrpc.NL_AUTH_SHA2_SIGNATURE
		fillHeader(&sig.SignatureAlgorithm, &sig.SealAlgorithm, &sig.Pad, &sig.Flags, header)
		copy(sig.SequenceNumber[:], seqField)
		copy(sig.Checksum[:], checksum)
		if seal {
			copy(sig.Confounder[:], confounder)
		}
		return sig.Marshal()
	}
	var sig msnrpc.NL_AUTH_SIGNATURE
	fillHeader(&sig.SignatureAlgorithm, &sig.SealAlgorithm, &sig.Pad, &sig.Flags, header)
	copy(sig.SequenceNumber[:], seqField)
	copy(sig.Checksum[:], checksum)
	if seal {
		copy(sig.Confounder[:], confounder)
	}
	return sig.Marshal()
}

// parseToken parses a token for the active cipher suite, returning its 8-byte header, the
// (still-encrypted) sequence number, the checksum, the (still-encrypted) confounder when
// present, and whether the token seals.
func (m *MessageSecurity) parseToken(token []byte) (header, seqField, checksum, confounder []byte, seal bool, err error) {
	if m.aes {
		var sig msnrpc.NL_AUTH_SHA2_SIGNATURE
		if err = sig.Unmarshal(token); err != nil {
			return
		}
		return sig.Header(), sig.SequenceNumber[:], sig.Checksum[:], sig.Confounder[:], sig.Sealed(), nil
	}
	var sig msnrpc.NL_AUTH_SIGNATURE
	if err = sig.Unmarshal(token); err != nil {
		return
	}
	return sig.Header(), sig.SequenceNumber[:], sig.Checksum[:], sig.Confounder[:], sig.Sealed(), nil
}

// fillHeader copies the four little-endian header fields out of an 8-byte header slice.
func fillHeader(sigAlg, sealAlg, pad, flags *uint16, header []byte) {
	*sigAlg = binary.LittleEndian.Uint16(header[0:2])
	*sealAlg = binary.LittleEndian.Uint16(header[2:4])
	*pad = binary.LittleEndian.Uint16(header[4:6])
	*flags = binary.LittleEndian.Uint16(header[6:8])
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

// newAESSealStream builds the AES-128-CFB8 stream used to seal (decrypt=false) or unseal
// (decrypt=true) the confounder and stub ([MS-NRPC] 3.3.4.2.1 step 8): keyed by the session
// key XOR 0xF0, IV = derivedSeq repeated. A single stream is returned so the caller processes
// the confounder and stub as one continuous segment.
func newAESSealStream(sessionKey [16]byte, derivedSeq []byte, decrypt bool) (cipher.Stream, error) {
	block, err := aes.NewCipher(xorKey(sessionKey))
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

// xorKey returns the session key with every octet XORed by 0xF0, the sealing-key derivation
// shared by the AES and RC4 suites ([MS-NRPC] 3.3.4.2.1 step 8).
func xorKey(sessionKey [16]byte) []byte {
	out := make([]byte, len(sessionKey))
	for i := range sessionKey {
		out[i] = sessionKey[i] ^ 0xf0
	}
	return out
}

// hmacMD5 returns HMAC-MD5(key, data), a building block of the legacy RC4 key derivations.
func hmacMD5(key, data []byte) []byte {
	mac := hmac.New(md5.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// rc4Apply runs data through a freshly keyed RC4 keystream (encryption and decryption are the
// same operation).
func rc4Apply(key, data []byte) ([]byte, error) {
	c, err := rc4.NewRC4WithKey(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	c.XORKeyStream(out, data)
	return out, nil
}

// nlAuthHeader8 is the number of leading token-header octets fed into the checksum.
const nlAuthHeader8 = 8
