package msnrpc

import (
	"encoding/binary"
	"fmt"
)

// nlAuthSHA2ReservedSize is the trailing Reserved field of NL_AUTH_SHA2_SIGNATURE. The
// sender sets it to zero and the receiver ignores it ([MS-NRPC] 2.2.1.3.3), but it is a
// real part of the wire layout — Windows and impacket both emit it, so it is included in
// the token length (48 octets integrity-only, 56 octets sealing).
const nlAuthSHA2ReservedSize = 24

// NL_AUTH_SHA2_SIGNATURE ([MS-NRPC] 2.2.1.3.3) is the per-message security token Netlogon
// places after the RPC sec_trailer when acting as its own security provider with the AES
// cipher suite negotiated: an HMAC-SHA256 checksum with AES-128 CFB8 sealing. Like
// NL_AUTH_SIGNATURE it is a fixed little-endian layout (not NDR) and carries its own
// Marshal/Unmarshal. It differs from the legacy token by the algorithm identifiers and a
// trailing 24-octet Reserved field.
//
// The Confounder is present on the wire only when the token seals the message
// (SealAlgorithm != NlSealNotEncrypted). On the wire the SequenceNumber and (when sealing)
// the Confounder are encrypted; the Checksum is carried in the clear ([MS-NRPC] 3.3.4.2.1).
type NL_AUTH_SHA2_SIGNATURE struct {
	SignatureAlgorithm uint16
	SealAlgorithm      uint16
	Pad                uint16 // MUST be 0xFFFF
	Flags              uint16 // MUST be 0x0000
	SequenceNumber     [8]byte
	Checksum           [8]byte
	Confounder         [8]byte // present on the wire only when Sealed() is true
	// Reserved [24]byte is always zero on send and ignored on receipt.
}

// Sealed reports whether the token seals (encrypts) the message, in which case the
// Confounder is part of the wire layout.
func (s *NL_AUTH_SHA2_SIGNATURE) Sealed() bool { return s.SealAlgorithm != NlSealNotEncrypted }

// Header returns the eight header octets (the four little-endian uint16s). These are the
// bytes fed to the HMAC-SHA256 checksum computation ([MS-NRPC] 3.3.4.2.1 step 7).
func (s *NL_AUTH_SHA2_SIGNATURE) Header() []byte {
	h := make([]byte, nlAuthHeaderSize)
	binary.LittleEndian.PutUint16(h[0:2], s.SignatureAlgorithm)
	binary.LittleEndian.PutUint16(h[2:4], s.SealAlgorithm)
	binary.LittleEndian.PutUint16(h[4:6], s.Pad)
	binary.LittleEndian.PutUint16(h[6:8], s.Flags)
	return h
}

// Marshal serializes the token: the 8-byte header, the 8-byte SequenceNumber, the 8-byte
// Checksum, the 8-byte Confounder (only when Sealed()), and the trailing 24 zero octets of
// Reserved. The result is 48 octets for an integrity-only token and 56 octets for a sealing
// token.
func (s *NL_AUTH_SHA2_SIGNATURE) Marshal() []byte {
	out := make([]byte, 0, 56)
	out = append(out, s.Header()...)
	out = append(out, s.SequenceNumber[:]...)
	out = append(out, s.Checksum[:]...)
	if s.Sealed() {
		out = append(out, s.Confounder[:]...)
	}
	out = append(out, make([]byte, nlAuthSHA2ReservedSize)...)
	return out
}

// Unmarshal parses a token from the front of data. Whether the Confounder is expected is
// decided by SealAlgorithm, so the header is read first. The trailing Reserved field is not
// required to be present (the receiver ignores it), so it is not enforced.
func (s *NL_AUTH_SHA2_SIGNATURE) Unmarshal(data []byte) error {
	if len(data) < nlAuthHeaderSize+16 {
		return fmt.Errorf("NL_AUTH_SHA2_SIGNATURE truncated: have %d bytes, need at least %d", len(data), nlAuthHeaderSize+16)
	}
	s.SignatureAlgorithm = binary.LittleEndian.Uint16(data[0:2])
	s.SealAlgorithm = binary.LittleEndian.Uint16(data[2:4])
	s.Pad = binary.LittleEndian.Uint16(data[4:6])
	s.Flags = binary.LittleEndian.Uint16(data[6:8])
	copy(s.SequenceNumber[:], data[8:16])
	copy(s.Checksum[:], data[16:24])
	if s.Sealed() {
		if len(data) < 32 {
			return fmt.Errorf("NL_AUTH_SHA2_SIGNATURE sealing token truncated: have %d bytes, need at least 32", len(data))
		}
		copy(s.Confounder[:], data[24:32])
	}
	return nil
}
