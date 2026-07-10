package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

// WriteTCPMessage writes payload to w framed for DNS-over-TCP transport. Per RFC
// 1035 §4.2.2 (which RFC 4795 §2.4 adopts for LLMNR over TCP on port 5355), the
// message is prefixed with a two-byte big-endian length field giving the length
// of the message, excluding the two length bytes themselves. The prefix and the
// payload are written together so the framing cannot be split across writes.
//
// It returns an error if payload does not fit in the 16-bit length field or if
// the underlying write fails.
func WriteTCPMessage(w io.Writer, payload []byte) error {
	if len(payload) > 0xFFFF {
		return fmt.Errorf("message too large for TCP framing: %d bytes (max %d)", len(payload), 0xFFFF)
	}

	framed := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(framed[:2], uint16(len(payload)))
	copy(framed[2:], payload)

	if _, err := w.Write(framed); err != nil {
		return fmt.Errorf("failed to write TCP message: %w", err)
	}
	return nil
}

// ReadTCPMessage reads a single DNS-over-TCP framed message from r and returns
// its payload (without the two-byte length prefix). Per RFC 1035 §4.2.2 it first
// reads the two-byte big-endian length field, then reads exactly that many bytes
// of message. io.ReadFull is used for both reads so a short read is reported as
// an error rather than silently returning a partial message.
//
// It returns an error if the length prefix or the announced number of message
// bytes cannot be fully read.
func ReadTCPMessage(r io.Reader) ([]byte, error) {
	lengthPrefix := make([]byte, 2)
	if _, err := io.ReadFull(r, lengthPrefix); err != nil {
		return nil, fmt.Errorf("failed to read TCP length prefix: %w", err)
	}

	length := binary.BigEndian.Uint16(lengthPrefix)
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("failed to read TCP message body of %d bytes: %w", length, err)
	}

	return payload, nil
}
