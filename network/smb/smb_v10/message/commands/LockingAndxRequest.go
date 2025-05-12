package commands

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/command_interface"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// LockingAndxRequest
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/b5c6eae7-976b-4444-b52e-c76c68c861ad
type LockingAndxRequest struct {
	command_interface.Command

	// Parameters

	// FID (2 bytes): This field MUST be a valid 16-bit unsigned integer indicating the file from which the data SHOULD be read.
	FID types.USHORT

	// TypeOfLock (1 byte): This field is an 8-bit unsigned integer bit mask indicating the nature of the lock request and the
	// format of the LOCKING_ANDX_RANGE data. If the negotiated protocol is NT LAN Manager or later, CAP_LARGE_FILES was negotiated
	// and LARGE_FILES bit is set, then the Locks and Unlocks arrays are in the large file 64-bit offset LOCKING_ANDX_RANGE format.
	// This allows specification of 64-bit offsets for very large files.

	// If TypeOfLock has the SHARED_LOCK bit set, the lock is specified as a shared read-only lock. If shared read-only locks cannot
	// be supported by a server, the server SHOULD map the lock to an exclusive lock for both read and write. Locks for both read and
	// write messages in which TypeOfLock bit READ_WRITE_LOCK is set SHOULD be prohibited by the server, and the server SHOULD return
	// an implementation-specific error to the client. If TypeOfLock has the CHANGE_LOCKTYPE bit set, the client is requesting that
	// the server atomically change the lock type from a shared lock to an exclusive lock, or vice versa. If the server cannot do this
	// in an atomic fashion, the server MUST reject this request and return an implementation-specific error to the client. Closing a
	// file with locks still in force causes the locks to be released in a nondeterministic order.
	// If the Locks vector contains one and only one entry (NumberOfRequestedLocks == 1) and TypeOfLock has the CANCEL_LOCK bit set, the
	// client is requesting that the server cancel a previously requested but unacknowledged lock. This allows the client to cancel lock
	// requests that can wait forever to complete (see Timeout below).
	TypeOfLock types.UCHAR

	// NewOpLockLevel (1 byte): This field is valid only in SMB_COM_LOCKING_ANDX (0x24) (section 2.2.4.32) SMB requests sent from the
	// server to the client in response to a change in an existing OpLock's state. This field is an 8-bit unsigned integer indicating the
	// OpLock level now in effect for the FID in the request. If NewOpLockLevel is 0x00, the client possesses no OpLocks on the file at all.
	// If NewOpLockLevel is 0x01, then the client possesses a Level II OpLock.
	NewOpLockLevel types.UCHAR

	// Timeout (4 bytes): This field is a 32-bit unsigned integer value. Timeout is the maximum amount of time to wait, in milliseconds,
	// for the byte range(s) specified in Locks to become locked. A Timeout value of 0x00000000 indicates that the server fails immediately
	// if any lock range specified is already locked and cannot be locked by this request. A Timeout value of -1 (0xFFFFFFFF) indicates that
	// the server waits as long as it takes (wait forever) for each byte range specified to become unlocked so that it can be locked by this
	// request. Any other value of Timeout specifies the maximum number of milliseconds to wait for all lock ranges specified in Locks to
	// become available and to be locked by this request.
	Timeout types.ULONG

	// NumberOfRequestedUnlocks (2 bytes): This field is a 16-bit unsigned integer value containing the number of entries in the Unlocks array.
	NumberOfRequestedUnlocks types.USHORT

	// NumberOfRequestedLocks (2 bytes): This field is a 16-bit unsigned integer value containing the number of entries in the Locks array.
	NumberOfRequestedLocks types.USHORT

	// Data

	// Unlocks (variable): An array of byte ranges to be unlocked.
	// If 32-bit offsets are being used, this field uses LOCKING_ANDX_RANGE32 (see below) and is (10 * NumberOfRequestedUnlocks) bytes in length.
	// If 64-bit offsets are being used, this field uses LOCKING_ANDX_RANGE64 (see below) and is (20 * NumberOfRequestedUnlocks) bytes in length.
	Unlocks []types.LOCKING_ANDX_RANGE64

	// Locks (variable): An array of byte ranges to be locked.
	// If 32-bit offsets are being used, this field uses LOCKING_ANDX_RANGE32 (see below) and is (10 * NumberOfRequestedLocks) bytes in length.
	// If 64-bit offsets are being used, this field uses LOCKING_ANDX_RANGE64 (see below) and is (20 * NumberOfRequestedLocks) bytes in length.
	Locks []types.LOCKING_ANDX_RANGE64
}

