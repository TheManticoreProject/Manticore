package data_structures

import "github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"

// RPC_UNICODE_STRING
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/94a16bb6-c610-4cb9-8db6-26f15f560061
type RPC_UNICODE_STRING struct {
	// Length: The length, in bytes, of the string pointed to by the Buffer member.
	// The length MUST be a multiple of 2. The length MUST equal the entire size of the buffer.
	Length data_types.WORD
	// MaximumLength: The maximum size, in bytes, of the string pointed to by Buffer. The size MUST be a multiple of 2.
	// If not, the size MUST be decremented by 1 prior to use. This value MUST not be less than Length.
	MaximumLength data_types.WORD
	// Buffer: A pointer to a string buffer. The string pointed to by the buffer member MUST NOT include a terminating null character.
	Buffer data_types.WCHAR
}

type PRPC_UNICODE_STRING *RPC_UNICODE_STRING
