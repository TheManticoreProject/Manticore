// Package pac parses, builds, and verifies the Microsoft Privilege Attribute
// Certificate (PAC, [MS-PAC]) carried in a Kerberos ticket's authorization data.
//
// A PAC is a PACTYPE container (little-endian, NOT NDR at the container level):
// a cBuffers/Version header followed by a table of PAC_INFO_BUFFER entries, each
// pointing at an 8-byte-aligned buffer. Individual buffers such as the logon
// info (KERB_VALIDATION_INFO) are themselves NDR-encoded; this package exposes
// those buffers as raw bytes and does not yet NDR-decode them.
//
// The security-relevant operations — the Server and KDC signatures ([MS-PAC]
// 2.8, "Generation of PAC Signatures") — are implemented: both are keyed
// checksums with key usage KERB_NON_KERB_CKSUM_SALT (17); the server signature
// covers the entire PAC with all signature buffers' Signature fields zeroed, and
// the KDC signature counter-signs the server signature.
package pac

import (
	"encoding/binary"
	"fmt"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// PAC_INFO_BUFFER ulType values ([MS-PAC] 2.4).
const (
	BufferLogonInfo           = 0x00000001 // KERB_VALIDATION_INFO (NDR)
	BufferCredentials         = 0x00000002 // PAC_CREDENTIAL_INFO
	BufferServerChecksum      = 0x00000006 // server signature
	BufferKDCChecksum         = 0x00000007 // KDC (krbtgt) signature
	BufferClientInfo          = 0x0000000A // PAC_CLIENT_INFO
	BufferConstrainedDeleg    = 0x0000000B // S4U_DELEGATION_INFO
	BufferUPNDNSInfo          = 0x0000000C // UPN_DNS_INFO
	BufferClientClaims        = 0x0000000D // client claims
	BufferDeviceInfo          = 0x0000000E // PAC_DEVICE_INFO
	BufferDeviceClaims        = 0x0000000F // device claims
	BufferTicketChecksum      = 0x00000010 // ticket signature
	BufferAttributesInfo      = 0x00000011 // PAC_ATTRIBUTES_INFO
	BufferRequestorSID        = 0x00000012 // PAC_REQUESTOR
	BufferExtendedKDCChecksum = 0x00000013 // extended KDC (full-PAC) signature
)

// PAC signature SignatureType values ([MS-PAC] 2.8). These are Kerberos checksum
// type numbers; the RC4 value 0xFFFFFF76 is -138 as an int32.
const (
	sigHMACMD5     uint32 = 0xFFFFFF76 // KERB_CHECKSUM_HMAC_MD5, 16 bytes
	sigHMACSHA1128 uint32 = 0x0000000F // HMAC_SHA1_96_AES128, 12 bytes
	sigHMACSHA1256 uint32 = 0x00000010 // HMAC_SHA1_96_AES256, 12 bytes
	sigHMACSHA2128 uint32 = 0x00000013 // HMAC_SHA256_128_AES128, 16 bytes (newer)
	sigHMACSHA2256 uint32 = 0x00000014 // HMAC_SHA384_192_AES256, 24 bytes (newer)
)

// checksumSize returns the signature byte length for a PAC SignatureType.
func checksumSize(sigType uint32) (int, bool) {
	switch sigType {
	case sigHMACMD5, sigHMACSHA2128:
		return 16, true
	case sigHMACSHA1128, sigHMACSHA1256:
		return 12, true
	case sigHMACSHA2256:
		return 24, true
	default:
		return 0, false
	}
}

// cksumType maps a PAC SignatureType to the checksum-type number used by the
// crypto package (sign-extending the RC4 value to -138).
func cksumType(sigType uint32) int {
	return int(int32(sigType))
}

// Buffer is one PAC_INFO_BUFFER and its contents.
type Buffer struct {
	Type   uint32
	Offset uint64 // offset from the start of the PAC (informational after Parse)
	Data   []byte
}

// PAC is a parsed PACTYPE container.
type PAC struct {
	Version uint32
	Buffers []Buffer
	raw     []byte // exact bytes Parse consumed (nil for a freshly built PAC)
}

// Buffer returns the first buffer of the given ulType.
func (p *PAC) Buffer(ulType uint32) (Buffer, bool) {
	for _, b := range p.Buffers {
		if b.Type == ulType {
			return b, true
		}
	}
	return Buffer{}, false
}

// Parse decodes a PACTYPE container. Buffer contents are copied out.
func Parse(data []byte) (*PAC, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("pac: too short (%d bytes)", len(data))
	}
	cBuffers := binary.LittleEndian.Uint32(data[0:])
	version := binary.LittleEndian.Uint32(data[4:])
	if version != 0 {
		return nil, fmt.Errorf("pac: unexpected version %d (want 0)", version)
	}
	if cBuffers > 64 {
		return nil, fmt.Errorf("pac: implausible buffer count %d", cBuffers)
	}
	need := 8 + int(cBuffers)*16
	if len(data) < need {
		return nil, fmt.Errorf("pac: truncated info-buffer table")
	}

	p := &PAC{Version: version, raw: append([]byte(nil), data...)}
	for i := 0; i < int(cBuffers); i++ {
		off := 8 + i*16
		ulType := binary.LittleEndian.Uint32(data[off:])
		cb := binary.LittleEndian.Uint32(data[off+4:])
		bufOff := binary.LittleEndian.Uint64(data[off+8:])
		if bufOff > uint64(len(data)) || bufOff+uint64(cb) > uint64(len(data)) {
			return nil, fmt.Errorf("pac: buffer %d (type 0x%x) out of range: off=%d size=%d", i, ulType, bufOff, cb)
		}
		p.Buffers = append(p.Buffers, Buffer{
			Type:   ulType,
			Offset: bufOff,
			Data:   append([]byte(nil), data[bufOff:bufOff+uint64(cb)]...),
		})
	}
	return p, nil
}