// NewLockingAndxRequest creates a new LockingAndxRequest structure
//
// Returns:
// - A pointer to the new LockingAndxRequest structure
func NewLockingAndxRequest() *LockingAndxRequest {
	c := &LockingAndxRequest{
		// Parameters
		FID:                      types.USHORT(0),
		TypeOfLock:               types.UCHAR(0),
		NewOpLockLevel:           types.UCHAR(0),
		Timeout:                  types.ULONG(0),
		NumberOfRequestedUnlocks: types.USHORT(0),
		NumberOfRequestedLocks:   types.USHORT(0),

		// Data
		Unlocks: []types.LOCKING_ANDX_RANGE64{},
		Locks:   []types.LOCKING_ANDX_RANGE64{},
	}

	c.Command.SetCommandCode(codes.SMB_COM_LOCKING_ANDX)

	return c
}

// IsAndX returns true if the command is an AndX
func (c *LockingAndxRequest) IsAndX() bool {
	return true
}

// Marshal marshals the LockingAndxRequest structure into a byte array
//
// Returns:
// - A byte array representing the LockingAndxRequest structure
// - An error if the marshaling fails
func (c *LockingAndxRequest) Marshal() ([]byte, error) {
	marshalledCommand := []byte{}

	// Create the Parameters structure if it is nil
	if c.GetParameters() == nil {
		c.SetParameters(parameters.NewParameters())
	}
	// Create the Data structure if it is nil
	if c.GetData() == nil {
		c.SetData(data.NewData())
	}

	// In case of AndX, we need to add the parameters to the Parameters structure first
	if c.IsAndX() {
		if c.GetAndX() == nil {
			c.SetAndX(andx.NewAndX())
			c.GetAndX().AndXCommand = c.GetCommandCode()
		}

		for _, parameter := range c.GetAndX().GetParameters() {
			c.GetParameters().AddWord(parameter)
		}
	}

	// First marshal the data and then the parameters
	// This is because some parameters are dependent on the data, for example the size of some fields within
	// the data will be stored in the parameters
	rawDataContent := []byte{}

	// Marshal Unlocks
	for _, unlock := range c.Unlocks {
		unlockBytes, err := unlock.Marshal()
		if err != nil {
			return nil, fmt.Errorf("error marshalling unlock: %v", err)
		}
		rawDataContent = append(rawDataContent, unlockBytes...)
	}

	// Marshal Locks
	for _, lock := range c.Locks {
		lockBytes, err := lock.Marshal()
		if err != nil {
			return nil, fmt.Errorf("error marshalling lock: %v", err)
		}
		rawDataContent = append(rawDataContent, lockBytes...)
	}

	// Then marshal the parameters
	rawParametersContent := []byte{}

	// Marshalling parameter FID
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.FID))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter TypeOfLock
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.TypeOfLock))

	// Marshalling parameter NewOpLockLevel
	rawParametersContent = append(rawParametersContent, types.UCHAR(c.NewOpLockLevel))

	// Marshalling parameter Timeout
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, uint32(c.Timeout))
	rawParametersContent = append(rawParametersContent, buf4...)

	// Marshalling parameter NumberOfRequestedUnlocks
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.NumberOfRequestedUnlocks))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameter NumberOfRequestedLocks
	buf2 = make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(c.NumberOfRequestedLocks))
	rawParametersContent = append(rawParametersContent, buf2...)

	// Marshalling parameters
	c.GetParameters().AddWordsFromBytesStream(rawParametersContent)
	marshalledParameters, err := c.GetParameters().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledParameters...)

	// Marshalling data
	c.GetData().Add(rawDataContent)
	marshalledData, err := c.GetData().Marshal()
	if err != nil {
		return nil, err
	}
	marshalledCommand = append(marshalledCommand, marshalledData...)

	return marshalledCommand, nil
}

