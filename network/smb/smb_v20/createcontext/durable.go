package createcontext

import (
	"encoding/binary"
	"fmt"
)

// Durable handle flags (MS-SMB2 2.2.13.2.11).
const (
	SMB2_DHANDLE_FLAG_PERSISTENT uint32 = 0x00000002
)

// --- V1 durable handles (SMB 2.x) ---

const (
	durableV1ReqDataSize   = 16
	durableV1ReconDataSize = 16
)

// MarshalDurableHandleRequestV1 encodes a V1 durable handle request (DHnQ).
// The 16-byte data field is reserved and MUST be zeros.
func MarshalDurableHandleRequestV1() CreateContext {
	return CreateContext{Name: NameDurableHandleReq, Data: make([]byte, durableV1ReqDataSize)}
}

// MarshalDurableHandleReconnectV1 encodes a V1 durable handle reconnect (DHnC).
// The data carries the FileId (persistent + volatile, 16 bytes) of the original open.
func MarshalDurableHandleReconnectV1(persistent, volatile uint64) CreateContext {
	data := make([]byte, durableV1ReconDataSize)
	binary.LittleEndian.PutUint64(data[0:8], persistent)
	binary.LittleEndian.PutUint64(data[8:16], volatile)
	return CreateContext{Name: NameDurableHandleRecon, Data: data}
}

// ParseDurableHandleReconnectV1 decodes the FileId from a V1 reconnect context.
func ParseDurableHandleReconnectV1(data []byte) (persistent, volatile uint64, err error) {
	if len(data) < durableV1ReconDataSize {
		return 0, 0, fmt.Errorf("durable handle reconnect V1 data too short: %d bytes, need %d", len(data), durableV1ReconDataSize)
	}
	return binary.LittleEndian.Uint64(data[0:8]), binary.LittleEndian.Uint64(data[8:16]), nil
}

// --- V2 durable handles (SMB 3.x) ---

const (
	durableV2ReqDataSize       = 32
	durableV2ReconDataSize     = 36
	durableV2ResponseDataSize  = 8
)

// DurableHandleRequestV2 is the parsed SMB2_CREATE_DURABLE_HANDLE_REQUEST_V2
// (DH2Q) create context data (MS-SMB2 2.2.13.2.11).
type DurableHandleRequestV2 struct {
	Timeout    uint32
	Flags      uint32
	CreateGuid [16]byte
}

// DurableHandleReconnectV2 is the parsed SMB2_CREATE_DURABLE_HANDLE_RECONNECT_V2
// (DH2C) create context data (MS-SMB2 2.2.13.2.12).
type DurableHandleReconnectV2 struct {
	Persistent uint64
	Volatile   uint64
	CreateGuid [16]byte
	Flags      uint32
}

// DurableHandleResponseV2 is the parsed SMB2_CREATE_DURABLE_HANDLE_RESPONSE_V2
// create context data from a CREATE response (MS-SMB2 2.2.14.2.12).
type DurableHandleResponseV2 struct {
	Timeout uint32
	Flags   uint32
}

// MarshalDurableHandleRequestV2 encodes a V2 durable handle request (DH2Q).
func MarshalDurableHandleRequestV2(timeout uint32, flags uint32, createGuid [16]byte) CreateContext {
	data := make([]byte, durableV2ReqDataSize)
	binary.LittleEndian.PutUint32(data[0:4], timeout)
	binary.LittleEndian.PutUint32(data[4:8], flags)
	// data[8:16] reserved
	copy(data[16:32], createGuid[:])
	return CreateContext{Name: NameDurableHandleReqV2, Data: data}
}

// ParseDurableHandleRequestV2 decodes a DH2Q create context.
func ParseDurableHandleRequestV2(data []byte) (*DurableHandleRequestV2, error) {
	if len(data) < durableV2ReqDataSize {
		return nil, fmt.Errorf("durable handle request V2 data too short: %d bytes, need %d", len(data), durableV2ReqDataSize)
	}
	d := &DurableHandleRequestV2{
		Timeout: binary.LittleEndian.Uint32(data[0:4]),
		Flags:   binary.LittleEndian.Uint32(data[4:8]),
	}
	copy(d.CreateGuid[:], data[16:32])
	return d, nil
}

// MarshalDurableHandleReconnectV2 encodes a V2 durable handle reconnect (DH2C).
func MarshalDurableHandleReconnectV2(persistent, volatile uint64, createGuid [16]byte, flags uint32) CreateContext {
	data := make([]byte, durableV2ReconDataSize)
	binary.LittleEndian.PutUint64(data[0:8], persistent)
	binary.LittleEndian.PutUint64(data[8:16], volatile)
	copy(data[16:32], createGuid[:])
	binary.LittleEndian.PutUint32(data[32:36], flags)
	return CreateContext{Name: NameDurableHandleReconV2, Data: data}
}

// ParseDurableHandleReconnectV2 decodes a DH2C create context.
func ParseDurableHandleReconnectV2(data []byte) (*DurableHandleReconnectV2, error) {
	if len(data) < durableV2ReconDataSize {
		return nil, fmt.Errorf("durable handle reconnect V2 data too short: %d bytes, need %d", len(data), durableV2ReconDataSize)
	}
	d := &DurableHandleReconnectV2{
		Persistent: binary.LittleEndian.Uint64(data[0:8]),
		Volatile:   binary.LittleEndian.Uint64(data[8:16]),
		Flags:      binary.LittleEndian.Uint32(data[32:36]),
	}
	copy(d.CreateGuid[:], data[16:32])
	return d, nil
}

// ParseDurableHandleResponseV2 decodes the DH2Q create context from a CREATE response.
func ParseDurableHandleResponseV2(data []byte) (*DurableHandleResponseV2, error) {
	if len(data) < durableV2ResponseDataSize {
		return nil, fmt.Errorf("durable handle response V2 data too short: %d bytes, need %d", len(data), durableV2ResponseDataSize)
	}
	return &DurableHandleResponseV2{
		Timeout: binary.LittleEndian.Uint32(data[0:4]),
		Flags:   binary.LittleEndian.Uint32(data[4:8]),
	}, nil
}
