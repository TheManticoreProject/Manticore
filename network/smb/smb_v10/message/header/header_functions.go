package header

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
)

// SetFlags sets the flags field in the SMB Header.
// The flags field is an 8-bit field of 1-bit flags describing various features in effect for the message.
//
// Parameters:
//   - flags: The byte value to set as flags
func (h *Header) SetFlags(value uint8) {
	h.Flags = flags.Flags(value)
}

// SetFlags2 sets the flags2 field in the SMB Header.
// The flags2 field is a 16-bit field of 1-bit flags that represent various features in effect for the message.
//
// Parameters:
//   - flags2: The uint16 value to set as flags2
func (h *Header) SetFlags2(value uint16) {
	h.Flags2 = flags2.Flags2(value)
}

// IsResponse returns true if the header is a response, false otherwise.
//
// Returns:
//   - bool: True if the header is a response, false otherwise
func (h *Header) IsResponse() bool {
	return h.Flags&flags.FLAGS_REPLY == flags.FLAGS_REPLY
}

// IsRequest returns true if the header is a request, false otherwise.
//
// Returns:
//   - bool: True if the header is a request, false otherwise
func (h *Header) IsRequest() bool {
	return !h.IsResponse()
}
