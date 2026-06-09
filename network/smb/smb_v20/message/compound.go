package message

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
)

// align8 rounds n up to the next multiple of 8.
func align8(n int) int {
	return (n + 7) &^ 7
}

// MarshalCompound serializes a sequence of SMB2 messages into a single compounded
// request/response.
//
// Each message contributes a 64-byte header followed by its command body. Every
// segment except the last is padded with zero bytes so that the following header
// starts on an 8-byte boundary, and each non-last header's NextCommand is set to
// the offset, in bytes, from the start of that header to the start of the next
// header. The last header's NextCommand is 0.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-smb2/fb188936-5050-48d3-b350-dc43059638a4
func MarshalCompound(messages []*Message) ([]byte, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages to marshal")
	}

	out := []byte{}
	for i, m := range messages {
		if m.Command == nil {
			return nil, fmt.Errorf("message %d has no command set", i)
		}

		body, err := m.Command.Marshal()
		if err != nil {
			return nil, err
		}

		segmentLen := header.SMB2_HEADER_SIZE + len(body)
		isLast := i == len(messages)-1

		m.Header.Command = m.Command.GetCommandCode()
		if isLast {
			m.Header.NextCommand = 0
		} else {
			m.Header.NextCommand = uint32(align8(segmentLen))
		}

		marshalledHeader, err := m.Header.Marshal()
		if err != nil {
			return nil, err
		}

		out = append(out, marshalledHeader...)
		out = append(out, body...)

		if !isLast {
			out = append(out, make([]byte, align8(segmentLen)-segmentLen)...)
		}
	}

	return out, nil
}

// UnmarshalCompound parses a compounded SMB2 message into its constituent
// messages by following each header's NextCommand offset. A single
// (non-compounded) message parses into a one-element slice.
func UnmarshalCompound(data []byte) ([]*Message, error) {
	messages := []*Message{}
	offset := 0

	for offset < len(data) {
		m := NewMessage()
		consumed, err := m.Unmarshal(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("compounded message at offset %d: %w", offset, err)
		}
		messages = append(messages, m)

		// NextCommand == 0 marks the last segment; Unmarshal reports it by
		// consuming the remainder of the slice.
		if m.Header.NextCommand == 0 {
			break
		}
		offset += consumed
	}

	return messages, nil
}
