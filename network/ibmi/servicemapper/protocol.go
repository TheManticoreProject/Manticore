package servicemapper

import (
	"encoding/binary"
	"errors"
	"fmt"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
)

// DefaultPort is the service-mapper TCP port.
const DefaultPort = 449

// ReplyLength is the size of a successful reply.
const ReplyLength = 5

const replySuccess = '+'

var (
	// ErrUnsupportedService is matched when a name is not one of the
	// supported mapper names.
	ErrUnsupportedService = errors.New("servicemapper: unsupported service name")
	// ErrNotFound is matched when the mapper answered but did not return a
	// port (its first byte was not '+').
	ErrNotFound = errors.New("servicemapper: service not available")
	// ErrProtocol is matched by malformed replies.
	ErrProtocol = errors.New("servicemapper: malformed reply")
)

// LookupError reports a mapper that answered without a port.
type LookupError struct {
	Name   string
	Status byte // first reply byte
}

func (e *LookupError) Error() string {
	return fmt.Sprintf("servicemapper: %s not available (status byte 0x%02X)", e.Name, e.Status)
}

// Is makes errors.Is(err, ErrNotFound) true.
func (e *LookupError) Is(target error) bool { return target == ErrNotFound }

// EncodeRequest validates name and returns its request bytes: the ASCII
// name alone, with no length prefix or terminator.
func EncodeRequest(name string) ([]byte, error) {
	if _, _, ok := ibminet.ServiceForMapperName(name); !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedService, name)
	}
	return []byte(name), nil
}

// DecodeReply decodes a complete reply for name. A first byte other than
// '+' is a LookupError; a short or long reply, port 0 or a port above 65535
// is an ErrProtocol error.
func DecodeReply(name string, b []byte) (int, error) {
	if len(b) == 0 {
		return 0, fmt.Errorf("%w: empty reply for %s", ErrProtocol, name)
	}
	if b[0] != replySuccess {
		return 0, &LookupError{Name: name, Status: b[0]}
	}
	if len(b) != ReplyLength {
		return 0, fmt.Errorf("%w: %d-byte reply for %s, want %d", ErrProtocol, len(b), name, ReplyLength)
	}
	port := binary.BigEndian.Uint32(b[1:])
	if port == 0 || port > 65535 {
		return 0, fmt.Errorf("%w: port %d for %s", ErrProtocol, port, name)
	}
	return int(port), nil
}
