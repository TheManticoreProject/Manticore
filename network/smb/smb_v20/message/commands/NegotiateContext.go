package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
)

// SMB2 NEGOTIATE context types (MS-SMB2 2.2.3.1), carried in the SMB 3.1.1
// NEGOTIATE Request/Response after the dialects array.
const (
	SMB2_PREAUTH_INTEGRITY_CAPABILITIES = 0x0001
	SMB2_ENCRYPTION_CAPABILITIES        = 0x0002
	SMB2_COMPRESSION_CAPABILITIES       = 0x0003
	SMB2_NETNAME_NEGOTIATE_CONTEXT_ID   = 0x0005
	SMB2_TRANSPORT_CAPABILITIES         = 0x0006
	SMB2_RDMA_TRANSFORM_CAPABILITIES    = 0x0007
	SMB2_SIGNING_CAPABILITIES           = 0x0008
)

// SMB2 pre-authentication integrity hash algorithm IDs (MS-SMB2 2.2.3.1.1).
const SMB2_PREAUTH_HASH_SHA_512 = 0x0001

// SMB2 encryption cipher IDs (MS-SMB2 2.2.3.1.2).
const (
	SMB2_ENCRYPTION_AES128_CCM = 0x0001
	SMB2_ENCRYPTION_AES128_GCM = 0x0002
	SMB2_ENCRYPTION_AES256_CCM = 0x0003
	SMB2_ENCRYPTION_AES256_GCM = 0x0004
)

// NegotiateContext is a single SMB2 NEGOTIATE context: a 2-byte type, a 2-byte
// data length, 4 reserved bytes, and the context-specific data. On the wire the
// contexts are 8-byte aligned relative to the start of the SMB2 header.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/15332256-522e-4a53-8cd7-0bd17678a2f7
type NegotiateContext struct {
	ContextType uint16
	Data        []byte
}

// NewPreauthIntegrityContext builds an SMB2_PREAUTH_INTEGRITY_CAPABILITIES
// context advertising the SHA-512 hash algorithm and the supplied salt.
func NewPreauthIntegrityContext(salt []byte) *NegotiateContext {
	data := make([]byte, 6+len(salt))
	binary.LittleEndian.PutUint16(data[0:2], 1) // HashAlgorithmCount
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(salt)))
	binary.LittleEndian.PutUint16(data[4:6], SMB2_PREAUTH_HASH_SHA_512)
	copy(data[6:], salt)
	return &NegotiateContext{ContextType: SMB2_PREAUTH_INTEGRITY_CAPABILITIES, Data: data}
}

// NewEncryptionContext builds an SMB2_ENCRYPTION_CAPABILITIES context
// advertising the supplied ciphers in preference order.
func NewEncryptionContext(ciphers []uint16) *NegotiateContext {
	data := make([]byte, 2+2*len(ciphers))
	binary.LittleEndian.PutUint16(data[0:2], uint16(len(ciphers)))
	for i, c := range ciphers {
		binary.LittleEndian.PutUint16(data[2+2*i:], c)
	}
	return &NegotiateContext{ContextType: SMB2_ENCRYPTION_CAPABILITIES, Data: data}
}

// marshalNegotiateContexts serializes a list of negotiate contexts, 8-byte
// aligning each one relative to the start of the SMB2 header. bodyOffset is the
// offset, within the message body, immediately after the dialects array (before
// any alignment). It returns the padding + context bytes to append to the body
// and the header-relative offset of the first context (the value for the
// NegotiateContextOffset field).
func marshalNegotiateContexts(bodyOffset int, contexts []*NegotiateContext) ([]byte, int) {
	var out []byte
	cur := bodyOffset

	pad := func() {
		if p := (8 - (header.SMB2_HEADER_SIZE+cur)%8) % 8; p > 0 {
			out = append(out, make([]byte, p)...)
			cur += p
		}
	}

	pad()
	first := header.SMB2_HEADER_SIZE + cur
	for i, ctx := range contexts {
		if i > 0 {
			pad()
		}
		var h [8]byte
		binary.LittleEndian.PutUint16(h[0:2], ctx.ContextType)
		binary.LittleEndian.PutUint16(h[2:4], uint16(len(ctx.Data)))
		out = append(out, h[:]...)
		out = append(out, ctx.Data...)
		cur += 8 + len(ctx.Data)
	}
	return out, first
}

// parseNegotiateContexts decodes count negotiate contexts from data, whose first
// context begins at headerRelativeOffset (relative to the start of the SMB2
// header). data begins at the message body (immediately after the 64-byte
// header). Each context is 8-byte aligned relative to the header.
func parseNegotiateContexts(data []byte, headerRelativeOffset int, count int) ([]*NegotiateContext, error) {
	contexts := make([]*NegotiateContext, 0, count)
	off := headerRelativeOffset - header.SMB2_HEADER_SIZE
	for i := 0; i < count; i++ {
		// Align to 8 relative to the header.
		if p := (8 - (header.SMB2_HEADER_SIZE+off)%8) % 8; p > 0 {
			off += p
		}
		if off+8 > len(data) {
			return nil, fmt.Errorf("negotiate context %d header out of bounds", i)
		}
		ctxType := binary.LittleEndian.Uint16(data[off : off+2])
		dataLen := int(binary.LittleEndian.Uint16(data[off+2 : off+4]))
		off += 8
		if off+dataLen > len(data) {
			return nil, fmt.Errorf("negotiate context %d data out of bounds", i)
		}
		cd := make([]byte, dataLen)
		copy(cd, data[off:off+dataLen])
		contexts = append(contexts, &NegotiateContext{ContextType: ctxType, Data: cd})
		off += dataLen
	}
	return contexts, nil
}

// SelectedCipher returns the cipher ID from an SMB2_ENCRYPTION_CAPABILITIES
// context in the list (the server echoes its single chosen cipher), or 0 if
// none is present.
func SelectedCipher(contexts []*NegotiateContext) uint16 {
	for _, ctx := range contexts {
		if ctx.ContextType == SMB2_ENCRYPTION_CAPABILITIES && len(ctx.Data) >= 4 {
			return binary.LittleEndian.Uint16(ctx.Data[2:4])
		}
	}
	return 0
}

// SelectedPreauthHash returns the hash algorithm ID from an
// SMB2_PREAUTH_INTEGRITY_CAPABILITIES context in the list, or 0 if none.
func SelectedPreauthHash(contexts []*NegotiateContext) uint16 {
	for _, ctx := range contexts {
		if ctx.ContextType == SMB2_PREAUTH_INTEGRITY_CAPABILITIES && len(ctx.Data) >= 8 {
			return binary.LittleEndian.Uint16(ctx.Data[4:6])
		}
	}
	return 0
}
