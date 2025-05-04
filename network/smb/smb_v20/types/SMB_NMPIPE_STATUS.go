package types

import "fmt"

const (
	// This bit field indicates the client read mode for the named pipe. This bit field has no effect on writes to the named pipe.
	// A value of zero indicates that the named pipe was opened in or set to byte mode by the client.
	// byte mode: One of two kinds of named pipe, the other of which is message mode. In byte mode, the data sent or received on the
	// named pipe does not have message boundaries but is treated as a continuous stream. [XOPEN-SMB] uses the term stream mode instead
	// of byte mode, and [SMB-LM1X] refers to byte mode as byte stream mode.
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/760f8b7f-9a8a-4f0c-a044-1501a83a933b#gt_8586fe9a-1cfa-458a-a145-8db64d69c69c
	SMB_NMPIPE_STATUS_READ_MODE_BYTE uint8 = 0
	// A value of 1 indicates that the client opened or set the named pipe to message mode.
	// message mode: A named pipe can be of two types: byte mode or message mode. In byte mode, the data sent or received on the named
	// pipe does not have messageboundaries but is treated as a continuous Stream. In message mode, message boundaries are enforced.
	// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/760f8b7f-9a8a-4f0c-a044-1501a83a933b#gt_c49a48e8-f1ac-4568-bc87-0672eb08868b
	SMB_NMPIPE_STATUS_READ_MODE_MESSAGE uint8 = 1
	// Reserved. Bit 0x0200 MUST be ignored.
	SMB_NMPIPE_STATUS_READ_MODE_2_RESERVED uint8 = 2
	// Reserved
	SMB_NMPIPE_STATUS_READ_MODE_3_RESERVED uint8 = 3

	// If not set, it indicates a client-side end of the named pipe. The SMB server MUST clear the Endpoint bit (set it to zero) when
	// responding to the client request because the CIFS client is a consumer requesting service from the named pipe. When this bit is clear,
	// it indicates that the client is accessing the consumer endpoint.
	// If set, it indicates the server end of the pipe.
	SMB_NMPIPE_STATUS_ENDPOINT_SERVER uint8 = 0x40

	// If not set:
	//
	// A named pipe read or raw read request will wait (block) until sufficient data to satisfy the read request becomes available, or until
	// the request is canceled.
	//
	// A named pipe write or raw write request blocks until its data is consumed, if the write request length is greater than zero.
	//
	// If set:
	//
	// A read or a raw read request returns all data available to be read from the named pipe, up to the maximum read size set in the request.
	// Write operations return after writing data to named pipes without waiting for the data to be consumed.
	// Named pipe non-blocking raw writes are not allowed. Raw writes MUST be performed in blocking mode.
	SMB_NMPIPE_STATUS_NONBLOCKING uint8 = 0x80
)

// SMB_NMPIPE_STATUS
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/6911a709-5dfb-4ffb-b090-3e8ef872f85c
// The SMB_NMPIPE_STATUS data type is a 16-bit field that encodes the status of a named pipe. Any combination of the following
// flags MUST be valid. The ReadMode and NamedPipeType bit fields are defined as 2-bit integers. Subfields marked Reserved SHOULD
// be set to zero by the server and MUST be ignored by the client.
type SMB_NMPIPE_STATUS struct {
	// An 8-bit unsigned integer that gives the maximum number of instances the named pipe can have.
	ICount uint8
	// A 8-bit integer that gives information about the named pipe.
	Flags uint8
}

// SetICount sets the ICount field of the SMB_NMPIPE_STATUS
//
// Parameters:
// - icount: The value to set the ICount field to
func (s *SMB_NMPIPE_STATUS) SetICount(icount uint8) {
	s.ICount = icount
}

// GetICount returns the ICount field of the SMB_NMPIPE_STATUS
//
// Returns:
// - The value of the ICount field
func (s SMB_NMPIPE_STATUS) GetICount() uint8 {
	return s.ICount
}

// SetNonBlockingStatus sets the non-blocking status of the SMB_NMPIPE_STATUS
//
// Parameters:
// - nonBlocking: The value to set the non-blocking status to
func (s *SMB_NMPIPE_STATUS) SetNonBlockingStatus(nonBlocking bool) {
	if nonBlocking {
		s.Flags = s.Flags | SMB_NMPIPE_STATUS_NONBLOCKING
	} else {
		s.Flags = s.Flags & ^SMB_NMPIPE_STATUS_NONBLOCKING
	}
}

// IsNonBlocking returns true if the non-blocking status is set
//
// Returns:
// - true if the non-blocking status is set, false otherwise
func (s SMB_NMPIPE_STATUS) IsNonBlocking() bool {
	return s.Flags&SMB_NMPIPE_STATUS_NONBLOCKING == SMB_NMPIPE_STATUS_NONBLOCKING
}

// GetReadMode returns the read mode of the SMB_NMPIPE_STATUS
//
// Returns:
// - The value of the read mode
func (s SMB_NMPIPE_STATUS) GetReadMode() uint8 {
	return s.Flags & 0b11
}

// GetNamedPipeType returns the named pipe type of the SMB_NMPIPE_STATUS
func (s SMB_NMPIPE_STATUS) String() string {
	return fmt.Sprintf("ICount: %d, Flags: %d", s.ICount, s.Flags)
}

// Marshal marshals the SMB_NMPIPE_STATUS into a byte array
//
// Returns:
// - The byte array representing the SMB_NMPIPE_STATUS
func (s SMB_NMPIPE_STATUS) Marshal() []byte {
	return []byte{s.ICount, s.Flags}
}

// Unmarshal unmarshals the SMB_NMPIPE_STATUS from a byte array
//
// Parameters:
// - data: The byte array to unmarshal the SMB_NMPIPE_STATUS from
//
// Returns:
// - The number of bytes unmarshalled
func (s *SMB_NMPIPE_STATUS) Unmarshal(data []byte) (int, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("data must be 2 bytes long")
	}

	s.ICount = data[0]
	s.Flags = data[1]

	return 2, nil
}
