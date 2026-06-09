package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

const (
	// LockRequestStructureSize is the fixed StructureSize value the client MUST
	// set for an SMB2 LOCK Request. It is 48 — the size with a single
	// SMB2_LOCK_ELEMENT — regardless of how many lock elements are sent.
	LockRequestStructureSize = 48

	// lockRequestFixedSize is the size, in bytes, of the portion of the body that
	// precedes the variable Locks array.
	lockRequestFixedSize = 24

	// LockElementSize is the wire size, in bytes, of one SMB2_LOCK_ELEMENT.
	LockElementSize = 24
)

// SMB2 lock element flags (SMB2_LOCK_ELEMENT Flags field).
const (
	SMB2_LOCKFLAG_SHARED_LOCK      = 0x00000001
	SMB2_LOCKFLAG_EXCLUSIVE_LOCK   = 0x00000002
	SMB2_LOCKFLAG_UNLOCK           = 0x00000004
	SMB2_LOCKFLAG_FAIL_IMMEDIATELY = 0x00000010
)

// LockElement is a single SMB2_LOCK_ELEMENT describing a byte range to lock or
// unlock.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/73e941c7-9b07-42f6-8b0f-31c1a2cbf0b2
type LockElement struct {
	// Offset (8 bytes): The byte offset into the file at which the range begins.
	Offset types.UINT64
	// Length (8 bytes): The length, in bytes, of the range.
	Length types.UINT64
	// Flags (4 bytes): Whether the range is locked shared/exclusive or unlocked.
	Flags types.ULONG
	// Reserved (4 bytes): The client MUST set this to 0.
	Reserved types.ULONG
}

// LockRequest is the SMB2 LOCK Request body, sent by the client to lock or unlock
// one or more byte ranges of the file identified by FileId.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/6178b960-48b6-4999-b589-669f88e9017d
type LockRequest struct {
	command_interface.Command

	// LockSequence (4 bytes): Reserved in SMB 2.0.2; lock sequence number/index otherwise.
	LockSequence types.ULONG

	// FileId (16 bytes): The file on which to perform the locks/unlocks.
	FileId types.SMB2_FILEID

	// Locks is the array of byte-range lock elements (at least one).
	Locks []LockElement
}

// NewLockRequest creates a new SMB2 LOCK Request.
func NewLockRequest() *LockRequest {
	c := &LockRequest{Locks: []LockElement{}}
	c.SetCommandCode(codes.SMB2_LOCK)
	c.StructureSize = LockRequestStructureSize
	return c
}

// Marshal serializes the LOCK Request body.
func (c *LockRequest) Marshal() ([]byte, error) {
	buf := make([]byte, lockRequestFixedSize)
	binary.LittleEndian.PutUint16(buf[0:2], LockRequestStructureSize)
	binary.LittleEndian.PutUint16(buf[2:4], uint16(len(c.Locks)))
	binary.LittleEndian.PutUint32(buf[4:8], c.LockSequence)

	fileId, err := c.FileId.Marshal()
	if err != nil {
		return nil, err
	}
	copy(buf[8:24], fileId)

	for _, lock := range c.Locks {
		element := make([]byte, LockElementSize)
		binary.LittleEndian.PutUint64(element[0:8], lock.Offset)
		binary.LittleEndian.PutUint64(element[8:16], lock.Length)
		binary.LittleEndian.PutUint32(element[16:20], lock.Flags)
		binary.LittleEndian.PutUint32(element[20:24], lock.Reserved)
		buf = append(buf, element...)
	}

	return buf, nil
}

// Unmarshal deserializes the LOCK Request body.
func (c *LockRequest) Unmarshal(data []byte) (int, error) {
	if len(data) < lockRequestFixedSize {
		return 0, fmt.Errorf("data too short to unmarshal SMB2 LOCK Request: have %d bytes, need at least %d", len(data), lockRequestFixedSize)
	}

	c.StructureSize = binary.LittleEndian.Uint16(data[0:2])
	lockCount := int(binary.LittleEndian.Uint16(data[2:4]))
	c.LockSequence = binary.LittleEndian.Uint32(data[4:8])
	if _, err := c.FileId.Unmarshal(data[8:24]); err != nil {
		return 0, err
	}

	offset := lockRequestFixedSize
	if len(data) < offset+lockCount*LockElementSize {
		return 0, fmt.Errorf("data too short for %d lock elements: have %d bytes from offset %d", lockCount, len(data)-offset, offset)
	}
	c.Locks = make([]LockElement, lockCount)
	for i := 0; i < lockCount; i++ {
		e := data[offset : offset+LockElementSize]
		c.Locks[i] = LockElement{
			Offset:   binary.LittleEndian.Uint64(e[0:8]),
			Length:   binary.LittleEndian.Uint64(e[8:16]),
			Flags:    binary.LittleEndian.Uint32(e[16:20]),
			Reserved: binary.LittleEndian.Uint32(e[20:24]),
		}
		offset += LockElementSize
	}

	return offset, nil
}
