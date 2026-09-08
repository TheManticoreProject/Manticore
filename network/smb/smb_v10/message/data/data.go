package data

import (
	"encoding/binary"
	"fmt"
)

// Data represents the data field in an SMB message
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/48b4bd5d-7206-4002-bde1-c34cf614b138
//
// ByteCount is a USHORT and no extension widens it, so it cannot describe a block
// of 0x10000 bytes or more. Two commands legitimately carry one — a large read
// response and a large write request, under CAP_LARGE_READX and CAP_LARGE_WRITEX
// — and they locate and size their own payload from DataOffset and
// DataLength/DataLengthHigh instead. [MS-SMB] section 3.3.5.8 says a server
// SHOULD reject a write whose data is not (DataLength | DataLengthHigh <<16)
// bytes, and Windows servers "ignore the ByteCount field" outright when working
// it out.
//
// ByteCount is not a reliable bound even below that. [MS-CIFS] section 2.2.4.43.1
// describes relocating a write's data to the end of an AndX chain and has
// ByteCount stay at "1 + SMB_Parameters.Words.DataLength" — counting bytes that
// are not in the block at all. So Length, not ByteCount, is what a caller should
// ask for the size of what is actually held.
type Data struct {
	ByteCount uint16
	Bytes     []byte
}

// NewData creates a new Data structure with an empty Bytes field and a ByteCount of 0
//
// This function creates a new Data structure with an empty Bytes field and a ByteCount of 0.
func NewData() *Data {
	return &Data{
		ByteCount: 0,
		Bytes:     []byte{},
	}
}

// Size returns the size of the Data structure
//
// This function returns the size of the Data structure. It returns the ByteCount field.
//
// It is the field as it goes on the wire, which for a block of 0x10000 bytes or
// more is the low 16 bits of the true length. Use Length for how many bytes are
// actually held.
func (d *Data) Size() uint16 {
	return d.ByteCount
}

// Length returns the number of bytes the Data structure holds.
//
// It differs from Size for a block that ByteCount cannot describe, which is the
// case a large read or write produces.
//
// Returns:
//   - The number of bytes held
func (d *Data) Length() int {
	return len(d.Bytes)
}

// Add appends bytes to the Data structure and updates the ByteCount
//
// This function appends the given bytes to the existing bytes in the Data structure and updates the ByteCount
// to reflect the new length of the Bytes field.
//
// Past 0xFFFF bytes ByteCount holds the low 16 bits of the length, because that
// is all a USHORT can carry. That is deliberate rather than an overflow: the two
// commands that produce such a block describe their payload with their own length
// and offset fields, and a receiver of one reads those.
func (d *Data) Add(bytes []byte) {
	d.Bytes = append(d.Bytes, bytes...)
	d.ByteCount = uint16(len(d.Bytes) & 0xFFFF)
}

// SetData sets the data for the message
//
// This function sets the data for the message. It sets the Bytes field to the given data and updates the ByteCount
// to reflect the length of the given data.
func (d *Data) SetData(data []byte) {
	d.Bytes = data
	// The low 16 bits, for the same reason as Add.
	d.ByteCount = uint16(len(data) & 0xFFFF)
}

// GetBytes returns the data for the message
//
// This function returns the data for the message. It returns the Bytes field.
func (d *Data) GetBytes() []byte {
	return d.Bytes
}

// Marshal marshals the Data structure into a byte array
//
// This function marshals the Data structure into a byte array. It creates a new byte array, appends the ByteCount
// and the Bytes field to it, and returns the resulting byte array.
func (d *Data) Marshal() ([]byte, error) {
	marshalled := []byte{}

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, d.ByteCount)
	marshalled = append(marshalled, buf...)

	marshalled = append(marshalled, d.Bytes...)

	return marshalled, nil
}

// Unmarshal unmarshals the Data structure from a byte array
//
// This function unmarshals the Data structure from a byte array. It reads the ByteCount from the first two bytes
// of the input byte array, and then reads the corresponding number of bytes into the Bytes field. It returns
// the number of bytes read and an error if the input byte array is empty or too short to unmarshal the Data structure.
func (d *Data) Unmarshal(data []byte) (int, error) {
	bytesRead := 0

	if len(data) < 2 {
		return bytesRead, fmt.Errorf("data too short to unmarshal SMB data byte count")
	}

	d.ByteCount = binary.LittleEndian.Uint16(data[:2])
	bytesRead += 2

	data = data[bytesRead:]

	if d.ByteCount > 0 {
		// Each word is 2 bytes
		if len(data) < (int(d.ByteCount)) {
			return bytesRead, fmt.Errorf("data too short to unmarshal SMB parameters")
		}

		d.Bytes = data[:int(d.ByteCount)]

		bytesRead += int(d.ByteCount)
	} else {
		d.Bytes = []byte{}
	}

	return bytesRead, nil
}
