package pdu

import (
	"encoding/binary"
	"fmt"
)

// CancelBodySize is the size, in octets, of a cl_cancel body.
const CancelBodySize = 8

// CancelAckBodySize is the size, in octets, of a non-empty cancel_ack body.
const CancelAckBodySize = 9

// CancelBody is the body of a cl_cancel PDU ([C706] section 12.6.4.4, cancel_body_t
// at body format version 0).
//
// Wire layout (little-endian):
//
//	0-3  vers       uint32  (body format version, 0)
//	4-7  cancel_id  uint32  (identifies the cancel event)
type CancelBody struct {
	Version  uint32
	CancelID uint32
}

// Marshal serializes the cl_cancel body into its 8-octet wire form.
func (b *CancelBody) Marshal() ([]byte, error) {
	buf := make([]byte, CancelBodySize)
	binary.LittleEndian.PutUint32(buf[0:4], b.Version)
	binary.LittleEndian.PutUint32(buf[4:8], b.CancelID)
	return buf, nil
}

// Unmarshal parses a cl_cancel body from data and returns the bytes consumed.
func (b *CancelBody) Unmarshal(data []byte) (int, error) {
	if len(data) < CancelBodySize {
		return 0, fmt.Errorf("cl_cancel body truncated: have %d bytes, need %d", len(data), CancelBodySize)
	}
	b.Version = binary.LittleEndian.Uint32(data[0:4])
	b.CancelID = binary.LittleEndian.Uint32(data[4:8])
	return CancelBodySize, nil
}

// CancelAckBody is the optional body of a cancel_ack PDU ([C706] section 12.6.4.3,
// cancel_ack_body_t at body format version 0). A cancel_ack with no body is an
// orphan acknowledgement; when a body is present it carries the fields below.
//
// Wire layout (little-endian):
//
//	0-3  vers                 uint32   (body format version, 0)
//	4-7  cancel_id            uint32   (cancel event being acknowledged)
//	8    server_is_accepting  boolean  (non-zero if the server accepts cancels)
type CancelAckBody struct {
	Version           uint32
	CancelID          uint32
	ServerIsAccepting bool
}

// Marshal serializes the cancel_ack body into its 9-octet wire form.
func (b *CancelAckBody) Marshal() ([]byte, error) {
	buf := make([]byte, CancelAckBodySize)
	binary.LittleEndian.PutUint32(buf[0:4], b.Version)
	binary.LittleEndian.PutUint32(buf[4:8], b.CancelID)
	if b.ServerIsAccepting {
		buf[8] = 1
	}
	return buf, nil
}

// Unmarshal parses a cancel_ack body from data and returns the bytes consumed.
func (b *CancelAckBody) Unmarshal(data []byte) (int, error) {
	if len(data) < CancelAckBodySize {
		return 0, fmt.Errorf("cancel_ack body truncated: have %d bytes, need %d", len(data), CancelAckBodySize)
	}
	b.Version = binary.LittleEndian.Uint32(data[0:4])
	b.CancelID = binary.LittleEndian.Uint32(data[4:8])
	b.ServerIsAccepting = data[8] != 0
	return CancelAckBodySize, nil
}
