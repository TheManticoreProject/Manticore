// Package createcontext implements the SMB2_CREATE_CONTEXT structure carried in
// the variable buffer of SMB2 CREATE requests and responses. A create context is
// a tagged name/data pair (durable-handle request, lease request, query
// maximal-access, allocation size, …); CREATE carries a chained list of them.
//
// The SMB2 CREATE command bodies keep the context list as a raw []byte
// (CreateRequest.CreateContexts / CreateResponse.CreateContexts); this package
// builds and parses that buffer into typed contexts.
//
// [MS-SMB2] 2.2.13.2 SMB2_CREATE_CONTEXT Request Values:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/93a3d702-90a0-4c8a-86c4-12dfa9230f5f
package createcontext

import (
	"encoding/binary"
	"fmt"
)

// Well-known SMB2_CREATE_CONTEXT name tags (MS-SMB2 2.2.13.2). The name is the
// ASCII tag the server matches against; most are 4 bytes.
var (
	NameExtendedAttributes = []byte("ExtA") // SMB2_CREATE_EA_BUFFER
	NameSecurityDescriptor = []byte("SecD") // SMB2_CREATE_SD_BUFFER
	NameDurableHandleReq   = []byte("DHnQ") // SMB2_CREATE_DURABLE_HANDLE_REQUEST
	NameDurableHandleRecon = []byte("DHnC") // SMB2_CREATE_DURABLE_HANDLE_RECONNECT
	NameAllocationSize     = []byte("AlSi") // SMB2_CREATE_ALLOCATION_SIZE
	NameQueryMaximalAccess = []byte("MxAc") // SMB2_CREATE_QUERY_MAXIMAL_ACCESS_REQUEST
	NameTimewarpToken      = []byte("TWrp") // SMB2_CREATE_TIMEWARP_TOKEN
	NameQueryOnDiskID      = []byte("QFid") // SMB2_CREATE_QUERY_ON_DISK_ID
	NameRequestLease       = []byte("RqLs") // SMB2_CREATE_REQUEST_LEASE
)

// headerSize is the fixed SMB2_CREATE_CONTEXT header: Next(4) NameOffset(2)
// NameLength(2) Reserved(2) DataOffset(2) DataLength(4).
const headerSize = 16

// CreateContext is a single SMB2_CREATE_CONTEXT: a name tag and its associated
// data buffer (either may drive the create behavior; Data is empty for contexts
// that carry only a name).
type CreateContext struct {
	Name []byte
	Data []byte
}

// align8 rounds n up to the next multiple of 8.
func align8(n int) int {
	return (n + 7) &^ 7
}

// Marshal encodes a list of create contexts into the chained, 8-byte-aligned
// buffer carried by CreateRequest/CreateResponse.CreateContexts. Each context's
// Name is placed immediately after the 16-byte header; its Data (when present)
// follows on the next 8-byte boundary. Every context except the last is padded so
// the next begins on an 8-byte boundary, and its Next field is set accordingly.
func Marshal(contexts []CreateContext) ([]byte, error) {
	if len(contexts) == 0 {
		return nil, nil
	}

	var out []byte
	for i, ctx := range contexts {
		nameOffset := headerSize
		dataOffset := 0
		// Data, when present, starts on the 8-byte boundary after the name.
		if len(ctx.Data) > 0 {
			dataOffset = align8(headerSize + len(ctx.Name))
		}

		segLen := headerSize + len(ctx.Name)
		if len(ctx.Data) > 0 {
			segLen = dataOffset + len(ctx.Data)
		}

		isLast := i == len(contexts)-1
		padded := segLen
		if !isLast {
			padded = align8(segLen)
		}

		seg := make([]byte, padded)
		if !isLast {
			binary.LittleEndian.PutUint32(seg[0:4], uint32(padded)) // Next
		}
		binary.LittleEndian.PutUint16(seg[4:6], uint16(nameOffset))
		binary.LittleEndian.PutUint16(seg[6:8], uint16(len(ctx.Name)))
		// seg[8:10] Reserved = 0
		if len(ctx.Data) > 0 {
			binary.LittleEndian.PutUint16(seg[10:12], uint16(dataOffset))
			binary.LittleEndian.PutUint32(seg[12:16], uint32(len(ctx.Data)))
			copy(seg[dataOffset:], ctx.Data)
		}
		copy(seg[nameOffset:], ctx.Name)

		out = append(out, seg...)
	}
	return out, nil
}

// Parse decodes a chained SMB2_CREATE_CONTEXT buffer into its contexts. All
// offsets and lengths are validated against the buffer, so a malformed
// (server-controlled) list is rejected rather than indexing out of range.
func Parse(buf []byte) ([]CreateContext, error) {
	var out []CreateContext
	offset := 0
	for offset < len(buf) {
		if len(buf)-offset < headerSize {
			return nil, fmt.Errorf("create context at offset %d shorter than header", offset)
		}
		ctx := buf[offset:]
		next := int(binary.LittleEndian.Uint32(ctx[0:4]))
		nameOffset := int(binary.LittleEndian.Uint16(ctx[4:6]))
		nameLength := int(binary.LittleEndian.Uint16(ctx[6:8]))
		dataOffset := int(binary.LittleEndian.Uint16(ctx[10:12]))
		dataLength := int(binary.LittleEndian.Uint32(ctx[12:16]))

		// The fields above index within this context. Bound the context to the
		// remaining buffer (or to Next when set) before slicing name/data.
		ctxEnd := len(buf) - offset
		if next != 0 {
			if next < headerSize || offset+next > len(buf) {
				return nil, fmt.Errorf("create context Next %d at offset %d out of bounds", next, offset)
			}
			ctxEnd = next
		}

		var name, data []byte
		if nameLength > 0 {
			if nameOffset < headerSize || nameOffset+nameLength > ctxEnd {
				return nil, fmt.Errorf("create context name [%d:%d] at offset %d out of bounds", nameOffset, nameOffset+nameLength, offset)
			}
			name = append([]byte{}, ctx[nameOffset:nameOffset+nameLength]...)
		}
		if dataLength > 0 {
			if dataOffset < headerSize || dataOffset+dataLength > ctxEnd {
				return nil, fmt.Errorf("create context data [%d:%d] at offset %d out of bounds", dataOffset, dataOffset+dataLength, offset)
			}
			data = append([]byte{}, ctx[dataOffset:dataOffset+dataLength]...)
		}
		out = append(out, CreateContext{Name: name, Data: data})

		if next == 0 {
			break
		}
		offset += next
	}
	return out, nil
}
