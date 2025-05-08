package types

import (
	"fmt"
)

// SMB_ResumeKey
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/239b0def-8370-4dc7-8391-ee60952901b1
// BufferFormat2 (1 byte): This field MUST be 0x05, which indicates a variable
// block is to follow.
// ResumeKeyLength (2 bytes): This field MUST be either 0x0000 or 21 (0x0015). If
// the value of this field is 0x0000, this is an initial search request. The server
// MUST allocate resources to maintain search state so that subsequent requests MAY
// be processed. If the value of this field is 21 (0x0015), this request MUST be
// the continuation of a previous search, and the next field MUST contain a
// ResumeKey previously returned by the server.
// ResumeKey (variable): SMB_Resume_Key If the value of ResumeKeyLength is 21
// (0x0015), this field MUST contain a ResumeKey returned by the server in response
// to a previous SMB_COM_SEARCH request. The ResumeKey contains data used by both
// the client and the server to maintain the state of the search. The structure of the
// ResumeKey follows:
type SMB_RESUME_KEY struct {
	SMB_STRING

	// Reserved (1 byte): This field is reserved and MUST NOT be modified by the client.
	// Older documentation is contradictory as to whether this field is reserved for
	// client side or server side use. New server implementations SHOULD avoid using or
	// modifying the content of this field.
	Reserved UCHAR

	// ServerState (16 bytes): This field is maintained by the server and MUST NOT be
	// modified by the client. The contents of this field are server-specific.
	ServerState [16]UCHAR

	// ClientState (4 bytes): This field MAY be used by the client to maintain state
	// across a series of SMB_COM_SEARCH calls. The value provided by the client MUST be
	// returned in each ResumeKey provided in the response. The contents of this field
	// are client-specific.
	ClientState [4]UCHAR
}

// NewSMB_RESUME_KEY creates a new SMB_RESUME_KEY structure
//
// Returns:
// - A pointer to the new SMB_RESUME_KEY structure
func NewSMB_RESUME_KEY() *SMB_RESUME_KEY {
	return &SMB_RESUME_KEY{
		SMB_STRING: SMB_STRING{
			// This field MUST be 0x05, which indicates that a variable block is to follow.
			BufferFormat: SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK,
			Length:       0,
			Buffer:       []byte{},
		},
		Reserved:    0,
		ServerState: [16]UCHAR{},
		ClientState: [4]UCHAR{},
	}
}

// Marshal marshals the SMB_RESUME_KEY structure
//
// Returns:
// - A byte array representing the SMB_RESUME_KEY structure
// - An error if the marshaling fails
func (r *SMB_RESUME_KEY) Marshal() ([]byte, error) {
	// This field MUST be 0x05, which indicates that a variable block is to follow.
	r.SMB_STRING.BufferFormat = SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK

	byteStream := []byte{}
	byteStream = append(byteStream, r.Reserved)
	byteStream = append(byteStream, r.ServerState[:]...)
	byteStream = append(byteStream, r.ClientState[:]...)
	r.SMB_STRING.Buffer = byteStream

	r.SMB_STRING.Length = uint16(len(byteStream))

	return r.SMB_STRING.Marshal()
}

// Unmarshal unmarshals the SMB_RESUME_KEY structure
//
// Parameters:
// - data: The byte array to unmarshal
//
// Returns:
// - The number of bytes unmarshalled
// - An error if the unmarshaling fails
func (r *SMB_RESUME_KEY) Unmarshal(data []byte) (int, error) {
	offset := 0

	bytesRead, err := r.SMB_STRING.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	fmt.Println(r.SMB_STRING.Buffer)
	if len(r.SMB_STRING.Buffer) < 21 {
		return 0, fmt.Errorf("SMB_STRING.Buffer length is not 21")
	}

	r.Reserved = r.SMB_STRING.Buffer[0]
	offset++

	copy(r.ServerState[:], r.SMB_STRING.Buffer[1:17])
	offset += 16

	copy(r.ClientState[:], r.SMB_STRING.Buffer[17:21])
	offset += 4

	return bytesRead, nil
}