// Unmarshal unmarshals a byte array into the command structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
func (c *LockingAndxRequest) Unmarshal(data []byte) (int, error) {
	offset := 0

	// First unmarshal the two structures
	bytesRead, err := c.GetParameters().Unmarshal(data)
	if err != nil {
		return 0, err
	}
	rawParametersContent := c.GetParameters().GetBytes()
	_, err = c.GetData().Unmarshal(data[bytesRead:])
	if err != nil {
		return 0, err
	}
	rawDataContent := c.GetData().GetBytes()

	// If the parameters and data are empty, this is a response containing an error code in
	// the SMB Header Status field
	if len(rawParametersContent) == 0 && len(rawDataContent) == 0 {
		return 0, nil
	}

	// First unmarshal the parameters
	offset = 0

	// Unmarshalling parameter FID
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for FID")
	}
	c.FID = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter TypeOfLock
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for TypeOfLock")
	}
	c.TypeOfLock = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter NewOpLockLevel
	if len(rawParametersContent) < offset+1 {
		return offset, fmt.Errorf("data too short for NewOpLockLevel")
	}
	c.NewOpLockLevel = types.UCHAR(rawParametersContent[offset])
	offset++

	// Unmarshalling parameter Timeout
	if len(rawParametersContent) < offset+4 {
		return offset, fmt.Errorf("rawParametersContent too short for Timeout")
	}
	c.Timeout = types.ULONG(binary.BigEndian.Uint32(rawParametersContent[offset : offset+4]))
	offset += 4

	// Unmarshalling parameter NumberOfRequestedUnlocks
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for NumberOfRequestedUnlocks")
	}
	c.NumberOfRequestedUnlocks = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Unmarshalling parameter NumberOfRequestedLocks
	if len(rawParametersContent) < offset+2 {
		return offset, fmt.Errorf("rawParametersContent too short for NumberOfRequestedLocks")
	}
	c.NumberOfRequestedLocks = types.USHORT(binary.BigEndian.Uint16(rawParametersContent[offset : offset+2]))
	offset += 2

	// Then unmarshal the data
	offset = 0

	// Unmarshal Unlocks
	for i := 0; i < int(c.NumberOfRequestedUnlocks); i++ {
		if len(rawDataContent) < offset+20 {
			return offset, fmt.Errorf("rawDataContent too short for Unlocks[%d]", i)
		}
		unlock := types.LOCKING_ANDX_RANGE64{}
		bytesRead, err := unlock.Unmarshal(rawDataContent[offset : offset+20])
		if err != nil {
			return offset, fmt.Errorf("error unmarshalling unlock: %v", err)
		}
		c.Unlocks = append(c.Unlocks, unlock)
		offset += bytesRead
	}

	// Unmarshal Locks
	for i := 0; i < int(c.NumberOfRequestedLocks); i++ {
		if len(rawDataContent) < offset+20 {
			return offset, fmt.Errorf("rawDataContent too short for Locks[%d]", i)
		}
		lock := types.LOCKING_ANDX_RANGE64{}
		bytesRead, err := lock.Unmarshal(rawDataContent[offset : offset+20])
		if err != nil {
			return offset, fmt.Errorf("error unmarshalling lock: %v", err)
		}
		c.Locks = append(c.Locks, lock)
		offset += bytesRead
	}

	return offset, nil
}
