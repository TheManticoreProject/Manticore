package header

import (
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// HasValidProtocolId reports whether the ProtocolId field holds the SMB2 marker
// 0xFE 'S' 'M' 'B'.
func (h *Header) HasValidProtocolId() bool {
	return h.ProtocolId == SMB2ProtocolId
}

// SetFlags sets the 32-bit Flags field.
func (h *Header) SetFlags(value flags.Flags) {
	h.Flags = value
}

// AddFlags sets the given flag bits without clearing the others.
func (h *Header) AddFlags(value flags.Flags) {
	h.Flags |= value
}

// IsResponse reports whether the message is a response
// (SMB2_FLAGS_SERVER_TO_REDIR set).
func (h *Header) IsResponse() bool {
	return h.Flags.IsServerToRedir()
}

// IsRequest reports whether the message is a request.
func (h *Header) IsRequest() bool {
	return !h.IsResponse()
}

// IsAsync reports whether the header uses the ASYNC form
// (SMB2_FLAGS_ASYNC_COMMAND set).
func (h *Header) IsAsync() bool {
	return h.Flags.IsAsync()
}

// GetCommand returns the command code.
func (h *Header) GetCommand() codes.CommandCode {
	return h.Command
}

// SetCommand sets the command code.
func (h *Header) SetCommand(command codes.CommandCode) {
	h.Command = command
}

// GetMessageId returns the 64-bit MessageId.
func (h *Header) GetMessageId() types.UINT64 {
	return h.MessageId
}

// SetMessageId sets the 64-bit MessageId.
func (h *Header) SetMessageId(messageId types.UINT64) {
	h.MessageId = messageId
}

// GetTreeId returns the 32-bit TreeId (SYNC form).
func (h *Header) GetTreeId() types.ULONG {
	return h.TreeId
}

// SetTreeId sets the 32-bit TreeId (SYNC form).
func (h *Header) SetTreeId(treeId types.ULONG) {
	h.TreeId = treeId
}

// GetSessionId returns the 64-bit SessionId.
func (h *Header) GetSessionId() types.UINT64 {
	return h.SessionId
}

// SetSessionId sets the 64-bit SessionId.
func (h *Header) SetSessionId(sessionId types.UINT64) {
	h.SessionId = sessionId
}

// GetAsyncId returns the 64-bit AsyncId (ASYNC form).
func (h *Header) GetAsyncId() types.UINT64 {
	return h.AsyncId
}

// SetAsyncId sets the 64-bit AsyncId (ASYNC form).
func (h *Header) SetAsyncId(asyncId types.UINT64) {
	h.AsyncId = asyncId
}
