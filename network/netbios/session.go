package netbios

import "fmt"

// Session message types
// Source: RFC 1002 page 30

type SESSION_MESSAGE_TYPE uint8

const (
	SESSION_MESSAGE           SESSION_MESSAGE_TYPE = 0x00
	SESSION_REQUEST           SESSION_MESSAGE_TYPE = 0x81
	SESSION_POSITIVE_RESPONSE SESSION_MESSAGE_TYPE = 0x82
	SESSION_NEGATIVE_RESPONSE SESSION_MESSAGE_TYPE = 0x83
	SESSION_RETARGET_RESPONSE SESSION_MESSAGE_TYPE = 0x84
	SESSION_KEEP_ALIVE        SESSION_MESSAGE_TYPE = 0x85
)

var SessionMessageTypeToString = map[SESSION_MESSAGE_TYPE]string{
	SESSION_MESSAGE:           "SESSION_MESSAGE",
	SESSION_REQUEST:           "SESSION_REQUEST",
	SESSION_POSITIVE_RESPONSE: "SESSION_POSITIVE_RESPONSE",
	SESSION_NEGATIVE_RESPONSE: "SESSION_NEGATIVE_RESPONSE",
	SESSION_RETARGET_RESPONSE: "SESSION_RETARGET_RESPONSE",
	SESSION_KEEP_ALIVE:        "SESSION_KEEP_ALIVE",
}

func (s SESSION_MESSAGE_TYPE) String() string {
	if str, exists := SessionMessageTypeToString[s]; exists {
		return str
	}
	return "UNKNOWN"
}

// NEGATIVE SESSION RESPONSE error codes carried in the one-byte ERROR_CODE field
// of a NEGATIVE SESSION RESPONSE (0x83).
// Source: RFC 1002 4.3.4

type NEGATIVE_SESSION_ERROR uint8

const (
	NEGATIVE_SESSION_NOT_LISTENING_ON_CALLED_NAME   NEGATIVE_SESSION_ERROR = 0x80 // Not listening on called name
	NEGATIVE_SESSION_NOT_LISTENING_FOR_CALLING_NAME NEGATIVE_SESSION_ERROR = 0x81 // Not listening for calling name
	NEGATIVE_SESSION_CALLED_NAME_NOT_PRESENT        NEGATIVE_SESSION_ERROR = 0x82 // Called name not present
	NEGATIVE_SESSION_INSUFFICIENT_RESOURCES         NEGATIVE_SESSION_ERROR = 0x83 // Called name present, but insufficient resources
	NEGATIVE_SESSION_UNSPECIFIED_ERROR              NEGATIVE_SESSION_ERROR = 0x8F // Unspecified error
)

var NegativeSessionErrorToString = map[NEGATIVE_SESSION_ERROR]string{
	NEGATIVE_SESSION_NOT_LISTENING_ON_CALLED_NAME:   "Not listening on called name",
	NEGATIVE_SESSION_NOT_LISTENING_FOR_CALLING_NAME: "Not listening for calling name",
	NEGATIVE_SESSION_CALLED_NAME_NOT_PRESENT:        "Called name not present",
	NEGATIVE_SESSION_INSUFFICIENT_RESOURCES:         "Called name present, but insufficient resources",
	NEGATIVE_SESSION_UNSPECIFIED_ERROR:              "Unspecified error",
}

func (e NEGATIVE_SESSION_ERROR) String() string {
	if str, exists := NegativeSessionErrorToString[e]; exists {
		return str
	}
	return "UNKNOWN"
}

// Error lets a NEGATIVE_SESSION_ERROR be returned directly as an error value.
func (e NEGATIVE_SESSION_ERROR) Error() string {
	return fmt.Sprintf("NEGATIVE SESSION RESPONSE: %s (0x%02X)", e.String(), uint8(e))
}
