package createcontext

import (
	"encoding/binary"
	"fmt"
)

// Lease state flags (MS-SMB2 2.2.13.2.8).
const (
	SMB2_LEASE_NONE           uint32 = 0x00
	SMB2_LEASE_READ_CACHING   uint32 = 0x01
	SMB2_LEASE_HANDLE_CACHING uint32 = 0x02
	SMB2_LEASE_WRITE_CACHING  uint32 = 0x04
)

// Lease flags (MS-SMB2 2.2.13.2.8 / 2.2.13.2.10).
const (
	SMB2_LEASE_FLAG_BREAK_IN_PROGRESS     uint32 = 0x02
	SMB2_LEASE_FLAG_PARENT_LEASE_KEY_SET  uint32 = 0x04
)

const (
	leaseV1DataSize = 32
	leaseV2DataSize = 52
)

// LeaseV1 is the SMB2_CREATE_REQUEST_LEASE create context data (MS-SMB2
// 2.2.13.2.8). It requests or reports a file lease: the client proposes a
// LeaseState, and the server grants the subset it can.
type LeaseV1 struct {
	LeaseKey      [16]byte
	LeaseState    uint32
	LeaseFlags    uint32
	LeaseDuration uint64
}

// LeaseV2 is the SMB2_CREATE_REQUEST_LEASE_V2 create context data (MS-SMB2
// 2.2.13.2.10). It extends V1 with a ParentLeaseKey for directory leasing and
// an Epoch for lease-state sequencing.
type LeaseV2 struct {
	LeaseKey       [16]byte
	LeaseState     uint32
	LeaseFlags     uint32
	LeaseDuration  uint64
	ParentLeaseKey [16]byte
	Epoch          uint16
}

// MarshalLeaseV1 encodes a V1 lease request into a create context.
func MarshalLeaseV1(key [16]byte, state uint32) CreateContext {
	data := make([]byte, leaseV1DataSize)
	copy(data[0:16], key[:])
	binary.LittleEndian.PutUint32(data[16:20], state)
	return CreateContext{Name: NameRequestLease, Data: data}
}

// MarshalLeaseV2 encodes a V2 lease request into a create context. When
// parentKey is non-zero, SMB2_LEASE_FLAG_PARENT_LEASE_KEY_SET is set
// automatically.
func MarshalLeaseV2(key [16]byte, state uint32, parentKey [16]byte, epoch uint16) CreateContext {
	data := make([]byte, leaseV2DataSize)
	copy(data[0:16], key[:])
	binary.LittleEndian.PutUint32(data[16:20], state)

	flags := uint32(0)
	zero := [16]byte{}
	if parentKey != zero {
		flags |= SMB2_LEASE_FLAG_PARENT_LEASE_KEY_SET
	}
	binary.LittleEndian.PutUint32(data[20:24], flags)

	copy(data[32:48], parentKey[:])
	binary.LittleEndian.PutUint16(data[48:50], epoch)
	return CreateContext{Name: NameRequestLease, Data: data}
}

// ParseLeaseV1 decodes a V1 lease context from raw data.
func ParseLeaseV1(data []byte) (*LeaseV1, error) {
	if len(data) < leaseV1DataSize {
		return nil, fmt.Errorf("lease V1 data too short: %d bytes, need %d", len(data), leaseV1DataSize)
	}
	l := &LeaseV1{
		LeaseState:    binary.LittleEndian.Uint32(data[16:20]),
		LeaseFlags:    binary.LittleEndian.Uint32(data[20:24]),
		LeaseDuration: binary.LittleEndian.Uint64(data[24:32]),
	}
	copy(l.LeaseKey[:], data[0:16])
	return l, nil
}

// ParseLeaseV2 decodes a V2 lease context from raw data.
func ParseLeaseV2(data []byte) (*LeaseV2, error) {
	if len(data) < leaseV2DataSize {
		return nil, fmt.Errorf("lease V2 data too short: %d bytes, need %d", len(data), leaseV2DataSize)
	}
	l := &LeaseV2{
		LeaseState:    binary.LittleEndian.Uint32(data[16:20]),
		LeaseFlags:    binary.LittleEndian.Uint32(data[20:24]),
		LeaseDuration: binary.LittleEndian.Uint64(data[24:32]),
		Epoch:         binary.LittleEndian.Uint16(data[48:50]),
	}
	copy(l.LeaseKey[:], data[0:16])
	copy(l.ParentLeaseKey[:], data[32:48])
	return l, nil
}

// ParseLeaseResponse parses the "RqLs" create context from a CREATE response,
// returning a *LeaseV1 or *LeaseV2 depending on the data length.
func ParseLeaseResponse(data []byte) (interface{}, error) {
	switch {
	case len(data) >= leaseV2DataSize:
		return ParseLeaseV2(data)
	case len(data) >= leaseV1DataSize:
		return ParseLeaseV1(data)
	default:
		return nil, fmt.Errorf("lease context data too short: %d bytes", len(data))
	}
}
