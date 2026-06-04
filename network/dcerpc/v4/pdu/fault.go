package pdu

import (
	"encoding/binary"
	"fmt"
)

// StatusBodySize is the size, in octets, of a fault or reject body.
const StatusBodySize = 4

// MarshalStatusBody encodes the body shared by the fault and reject PDUs: a single
// NDR unsigned long status code indicating why the operation faulted or why the
// request was rejected ([C706] section 12.6.4.6 fault, 12.6.4.7 reject).
func MarshalStatusBody(status uint32) []byte {
	buf := make([]byte, StatusBodySize)
	binary.LittleEndian.PutUint32(buf, status)
	return buf
}

// UnmarshalStatusBody decodes a fault or reject body into its status code.
func UnmarshalStatusBody(data []byte) (uint32, error) {
	if len(data) < StatusBodySize {
		return 0, fmt.Errorf("fault/reject body truncated: have %d bytes, need %d", len(data), StatusBodySize)
	}
	return binary.LittleEndian.Uint32(data[:StatusBodySize]), nil
}