// align8 rounds n up to a multiple of 8.
func align8(n int) int { return (n + 7) &^ 7 }

// Marshal lays out the PAC in wire form: header, PAC_INFO_BUFFER table, then each
// buffer's data 8-byte aligned. It uses the buffers' current Data verbatim (so
// signature buffers should already be sized and, if unsigned, zero-filled).
func (p *PAC) Marshal() ([]byte, error) {
	n := len(p.Buffers)
	headerLen := 8 + n*16
	total := headerLen
	offsets := make([]int, n)
	for i, b := range p.Buffers {
		total = align8(total)
		offsets[i] = total
		total += len(b.Data)
	}
	out := make([]byte, total)
	binary.LittleEndian.PutUint32(out[0:], uint32(n))
	binary.LittleEndian.PutUint32(out[4:], 0) // Version
	for i, b := range p.Buffers {
		e := 8 + i*16
		binary.LittleEndian.PutUint32(out[e:], b.Type)
		binary.LittleEndian.PutUint32(out[e+4:], uint32(len(b.Data)))
		binary.LittleEndian.PutUint64(out[e+8:], uint64(offsets[i]))
		copy(out[offsets[i]:], b.Data)
	}
	return out, nil
}

// signatureBufferTypes lists the buffer types whose Signature fields are zeroed
// before the server signature is computed ([MS-PAC] "Generation of PAC
// Signatures": server, KDC, and extended-KDC signatures are all zeroed).
var signatureBufferTypes = map[uint32]bool{
	BufferServerChecksum:      true,
	BufferKDCChecksum:         true,
	BufferExtendedKDCChecksum: true,
}

// zeroedForServerSig returns a copy of the marshaled PAC with the Signature
// fields (the bytes after the 4-byte SignatureType) of every signature buffer
// set to zero, as required before computing the server signature.
func zeroedForServerSig(marshaled []byte, buffers []Buffer, offsets []int) ([]byte, error) {
	out := append([]byte(nil), marshaled...)
	for i, b := range buffers {
		if !signatureBufferTypes[b.Type] {
			continue
		}
		if len(b.Data) < 4 {
			return nil, fmt.Errorf("pac: signature buffer 0x%x too short", b.Type)
		}
		sigType := binary.LittleEndian.Uint32(b.Data[0:])
		sz, ok := checksumSize(sigType)
		if !ok {
			return nil, fmt.Errorf("pac: unsupported signature type 0x%x", sigType)
		}
		start := offsets[i] + 4
		if start+sz > len(out) {
			return nil, fmt.Errorf("pac: signature buffer 0x%x overruns PAC", b.Type)
		}
		for j := 0; j < sz; j++ {
			out[start+j] = 0
		}
	}
	return out, nil
}

// layout re-marshals and returns the bytes plus each buffer's offset.
func (p *PAC) layout() ([]byte, []int) {
	n := len(p.Buffers)
	headerLen := 8 + n*16
	total := headerLen
	offsets := make([]int, n)
	for i, b := range p.Buffers {
		total = align8(total)
		offsets[i] = total
		total += len(b.Data)
	}
	m, _ := p.Marshal()
	return m, offsets
}

