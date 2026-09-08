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
// The BufferFormat2 and ResumeKeyLength fields described above are part of the
// SMB_COM_SEARCH *request's* data block, not of this structure. Inside an
// SMB_DIRECTORY_INFORMATION entry the resume key is a bare 21-byte field with no
// format byte and no length ahead of it, so this structure marshals only its own
// three fields and the request writes its own wrapper.
type SMB_RESUME_KEY struct {
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
	byteStream := make([]byte, 0, SMB_RESUME_KEY_SIZE)
	byteStream = append(byteStream, r.Reserved)
	byteStream = append(byteStream, r.ServerState[:]...)
	byteStream = append(byteStream, r.ClientState[:]...)
	return byteStream, nil
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
	if len(data) < SMB_RESUME_KEY_SIZE {
		return 0, fmt.Errorf("data too short for SMB_RESUME_KEY (need %d bytes, have %d)",
			SMB_RESUME_KEY_SIZE, len(data))
	}

	r.Reserved = data[0]
	copy(r.ServerState[:], data[1:17])
	copy(r.ClientState[:], data[17:SMB_RESUME_KEY_SIZE])

	return SMB_RESUME_KEY_SIZE, nil
}

// SMB_RESUME_KEY_SIZE is the size of the structure on the wire: Reserved(1)
// ServerState(16) ClientState(4), per [MS-CIFS] section 2.2.4.58.1.
const SMB_RESUME_KEY_SIZE = 21
