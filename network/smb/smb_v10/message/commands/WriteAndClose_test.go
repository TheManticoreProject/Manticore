package commands

import (
	"testing"
)

// TestWriteAndCloseRequestUnmarshalRejectsOversizeCount asserts a
// CountOfBytesToWrite larger than the data block is refused instead of slicing
// past the end of the buffer.
//
// The field is client-supplied and was sliced unchecked, so any peer could crash
// the process with one 100-byte message. Every other field in this decoder is
// length-checked; this one was missed.
func TestWriteAndCloseRequestUnmarshalRejectsOversizeCount(t *testing.T) {
	// WordCount 6, then the six parameter words: FID, CountOfBytesToWrite,
	// WriteOffsetInBytes (2 words) and LastWriteTime (2 words).
	// CountOfBytesToWrite claims 0x3031 bytes.
	parameters := []byte{
		0x06,       // WordCount
		0x01, 0x00, // FID
		0x31, 0x30, // CountOfBytesToWrite = 0x3031
		0x00, 0x00, 0x00, 0x00, // WriteOffsetInBytes
		0x00, 0x00, 0x00, 0x00, // LastWriteTime
	}
	// A data block carrying only the Pad byte and a single data byte, far short
	// of the 0x3031 bytes the parameters claim.
	data := []byte{0x02, 0x00, 0x00, 0x41}

	request := NewWriteAndCloseRequest()
	if _, err := request.Unmarshal(append(parameters, data...)); err == nil {
		t.Fatal("Unmarshal() accepted a CountOfBytesToWrite larger than the data block")
	}
}

// TestWriteAndCloseRequestUnmarshalAcceptsExactCount asserts the bounds check
// does not reject a well-formed message whose data block is exactly the declared
// length.
func TestWriteAndCloseRequestUnmarshalAcceptsExactCount(t *testing.T) {
	payload := []byte("hello")

	parameters := []byte{
		0x06,       // WordCount
		0x07, 0x00, // FID
		byte(len(payload)), 0x00, // CountOfBytesToWrite
		0x00, 0x00, 0x00, 0x00, // WriteOffsetInBytes
		0x00, 0x00, 0x00, 0x00, // LastWriteTime
	}
	// ByteCount covers the Pad byte plus the payload.
	byteCount := 1 + len(payload)
	data := []byte{byte(byteCount), 0x00, 0x00}
	data = append(data, payload...)

	request := NewWriteAndCloseRequest()
	if _, err := request.Unmarshal(append(parameters, data...)); err != nil {
		t.Fatalf("Unmarshal() on a well-formed message error = %v", err)
	}
	if got := string(request.Data); got != string(payload) {
		t.Fatalf("Data = %q, want %q", got, payload)
	}
	if request.CountOfBytesToWrite != 5 {
		t.Fatalf("CountOfBytesToWrite = %d, want 5", request.CountOfBytesToWrite)
	}
}
