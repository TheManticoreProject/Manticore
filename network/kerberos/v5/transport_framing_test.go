package kerberos

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"
)

// TestTCPFramingRoundTrip confirms writeTCPFramed emits the RFC 4120 Section
// 7.2.2 4-byte big-endian length prefix and that readTCPFramed recovers exactly
// the framed payload from the resulting stream.
func TestTCPFramingRoundTrip(t *testing.T) {
	for _, msg := range [][]byte{
		[]byte("A"),
		bytes.Repeat([]byte{0xAB}, 1500), // larger than a UDP datagram
		{},
	} {
		var buf bytes.Buffer
		if err := writeTCPFramed(&buf, msg); err != nil {
			t.Fatalf("writeTCPFramed(%d bytes): %v", len(msg), err)
		}
		// Prefix must be the big-endian length.
		if got := binary.BigEndian.Uint32(buf.Bytes()[:4]); got != uint32(len(msg)) {
			t.Errorf("length prefix = %d, want %d", got, len(msg))
		}
		if buf.Len() != 4+len(msg) {
			t.Errorf("framed size = %d, want %d", buf.Len(), 4+len(msg))
		}

		if len(msg) == 0 {
			// A zero-length frame is rejected on read (see below); skip readback.
			continue
		}
		got, err := readTCPFramed(&buf)
		if err != nil {
			t.Fatalf("readTCPFramed: %v", err)
		}
		if !bytes.Equal(got, msg) {
			t.Errorf("readTCPFramed = %X, want %X", got, msg)
		}
	}
}

// TestReadTCPFramedErrors covers the defensive checks in readTCPFramed: a zero
// length, an implausibly large length, a truncated length prefix, and a body
// shorter than its declared length are all rejected.
func TestReadTCPFramedErrors(t *testing.T) {
	frame := func(length uint32, body []byte) []byte {
		b := make([]byte, 4+len(body))
		binary.BigEndian.PutUint32(b[:4], length)
		copy(b[4:], body)
		return b
	}

	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"zero length", frame(0, nil), "empty TCP response"},
		{"too large", frame(maxTCPResponse+1, nil), "response too large"},
		{"short prefix", []byte{0x00, 0x01}, "TCP read length"},
		{"short body", frame(10, []byte("short")), "TCP read body"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readTCPFramed(bytes.NewReader(tt.data))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

// errWriter is an io.Writer that always fails, to exercise writeTCPFramed's
// send-error path.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("boom") }

func TestWriteTCPFramedError(t *testing.T) {
	if err := writeTCPFramed(errWriter{}, []byte("hi")); err == nil ||
		!strings.Contains(err.Error(), "TCP send") {
		t.Errorf("writeTCPFramed error = %v, want a TCP send error", err)
	}
}
