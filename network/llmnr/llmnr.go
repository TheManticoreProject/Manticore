package llmnr

import (
	"errors"
)

const (
	// LLMNR uses port 5355 as specified in RFC 4795
	LLMNRPort = 5355

	// Multicast addresses for LLMNR
	IPv4MulticastAddr = "224.0.0.252"
	IPv6MulticastAddr = "FF02::1:3"

	MaxLabelLength  = 63  // Maximum length of a single label
	MaxDomainLength = 255 // Maximum length of entire domain name
	HeaderSize      = 12  // Size of LLMNR header in bytes

	// DNS wire format related constants
	labelPointer  = 0xC0
	MaxPacketSize = 512
)

// LLMNR Header Flags
const (
	FlagQR = 1 << 15 // Query/Response flag
	FlagOP = 1 << 14 // Operation code
	FlagC  = 1 << 13 // Conflict flag
	FlagTC = 1 << 12 // Truncation flag
	FlagT  = 1 << 11 // Tentative flag
)

func FlagToString(f uint16) string {
	switch f {
	case FlagQR:
		return "QR"
	case FlagOP:
		return "OP"
	case FlagC:
		return "C"
	case FlagTC:
		return "TC"
	case FlagT:
		return "T"
	}
	return "Unknown"
}

// Common errors
var (
	ErrInvalidQuestion = errors.New("invalid question format")
	ErrInvalidMessage  = errors.New("invalid message format")
)