// Sign fills the server signature (buffer 0x06) and then the KDC signature
// (buffer 0x07) in place, using the given keys. Both use key usage
// KERB_NON_KERB_CKSUM_SALT. The buffers must already exist, sized to
// 4 + checksumSize(their SignatureType). Returns the fully signed marshaled PAC.
func (p *PAC) Sign(serverKey, kdcKey []byte) ([]byte, error) {
	srvIdx, kdcIdx := -1, -1
	for i, b := range p.Buffers {
		switch b.Type {
		case BufferServerChecksum:
			srvIdx = i
		case BufferKDCChecksum:
			kdcIdx = i
		}
	}
	if srvIdx < 0 || kdcIdx < 0 {
		return nil, fmt.Errorf("pac: Sign requires both server (0x06) and KDC (0x07) checksum buffers")
	}

	// Server signature: over the whole PAC with signature fields zeroed.
	marshaled, offsets := p.layout()
	zeroed, err := zeroedForServerSig(marshaled, p.Buffers, offsets)
	if err != nil {
		return nil, err
	}
	srvType := binary.LittleEndian.Uint32(p.Buffers[srvIdx].Data[0:])
	srvSig, err := kerbcrypto.GetChecksum(cksumType(srvType), serverKey, iana.KeyUsageKerbNonKerbCksumSalt, zeroed)
	if err != nil {
		return nil, fmt.Errorf("pac: server signature: %w", err)
	}
	copy(p.Buffers[srvIdx].Data[4:], srvSig)

	// KDC signature: counter-sign the server Signature bytes.
	kdcType := binary.LittleEndian.Uint32(p.Buffers[kdcIdx].Data[0:])
	kdcSig, err := kerbcrypto.GetChecksum(cksumType(kdcType), kdcKey, iana.KeyUsageKerbNonKerbCksumSalt, srvSig)
	if err != nil {
		return nil, fmt.Errorf("pac: KDC signature: %w", err)
	}
	copy(p.Buffers[kdcIdx].Data[4:], kdcSig)

	return p.Marshal()
}

// VerifyServerSignature recomputes the server signature with serverKey and
// compares it, in constant time, to the stored value. The PAC must have been
// produced by Parse (so the exact issued bytes are available).
func (p *PAC) VerifyServerSignature(serverKey []byte) error {
	if p.raw == nil {
		return fmt.Errorf("pac: VerifyServerSignature requires a parsed PAC")
	}
	srv, ok := p.Buffer(BufferServerChecksum)
	if !ok {
		return fmt.Errorf("pac: no server signature buffer")
	}
	if len(srv.Data) < 4 {
		return fmt.Errorf("pac: server signature buffer too short")
	}
	sigType := binary.LittleEndian.Uint32(srv.Data[0:])
	sz, ok := checksumSize(sigType)
	if !ok {
		return fmt.Errorf("pac: unsupported server signature type 0x%x", sigType)
	}
	stored := srv.Data[4 : 4+sz]

	zeroed, err := p.rawZeroedForServerSig()
	if err != nil {
		return err
	}
	if !kerbcrypto.VerifyChecksum(cksumType(sigType), serverKey, iana.KeyUsageKerbNonKerbCksumSalt, zeroed, stored) {
		return fmt.Errorf("pac: server signature verification failed")
	}
	return nil
}

// VerifyKDCSignature recomputes the KDC signature with the krbtgt key (over the
// stored server-signature bytes) and compares it to the stored value.
func (p *PAC) VerifyKDCSignature(krbtgtKey []byte) error {
	srv, ok := p.Buffer(BufferServerChecksum)
	if !ok {
		return fmt.Errorf("pac: no server signature buffer")
	}
	kdc, ok := p.Buffer(BufferKDCChecksum)
	if !ok {
		return fmt.Errorf("pac: no KDC signature buffer")
	}
	srvType := binary.LittleEndian.Uint32(srv.Data[0:])
	srvSz, ok := checksumSize(srvType)
	if !ok || len(srv.Data) < 4+srvSz {
		return fmt.Errorf("pac: bad server signature buffer")
	}
	serverSig := srv.Data[4 : 4+srvSz]

	kdcType := binary.LittleEndian.Uint32(kdc.Data[0:])
	kdcSz, ok := checksumSize(kdcType)
	if !ok || len(kdc.Data) < 4+kdcSz {
		return fmt.Errorf("pac: bad KDC signature buffer")
	}
	stored := kdc.Data[4 : 4+kdcSz]

	if !kerbcrypto.VerifyChecksum(cksumType(kdcType), krbtgtKey, iana.KeyUsageKerbNonKerbCksumSalt, serverSig, stored) {
		return fmt.Errorf("pac: KDC signature verification failed")
	}
	return nil
}

// rawZeroedForServerSig zeroes the signature fields within the originally parsed
// bytes, locating each signature buffer by its recorded Offset.
func (p *PAC) rawZeroedForServerSig() ([]byte, error) {
	out := append([]byte(nil), p.raw...)
	for _, b := range p.Buffers {
		if !signatureBufferTypes[b.Type] {
			continue
		}
		if len(b.Data) < 4 {
			return nil, fmt.Errorf("pac: signature buffer 0x%x too short", b.Type)
		}
		sigType := binary.LittleEndian.Uint32(b.Data[0:])
		sz, ok := checksumSize(sigType)
		if !ok {
			return nil, fmt.Errorf("pac: unsupported signature type 0x%x", sigType)
		}
		start := int(b.Offset) + 4
		if start+sz > len(out) {
			return nil, fmt.Errorf("pac: signature buffer 0x%x overruns PAC", b.Type)
		}
		for j := 0; j < sz; j++ {
			out[start+j] = 0
		}
	}
	return out, nil
}
