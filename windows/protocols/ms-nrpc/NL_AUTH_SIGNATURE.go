package msnrpc

import (
	"encoding/binary"
	"fmt"
)

// Netlogon signature and seal algorithm identifiers ([MS-NRPC] 2.2.1.3.2/2.2.1.3.3), the
// 16-bit little-endian values placed in the SignatureAlgorithm and SealAlgorithm fields of
// the per-message tokens. HMAC-SHA256/AES-128 are the AES-negotiated pair; HMAC-MD5/RC4 the
// legacy pair. NlSealNotEncrypted marks an integrity-only (signed, not sealed) token, in
// which case the Confounder is omitted from the wire structure.
const (
	NlSignatureHMACMD5    uint16 = 0x0077 // HMAC-MD5 (legacy)
	NlSignatureHMACSHA256 uint16 = 0x0013 // HMAC-SHA256 (AES)
	NlSealNotEncrypted    uint16 = 0xffff // not sealed; Confounder absent
	NlSealRC4             uint16 = 0x007a // RC4 (legacy)
	NlSealAES128          uint16 = 0x001a // AES-128 CFB8
)

// nlAuthHeaderSize is the size of the fixed four-uint16 header (SignatureAlgorithm,
// SealAlgorithm, Pad, Flags) that opens both token structures. Its first eight octets are
// the exact bytes fed into the checksum computation ([MS-NRPC] 3.3.4.2.1 step 7).
const nlAuthHeaderSize = 8

// NL_AUTH_SIGNATURE ([MS-NRPC] 2.2.1.3.2) is the per-message security token Netlogon places
// after the RPC sec_trailer when acting as its own security provider with the legacy
// (non-AES) cipher suite: an HMAC-MD5 checksum with RC4 sealing. It is NOT an NDR structure
// — it is a fixed little-endian layout, so it carries its own Marshal/Unmarshal rather than
// relying on the ndr package. The AES cipher suite uses NL_AUTH_SHA2_SIGNATURE instead.
//
// The Confounder is present on the wire only when the token seals (encrypts) the message,
// i.e. when SealAlgorithm != NlSealNotEncrypted; for an integrity-only token it is omitted.
// On the wire the SequenceNumber and (when sealing) the Confounder are encrypted, whereas
// the Checksum is carried in the clear ([MS-NRPC] 3.3.4.2.1).
type NL_AUTH_SIGNATURE struct {
	SignatureAlgorithm uint16
	SealAlgorithm      uint16
	Pad                uint16 // MUST be 0xFFFF
	Flags              uint16 // MUST be 0x0000
	SequenceNumber     [8]byte
	Checksum           [8]byte
	Confounder         [8]byte // present on the wire only when Sealed() is true
}

// Sealed reports whether the token seals (encrypts) the message, in which case the
// Confounder is part of the wire layout.
func (s *NL_AUTH_SIGNATURE) Sealed() bool { return s.SealAlgorithm != NlSealNotEncrypted }

// Header returns the eight header octets (the four little-endian uint16s). These are the
// bytes fed to the checksum computation ([MS-NRPC] 3.3.4.2.1 step 7).
func (s *NL_AUTH_SIGNATURE) Header() []byte {
	h := make([]byte, nlAuthHeaderSize)
	binary.LittleEndian.PutUint16(h[0:2], s.SignatureAlgorithm)
	binary.LittleEndian.PutUint16(h[2:4], s.SealAlgorithm)
	binary.LittleEndian.PutUint16(h[4:6], s.Pad)
	binary.LittleEndian.PutUint16(h[6:8], s.Flags)
	return h
}

// Marshal serializes the token: the 8-byte header, the 8-byte SequenceNumber, the 8-byte
// Checksum, and — only when Sealed() — the 8-byte Confounder. The result is 24 octets for
// an integrity-only token and 32 octets for a sealing token.
func (s *NL_AUTH_SIGNATURE) Marshal() []byte {
	out := make([]byte, 0, 32)
	out = append(out, s.Header()...)
	out = append(out, s.SequenceNumber[:]...)
	out = append(out, s.Checksum[:]...)
	if s.Sealed() {
		out = append(out, s.Confounder[:]...)
	}
	return out
}

// Unmarshal parses a token from the front of data. Whether the Confounder is expected is
// decided by SealAlgorithm, so the header is read first.
func (s *NL_AUTH_SIGNATURE) Unmarshal(data []byte) error {
	if len(data) < nlAuthHeaderSize+16 {
		return fmt.Errorf("NL_AUTH_SIGNATURE truncated: have %d bytes, need at least %d", len(data), nlAuthHeaderSize+16)
	}
	s.SignatureAlgorithm = binary.LittleEndian.Uint16(data[0:2])
	s.SealAlgorithm = binary.LittleEndian.Uint16(data[2:4])
	s.Pad = binary.LittleEndian.Uint16(data[4:6])
	s.Flags = binary.LittleEndian.Uint16(data[6:8])
	copy(s.SequenceNumber[:], data[8:16])
	copy(s.Checksum[:], data[16:24])
	if s.Sealed() {
		if len(data) < 32 {
			return fmt.Errorf("NL_AUTH_SIGNATURE sealing token truncated: have %d bytes, need 32", len(data))
		}
		copy(s.Confounder[:], data[24:32])
	}
	return nil
}
